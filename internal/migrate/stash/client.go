package stash

import (
	"context"

	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

type Client interface {
	ListPRComments(ctx context.Context, repoSlug string, prNumber int, opts types.ListOptions, tracer tracer.Tracer) ([]*types.PRComment, *scm.Response, error)
	ListBranchRules(ctx context.Context, repoSlug string, opts types.ListOptions) ([]*types.BranchRule, *scm.Response, error)
}
