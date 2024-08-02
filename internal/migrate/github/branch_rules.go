package github

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
	allRulesets := []*types.BranchRule{}

	// classic branch protection rules
	for {
		rules, res, err := e.github.ListBranchRules(ctx, repoSlug, e.fileLogger, opts)
		if err != nil {
			e.tracer.LogError(common.ErrListBranchRules, repoSlug, err)
			e.tracer.Stop(common.MsgFailedExportBranchRules, repoSlug)
			return nil, fmt.Errorf(common.ErrListBranchRules, repoSlug, err)
		}
		allRules = append(allRules, rules...)

		if res.Page.NextURL == "" {
			break
		}
		opts.URL = res.Page.NextURL
	}

	// rulesets
	for {
		rulesets, _, err := e.github.ListBranchRulesets(ctx, repoSlug, opts)
		if err != nil {
			e.tracer.LogError(common.ErrListBranchRulesets, repoSlug, err)
			e.tracer.Stop(common.MsgFailedExportBranchRules, repoSlug)
			return nil, fmt.Errorf(common.ErrListBranchRulesets, repoSlug, err)
		}
		allRulesets = append(allRulesets, rulesets...)

		if len(rulesets) == 0 {
			break
		}
		opts.Page += 1
	}
	for _, r := range allRulesets {
		rule, _, err := e.github.FindBranchRuleset(ctx, repoSlug, e.fileLogger, r.ID)
		if err != nil {
			e.tracer.LogError(common.ErrFetchBranchRuleset, r.ID, repoSlug, err)
			e.tracer.Stop(common.MsgFailedExportBranchRules, repoSlug)
			return nil, fmt.Errorf(common.ErrFetchBranchRuleset, r.ID, repoSlug, err)
		}
		allRules = append(allRules, rule)
	}

	e.tracer.Stop(common.MsgCompleteExportBranchRules, len(allRules), repoSlug)
	return allRules, nil
}
