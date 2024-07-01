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

package types

import "github.com/drone/go-scm/scm"

const (
	InfoFileName         = "info.json"
	ExporterLogsFileName = "ExporterLogs.log"
	PRDir                = "pr"
	GitDir               = "git"
	WebhookFileName      = "webhooks.json"
	BranchRulesFileName  = "rules.json"
	UsersFileName        = "users.json"
)

type (
	Repository struct {
		scm.Repository
		Slug string
	}

	PR struct {
		scm.PullRequest
	}

	Hunk struct {
		Header string
		Lines  []string
	}

	CodeComment struct {
		Path         string
		CodeSnippet  Hunk
		Side         string
		HunkHeader   string
		SourceSha    string
		MergeBaseSha string
	}

	Comment struct {
		scm.Comment
		ParentID    int
		CodeComment *CodeComment
	}

	BranchRule struct {
		ID               int            `json:"id"`
		Name             string         `json:"name"`
		RuleDefinition   RuleDefinition `json:"definition"`
		IncludeDefault   bool           `json:"includeDefault"`
		IncludedPatterns []string       `json:"included_patterns"`
		ExcludedPatterns []string       `json:"excluded_patterns"`
		BypassGroups     []string       `json:"bypass_groups"`
		BypassKeys       []string       `json:"bypass_keys"`
	}

	RepositoryData struct {
		Repository      Repository         `json:"repository"`
		PullRequestData []*PullRequestData `json:"pull_request_data"`
		BranchRules     []BranchRule       `json:"branch_rules"`
		Webhooks        WebhookData        `json:"webhooks"`
	}

	PullRequestData struct {
		PullRequest PR        `json:"pull_request"`
		Comments    []Comment `json:"comments"`
	}

	WebhookData struct {
		Hooks []*scm.Hook `json:"hooks"`
	}

	Bypass struct {
		UserIDs    []string `json:"user_ids,omitempty"`
		RepoOwners bool     `json:"repo_owners,omitempty"`
	}

	Approvals struct {
		RequireCodeOwners      bool `json:"require_code_owners,omitempty"`
		RequireMinimumCount    int  `json:"require_minimum_count,omitempty"`
		RequireLatestCommit    bool `json:"require_latest_commit,omitempty"`
		RequireNoChangeRequest bool `json:"require_no_change_request,omitempty"`
	}

	Comments struct {
		RequireResolveAll bool `json:"require_resolve_all,omitempty"`
	}

	Merge struct {
		StrategiesAllowed []string `json:"strategies_allowed,omitempty"`
		DeleteBranch      bool     `json:"delete_branch,omitempty"`
	}

	StatusChecks struct {
		RequireUIDs []string `json:"require_uids,omitempty"`
	}

	PullReq struct {
		Approvals    Approvals    `json:"approvals,omitempty"`
		Comments     Comments     `json:"comments,omitempty"`
		Merge        Merge        `json:"merge,omitempty"`
		StatusChecks StatusChecks `json:"status_checks,omitempty"`
	}

	// Lifecycle represents the lifecycle rules for branches.
	Lifecycle struct {
		CreateForbidden bool `json:"create_forbidden,omitempty"`
		DeleteForbidden bool `json:"delete_forbidden,omitempty"`
		UpdateForbidden bool `json:"update_forbidden,omitempty"`
	}

	RuleDefinition struct {
		Bypass    Bypass    `json:"bypass,omitempty"`
		PullReq   PullReq   `json:"pullreq,omitempty"`
		Lifecycle Lifecycle `json:"lifecycle,omitempty"`
	}
)
