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

import (
	"time"
)

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
		Slug       string     `json:"slug"`
		ID         string     `json:"id"`
		Namespace  string     `json:"namespace"`
		Name       string     `json:"name"`
		Branch     string     `json:"branch"`
		Archived   bool       `json:"archived"`
		Private    bool       `json:"private"`
		Visibility Visibility `json:"visibility"`
		Clone      string     `json:"clone"`
		CloneSSH   string     `json:"clone_ssh"`
		Link       string     `json:"link"`
		Created    time.Time  `json:"created"`
		Updated    time.Time  `json:"updated"`
	}

	Perm struct {
		Pull  bool `json:"pull"`
		Push  bool `json:"push"`
		Admin bool `json:"admin"`
	}

	Visibility int

	PR struct {
		Number  int       `json:"number"`
		Title   string    `json:"title"`
		Body    string    `json:"body"`
		Sha     string    `json:"sha"`
		Ref     string    `json:"ref"`
		Source  string    `json:"source"`
		Target  string    `json:"target"`
		Fork    string    `json:"fork"`
		Link    string    `json:"link"`
		Diff    string    `json:"diff"`
		Closed  bool      `json:"closed"`
		Merged  bool      `json:"merged"`
		Merge   string    `json:"merge"`
		Base    Reference `json:"base"`
		Head    Reference `json:"head"`
		Author  User      `json:"author"`
		Created time.Time `json:"created"`
		Updated time.Time `json:"updated"`
		Labels  []Label   `json:"labels"`
	}

	Label struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	// User represents a user account.
	User struct {
		ID      string    `json:"id"`
		Login   string    `json:"login"`
		Name    string    `json:"name"`
		Email   string    `json:"email"`
		Avatar  string    `json:"avatar"`
		Created time.Time `json:"created"`
		Updated time.Time `json:"updated"`
	}

	// Reference represents a git reference.
	Reference struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Sha  string `json:"sha"`
	}

	Hunk struct {
		Header string   `json:"header"`
		Lines  []string `json:"lines"`
	}

	CodeComment struct {
		Path         string `json:"path"`
		CodeSnippet  Hunk   `json:"code_snippet"`
		Side         string `json:"side"`
		HunkHeader   string `json:"hunk_header"`
		SourceSha    string `json:"source_sha"`
		MergeBaseSha string `json:"merge_base_sha"`
	}

	Comment struct {
		ID          int          `json:"id"`
		Body        string       `json:"body"`
		Author      User         `json:"author"`
		Created     time.Time    `json:"created"`
		Updated     time.Time    `json:"updated"`
		ParentID    int          `json:"parent_id"`
		CodeComment *CodeComment `json:"code_comment"`
	}

	BranchRule struct {
		ID               int      `json:"id"`
		Name             string   `json:"name"`
		Type             string   `json:"type"`
		IncludeDefault   bool     `json:"include_default"`
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
		Hooks []*Hook `json:"hooks"`
	}

	// Hook represents a repository hook.
	Hook struct {
		ID         string   `json:"id"`
		Name       string   `json:"name"`
		Target     string   `json:"target"`
		Events     []string `json:"events"`
		Active     bool     `json:"active"`
		SkipVerify bool     `json:"skip_verify"`
	}
)

// Role values.
const (
	VisibilityUndefined Visibility = iota
	VisibilityPublic
	VisibilityInternal
	VisibilityPrivate
)
