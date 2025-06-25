// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitexporter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/harness/harness-migrate/internal/command"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/util"
	"github.com/harness/harness-migrate/types"

	"github.com/drone/go-scm/scm"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type gitCloner interface {
	clone(ctx context.Context) (bool, error)
}

type cloneParams struct {
	repoData   scm.Repository
	gitPath    string
	repoSlug   string
	pullreqRef []config.RefSpec
	auth       credentials
}

type credentials struct {
	username string
	token    string
}

type nativeGitCloner struct {
	params cloneParams
	tracer tracer.Tracer
}

type goGitCloner struct {
	params cloneParams
	tracer tracer.Tracer
}

func (e *Exporter) CloneRepository(
	ctx context.Context,
	repoData scm.Repository,
	repoPath string,
	repoSlug string,
	pullreqRef []config.RefSpec,
	tracer tracer.Tracer,
) (bool, int, error) {
	tracer.Start(common.MsgStartGitClone, repoSlug)

	gitPath := filepath.Join(repoPath, types.GitDir)
	params := cloneParams{
		repoData:   repoData,
		gitPath:    gitPath,
		repoSlug:   repoSlug,
		pullreqRef: pullreqRef,
		auth: credentials{
			username: e.ScmLogin,
			token:    e.ScmToken,
		},
	}

	cloner := e.selectCloner(params)
	isEmpty, err := cloner.clone(ctx)
	if err != nil {
		tracer.Stop(common.ErrGitClone, err)
		return false, 0, fmt.Errorf("failed to clone repo %s: %w", repoSlug, err)
	}

	var lfsObjectCount int

	if e.flags.NoLFS {
		tracer.Stop(common.MsgCompleteGitClone, repoSlug)
		return isEmpty, lfsObjectCount, nil
	}

	if !isEmpty {
		err = command.FetchLFSObjects(ctx, gitPath)
		if err != nil {
			tracer.Stop(common.ErrFetchLFSObjects, err)
			return isEmpty, lfsObjectCount, fmt.Errorf("failed to pull LFS objects for repo %s: %w", repoSlug, err)
		}

		lfsObjectCount, err = e.countLFSObjects(ctx, gitPath, repoSlug, tracer)
		if err != nil {
			tracer.Stop(common.ErrFetchLFSObjects, err)
			return isEmpty, 0, err
		}
	}

	e.Report[repoSlug].ReportMetric(report.ReportTypeGitLFSObjects, lfsObjectCount)
	tracer.Stop(common.MsgCompleteGitClone, repoSlug)
	return isEmpty, lfsObjectCount, nil
}

func (e *Exporter) selectCloner(params cloneParams) gitCloner {
	if e.flags.NoLFS {
		return &goGitCloner{params: params, tracer: e.Tracer}
	}
	return &nativeGitCloner{params: params, tracer: e.Tracer}
}

func (e *Exporter) countLFSObjects(
	ctx context.Context,
	gitPath string,
	repoSlug string,
	tracer tracer.Tracer,
) (int, error) {
	lfsObjectCount, err := command.HasLFSObjects(ctx, gitPath)
	if err != nil {
		tracer.LogError("git-lfs-check", repoSlug, err)
		return 0, fmt.Errorf("failed to check LFS objects for repo %s: %w", repoSlug, err)
	}

	return lfsObjectCount, nil
}

func refSpecsToStrings(refs []config.RefSpec) []string {
	result := make([]string, len(refs))
	for i, ref := range refs {
		result[i] = ref.String()
	}
	return result
}

func (c *nativeGitCloner) clone(ctx context.Context) (bool, error) {
	if err := util.CreateFolder(c.params.gitPath); err != nil {
		return false, err
	}

	cloneURL := c.params.repoData.Clone
	var prefix string
	switch {
	case strings.HasPrefix(cloneURL, "https://"):
		prefix = "https://"
	case strings.HasPrefix(cloneURL, "http://"):
		prefix = "http://"
	}

	if prefix != "" {
		cloneURL = fmt.Sprintf("%s%s:%s@%s",
			prefix,
			c.params.auth.username,
			c.params.auth.token,
			strings.TrimPrefix(cloneURL, prefix),
		)
	}

	output, err := command.RunGitCommand(ctx, c.params.gitPath,
		[]string{},
		"clone", "--bare",
		cloneURL, ".")
	if err != nil {
		c.tracer.LogError(common.ErrGitClone, c.params.repoSlug, err, string(output))
		return false, fmt.Errorf("failed to clone repo %s: %w", c.params.repoSlug, err)
	}

	// check if the repository is empty by looking for refs
	refsOutput, _ := command.RunGitCommand(ctx, c.params.gitPath, []string{}, "show-ref")
	if len(strings.TrimSpace(string(refsOutput))) == 0 {
		c.tracer.Stop(common.MsgGitCloneEmptyRepo, c.params.repoSlug)
		return true, nil
	}

	fetchArgs := []string{
		"fetch",
		"origin",
		"refs/heads/*:refs/heads/*",
		"refs/tags/*:refs/tags/*",
	}
	fetchArgs = append(fetchArgs, refSpecsToStrings(c.params.pullreqRef)...)

	output, err = command.RunGitCommand(ctx, c.params.gitPath, []string{}, fetchArgs...)
	if err != nil {
		c.tracer.LogError(common.ErrGitFetch, c.params.repoSlug, err, string(output))
		return false, fmt.Errorf("failed to fetch refs for %s: %w", c.params.repoSlug, err)
	}

	// remove local config to prevent credential leak
	if err := os.Remove(filepath.Join(c.params.gitPath, "config")); err != nil {
		c.tracer.LogError("git-config-remove", c.params.repoSlug, err)
		return false, fmt.Errorf("failed to remove config for %s: %w", c.params.repoSlug, err)
	}

	return false, nil
}

func (c *goGitCloner) clone(ctx context.Context) (bool, error) {
	if err := util.CreateFolder(c.params.gitPath); err != nil {
		return false, err
	}

	var cloneOutput bytes.Buffer
	repo, err := git.PlainCloneContext(ctx, c.params.gitPath, true, &git.CloneOptions{
		URL: c.params.repoData.Clone,
		Auth: &http.BasicAuth{
			Username: c.params.auth.username,
			Password: c.params.auth.token,
		},
		SingleBranch: false,
		Tags:         git.AllTags,
		NoCheckout:   true,
		Progress:     &cloneOutput,
	})

	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		c.tracer.Log(common.MsgRepoAlreadyExists, c.params.repoSlug)
		return false, nil
	}

	if errors.Is(err, plumbing.ErrReferenceNotFound) || errors.Is(err, transport.ErrEmptyRemoteRepository) {
		c.tracer.Stop(common.MsgGitCloneEmptyRepo, c.params.repoSlug)
		return true, nil
	}

	if err != nil {
		c.tracer.LogError(common.ErrGitClone, c.params.repoSlug, err, cloneOutput.String())
		return false, fmt.Errorf("failed to clone repo %s from %q: %w", c.params.repoSlug, c.params.repoData.Clone, err)
	}

	refSpecs := []config.RefSpec{"refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*"}
	refSpecs = append(refSpecs, c.params.pullreqRef...)

	var fetchOutput bytes.Buffer
	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: refSpecs,
		Auth: &http.BasicAuth{
			Username: c.params.auth.username,
			Password: c.params.auth.token,
		},
		Force:    true,
		Progress: &fetchOutput,
	})

	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		c.tracer.LogError(common.ErrGitFetch, c.params.repoSlug, err, fetchOutput.String())
		return false, fmt.Errorf("failed to sync repo %s from %q: %w", c.params.repoSlug, c.params.repoData.Clone, err)
	}

	return false, nil
}
