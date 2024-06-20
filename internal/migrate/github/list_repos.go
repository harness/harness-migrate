package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

func (e *Export) ListRepositories(
	ctx context.Context,
	_ types.ListOptions,
) ([]types.RepoResponse, error) {
	e.tracer.Start(common.MsgStartRepoList, "github", e.org)
	opts := scm.ListOptions{Page: 1, Size: 25}
	var allRepos []*scm.Repository

	if e.repository != "" {
		repoSlug := strings.Join([]string{e.org, e.repository}, "/")
		repo, _, err := e.github.Repositories.Find(ctx, repoSlug)
		if err != nil {
			e.tracer.LogError(common.ErrRepoList, err)
			return nil, fmt.Errorf("failed to get the repo %s: %w", repoSlug, err)
		}
		e.tracer.Stop(common.MsgCompleteRepoList, 1)
		return common.MapRepository([]*scm.Repository{repo}), nil
	}

	for {
		repos, resp, err := e.github.Repositories.ListNamespace(ctx, e.org, opts)
		if err != nil {
			e.tracer.LogError(common.ErrRepoList, err)
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
