package types

import "github.com/drone/go-scm/scm"

const (
	InfoFileName         = "info.json"
	ExporterLogsFileName = "ExporterLogs.log"
	PRDir                = "pr"
	GitDir               = "git"
	WebhookDir           = "webhooks"
)

type Repository struct {
	scm.Repository
	Slug string
}

type PR struct {
	scm.PullRequest
}

type CommentMetadata struct {
	ParentID     int
	Path         string
	Line         int
	LineSpan     int
	SourceSha    string
	MergeBaseSha string
}

type PRComment struct {
	scm.Comment
	CommentMetadata
}

type RepositoryData struct {
	Repository      Repository
	PullRequestData []*PullRequestData
	Webhooks        WebhookData
}

type PullRequestData struct {
	PullRequest PR
	Comments    []PRComment
}

type WebhookData struct {
	ConvertedHooks []*scm.Hook
}
