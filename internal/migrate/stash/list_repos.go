package stash

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

func (e *Export) ListRepositories(
	ctx context.Context,
	params types.ListOptions,
) ([]types.RepoResponse, error) {
	e.tracer.Start(common.MsgStartRepoList, "stash", e.project)
	opts := scm.ListOptions{Page: params.Page, Size: params.Size}
	var allRepos []*scm.Repository

	if e.repository != "" {
		repoSlug := strings.Join([]string{e.project, e.repository}, "/")
		repo, _, err := e.stash.Repositories.Find(ctx, repoSlug)
		// stash returns EOF if call find for an empty repo
		if err != nil && !errors.Is(err, io.EOF) {
			e.tracer.LogError(common.ErrRepoList, err)
			return nil, fmt.Errorf("failed to get the repo %s: %w", repoSlug, err)
		}
		e.tracer.Stop(common.MsgCompleteRepoList, 1)
		return common.MapRepository([]*scm.Repository{repo}), nil
	}

	for {
		repos, resp, err := e.stash.Repositories.ListNamespace(ctx, e.project, opts)
		if err != nil {
			e.tracer.LogError(common.ErrRepoList, err)
			return nil, fmt.Errorf("failed to get repos for project %s: %w", e.project, err)
		}
		allRepos = append(allRepos, repos...)

		if resp.Page.Next == 0 {
			break
		}
		opts.Page = resp.Page.Next
	}

	e.tracer.Stop(common.MsgCompleteRepoList, len(allRepos))
	return common.MapRepository(allRepos), nil
}
