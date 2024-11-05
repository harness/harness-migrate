// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bitbucket

import (
	"context"
	"fmt"
	"log"

	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/migrate"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/types/enum"
)

const logMessage = "[%s] Skipped mapping %q branch rule for pattern %q of repo %q as we do not support it as of now."

func (e *Export) convertBranchRules(ctx context.Context, from []*branchRule, repo string) []*types.BranchRule {
	var rules []*types.BranchRule
	for _, rule := range from {
		rules = append(rules, e.convertBranchRule(ctx, rule, repo)...)
	}
	return rules
}

func (e *Export) convertBranchRule(
	ctx context.Context,
	from *branchRule,
	repo string,
) []*types.BranchRule {
	var rules []*types.BranchRule
	var logs []string
	var warningMsg string

	rule := &types.BranchRule{
		Name:  migrate.DisplayNameToIdentifier(from.Pattern, "rule", ""),
		State: enum.RuleStateActive,
		Pattern: types.Pattern{
			IncludedPatterns: []string{from.Pattern},
		},
	}

	if from.Kind.Force != nil {
		rule.UpdateForceForbidden = true
	}

	if from.Kind.Delete != nil {
		rule.DeleteForbidden = true
	}

	if from.Kind.RequireTasksToBeCompleted != nil {
		rule.UpdateForbidden = true
		rule.RequireResolveAll = true
	}

	if from.Kind.RequireApprovalsToMerge != nil {
		rule.UpdateForbidden = true
		rule.RequireMinimumCount = *from.Kind.RequireApprovalsToMerge.Value
	}

	if from.Kind.ResetPullRequestApprovalsOnChange != nil {
		rule.UpdateForbidden = true
		rule.RequireLatestCommit = true
	}

	if from.Kind.RequireNoChangesRequested != nil {
		rule.RequireNoChangeRequest = true
	}

	if from.Kind.RequireDefaultReviewerApprovals != nil {
		rule.RequireCodeOwners = true
	}

	if from.Kind.RequireCommitsBehind != nil {
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "Maximum number of commits behind", from.Pattern, repo)
		logs = append(logs, warningMsg)
	}

	if from.Kind.ResetPRChangesRequestedOnChange != nil {
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "Reset requested changes when the source branch is modified", from.Pattern, repo)
		logs = append(logs, warningMsg)
	}

	if from.Kind.EnforceMergeChecks != nil {
		warningMsg = fmt.Sprintf("[%s] Skipped adding enforced status checks. Please create the status checks' pipelines in branch rule %q for repo %q and reconfigure the branch rule.",
			enum.LogLevelWarning, from.Pattern, repo)
		logs = append(logs, warningMsg)
	}

	if from.Kind.Push != nil {
		rule.UpdateForbidden = true
		for _, usr := range from.Kind.Push.Users {
			email, err := e.FindEmailByUUID(ctx, usr.UUID)
			if err != nil {
				e.tracer.LogError("failed to get user email with UUID %d for branch rule bypass list: %w", usr.UUID, err)
				continue
			}
			rule.Bypass.UserEmails = append(rule.Bypass.UserEmails, email)

			if len(from.Kind.Push.Groups) != 0 {
				warningMsg = fmt.Sprintf("[%s] Skipped adding group IDs to bypass list for branch %q rule"+
					" of repository %q as we do not support it as of now.", enum.LogLevelWarning, from.Pattern, repo)
				logs = append(logs, warningMsg)
			}
		}
	}
	rules = append(rules, rule)

	if from.Kind.RestrictMerges != nil {
		r := &types.BranchRule{
			Name:  migrate.DisplayNameToIdentifier(from.Pattern, "mergerule", ""),
			State: enum.RuleStateActive,
			Pattern: types.Pattern{
				IncludedPatterns: []string{from.Pattern},
			},
		}
		r.UpdateForbidden = true
		r.PullReq.Merge.Block = true
		r.UpdateForceForbidden = true

		for _, usr := range from.Kind.Push.Users {
			email, err := e.FindEmailByUUID(ctx, usr.UUID)
			if err != nil {
				e.tracer.LogError("failed to get user email with UUID %d for branch rule bypass list: %w", usr.UUID, err)
				continue
			}
			r.Bypass.UserEmails = append(r.Bypass.UserEmails, email)
		}
		if len(from.Kind.Push.Groups) != 0 {
			warningMsg = fmt.Sprintf("[%s] Skipped adding group IDs to bypass list for branch %q rule"+
				" of repository %q as we do not support it as of now.", enum.LogLevelWarning, from.Pattern, repo)
			logs = append(logs, warningMsg)
		}
		rules = append(rules, r)
	}

	for _, l := range logs {
		if err := e.fileLogger.Log(l); err != nil {
			log.Default().Printf("failed to log the not supported branch rules for project %q: %v", repo, err)
			return rules
		}
	}

	e.report[repo].ReportErrors(gitexporter.ReportTypeBranchRules, repo, logs)
	return rules
}
