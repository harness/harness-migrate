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
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

func New(
	client *scm.Client,
	org string,
	repo string,
	checkpointer *checkpoint.CheckpointManager,
	logger *gitexporter.FileLogger,
	tracer tracer.Tracer,
	report map[string]*report.Report,
) *Export {
	ckpt := make(map[string]types.User)
	c, ok, err := checkpoint.GetCheckpointData[map[string]types.User](checkpointer, CheckpointKeyUsers)
	if err != nil {
		tracer.LogError("cannot load checkpoint userdata: %v", err)
	}
	if ok {
		ckpt = c
	}
	return &Export{
		github:            client,
		org:               org,
		repository:        repo,
		checkpointManager: checkpointer,
		tracer:            tracer,
		userMap:           ckpt,
		fileLogger:        logger,
		report:            report,
	}
}

// PullRequestReviewers implements gitexporter.Interface.
func (e *Export) PullRequestReviewers(ctx context.Context, prNumber int) error {
	// Mock implementation
	fmt.Printf("Fetching reviewers for pull request #%d\n", prNumber)
	return nil
}
