package types

import "github.com/drone/go-scm/scm"

const (
	InfoFileName         = "info.json"
	ExporterLogsFileName = "ExporterLogs.log"
	PRDir                = "pr"
	GitDir               = "git"
	WebhookDir           = "webhooks"
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
		ID               int
		Name             string
		Type             string
		IncludeDefault   bool
		Branches         []string
		IncludedPatterns []string
		ExcludedPatterns []string
		BypassUsers      []string
		BypassGroups     []string
		BypassKeys       []string
	}

	RepositoryData struct {
		Repository      Repository
		PullRequestData []*PullRequestData
		BranchRules     []BranchRule
		Webhooks        WebhookData
	}

	PullRequestData struct {
		PullRequest PR
		Comments    []Comment
	}

	WebhookData struct {
		ConvertedHooks []*scm.Hook
	}
)
