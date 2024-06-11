package types

import "github.com/drone/go-scm/scm"

const (
	InfoFileName = "info.json"
	PRFolderName = "pr"
)

type Repository struct {
	scm.Repository
	Slug string
}

type PR struct {
	scm.PullRequest
}

type Comments struct {
}

type RepositoryData struct {
	Repository      Repository
	PullRequestData []*PullRequestData
}

type PullRequestData struct {
	PullRequest PR
	Comments    []Comments
}
