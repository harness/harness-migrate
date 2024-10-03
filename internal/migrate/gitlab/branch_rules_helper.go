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
	"strconv"

	"github.com/harness/harness-migrate/internal/migrate"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/types/enum"
)

func (e *Export) convertBranchRules(ctx context.Context, from []*branchRule, MRRules mergeRequestRules, prj string) []*types.BranchRule {
	rules := make([]*types.BranchRule, len(from))
	for i, rule := range from {
		rules[i] = e.convertBranchRule(ctx, rule, MRRules, prj)
	}
	return rules
}

func (e *Export) convertBranchRule(
	ctx context.Context,
	from *branchRule,
	MRRules mergeRequestRules,
	prj string,
) *types.BranchRule {
	var logs []string
	var warningMsg string

	bypass := types.Bypass{}

	pullReq := types.PullReq{
		Approvals: types.Approvals{
			RequireCodeOwners: from.CodeOwnerRequired,
		},
		Comments: types.Comments{
			RequireResolveAll: MRRules.RequireResolveComments,
		},
		Merge: types.Merge{
			StrategiesAllowed: mapMergeMethod(MRRules),
			DeleteBranch:      MRRules.DeleteBranch,
		},
	}

	// by default Gitlab's protected branches prevent merge
	if len(from.MergeAccess) == 0 {
		pullReq.Block = true
	}

	for _, mergeRule := range from.MergeAccess {
		if mergeRule.AccessLevel == levelNoAccess {
			pullReq.Block = true
		} else {
			// Other Access Levels are "Developer, Maintainer, Admin"
			// From Gitlab docs:
			// "the most permissive rule determines the level of protection for the branch"
			// https://docs.gitlab.com/ee/user/project/repository/branches/protected.html
			pullReq.Block = false
		}
		if mergeRule.GroupID != nil || mergeRule.DeployeKeyID != nil {
			warningMsg = fmt.Sprintf("[%s] Skipped adding group IDs and/or deploy key IDs to bypass list for branch %q rule"+
				" of project %q as we do not support it as of now.", enum.LogLevelWarning, from.Name, prj)
			logs = append(logs, warningMsg)
		}

		if mergeRule.UserID != nil {
			user, _, err := e.GetUserByID(ctx, *mergeRule.UserID)
			if err != nil {
				e.tracer.LogError("failed to get user email with ID %d for branch rule bypass list: %w", mergeRule.UserID, err)
				continue
			}
			bypass.UserEmails = append(bypass.UserEmails, user.PublicEmail)
		}

		if mergeRule.AccessLevel == levelAdmin {
			bypass.RepoOwners = true
		}
	}

	for _, pushRule := range from.PushAccess {
		if pushRule.GroupID != nil || pushRule.DeployeKeyID != nil {
			warningMsg = fmt.Sprintf("[%s] Skipped adding group IDs and/or deploy key IDs to bypass list for branch %q rule"+
				" of project %q as we do not support it as of now.", enum.LogLevelWarning, from.Name, prj)
			logs = append(logs, warningMsg)
		}

		if pushRule.UserID != nil {
			user, _, err := e.GetUserByID(ctx, *pushRule.UserID)
			if err != nil {
				e.tracer.LogError("failed to get user email with ID %d for branch rule bypass list: %w", pushRule.UserID, err)
				continue
			}
			bypass.UserEmails = append(bypass.UserEmails, user.PublicEmail)
		}

		if pushRule.AccessLevel == levelAdmin {
			bypass.RepoOwners = true
		}
	}

	definition := types.Definition{
		Bypass:  bypass,
		PullReq: pullReq,
		Lifecycle: types.Lifecycle{
			UpdateForbidden:      true,
			UpdateForceForbidden: !from.AllowForcePush,
		},
	}

	return &types.BranchRule{
		ID:         from.Id,
		Name:       migrate.DisplayNameToIdentifier(from.Name, "rule", strconv.Itoa(from.Id)),
		State:      enum.RuleStateActive, // all gitlab rules are active
		Definition: definition,
		Pattern: types.Pattern{
			IncludedPatterns: []string{from.Name},
		},
	}
}

func mapMergeMethod(from mergeRequestRules) []string {
	strategiesAllowed := []string{"merge"}
	switch from.SquashOption {
	case squashOptionAlways:
		return []string{"squash"}
	case squashOptionOn, squashOptionOff:
		strategiesAllowed = append(strategiesAllowed, "squash")
	}

	switch from.MergeMethod {
	case mergeMethodRebase:
		strategiesAllowed = append(strategiesAllowed, "rebase")
	case mergeMethodFastForward:
		strategiesAllowed = append(strategiesAllowed, []string{"rebase", "fast-forward"}...)
	}

	return strategiesAllowed
}
