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
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"

	"github.com/harness/harness-migrate/internal/command"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/types/enum"
)

// mergeBaseChecker reports whether a pull request's source and target share
// exactly one merge base. It has a native git and a go-git implementation,
// selected the same way as the cloner/pusher (native git unless LFS is
// disabled, which is also when git is unavailable).
type mergeBaseChecker interface {
	singleMergeBase(ctx context.Context, pr *types.PRResponse) (bool, error)
}

func (e *Exporter) selectMergeBaseChecker(gitPath string) mergeBaseChecker {
	if e.flags.NoLFS {
		return &goGitMergeBaseChecker{gitPath: gitPath}
	}
	return &nativeMergeBaseChecker{gitPath: gitPath}
}

// checkMergeBase closes an open PR whose source and target do not share exactly
// one merge base, so the import into Harness CODE is not rejected (CODE refuses
// to create an open PR with multiple/zero merge bases and that halts the whole
// migration), and reports whether it closed the PR.
//
// Fail-safe: zero/multiple merge bases or a PR-specific resolution failure
// (e.g. a fork whose source commit is absent) all result in closing the PR. All
// other PR metadata is preserved; it simply lands as closed in CODE.
func (e *Exporter) checkMergeBase(ctx context.Context, checker mergeBaseChecker, pr *types.PRResponse) bool {
	if pr.Closed || pr.Merged { // CODE only validates open PRs
		return false
	}
	if ok, err := checker.singleMergeBase(ctx, pr); ok && err == nil {
		return false
	}
	pr.Closed = true
	return true
}

func mergeBaseClosedLog(repoSlug string, prNumber int) string {
	return fmt.Sprintf("[%s] Migrating pull request #%d of repo %q as closed: "+
		"source and target have no single merge base (none or multiple - e.g. forked or "+
		"criss-cross history), which Harness CODE rejects for open pull requests.",
		enum.LogLevelWarning, prNumber, repoSlug)
}

// flushMergeBaseClosures writes the PRs that were closed due to an invalid merge
// base to the exporter log file and the repo report. Called once per repo.
func (e *Exporter) flushMergeBaseClosures(repoSlug string, logs []string) {
	if len(logs) == 0 {
		return
	}
	logger := &FileLogger{Location: e.zipLocation}
	for _, l := range logs {
		if err := logger.Log(l); err != nil {
			log.Default().Printf("failed to log closed PRs with invalid merge base for repo %q: %v", repoSlug, err)
			break
		}
	}
	e.Report[repoSlug].ReportErrors(report.ReportTypePRs, repoSlug, logs)
}

// prSourceRevision returns the revision identifying the PR's source tip. The
// head SHA is preferred (stable, present in the clone even for forks via the
// fetched PR-head refs); the PR-head ref is a fallback.
func prSourceRevision(pr *types.PRResponse) string {
	if pr.Sha != "" {
		return pr.Sha
	}
	return fmt.Sprintf("refs/pullreq/%d/head", pr.Number)
}

// nativeMergeBaseChecker uses `git merge-base --all`, a C-level graph walk that
// is fast and memory-light. Each call spawns its own process, so it is safe for
// concurrent use without additional locking.
type nativeMergeBaseChecker struct {
	gitPath string
}

func (c *nativeMergeBaseChecker) singleMergeBase(ctx context.Context, pr *types.PRResponse) (bool, error) {
	source := prSourceRevision(pr)
	target := "refs/heads/" + pr.Target

	out, err := command.RunGitCommand(ctx, c.gitPath, nil, "merge-base", "--all", source, target)
	if err != nil {
		// Exit code 1 means "no common ancestor" - not an error for us, just
		// zero merge bases. Any other exit code is a real, PR-specific failure.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("git merge-base for PR #%d: %w", pr.Number, err)
	}
	return countNonEmptyLines(out) == 1, nil
}

// goGitMergeBaseChecker uses go-git so the check still works on clients without
// a git binary. The repository is opened once and reused across PRs. It is used
// from a single goroutine (the PR result reader / the sequential no-metadata
// loop), so it needs no locking.
type goGitMergeBaseChecker struct {
	gitPath string
	repo    *git.Repository
}

func (c *goGitMergeBaseChecker) singleMergeBase(_ context.Context, pr *types.PRResponse) (bool, error) {
	if c.repo == nil {
		repo, err := git.PlainOpen(c.gitPath)
		if err != nil {
			return false, fmt.Errorf("open git repo with go-git: %w", err)
		}
		c.repo = repo
	}

	source, err := c.commit(prSourceRevision(pr))
	if err != nil {
		return false, fmt.Errorf("resolve source of PR #%d: %w", pr.Number, err)
	}
	target, err := c.commit("refs/heads/" + pr.Target)
	if err != nil {
		return false, fmt.Errorf("resolve target of PR #%d: %w", pr.Number, err)
	}
	bases, err := source.MergeBase(target)
	if err != nil {
		return false, fmt.Errorf("merge-base for PR #%d: %w", pr.Number, err)
	}
	return len(bases) == 1, nil
}

func (c *goGitMergeBaseChecker) commit(rev string) (*object.Commit, error) {
	h, err := c.repo.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return nil, err
	}
	return c.repo.CommitObject(*h)
}

func countNonEmptyLines(b []byte) int {
	n := 0
	for _, line := range bytes.Split(b, []byte{'\n'}) {
		if len(bytes.TrimSpace(line)) > 0 {
			n++
		}
	}
	return n
}
