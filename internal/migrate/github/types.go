package github

import (
	"time"

	"github.com/harness/harness-migrate/internal/null"
)

// Error represents a Github error.
type Error struct {
	Message string `json:"message"`
}

type user struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Type      string `json:"type"`
	SiteAdmin bool   `json:"site_admin"`
}

type prComment struct {
	URL               string      `json:"url"`
	ID                int         `json:"id"`
	DiffHunk          string      `json:"diff_hunk"`
	Path              string      `json:"path"`
	CommitID          string      `json:"commit_id"`
	OriginalCommitID  string      `json:"original_commit_id"`
	User              user        `json:"user"`
	Body              string      `json:"body"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	HTMLURL           string      `json:"html_url"`
	AuthorAssociation string      `json:"author_association"`
	StartLine         null.Int    `json:"start_line"`
	OriginalStartLine null.Int    `json:"original_start_line"`
	StartSide         null.String `json:"start_side"`
	Line              int         `json:"line"`
	OriginalLine      int         `json:"original_line"`
	Side              string      `json:"side"`
	InReplyToID       int         `json:"in_reply_to_id"`
	OriginalPosition  int         `json:"original_position"`
	Position          int         `json:"position"`
	SubjectType       string      `json:"subject_type"`
}
