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

// Package github provides automatic migration tools from Github to Harness.
package github

import (
	"context"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
	git "github.com/go-git/go-git/v5"
)

type (
	// wrapper wraps the Client to provide high level helper functions
	// for making http requests and unmarshalling the response.
	wrapper struct {
		*scm.Client
	}

	Export struct {
		stash      *wrapper
		org        string
		repository string

		checkpointManager *checkpoint.CheckpointManager

		tracer tracer.Tracer
	}
)

func New(
	client *scm.Client,
	org string,
	repo string,
	checkpointer *checkpoint.CheckpointManager,
	tracer tracer.Tracer,
) *Export {
	return &Export{
		stash:             &wrapper{client},
		org:               org,
		repository:        repo,
		checkpointManager: checkpointer,
		tracer:            tracer,
	}
}

// FetchPullRequestRefs implements gitexporter.Interface.
func (e *Export) FetchPullRequestRefs(ctx context.Context, repo *git.Repository, repoSlug string, scmLogin string, scmToken string) error {
	panic("unimplemented")
}

// ListBranchRules implements gitexporter.Interface.
func (e *Export) ListBranchRules(ctx context.Context, repoSlug string, opts types.ListOptions) ([]*types.BranchRule, error) {
	panic("unimplemented")
}

// ListPullRequestComments implements gitexporter.Interface.
func (e *Export) ListPullRequestComments(ctx context.Context, repoSlug string, prNumber int, opts types.ListOptions) ([]*types.PRComment, error) {
	panic("unimplemented")
}

// ListPullRequests implements gitexporter.Interface.
func (e *Export) ListPullRequests(ctx context.Context, repoSlug string, opts types.PullRequestListOptions) ([]types.PRResponse, error) {
	panic("unimplemented")
}

// ListRepositories implements gitexporter.Interface.
func (e *Export) ListRepositories(ctx context.Context, opts types.ListOptions) ([]types.RepoResponse, error) {
	panic("unimplemented")
}

// ListWebhooks implements gitexporter.Interface.
func (e *Export) ListWebhooks(ctx context.Context, repoSlug string, logger gitexporter.Logger, opts types.WebhookListOptions) (types.WebhookData, error) {
	panic("unimplemented")
}

// PullRequestReviewers implements gitexporter.Interface.
func (e *Export) PullRequestReviewers(ctx context.Context, prNumber int) error {
	panic("unimplemented")
}
