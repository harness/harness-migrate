package stash

import (
	"context"

	"github.com/harness/harness-migrate/internal/types"
)

// ListRequestedReviewers implements gitexporter.Interface.
func (e *Export) ListRequestedReviewers(ctx context.Context, repoSlug string, prNumber int) ([]*types.PRReviewer, error) {
	// Stash does not support requested reviewers concept
	return []*types.PRReviewer{}, nil
}
