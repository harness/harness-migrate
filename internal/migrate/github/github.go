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
	"fmt"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
	"github.com/go-git/go-git/v5/config"
)

type (
	// wrapper wraps the Client to provide high level helper functions
	// for making http requests and unmarshalling the response.
	wrapper struct {
		*scm.Client
	}

	Export struct {
		github     *wrapper
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
		github:            &wrapper{client},
		org:               org,
		repository:        repo,
		checkpointManager: checkpointer,
		tracer:            tracer,
	}
}

// FetchPullRequestRefs implements gitexporter.Interface.
func (e *Export) PullRequestRefs() []config.RefSpec {
	// Mock implementation
	fmt.Printf("Fetching pull request refs\n")
	return nil
}

// ListBranchRules implements gitexporter.Interface.
func (e *Export) ListBranchRules(ctx context.Context, repoSlug string, opts types.ListOptions) ([]*types.BranchRule, error) {
	// Mock implementation
	fmt.Printf("Listing branch rules for repo: %s\n", repoSlug)
	rules := []*types.BranchRule{}
	return rules, nil
}

// ListPullRequestComments implements gitexporter.Interface.
func (e *Export) ListPullRequestComments(ctx context.Context, repoSlug string, prNumber int, opts types.ListOptions) ([]*types.PRComment, error) {
	// Mock implementation
	fmt.Printf("Listing comments for pull request #%d in repo: %s\n", prNumber, repoSlug)
	comments := []*types.PRComment{}
	return comments, nil
}

// ListWebhooks implements gitexporter.Interface.
func (e *Export) ListWebhooks(ctx context.Context, repoSlug string, logger gitexporter.Logger, opts types.WebhookListOptions) (types.WebhookData, error) {
	// Mock implementation
	fmt.Printf("Listing webhooks for repo: %s\n", repoSlug)
	webhooks := types.WebhookData{}
	return webhooks, nil
}

// PullRequestReviewers implements gitexporter.Interface.
func (e *Export) PullRequestReviewers(ctx context.Context, prNumber int) error {
	// Mock implementation
	fmt.Printf("Fetching reviewers for pull request #%d\n", prNumber)
	return nil
}
