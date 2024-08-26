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

// Package types provides common types for data export and import.

package types

import (
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/types/enum"
)

// Org defines an organization.
type (
	Org struct {
		Name string `json:"name"`

		Projects []*Project `json:"project,omitempty"`
		Secrets  []*Secret  `json:"secrets,omitempty"`
	}

	// Project defines a project.
	Project struct {
		Name   string `json:"name"`
		Desc   string `json:"desc,omitempty"`
		Repo   string `json:"repo,omitempty"`
		Branch string `json:"branch,omitempty"`
		Type   string `json:"type"` // github, gitlab, bitbucket
		Yaml   []byte `json:"yaml"`

		Secrets   []*Secret   `json:"secrets,omitempty"`
		Pipelines []*Pipeline `json:"pipelines,omitempty"`
	}

	// Pipeline defines a pipeline.
	Pipeline struct {
		Name   string `json:"name"`
		Repo   string `json:"repo,omitempty"`
		Branch string `json:"branch,omitempty"`
		Type   string `json:"type"` // github, gitlab, bitbucket
		Yaml   string `json:"yaml"`
	}

	// Secret defines a secret.
	Secret struct {
		Name  string `json:"name"`
		Desc  string `json:"desc,omitempty"`
		Value string `json:"value,omitempty"`
	}

	PullRequestListOptions struct {
		Page   int
		Size   int
		Open   bool
		Closed bool
	}

	WebhookListOptions struct {
		Page int
		Size int
	}

	RepoResponse struct {
		scm.Repository
		RepoSlug string
		IsEmpty  bool
	}

	ListOptions struct {
		URL  string
		Page int
		Size int
	}

	PRResponse struct {
		scm.PullRequest
	}

	WebhookData struct {
		ConvertedHooks []*scm.Hook
	}

	RepoData struct {
		Repository      RepoResponse
		PullRequestData []*PullRequestData
		Webhooks        WebhookData
		BranchRules     []*BranchRule
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
		SourceSHA    string
		MergeBaseSHA string
		Outdated     bool
	}

	PRComment struct {
		scm.Comment
		ParentID    int
		CodeComment *CodeComment
	}

	PullRequestData struct {
		PullRequest PRResponse
		Comments    []*PRComment
	}

	BranchRule struct {
		ID    int
		Name  string
		State enum.RuleState
		Definition
		Pattern
		Created time.Time
		Updated time.Time
	}

	Definition struct {
		Bypass
		PullReq
		Lifecycle
	}

	Pattern struct {
		IncludeDefault   bool
		IncludedPatterns []string
		ExcludedPatterns []string
	}

	Bypass struct {
		UserEmails []string
		RepoOwners bool
	}

	Approvals struct {
		RequireCodeOwners      bool
		RequireMinimumCount    int
		RequireLatestCommit    bool
		RequireNoChangeRequest bool
	}

	Comments struct {
		RequireResolveAll bool
	}

	Merge struct {
		StrategiesAllowed []string
		DeleteBranch      bool
	}

	StatusChecks struct {
		RequireIdentifiers []string
	}

	PullReq struct {
		Approvals
		Comments
		Merge
		StatusChecks
	}

	Lifecycle struct {
		CreateForbidden bool
		DeleteForbidden bool
		UpdateForbidden bool
	}

	User struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}
)
