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
	var allRules []*types.BranchRule
	var allRuleSets []*types.BranchRule

	// classic branch protection rules
	for {
		rules, res, err := e.ListBranchRulesInternal(ctx, repoSlug, opts)
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
		ruleSets, _, err := e.ListBranchRuleSets(ctx, repoSlug, opts)
		if err != nil {
			e.tracer.LogError(common.ErrListBranchRulesets, repoSlug, err)
			e.tracer.Stop(common.MsgFailedExportBranchRules, repoSlug)
			return nil, fmt.Errorf(common.ErrListBranchRulesets, repoSlug, err)
		}
		allRuleSets = append(allRuleSets, ruleSets...)

		if len(ruleSets) == 0 {
			break
		}
		opts.Page += 1
	}
	for _, r := range allRuleSets {
		rule, _, err := e.FindBranchRuleset(ctx, repoSlug, r.ID)
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
