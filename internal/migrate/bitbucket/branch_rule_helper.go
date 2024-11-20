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

func (e *Export) convertBranchRules(ctx context.Context, from rules, repo string) []*types.BranchRule {
	var rules []*types.BranchRule
	for _, rule := range from.Values {
		rules = append(rules, e.convertBranchRule(ctx, rule, repo)...)
	}
	return rules
}

func (e *Export) convertBranchRule(
	ctx context.Context,
	from branchRule,
	repo string,
) []*types.BranchRule {
	var rules []*types.BranchRule
	var logs []string
	var warningMsg string

	convertable := true

	var pattern string
	switch from.BranchMatchKind {
	case "glob":
		pattern = from.Pattern
	case "branching_model":
		pattern = from.BranchType + "/*"
	}

	rule := &types.BranchRule{
		ID:    from.ID,
		Name:  migrate.DisplayNameToIdentifier(pattern),
		State: enum.RuleStateActive,
		Pattern: types.Pattern{
			IncludedPatterns: []string{pattern},
		},
	}

	switch from.Kind {
	case KindPush:
		rule.UpdateForbidden = true

	case KindRestrictMerges:
		rule.UpdateForbidden = true
		rule.PullReq.Merge.Block = true
		rule.UpdateForceForbidden = true

	case KindForce:
		rule.UpdateForceForbidden = true

	case KindDelete:
		rule.DeleteForbidden = true

	case KindRequireTasksToBeCompleted:
		rule.RequireResolveAll = true

	case KindRequireApprovalsToMerge:
		rule.UpdateForbidden = true
		rule.RequireMinimumCount = *from.Value

	case KindResetPullRequestApprovalsOnChange:
		rule.UpdateForbidden = true
		rule.RequireLatestCommit = true

	case KindRequireNoChangesRequested:
		rule.RequireNoChangeRequest = true

	case KindRequireDefaultReviewerApprovals:
		rule.RequireCodeOwners = true

	case KindRequireCommitsBehind:
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "Maximum number of commits behind", from.Pattern, repo)
		logs = append(logs, warningMsg)
		convertable = false

	case KindResetPullRequestChangesRequested:
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "Reset requested changes when the source branch is modified", from.Pattern, repo)
		logs = append(logs, warningMsg)
		convertable = false

	case KindEnforceMergeChecks:
		warningMsg = fmt.Sprintf("[%s] Skipped adding enforced status checks. Please create the status checks' pipelines in branch rule %d for repo %q and reconfigure the branch rule.",
			enum.LogLevelWarning, from.ID, repo)
		logs = append(logs, warningMsg)
		convertable = false

	case KindRequirePassingBuildsToMerge, KindAllowAutoMergeWhenBuildsPass:
		warningMsg = fmt.Sprintf("[%s] Skipped adding restrictions on build checks in branch rule %d for repo %q.",
			enum.LogLevelWarning, from.ID, repo)
		logs = append(logs, warningMsg)
		convertable = false
	}

	for _, usr := range from.Users {
		email, err := e.FindEmailByAccountID(ctx, usr.AccountID)
		if err != nil {
			e.tracer.LogError("failed to get user email with account ID %s for branch rule bypass list: %w", usr.AccountID, err)
			continue
		}
		rule.Bypass.UserEmails = append(rule.Bypass.UserEmails, email)

		if len(from.Groups) != 0 {
			warningMsg = fmt.Sprintf("[%s] Skipped adding group IDs to bypass list for branch %q rule"+
				" of repository %q as we do not support it as of now.", enum.LogLevelWarning, from.Pattern, repo)
			logs = append(logs, warningMsg)
		}
	}

	if convertable {
		rules = append(rules, rule)
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
