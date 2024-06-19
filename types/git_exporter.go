package types

import "github.com/drone/go-scm/scm"

const (
	InfoFileName         = "info.json"
	ExporterLogsFileName = "ExporterLogs.log"
	PRDir                = "pr"
	GitDir               = "git"
	WebhookFileName      = "webhooks.json"
	BranchRulesFileName  = "rules.json"
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
