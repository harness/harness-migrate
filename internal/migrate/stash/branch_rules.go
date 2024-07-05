package stash

import (
	"context"
	"fmt"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/types"
)

func (e *Export) ListBranchRules(
	ctx context.Context,
	repoSlug string,
	logger gitexporter.Logger,
	opts types.ListOptions,
) ([]*types.BranchRule, error) {
	e.tracer.Start(common.MsgStartBranchRulesExport, repoSlug)
	allRules := []*types.BranchRule{}
	defer func() {
		e.tracer.Stop(common.MsgCompleteBranchRulesExport, repoSlug, len(allRules))
	}()
	for {
		rules, res, err := e.stash.ListBranchRules(ctx, repoSlug, logger, opts)
		if err != nil {
			e.tracer.LogError(common.ErrBranchRulesList, repoSlug, err)
			return nil, fmt.Errorf(common.ErrBranchRulesList, repoSlug, err)
		}
		allRules = append(allRules, rules...)
		if res.Page.Next == 0 {
			break
		}
		opts.Page = res.Page.Next
	}
	return allRules, nil
}
