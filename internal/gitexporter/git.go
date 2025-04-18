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
	tracer     tracer.Tracer
}

type credentials struct {
	username string
	token    string
}

type nativeGitCloner struct {
	params cloneParams
}

type goGitCloner struct {
	params cloneParams
}

func (e *Exporter) CloneRepository(
	ctx context.Context,
	repoData scm.Repository,
	repoPath string,
	repoSlug string,
	pullreqRef []config.RefSpec,
	noGitLFS bool,
	tracer tracer.Tracer,
) (bool, int64, error) {
	tracer.Start(common.MsgStartGitClone, repoSlug)

	if !noGitLFS {
		if err := command.CheckGitLFSInstallation(); err != nil {
			return false, 0, err
		}
	}

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
		tracer: tracer,
	}

	cloner := e.selectCloner(params)
	isEmpty, err := cloner.clone(ctx)
	if err != nil {
		return false, 0, fmt.Errorf("failed to clone repo %s: %w", repoSlug, err)
	}

	lfsObjectCount, err := e.checkLFSObjects(ctx, gitPath, repoSlug, noGitLFS, tracer)
	if err != nil {
		return isEmpty, 0, err
	}

	tracer.Stop(common.MsgCompleteGitClone, repoSlug)
	return isEmpty, lfsObjectCount, nil
}

func (e *Exporter) selectCloner(params cloneParams) gitCloner {
	if err := command.CheckGitInstallation(); err == nil {
		return &nativeGitCloner{params: params}
	}
	return &goGitCloner{params: params}
}

func (e *Exporter) checkLFSObjects(
	ctx context.Context,
	gitPath string,
	repoSlug string,
	noGitLFS bool,
	tracer tracer.Tracer,
) (int64, error) {
	if noGitLFS {
		return 0, nil
	}

	lfsObjectCount, err := command.HasLFSObjects(ctx, gitPath, []string{})
	if err != nil {
		tracer.LogError("git-lfs-check", repoSlug, err)
		return 0, fmt.Errorf("failed to check LFS objects for repo %s: %w", repoSlug, err)
	}

	return lfsObjectCount, nil
}

func (c *nativeGitCloner) clone(ctx context.Context) (bool, error) {
	if err := util.CreateFolder(c.params.gitPath); err != nil {
		return false, err
	}

	if _, err := os.Stat(filepath.Join(c.params.gitPath, ".git")); err == nil {
		c.params.tracer.Log(common.MsgRepoAlreadyExists, c.params.repoSlug)
		return false, nil
	}

	gitEnv := []string{
		"GIT_TERMINAL_PROMPT=0",
		fmt.Sprintf("GIT_USERNAME=%s", c.params.auth.username),
		fmt.Sprintf("GIT_PASSWORD=%s", c.params.auth.token),
	}

	cloneArgs := []string{
		"clone",
		"--mirror",
		"--no-checkout",
		c.params.repoData.Clone,
		".",
	}

	output, err := command.RunGitCommand(ctx, c.params.gitPath, gitEnv, cloneArgs...)
	if err != nil {
		if strings.Contains(string(output), "You appear to have cloned an empty repository.") {
			c.params.tracer.Stop(common.MsgGitCloneEmptyRepo, c.params.repoSlug)
			return true, nil
		}

		c.params.tracer.LogError(common.ErrGitClone, c.params.repoSlug, err, string(output))
		return false, fmt.Errorf("failed to clone repo %s from %q: %w", c.params.repoSlug, c.params.repoData.Clone, err)
	}

	output, err = command.RunGitCommand(ctx, c.params.gitPath, gitEnv, "fetch", "--all", "--force", "--prune")
	if err != nil {
		c.params.tracer.LogError(common.ErrGitFetch, c.params.repoSlug, err, string(output))
		return false, fmt.Errorf("failed to sync repo %s from %q: %w", c.params.repoSlug, c.params.repoData.Clone, err)
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
		c.params.tracer.Log(common.MsgRepoAlreadyExists, c.params.repoSlug)
		return false, nil
	}

	if errors.Is(err, plumbing.ErrReferenceNotFound) || errors.Is(err, transport.ErrEmptyRemoteRepository) {
		c.params.tracer.Stop(common.MsgGitCloneEmptyRepo, c.params.repoSlug)
		return true, nil
	}

	if err != nil {
		c.params.tracer.LogError(common.ErrGitClone, c.params.repoSlug, err, cloneOutput.String())
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
		c.params.tracer.LogError(common.ErrGitFetch, c.params.repoSlug, err, fetchOutput.String())
		return false, fmt.Errorf("failed to sync repo %s from %q: %w", c.params.repoSlug, c.params.repoData.Clone, err)
	}

	return false, nil
}
