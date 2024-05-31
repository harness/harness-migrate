package stash

import (
	"context"
	"fmt"
	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/codeerror"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
	"strings"
)

type Export struct {
	Stash    *scm.Client
	StashOrg string

	Tracer tracer.Tracer
}

func (e Export) ListRepositories(
	ctx context.Context,
	_ types.ListRepoOptions,
) ([]types.RepoResponse, error) {
	e.Tracer.Start("starting repositories listing for bitbucket org %s", e.StashOrg)
	opts := scm.ListOptions{Size: 25}
	var allRepos []*scm.Repository

	for {
		repos, resp, err := e.Stash.Repositories.ListNamespace(ctx, e.StashOrg, opts)
		if err != nil {
			e.Tracer.LogError("cannot fetch repositories due to error :%w", err)
		}
		allRepos = append(allRepos, repos...)

		if resp.Page.Next == 0 {
			break
		}
		opts.Page = resp.Page.Next
	}
	e.Tracer.Stop("repositories listing complete. Total repos: %d", len(allRepos))

	return mapRepository(allRepos), nil
}

func (e Export) ListPullRequest(
	ctx context.Context,
	repoSlug string,
	_ types.PullRequestListOptions,
) ([]types.PRResponse, error) {
	e.Tracer.Start("Starting Pull request import for repo %s", repoSlug)
	var allPrs []*scm.PullRequest
	opts := scm.PullRequestListOptions{Size: 25}
	for {
		prs, resp, err := e.Stash.PullRequests.List(ctx, repoSlug, opts)
		if err != nil {
			e.Tracer.LogError("cannot list pr due to error :%w", err)
			return nil, fmt.Errorf("cannot list pr: %w", err)
		}
		allPrs = append(allPrs, prs...)
		if resp.Page.Next == 0 {
			break
		}
		opts.Page = resp.Page.Next
	}
	e.Tracer.Stop("Pull request import finished for repo %s with total %d PRs", repoSlug, len(allPrs))

	return common.MapPullRequest(allPrs), nil
}

func (e Export) PullRequestReviewers(
	context.Context,
	int) error {
	return &codeerror.ErrorOpNotSupported{Name: "pullreqreview"}
}

func (e Export) PullRequestComments(
	context.Context,
	int,
) error {
	return &codeerror.ErrorOpNotSupported{Name: "pullreqcomment"}
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
