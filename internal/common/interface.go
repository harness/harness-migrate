package common

import (
	"context"

	"github.com/harness/harness-migrate/internal/types"
)

// Interface helps to support a generic way of doing export for all git providers
type Interface interface {
	ListRepositories(ctx context.Context, opts types.ListRepoOptions) ([]types.RepoResponse, error)
	ListPullRequest(ctx context.Context, repoSlug string, opts types.PullRequestListOptions) ([]types.PRResponse, error)
	PullRequestReviewers(ctx context.Context, prNumber int) error
	PullRequestComments(ctx context.Context, prNumber int) error
}
