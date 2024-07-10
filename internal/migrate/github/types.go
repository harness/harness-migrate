package github

import "time"

// Error represents a Github error.
type Error struct {
	Message string `json:"message"`
}

type user struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	AvatarURL string `json:"avatar_url"`
	Type      string `json:"type"`
	SiteAdmin bool   `json:"site_admin"`
}

type codeComment struct {
	URL               string    `json:"url"`
	ID                int       `json:"id"`
	DiffHunk          string    `json:"diff_hunk"`
	Path              string    `json:"path"`
	CommitID          string    `json:"commit_id"`
	OriginalCommitID  string    `json:"original_commit_id"`
	User              user      `json:"user"`
	Body              string    `json:"body"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	AuthorAssociation string    `json:"author_association"`
	StartLine         *int      `json:"start_line"`
	OriginalStartLine *int      `json:"original_start_line"`
	StartSide         *string   `json:"start_side"`
	Line              *int      `json:"line"`
	OriginalLine      *int      `json:"original_line"`
	Side              string    `json:"side"`
	InReplyToID       int       `json:"in_reply_to_id"`
	OriginalPosition  int       `json:"original_position"`
	Position          *int      `json:"position"`
	SubjectType       string    `json:"subject_type"`
}

type HunkHeader struct {
	OldLine int
	OldSpan int
	NewLine int
	NewSpan int
	Text    string
}
