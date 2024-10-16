// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitlab

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/migrate"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/types/enum"
)

func (e *Export) convertMergeRequestRule(from mergeRequestRules) *types.BranchRule {
	return &types.BranchRule{
		ID:    0,
		Name:  migrate.DisplayNameToIdentifier("", "merge_request", "rule"),
		State: enum.RuleStateActive,
		Pattern: types.Pattern{
			IncludedPatterns: []string{}, // set no pattern to apply on all branches
		},
		Definition: types.Definition{
			PullReq: types.PullReq{
				Comments: types.Comments{
					RequireResolveAll: from.RequireResolveComments,
				},
				Merge: types.Merge{
					StrategiesAllowed: mapMergeMethod(from),
					DeleteBranch:      from.DeleteBranch,
				},
			},
		},
	}
}

func (e *Export) convertBranchRules(ctx context.Context, from []*branchRule, prj string) []*types.BranchRule {
	var rules []*types.BranchRule
	for _, rule := range from {
		rules = append(rules, e.convertBranchRule(ctx, rule, prj)...)
	}
	return rules
}

func (e *Export) convertBranchRule(
	ctx context.Context,
	from *branchRule,
	prj string,
) []*types.BranchRule {
	var rules []*types.BranchRule
	var logs []string
	var warningMsg string

	rule := &types.BranchRule{
		ID:    from.Id,
		Name:  migrate.DisplayNameToIdentifier(from.Name, "rule", strconv.Itoa(from.Id)),
		State: enum.RuleStateActive, // all gitlab rules are active
		Pattern: types.Pattern{
			IncludedPatterns: []string{from.Name},
		},
	}

	rule.Definition.PullReq = types.PullReq{
		Approvals: types.Approvals{
			RequireCodeOwners: from.CodeOwnerRequired,
		},
	}

	if len(from.MergeAccess) >= 0 { // No one is allowed to merge changes.
		rule.UpdateForbidden = true
		rule.PullReq.Merge.Block = true
		rule.UpdateForceForbidden = true
	}

	for _, mergeRule := range from.MergeAccess {
		if mergeRule.AccessLevel == levelAdmin {
			rule.Bypass.RepoOwners = true
		}

		if mergeRule.GroupID != nil || mergeRule.DeployeKeyID != nil {
			warningMsg = fmt.Sprintf("[%s] Skipped adding group IDs and/or deploy key IDs to bypass list for branch %q rule"+
				" of project %q as we do not support it as of now.", enum.LogLevelWarning, from.Name, prj)
			logs = append(logs, warningMsg)
		}

		if mergeRule.UserID != nil {
			email, err := e.FindEmailByUserID(ctx, *mergeRule.UserID)
			if err != nil {
				e.tracer.LogError("failed to get user email with ID %d for branch rule bypass list: %w", mergeRule.UserID, err)
				continue
			}
			rule.Bypass.UserEmails = append(rule.Bypass.UserEmails, email)
		}
	}

	if len(from.PushAccess) == 0 { // Changes require a merge request.
		rule.UpdateForbidden = true
	}

	rules = append(rules, rule)

	if len(from.PushAccess) > 0 {
		r := &types.BranchRule{
			ID:    from.Id,
			Name:  migrate.DisplayNameToIdentifier(from.Name, "", "pushrule_"+strconv.Itoa(from.Id)),
			State: enum.RuleStateActive, // all gitlab rules are active
			Pattern: types.Pattern{
				IncludedPatterns: []string{from.Name},
			},
			Definition: types.Definition{},
		}
		r.UpdateForbidden = true
		r.UpdateForceForbidden = true

		for _, pushRule := range from.PushAccess {
			if pushRule.AccessLevel == levelAdmin {
				r.Bypass.RepoOwners = true
			}

			if pushRule.GroupID != nil || pushRule.DeployeKeyID != nil {
				warningMsg = fmt.Sprintf("[%s] Skipped adding group IDs and/or deploy key IDs to bypass list for branch %q rule"+
					" of project %q as we do not support it as of now.", enum.LogLevelWarning, from.Name, prj)
				logs = append(logs, warningMsg)
			}

			if pushRule.UserID != nil {
				email, err := e.FindEmailByUserID(ctx, *pushRule.UserID)
				if err != nil {
					e.tracer.LogError("failed to get user email with ID %d for branch rule bypass list: %w", pushRule.UserID, err)
					continue
				}
				r.Bypass.UserEmails = append(r.Bypass.UserEmails, email)
			}
		}

		rules = append(rules, r)
	}

	for _, l := range logs {
		if err := e.fileLogger.Log(l); err != nil {
			log.Default().Printf("failed to log the not supported branch rules for project %q: %v", prj, err)
			return rules
		}
	}

	e.report[prj].ReportErrors(gitexporter.ReportTypeBranchRules, from.Name, logs)
	return rules
}

func mapMergeMethod(from mergeRequestRules) []string {
	strategiesAllowed := []string{}

	switch from.MergeMethod {
	case mergeMethodMerge:
		strategiesAllowed = append(strategiesAllowed, "merge")
	case mergeMethodRebase:
		strategiesAllowed = append(strategiesAllowed, []string{"merge", "rebase"}...)
	case mergeMethodFastForward:
		strategiesAllowed = append(strategiesAllowed, []string{"rebase", "fast-forward"}...)
	}

	switch from.SquashOption {
	case squashOptionAlways:
		return []string{"squash"}
	case squashOptionOn, squashOptionOff:
		strategiesAllowed = append(strategiesAllowed, "squash")
	}

	return strategiesAllowed
}
