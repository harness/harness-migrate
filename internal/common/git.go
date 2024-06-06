package common

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/drone/go-scm/scm"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/util"
)

func (e *Exporter) CloneRepository(
	ctx context.Context,
	repoData scm.Repository,
	repoPath string,
	repoSlug string, //for logging
	tracer tracer.Tracer,
) (*git.Repository, error) {
	tracer.Start(MsgStartGitClone, repoSlug)
	gitPath := filepath.Join(repoPath, "git")
	if err := util.CreateFolder(gitPath); err != nil {
		return nil, err
	}

	repo, err := git.PlainCloneContext(ctx, gitPath, true, &git.CloneOptions{
		URL: repoData.Clone,
		Auth: &http.BasicAuth{
			Username: e.ScmLogin,
			Password: e.ScmToken,
		},
		SingleBranch: false,
		Tags:         git.AllTags,
		NoCheckout:   true,
	})
	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		tracer.Log(MsgRepoAlreadyExists, repoSlug)
		return nil, nil
	}
	if err != nil && !errors.Is(err, transport.ErrEmptyRemoteRepository) {
		tracer.LogError(ErrGitCloneMsg, repoSlug, err)
		return nil, fmt.Errorf("failed to clone repo %s from %q: %w", repoSlug, repoData.Clone, err)
	}

	refSpecs := []config.RefSpec{"refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*"}
	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: refSpecs,
		Auth: &http.BasicAuth{
			Username: e.ScmLogin,
			Password: e.ScmToken,
		},
		Force: true,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		tracer.LogError(ErrGitCloneMsg, repoSlug, err)
		return nil, fmt.Errorf("failed to sync repo %s: %w", repoSlug, err)
	}

	tracer.Stop(MsgCompleteGitClone, repoSlug)

	return repo, nil
}
