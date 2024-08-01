package migrate

import (
	"context"

	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

type ClientInterface interface {
	ListPRComments(ctx context.Context, repoSlug string, prNumber int, opts types.ListOptions) ([]*types.PRComment, *scm.Response, error)
	ListBranchRules(ctx context.Context, repoSlug string, logger gitexporter.Logger, opts types.ListOptions) ([]*types.BranchRule, *scm.Response, error)
	ListBranchRulesets(ctx context.Context, repoSlug string, opts types.ListOptions) (*[]types.BranchRule, *scm.Response, error)
	FindBranchRuleset(ctx context.Context, repoSlug string, logger gitexporter.Logger, ruleID int) (*types.BranchRule, *scm.Response, error)
}
