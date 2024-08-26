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

package github

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/migrate"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/types/enum"
)

func (e *Export) convertBranchRulesList(from *branchProtectionRulesResponse, repo string) []*types.BranchRule {
	var to []*types.BranchRule
	for _, edge := range from.Data.Repository.BranchProtectionRules.Edges {
		to = append(to, e.convertBranchRule(edge.Node, repo)...)
	}
	return to
}

func (e *Export) convertBranchRuleSetsList(from []*ruleSet) []*types.BranchRule {
	var to []*types.BranchRule
	for _, rule := range from {
		if rule.Target == "branch" {
			to = append(to, &types.BranchRule{ID: rule.ID})
		}
	}
	return to
}

func (e *Export) convertBranchRule(from branchProtectionRule, repo string) []*types.BranchRule {
	var logs []string
	var warningMsg string
	var rules []*types.BranchRule
	// randomize is set as true as rulesets might have same pattern

	rule := &types.BranchRule{
		ID:    from.DatabaseID,
		Name:  migrate.DisplayNameToIdentifier(from.Pattern, "rule", from.ID),
		State: enum.RuleStateActive,
		Pattern: types.Pattern{
			IncludedPatterns: []string{from.Pattern},
		},
	}

	if !from.AllowsDeletions {
		rule.DeleteForbidden = true
	}
	if from.AllowsForcePushes {
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "force push allowances", from.Pattern, repo)
		logs = append(logs, warningMsg)
	}
	if from.BlocksCreations {
		rule.Lifecycle.CreateForbidden = true
	}
	if from.BypassPullRequestAllowances.TotalCount > 0 {
		actorNotUser := false
		for _, actor := range from.BypassPullRequestAllowances.Edges {
			if actor.Node.Actor.Email != "" {
				rule.Bypass.UserEmails = append(rule.Bypass.UserEmails, actor.Node.Actor.Email)
			} else {
				actorNotUser = true
			}
		}
		if actorNotUser {
			warningMsg = fmt.Sprintf("[%s] Skipped adding teams and apps to bypass list for pattern %q of repo %q as we do not support it as of now.",
				enum.LogLevelWarning, from.Pattern, repo)
			logs = append(logs, warningMsg)
		}
	}
	if from.DismissesStaleReviews {
		rule.UpdateForbidden = true
		rule.RequireLatestCommit = true
	}
	if !from.IsAdminEnforced {
		rule.Definition.Bypass.RepoOwners = true
	}
	if from.LockBranch {
		rule.Lifecycle.UpdateForbidden = true
		rule.Lifecycle.DeleteForbidden = true
	}
	if from.RequireLastPushApproval {
		rule.UpdateForbidden = true
		rule.RequireLatestCommit = true
	}
	if from.RequiresApprovingReviews {
		rule.UpdateForbidden = true
		rule.RequireMinimumCount = from.RequiredApprovingReviewCount
	}
	if from.RequiresCodeOwnerReviews {
		rule.UpdateForbidden = true
		rule.RequireCodeOwners = true
	}
	if from.RequiresCommitSignatures {
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "required commit signatures", from.Pattern, repo)
		logs = append(logs, warningMsg)
	}
	if from.RequiresConversationResolution {
		rule.UpdateForbidden = true
		rule.RequireResolveAll = true
	}
	if from.RequiresDeployments {
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "required deployments", from.Pattern, repo)
		logs = append(logs, warningMsg)
	}
	if from.RequiresLinearHistory {
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "required linear history", from.Pattern, repo)
		logs = append(logs, warningMsg)
	}
	if from.RequiresStatusChecks {
		warningMsg = fmt.Sprintf("[%s] Skipped adding status checks. Create the status checks' pipelines as in branch rule %q for repo %q and reconfigure the branch rule.",
			enum.LogLevelWarning, from.Pattern, repo)
		logs = append(logs, warningMsg)
	}
	if from.RestrictsPushes {
		r := &types.BranchRule{
			ID:    from.DatabaseID,
			Name:  migrate.DisplayNameToIdentifier(from.Pattern, "rule", from.ID),
			State: enum.RuleStateActive,
			Pattern: types.Pattern{
				IncludedPatterns: []string{from.Pattern},
			},
		}
		r.UpdateForbidden = true
		if !from.IsAdminEnforced {
			r.Definition.Bypass.RepoOwners = true
		}
		actorNotUser := false
		for _, actor := range from.PushAllowances.Edges {
			if actor.Node.Actor.Email != "" {
				rule.Bypass.UserEmails = append(rule.Bypass.UserEmails, actor.Node.Actor.Email)
			} else {
				actorNotUser = true
			}
		}
		if actorNotUser {
			warningMsg = fmt.Sprintf("[%s] Skipped adding teams and apps to bypass list for branch rule with"+
				" pattern %q of repo %q as we do not support it as of now.", enum.LogLevelWarning, from.Pattern, repo)
			logs = append(logs, warningMsg)
		}
		rules = append(rules, r)
	}
	if from.RestrictsReviewDismissals {
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "restricts review dismissals", from.Pattern, repo)
		logs = append(logs, warningMsg)
	}

	rules = append(rules, rule)
	for _, l := range logs {
		if err := e.fileLogger.Log(l); err != nil {
			log.Default().Printf("failed to log the not supported branch rules for repo %q: %v", repo, err)
			return rules
		}
	}
	e.report[repo].ReportErrors(gitexporter.ReportTypeBranchRules, from.ID, logs)

	return rules
}

func (e *Export) convertBranchRuleset(from *detailedRuleSet, repo string) *types.BranchRule {
	includedPatterns, includeDefault := mapPatterns(from.Conditions.RefName.Include)
	excludedPatterns, _ := mapPatterns(from.Conditions.RefName.Exclude)
	return &types.BranchRule{
		ID:         from.ID,
		Name:       migrate.DisplayNameToIdentifier(from.Name, "ruleset", strconv.Itoa(from.ID)),
		State:      mapRuleEnforcement(from.Enforcement),
		Definition: e.mapRuleDefinition(from, repo),
		Pattern: types.Pattern{
			IncludeDefault:   includeDefault,
			IncludedPatterns: includedPatterns,
			ExcludedPatterns: excludedPatterns,
		},
		Created: from.CreatedAt,
		Updated: from.UpdatedAt,
	}
}

func mapRuleEnforcement(enforcement string) enum.RuleState {
	switch enforcement {
	case "active":
		return enum.RuleStateActive
	case "evaluate":
		return enum.RuleStateMonitor
	case "disabled":
		return enum.RuleStateDisabled
	default:
		return enum.RuleStateDisabled
	}
}

func mapPatterns(branches []string) ([]string, bool) {
	includeDefault := false
	var res []string
	for _, b := range branches {
		switch b {
		case "~DEFAULT_BRANCH":
			includeDefault = true
		// This will include all the branches in the repo
		case "~ALL":
			res = append(res, "**")
		default:
			res = append(res, strings.TrimPrefix(b, "refs/heads/"))
		}
	}
	return res, includeDefault
}

func (e *Export) mapRuleDefinition(from *detailedRuleSet, repo string) types.Definition {
	definition := types.Definition{}
	var logs []string
	var warningMsg string

	for _, r := range from.Rules {
		switch r.Type {
		case "creation":
			definition.CreateForbidden = true
		case "update":
			definition.UpdateForbidden = true
		case "deletion":
			definition.DeleteForbidden = true
		case "required_linear_history":
			definition.UpdateForbidden = true
		case "pull_request":
			definition.UpdateForbidden = true
			parameters := extractPullRequestParameters(r.Parameters)
			if parameters.RequiredApprovingReviewCount > 0 {
				definition.RequireMinimumCount = parameters.RequiredApprovingReviewCount
			}
			definition.RequireCodeOwners = parameters.RequireCodeOwnerReview
			definition.RequireLatestCommit = parameters.RequireLastPushApproval
			definition.RequireResolveAll = parameters.RequiredReviewThreadResolution
			definition.RequireLatestCommit = parameters.DismissStaleReviewsOnPush
		case "required_status_checks":
			warningMsg = fmt.Sprintf("[%s] Skipped adding status checks. Create the status checks' "+
				"pipelines as in branch rule %q for repo %q and reconfigure the branch rule.",
				enum.LogLevelWarning, from.Name, repo)
			logs = append(logs, warningMsg)

		case "non_fast_forward":
			definition.UpdateForbidden = true
		default:
			warningMsg = fmt.Sprintf("[%s] Skipped mapping rule type %q for branch rule %q of repo %q as we "+
				"do not support it as of now.", enum.LogLevelWarning, r.Type, from.Name, repo)
			logs = append(logs, warningMsg)
		}
	}
	if len(from.BypassActors) != 0 {
		warningMsg = fmt.Sprintf("[%s] Couldn't map bypass list for branch rule %q of repo %q. Need to reconfigure "+
			"the bypass list for the branch rule.", enum.LogLevelWarning, from.Name, repo)
		logs = append(logs, warningMsg)
	}

	for _, l := range logs {
		if err := e.fileLogger.Log(l); err != nil {
			log.Default().Printf("failed to log the bypass actors for repo %q: %v", repo, err)
			return definition
		}
	}
	e.report[repo].ReportErrors(gitexporter.ReportTypeBranchRules, from.Name, logs)

	return definition
}

func extractPullRequestParameters(params map[string]interface{}) pullRequestParameters {
	jsonData, err := json.Marshal(params)
	if err != nil {
		log.Default().Printf("failed to marshal branch rule pull request parameters: %v", err)
		return pullRequestParameters{}
	}

	var parameters pullRequestParameters
	if err := json.Unmarshal(jsonData, &parameters); err != nil {
		log.Default().Printf("failed to unmarshal branch rule pull request parameters: %v", err)
		return pullRequestParameters{}
	}

	return parameters
}

func encodeListOptions(opts types.ListOptions) string {
	params := url.Values{}
	limit := common.DefaultLimit
	if opts.Size != 0 {
		limit = opts.Size
	}
	if opts.Page != 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}
	params.Set("per_page", strconv.Itoa(limit))
	return params.Encode()
}
