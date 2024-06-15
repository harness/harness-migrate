package gitexporter

import (
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
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func (e *Exporter) CloneRepository(
	ctx context.Context,
	repoData scm.Repository,
	repoPath string,
	repoSlug string, //for logging
	tracer tracer.Tracer,
) (*git.Repository, error) {
	tracer.Start(common.MsgStartGitClone, repoSlug)
	gitPath := filepath.Join(repoPath, types.GitDir)
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
		tracer.Log(common.MsgRepoAlreadyExists, repoSlug)
		return nil, nil
	}
	if err != nil && !errors.Is(err, transport.ErrEmptyRemoteRepository) {
		tracer.LogError(common.ErrGitClone, repoSlug, err)
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
		tracer.LogError(common.ErrGitFetch, repoSlug, err)
		return nil, fmt.Errorf("failed to sync repo %s from %q: %w", repoSlug, repoData.Clone, err)
	}

	tracer.Stop(common.MsgCompleteGitClone, repoSlug)

	return repo, nil
}
