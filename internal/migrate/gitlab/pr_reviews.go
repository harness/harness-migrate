// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package gitlab provides automatic migration tools from GitLab to Harness.
package gitlab

import (
	"context"

	"github.com/harness/harness-migrate/internal/codeerror"
	"github.com/harness/harness-migrate/internal/types"
)

func (e *Export) ListPullRequestReviews(ctx context.Context, repoSlug string, prNumber int, opts types.ListOptions) ([]*types.PRReview, error) {
	return nil, &codeerror.OpNotSupportedError{Name: "ListPullRequestReviews"}
}

func (e *Export) ListRequestedReviewers(ctx context.Context, repoSlug string, prNumber int) ([]*types.PRReviewer, error) {
	return nil, &codeerror.OpNotSupportedError{Name: "ListRequestedReviewers"}
}
