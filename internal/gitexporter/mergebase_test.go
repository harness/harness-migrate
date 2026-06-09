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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harness/harness-migrate/internal/command"
	"github.com/harness/harness-migrate/internal/types"
)

func TestCountNonEmptyLines(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"single", "abc\n", 1},
		{"single no newline", "abc", 1},
		{"two", "abc\ndef\n", 2},
		{"blank lines ignored", "\n\nabc\n\n", 1},
		{"trailing spaces", "  \nabc\n", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countNonEmptyLines([]byte(tt.in)); got != tt.want {
				t.Errorf("countNonEmptyLines(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

// TestSingleMergeBase verifies both the native git and go-git checkers agree on
// the same fixture repository across the relevant scenarios.
func TestSingleMergeBase(t *testing.T) {
	if _, err := command.RunGitCommand(context.Background(), t.TempDir(), nil, "version"); err != nil {
		t.Skip("git not available to build the fixture repository")
	}

	dir := t.TempDir()
	g := func(args ...string) string {
		t.Helper()
		out, err := command.RunGitCommand(context.Background(), dir, gitTestEnv(), args...)
		if err != nil {
			t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, out)
		}
		return strings.TrimSpace(string(out))
	}
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	g("init", "-q")
	g("symbolic-ref", "HEAD", "refs/heads/main")

	// A: common root on main.
	write("a", "a")
	g("add", "-A")
	g("commit", "-q", "-m", "A")
	g("branch", "topic")

	// topic gets commit Y (different file -> no conflict on later merge).
	g("checkout", "-q", "topic")
	write("y", "y")
	g("add", "-A")
	g("commit", "-q", "-m", "Y")
	shaTopicSingle := g("rev-parse", "topic")

	// main gets commit X.
	g("checkout", "-q", "main")
	write("x", "x")
	g("add", "-A")
	g("commit", "-q", "-m", "X")
	shaX := g("rev-parse", "HEAD")

	// Orphan branch: shares no history with main.
	g("checkout", "-q", "--orphan", "orphan")
	write("o", "o")
	g("add", "-A")
	g("commit", "-q", "-m", "orphan root")
	shaOrphan := g("rev-parse", "HEAD")

	t.Run("single", func(t *testing.T) {
		// main (A->X) and topic (A->Y) share exactly A.
		assertBothCheckers(t, dir, shaTopicSingle, "main", true)
	})

	t.Run("unrelated histories", func(t *testing.T) {
		assertBothCheckers(t, dir, shaOrphan, "main", false)
	})

	t.Run("missing target ref", func(t *testing.T) {
		assertBothCheckers(t, dir, shaTopicSingle, "does-not-exist", false)
	})

	// Build the criss-cross: main merges topic, topic merges the pre-merge X.
	g("checkout", "-q", "main")
	g("merge", "-q", "--no-ff", "-m", "merge topic into main", "topic")
	g("checkout", "-q", "topic")
	g("merge", "-q", "--no-ff", "-m", "merge X into topic", shaX)
	shaTopicCross := g("rev-parse", "topic")

	t.Run("multiple merge bases", func(t *testing.T) {
		assertBothCheckers(t, dir, shaTopicCross, "main", false)
	})
}

// assertBothCheckers runs the native and go-git checkers and asserts each
// reports the expected single-merge-base result for the given source/target.
func assertBothCheckers(t *testing.T, dir, sourceSHA, target string, wantSingle bool) {
	t.Helper()
	pr := &types.PRResponse{}
	pr.Sha = sourceSHA
	pr.Target = target

	native := &nativeMergeBaseChecker{gitPath: dir}
	if got := single(native, pr); got != wantSingle {
		t.Errorf("native checker: got single=%v, want %v", got, wantSingle)
	}

	goGit := &goGitMergeBaseChecker{gitPath: dir}
	if got := single(goGit, pr); got != wantSingle {
		t.Errorf("go-git checker: got single=%v, want %v", got, wantSingle)
	}
}

// single collapses (ok, err) into the effective "has a single merge base"
// decision used by the exporter: an error is treated as not-single (close).
func single(c mergeBaseChecker, pr *types.PRResponse) bool {
	ok, err := c.singleMergeBase(context.Background(), pr)
	return ok && err == nil
}

func gitTestEnv() []string {
	return []string{
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
	}
}
