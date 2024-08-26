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

package stash

import (
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

func (e *Export) convertBranchRulesList(
	from []*branchPermission,
	m map[string]modelValue,
	repoSlug string,
) []*types.BranchRule {
	var rules []*types.BranchRule
	for _, p := range from {
		rules = append(rules, e.convertBranchRule(p, m, repoSlug))
	}
	return rules
}

func (e *Export) convertBranchRule(
	from *branchPermission,
	m map[string]modelValue,
	repoSlug string,
) *types.BranchRule {
	includeDefault := false
	var includedPatterns []string

	switch from.Matcher.Type.ID {
	case matcherTypeBranch:
		// displayID will give just branch name main and ID will give refs/heads/main
		includedPatterns = append(includedPatterns, from.Matcher.DisplayID)
	case matcherTypePattern:
		includedPatterns = append(includedPatterns, convertIntoGlobstar(from.Matcher.ID))
	case matcherTypeModelBranch:
		v := m[from.Matcher.ID]
		if v.UseDefault {
			includeDefault = true
		} else {
			includedPatterns = append(includedPatterns, strings.TrimPrefix(v.RefID, "refs/heads/"))
		}
	case matcherTypeModelCategory:
		includedPatterns = append(includedPatterns, convertIntoGlobstar(m[from.Matcher.ID].Prefix))
	}

	var warningMsg string
	var logs []string
	var keys []string
	for _, key := range from.AccessKeys {
		keys = append(keys, key.Key.Label)
	}
	if len(from.Groups) != 0 {
		warningMsg = fmt.Sprintf("[%s] Skipped adding user group(s) [%q] to %q branch rule's bypass list of repo %q",
			enum.LogLevelWarning, strings.Join(from.Groups, ", "), from.Matcher.DisplayID, repoSlug)
		logs = append(logs, warningMsg)
	}
	if len(keys) != 0 {
		warningMsg = fmt.Sprintf("[%s] Skipped adding access key(s) [%q] to %q branch rule's bypass list of repo %q",
			enum.LogLevelWarning, strings.Join(keys, ", "), from.Matcher.DisplayID, repoSlug)
		logs = append(logs, warningMsg)
	}

	for _, l := range logs {
		if err := e.fileLogger.Log(l); err != nil {
			log.Default().Printf("failed to log the exemptions from bypass list of branch rules for repo %q: %v",
				repoSlug, err)
		}
	}
	e.report[repoSlug].ReportErrors(gitexporter.ReportTypeBranchRules, strconv.Itoa(from.ID), logs)

	return &types.BranchRule{
		ID:         from.ID,
		Name:       migrate.DisplayNameToIdentifier(from.Matcher.DisplayID, "rule", strconv.Itoa(from.ID)),
		State:      enum.RuleStateActive,
		Definition: mapRuleDefinition(from.Type, from.Users),
		Pattern: types.Pattern{
			IncludeDefault:   includeDefault,
			IncludedPatterns: includedPatterns,
		},
	}
}

func convertBranchModelsMap(from branchModels) map[string]modelValue {
	m := map[string]modelValue{}
	m[branchDevelopment] = modelValue{modelBranch: from.Development}
	m[branchProduction] = modelValue{modelBranch: from.Production}
	for _, c := range from.Categories {
		m[c.ID] = modelValue{Prefix: c.Prefix}
	}
	return m
}

func mapRuleDefinition(t string, bypassUsers []author) types.Definition {
	var emails []string
	for _, u := range bypassUsers {
		emails = append(emails, u.EmailAddress)
	}

	lifecycle := types.Lifecycle{}
	switch t {
	case "read-only":
		lifecycle = types.Lifecycle{
			CreateForbidden: true,
			UpdateForbidden: true,
			DeleteForbidden: true,
		}
	case "no-deletes":
		lifecycle.DeleteForbidden = true
	case "pull-request-only", "fast-forward-only":
		lifecycle.UpdateForbidden = true
	}

	return types.Definition{
		Lifecycle: lifecycle,
		Bypass: types.Bypass{
			UserEmails: emails,
		},
	}
}

func convertIntoGlobstar(s string) string {
	if strings.HasSuffix(s, "/") {
		return s + "**"
	}
	return s
}

func (e *Error) Error() string {
	if len(e.Errors) == 0 {
		if len(e.Message) > 0 {
			return fmt.Sprintf("bitbucket: status: %d message: %s", e.Status, e.Message)
		}
		return "bitbucket: undefined error"
	}
	return e.Errors[0].Message
}

func encodeListOptions(opts types.ListOptions) string {
	params := url.Values{}
	limit := common.DefaultLimit
	if opts.Size != 0 {
		limit = opts.Size
	}
	params.Set("limit", strconv.Itoa(limit))

	if opts.Page > 0 {
		params.Set("start", strconv.Itoa(
			(opts.Page-1)*limit),
		)
	}
	return params.Encode()
}
