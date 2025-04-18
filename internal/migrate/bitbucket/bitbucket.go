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

// Package bitbucket provides automatic migration tools from Bitbucket Cloud to Harness.
package bitbucket

import (
	"context"

	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/codeerror"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/internal/tracer"
)

func New(
	client *scm.Client,
	workspace string,
	repository string,
	checkpointer *checkpoint.CheckpointManager,
	logger *gitexporter.FileLogger,
	tracer tracer.Tracer,
	report map[string]*report.Report,
) *Export {
	ckpt := make(map[string]user)
	c, ok, err := checkpoint.GetCheckpointData[map[string]user](checkpointer, CheckpointKeyUsers)
	if err != nil {
		tracer.LogError("cannot load checkpoint userdata: %v", err)
	}
	if ok {
		ckpt = c
	}
	return &Export{
		bitbucket:         client,
		workspace:         workspace,
		repository:        repository,
		checkpointManager: checkpointer,
		tracer:            tracer,
		userMap:           ckpt,
		fileLogger:        logger,
		report:            report,
	}
}

// PullRequestReviewers implements gitexporter.Interface.
func (e *Export) PullRequestReviewers(ctx context.Context, prNumber int) error {
	return &codeerror.OpNotSupportedError{Name: "pullreqreview"}
}
