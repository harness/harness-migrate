package stash

import (
	"context"
	"errors"
	"fmt"

	"github.com/harness/harness-migrate/internal/common"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func (e *Export) FetchPullRequestRefs(_ context.Context, repo *git.Repository, repoSlug string, stashLogin string, stashToken string) error {
	e.tracer.Start(common.MsgGitFetchRef, repoSlug)
	// skip fetching references for empty repos.
	if repo == nil {
		e.tracer.Stop(common.MsgCompleteGitFetchRef, repoSlug)
		return nil
	}

	refSpecs := []config.RefSpec{"refs/pull-requests/*/from:refs/pullreq/*/head"}
	err := repo.Fetch(&git.FetchOptions{
		RefSpecs: refSpecs,
		Auth: &http.BasicAuth{
			Username: stashLogin,
			Password: stashToken,
		},
		Force: true,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		e.tracer.LogError(common.ErrGitFetch, repoSlug, err)
		e.tracer.Stop(common.ErrGitFetch, repoSlug, err)
		return fmt.Errorf("failed to fetch repo pull requests references %s: %w", repoSlug, err)
	}

	e.tracer.Stop(common.MsgCompleteGitFetchRef, repoSlug)
	return nil
}
