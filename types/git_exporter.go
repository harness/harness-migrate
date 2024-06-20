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

	CommentMetadata struct {
		ParentID     int
		Path         string
		Line         int
		LineSpan     int
		SourceSha    string
		MergeBaseSha string
	}

	Comment struct {
		scm.Comment
		CommentMetadata
	}

	BranchRule struct {
		ID               int      `json:"id"`
		Name             string   `json:"name"`
		Type             string   `json:"type"`
		IncludeDefault   bool     `json:"includeDefault"`
		Branches         []string `json:"branches"`
		IncludedPatterns []string `json:"included_patterns"`
		ExcludedPatterns []string `json:"excluded_patterns"`
		BypassUsers      []string `json:"bypass_users"`
		BypassGroups     []string `json:"bypass_groups"`
		BypassKeys       []string `json:"bypass_keys"`
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
)
