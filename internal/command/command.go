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

package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

const (
	MinVersionGit    = 2.45
	MinVersionGitLFS = 3.5
)

var (
	ErrGitNotInstalled     = errors.New("git is not installed")
	ErrGitLFSNotInstalled  = errors.New("git-lfs is not installed")
	ErrGitVersionTooOld    = fmt.Errorf("git version must be %f or higher", MinVersionGit)
	ErrGitLFSVersionTooOld = fmt.Errorf("git-lfs version must be %f or higher", MinVersionGitLFS)

	// handles both "git version 2.45.1" and "git-lfs/3.6.1"
	versionRegex = regexp.MustCompile(`(?:version\s+|/)((?:\d+\.){1,2}\d+)`)
)

// extractVersion extracts major and minor version numbers
func extractVersion(input string) (major, minor int) {
	matches := versionRegex.FindStringSubmatch(input)
	if matches == nil || len(matches) < 2 {
		return 0, 0
	}

	parts := regexp.MustCompile(`\.`).Split(matches[1], 3)
	if len(parts) < 2 {
		return 0, 0
	}

	major, _ = strconv.Atoi(parts[0])
	minor, _ = strconv.Atoi(parts[1])
	return major, minor
}

func isVersionGreaterOrEqual(major1, minor1, major2, minor2 int) bool {
	if major1 != major2 {
		return major1 > major2
	}
	return minor1 >= minor2
}

// CheckGitInstallation verifies if git is installed and meets minimum version requirements
func CheckGitInstallation() error {
	output, err := exec.Command("git", "version").Output()
	if err != nil {
		return ErrGitNotInstalled
	}

	major, minor := extractVersion(string(output))
	if major == 0 && minor == 0 {
		return ErrGitNotInstalled
	}

	if !isVersionGreaterOrEqual(major, minor, 2, 45) {
		return ErrGitVersionTooOld
	}

	return nil
}

// CheckGitLFSInstallation verifies if git-lfs is installed and meets minimum version requirements
func CheckGitLFSInstallation() error {
	output, err := exec.Command("git-lfs", "version").Output()
	if err != nil {
		// Try alternative command
		output, err = exec.Command("git", "lfs", "version").Output()
		if err != nil {
			return ErrGitLFSNotInstalled
		}
	}

	major, minor := extractVersion(string(output))
	if major == 0 && minor == 0 {
		return ErrGitLFSNotInstalled
	}

	if !isVersionGreaterOrEqual(major, minor, 3, 5) {
		return ErrGitLFSVersionTooOld
	}

	return nil
}

// RunGitCommand executes a git command and returns its output and error
func RunGitCommand(ctx context.Context, dir string, env []string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	return cmd.CombinedOutput()
}

// RunGitLFSCommand executes a git-lfs command with credentials and returns its output and error
func RunGitLFSCommand(ctx context.Context, dir string, env []string, args ...string) ([]byte, error) {
	// Add git-lfs prefix to args
	lfsArgs := append([]string{"lfs"}, args...)
	return RunGitCommand(ctx, dir, env, lfsArgs...)
}

// HasLFSObjects checks if the repository has any Git LFS objects and returns the count
func HasLFSObjects(ctx context.Context, dir string, env []string) (int64, error) {
	output, err := RunGitLFSCommand(ctx, dir, env, "ls-files")
	if err != nil {
		return 0, err
	}

	return int64(bytes.Count(output, []byte{'\n'})), nil
}
