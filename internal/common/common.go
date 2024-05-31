package common

import (
	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/types"
)

func MapPullRequest(prs []*scm.PullRequest) []types.PRResponse {
	r := make([]types.PRResponse, len(prs))
	for i, pr := range prs {
		r[i] = types.PRResponse{PullRequest: *pr}
	}
	return r
}
