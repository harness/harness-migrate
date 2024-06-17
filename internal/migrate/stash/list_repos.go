package stash

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
	e.tracer.Start(common.MsgStartRepoList, "bitbucket", e.org)
	opts := scm.ListOptions{Page: 1, Size: 25}
	var allRepos []*scm.Repository

	if e.repository != "" {
		repoSlug := strings.Join([]string{e.org, e.repository}, "/")
		repo, _, err := e.stash.Repositories.Find(ctx, repoSlug)
		if err != nil {
			e.tracer.LogError(common.ErrRepoList, err)
			return nil, fmt.Errorf("failed to get the repo %s: %w", repoSlug, err)
		}
		e.tracer.Stop(common.MsgCompleteRepoList, 1)
		return mapRepository([]*scm.Repository{repo}), nil
	}

	for {
		repos, resp, err := e.stash.Repositories.ListNamespace(ctx, e.org, opts)
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
	return mapRepository(allRepos), nil
}

func mapRepository(repos []*scm.Repository) []types.RepoResponse {
	r := make([]types.RepoResponse, len(repos))
	for i, repo := range repos {
		repoSlug := make([]string, 0)

		if repo.Namespace != "" {
			repoSlug = append(repoSlug, repo.Namespace)
		}
		if repo.Name != "" {
			repoSlug = append(repoSlug, repo.Name)
		}

		r[i] = types.RepoResponse{Repository: *repo, RepoSlug: strings.Join(repoSlug, "/")}
	}
	return r
}
