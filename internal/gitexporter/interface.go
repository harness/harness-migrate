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

package gitexporter

import (
	"context"

	"github.com/harness/harness-migrate/internal/types"

	"github.com/go-git/go-git/v5/config"
)

// Interface helps to support a generic way of doing export for all git providers
type Interface interface {
	ListRepositories(ctx context.Context, opts types.ListOptions) ([]types.RepoResponse, error)

	ListPullRequests(ctx context.Context, repoSlug string, opts types.PullRequestListOptions) ([]types.PRResponse, error)

	ListPullRequestComments(ctx context.Context, repoSlug string, prNumber int, opts types.ListOptions) ([]*types.PRComment, error)

	PullRequestReviewers(ctx context.Context, prNumber int) error

	PullRequestRefs() []config.RefSpec

	ListWebhooks(ctx context.Context, repoSlug string, opts types.WebhookListOptions) (types.WebhookData, error)

	ListBranchRules(ctx context.Context, repoSlug string, opts types.ListOptions) ([]*types.BranchRule, error)
}
