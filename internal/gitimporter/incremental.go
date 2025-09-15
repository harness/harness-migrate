// Copyright 2024 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitimporter

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/harness/harness-migrate/internal/command"
	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/tracer"
)

const (
	// Large offset used during reference updates to avoid conflicts
	tempPROffset = 10000000
)

type IncrementalMigrationHandler struct {
	harnessClient harness.Client
	tracer        tracer.Tracer
	repoRef       string
}

func NewIncrementalMigrationHandler(
	harnessClient harness.Client,
	tracer tracer.Tracer,
	repoRef string,
) *IncrementalMigrationHandler {
	return &IncrementalMigrationHandler{
		harnessClient: harnessClient,
		tracer:        tracer,
		repoRef:       repoRef,
	}
}

func (h *IncrementalMigrationHandler) CheckRepositoryExists(ctx context.Context) error {
	_, err := h.harnessClient.FindRepoSettings(h.repoRef)
	if err != nil {
		return fmt.Errorf("repository %s does not exist on target server: %w", h.repoRef, err)
	}

	return nil
}

func (h *IncrementalMigrationHandler) GetPROffset(ctx context.Context) (int, error) {
	metadata, err := h.harnessClient.GetRepository(h.repoRef)
	if err != nil {
		return 0, fmt.Errorf("failed to get repository metadata: %w", err)
	}

	return metadata.PullRequestNumber, nil
}

func (h *IncrementalMigrationHandler) UpdatePRReferences(ctx context.Context, gitDir string, offset int) error {
	refs, err := h.listPRReferences(ctx, gitDir)
	if err != nil {
		return fmt.Errorf("failed to list PR references: %w", err)
	}

	if len(refs) == 0 {
		h.tracer.Log("No PR references found to update")
		return nil
	}

	// move all PR references to temporary high-numbered refs to avoid conflicts.
	if err := h.moveReferences(ctx, gitDir, refs, 0, tempPROffset); err != nil {
		return fmt.Errorf("failed to move references to temporary: %w", err)
	}

	// move all PR references to final refs with offset added.
	if err := h.moveReferences(ctx, gitDir, refs, tempPROffset, offset); err != nil {
		return fmt.Errorf("failed to move references to final: %w", err)
	}

	h.tracer.Log("Successfully updated PR references with offset from %d to %d", offset+1, offset+len(refs))
	return nil
}

func (h *IncrementalMigrationHandler) listPRReferences(
	ctx context.Context,
	gitDir string,
) ([]string, error) {
	output, err := command.RunGitCommand(ctx, gitDir, []string{},
		"for-each-ref", "--format=%(refname)", "refs/pullreq/*/head")
	if err != nil {
		return nil, err
	}

	refs := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(refs) == 1 && refs[0] == "" {
		return []string{}, nil
	}

	return refs, nil
}

func (h *IncrementalMigrationHandler) extractPRNumber(ref string) (int, error) {
	parts := strings.Split(ref, "/")
	if len(parts) != 4 || parts[0] != "refs" || parts[1] != "pullreq" || parts[3] != "head" {
		return 0, fmt.Errorf("invalid PR reference format: %s", ref)
	}

	prNumber, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, fmt.Errorf("invalid PR number in reference %s: %w", ref, err)
	}

	return prNumber, nil
}

func (h *IncrementalMigrationHandler) getCommitSHA(ctx context.Context, gitDir, ref string) (string, error) {
	commitOutput, err := command.RunGitCommand(ctx, gitDir, []string{}, "rev-parse", ref)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(commitOutput)), nil
}

// moveReferences is a helper function to move PR references from source to target
// srcOffset: offset to add to original PR number for source reference
// targetOffset: offset to add to original PR number for target reference
func (h *IncrementalMigrationHandler) moveReferences(
	ctx context.Context,
	gitDir string,
	refs []string,
	srcOffset, targetOffset int,
) error {
	for _, ref := range refs {
		if ref == "" {
			continue
		}

		prNumber, err := h.extractPRNumber(ref)
		if err != nil {
			h.tracer.Log("Skipping invalid reference %s: %v", ref, err)
			continue
		}

		srcRef := fmt.Sprintf("refs/pullreq/%d/head", prNumber+srcOffset)
		targetRef := fmt.Sprintf("refs/pullreq/%d/head", prNumber+targetOffset)

		commit, err := h.getCommitSHA(ctx, gitDir, srcRef)
		if err != nil {
			h.tracer.Log("Failed to get commit for %s: %v", srcRef, err)
			continue
		}

		if err := h.createReference(ctx, gitDir, targetRef, commit); err != nil {
			h.tracer.Log("Failed to create target ref %s: %v", targetRef, err)
			continue
		}

		if err := h.deleteReference(ctx, gitDir, srcRef); err != nil {
			h.tracer.Log("Failed to delete source ref %s: %v", srcRef, err)
			continue
		}
	}

	return nil
}

// extractPRNumberFromRef extracts PR number from a ref string (helper for logging)
func (h *IncrementalMigrationHandler) extractPRNumberFromRef(ref string) int {
	parts := strings.Split(ref, "/")
	if len(parts) >= 3 {
		if prNum, err := strconv.Atoi(parts[2]); err == nil {
			return prNum
		}
	}
	return 0
}

func (h *IncrementalMigrationHandler) createReference(ctx context.Context, gitDir, ref, commit string) error {
	_, err := command.RunGitCommand(ctx, gitDir, []string{}, "update-ref", ref, commit)
	return err
}

func (h *IncrementalMigrationHandler) deleteReference(ctx context.Context, gitDir, ref string) error {
	_, err := command.RunGitCommand(ctx, gitDir, []string{}, "update-ref", "-d", ref)
	return err
}
