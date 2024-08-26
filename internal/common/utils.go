package common

import (
	"strconv"
	"strings"

	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

func MapRepository(repos []*scm.Repository) []types.RepoResponse {
	r := make([]types.RepoResponse, len(repos))
	for i, repo := range repos {
		repoSlug := make([]string, 0)

		if repo.Namespace != "" {
			repoSlug = append(repoSlug, repo.Namespace)
		}
		if repo.Name != "" {
			repoSlug = append(repoSlug, repo.Name)
		}

		r[i] = types.RepoResponse{Repository: *repo, RepoSlug: strings.Join(repoSlug, "/")}
	}
	return r
}

func MapPullRequest(prs []*scm.PullRequest) []types.PRResponse {
	r := make([]types.PRResponse, len(prs))
	for i, pr := range prs {
		r[i] = types.PRResponse{PullRequest: *pr}
	}
	return r
}

func MapPRComment(comments []*scm.Comment) []*types.PRComment {
	r := make([]*types.PRComment, len(comments))
	for i, c := range comments {
		r[i] = &types.PRComment{Comment: *c}
	}
	return r
}

func FormatHunkHeader(source, sourceSpan, destination, destinationSpan int, sectionHeading string) string {
	sb := strings.Builder{}
	sb.Grow(20 + len(sectionHeading))

	sb.WriteString("@@ -")
	sb.WriteString(strconv.Itoa(source))
	if sourceSpan != 1 {
		sb.WriteByte(',')
		sb.WriteString(strconv.Itoa(sourceSpan))
	}
	sb.WriteString(" +")
	sb.WriteString(strconv.Itoa(destination))
	if destinationSpan != 1 {
		sb.WriteByte(',')
		sb.WriteString(strconv.Itoa(destinationSpan))
	}
	sb.WriteString(" @@")

	if sectionHeading != "" {
		sb.WriteByte(' ')
		sb.WriteString(sectionHeading)
	}

	return sb.String()
}
