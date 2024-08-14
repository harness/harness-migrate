package github

import (
	"context"
	"fmt"

	"github.com/harness/harness-migrate/internal/checkpoint"
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

	checkpointDataKey := fmt.Sprintf(common.RuleCheckpointData, repoSlug)
	val, ok, err := checkpoint.GetCheckpointData[[]*types.BranchRule](e.checkpointManager, checkpointDataKey)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointDataRead, err)
	}
	if ok && val != nil {
		allRules = append(allRules, val...)
	}

	checkpointPageKey := fmt.Sprintf(common.RuleCheckpointPage, repoSlug)
	checkpointPageIntfc, ok := e.checkpointManager.GetCheckpoint(checkpointPageKey)
	var checkpointPage int
	if ok && checkpointPageIntfc != nil {
		checkpointPage = int(checkpointPageIntfc.(float64))
		opts.Page = checkpointPage
	}

	// all rules pages are done
	if checkpointPage == -1 {
		rules, err := e.fetchRulesets(ctx, repoSlug, opts)
		if err != nil {
			e.tracer.Stop(common.MsgFailedExportBranchRules, repoSlug)
			return nil, fmt.Errorf(common.ErrListBranchRulesets, repoSlug, err)
		}
		allRules = append(allRules, rules...)

		e.tracer.Stop(common.MsgCompleteExportBranchRules, len(allRules), repoSlug)
		return allRules, nil
	}
	// classic branch protection rules
	for {
		rules, resp, err := e.ListBranchRulesInternal(ctx, repoSlug, opts)
		if err != nil {
			e.tracer.LogError(common.ErrListBranchRules, repoSlug, err)
			e.tracer.Stop(common.MsgFailedExportBranchRules, repoSlug)
			return nil, fmt.Errorf(common.ErrListBranchRules, repoSlug, err)
		}
		allRules = append(allRules, rules...)

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allRules)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointRulesDataSave, repoSlug, err)
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, resp.Page.Next)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointRulesPageSave, repoSlug, err)
		}

		if resp.Page.NextURL == "" {
			break
		}
		opts.URL = resp.Page.NextURL
	}

	err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointRulesPageSave, repoSlug, err)
	}

	// rule sets
	rules, err := e.fetchRulesets(ctx, repoSlug, opts)
	if err != nil {
		e.tracer.Stop(common.MsgFailedExportBranchRules, repoSlug)
		return nil, fmt.Errorf(common.ErrListBranchRulesets, repoSlug, err)
	}
	allRules = append(allRules, rules...)

	e.tracer.Stop(common.MsgCompleteExportBranchRules, len(allRules), repoSlug)
	return allRules, nil
}

func (e *Export) fetchRulesets(ctx context.Context,
	repoSlug string,
	opts types.ListOptions,
) ([]*types.BranchRule, error) {
	allRulesets := []*types.BranchRule{}
	allRules := []*types.BranchRule{}

	checkpointDataKey := fmt.Sprintf(common.RuleSetCheckpointData, repoSlug)
	val, ok, err := checkpoint.GetCheckpointData[[]*types.BranchRule](e.checkpointManager, checkpointDataKey)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointDataRead, err)
	}
	if ok && val != nil {
		allRules = append(allRules, val...)
	}

	checkpointPageKey := fmt.Sprintf(common.RuleSetCheckpointPage, repoSlug)
	checkpointPageIntfc, ok := e.checkpointManager.GetCheckpoint(checkpointPageKey)
	var checkpointPage int
	if ok && checkpointPageIntfc != nil {
		checkpointPage = int(checkpointPageIntfc.(float64))
		opts.Page = checkpointPage
	}

	// all ruleset pages are done
	if checkpointPage == -1 {
		return allRules, nil
	}

	for {
		ruleSets, _, err := e.ListBranchRuleSets(ctx, repoSlug, opts)
		if err != nil {
			return nil, fmt.Errorf(common.ErrListBranchRulesets, repoSlug, err)
		}
		allRulesets = append(allRulesets, ruleSets...)

		if len(ruleSets) == 0 {
			break
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allRulesets)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointRulesDataSave, repoSlug, err)
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, opts.Page+1)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointRulesPageSave, repoSlug, err)
		}

		opts.Page += 1
	}

	err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointRulesPageSave, repoSlug, err)
	}

	for _, r := range allRulesets {
		rule, _, err := e.FindBranchRuleset(ctx, repoSlug, r.ID)
		if err != nil {
			e.tracer.LogError(common.ErrFetchBranchRuleset, r.ID, repoSlug, err)
			return nil, fmt.Errorf(common.ErrFetchBranchRuleset, r.ID, repoSlug, err)
		}
		allRules = append(allRules, rule)
	}

	return allRules, nil
}
