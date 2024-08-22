package gitexporter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/util"
	"github.com/harness/harness-migrate/types"

	"github.com/drone/go-scm/scm"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func (e *Exporter) CloneRepository(
	ctx context.Context,
	repoData scm.Repository,
	repoPath string,
	repoSlug string, //for logging
	pullreqRef []config.RefSpec,
	tracer tracer.Tracer,
) (bool, error) {
	var isEmpty bool
	tracer.Start(common.MsgStartGitClone, repoSlug)
	gitPath := filepath.Join(repoPath, types.GitDir)
	if err := util.CreateFolder(gitPath); err != nil {
		return isEmpty, err
	}

	var cloneOutput bytes.Buffer
	repo, err := git.PlainCloneContext(ctx, gitPath, true, &git.CloneOptions{
		URL: repoData.Clone,
		Auth: &http.BasicAuth{
			Username: e.ScmLogin,
			Password: e.ScmToken,
		},
		SingleBranch: false,
		Tags:         git.AllTags,
		NoCheckout:   true,
		Progress:     &cloneOutput,
	})

	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		tracer.Log(common.MsgRepoAlreadyExists, repoSlug)
		return isEmpty, nil
	}

	// no need to fetch ref if repo is empty
	// empty Github repos resulted in plumbing.ErrReferenceNotFound
	// empty stash repos resulted in transport.ErrEmptyRemoteRepository
	if errors.Is(err, plumbing.ErrReferenceNotFound) || errors.Is(err, transport.ErrEmptyRemoteRepository) {
		isEmpty = true
		tracer.Stop(common.MsgGitCloneEmptyRepo, repoSlug)
		return isEmpty, nil
	}

	if err != nil {
		tracer.LogError(common.ErrGitClone, repoSlug, err, cloneOutput.String())
		return isEmpty, fmt.Errorf("failed to clone repo %s from %q: %w", repoSlug, repoData.Clone, err)
	}

	refSpecs := []config.RefSpec{"refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*"}
	refSpecs = append(refSpecs, pullreqRef...)

	var fetchOutput bytes.Buffer
	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: refSpecs,
		Auth: &http.BasicAuth{
			Username: e.ScmLogin,
			Password: e.ScmToken,
		},
		Force:    true,
		Progress: &fetchOutput,
	})

	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		tracer.LogError(common.ErrGitFetch, repoSlug, err, fetchOutput.String())
		return isEmpty, fmt.Errorf("failed to sync repo %s from %q: %w", repoSlug, repoData.Clone, err)
	}

	tracer.Stop(common.MsgCompleteGitClone, repoSlug)

	return isEmpty, nil
}
