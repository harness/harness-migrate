package stash

import (
	"context"
	"fmt"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"
)

func (e *Export) ListBranchRules(
	ctx context.Context,
	repoSlug string,
	opts types.ListOptions,
) ([]*types.BranchRule, error) {
	e.tracer.Start(common.MsgStartExportBranchRules, repoSlug)
	allRules := []*types.BranchRule{}

	for {
		rules, res, err := e.stash.ListBranchRules(ctx, repoSlug, e.fileLogger, opts)
		if err != nil {
			e.tracer.LogError(common.ErrListBranchRules, repoSlug, err)
			e.tracer.Stop(common.MsgFailedExportBranchRules, repoSlug)
			return nil, fmt.Errorf(common.ErrListBranchRules, repoSlug, err)
		}
		allRules = append(allRules, rules...)
		if res.Page.Next == 0 {
			break
		}
		opts.Page = res.Page.Next
	}

	e.tracer.Stop(common.MsgCompleteExportBranchRules, len(allRules), repoSlug)
	return allRules, nil
}
