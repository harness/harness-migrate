package stash

import (
	"context"

	"github.com/harness/harness-migrate/internal/codeerror"
)

func (e *Export) PullRequestReviewers(
	context.Context,
	int) error {
	return &codeerror.OpNotSupportedError{Name: "pullreqreview"}
}
