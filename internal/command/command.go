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

type Credentials struct {
	Username string
	Password string
}

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

func CheckGitDependencies() error {
	if err := checkGitInstallation(); err != nil {
		return err
	}

	if err := checkGitLFSInstallation(); err != nil {
		return err
	}

	return nil
}

// checkGitInstallation verifies if git is installed and meets minimum version requirements
func checkGitInstallation() error {
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

// checkGitLFSInstallation verifies if git-lfs is installed and meets minimum version requirements
func checkGitLFSInstallation() error {
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

func RunGitCommandWithAuth(ctx context.Context, dir string, auth Credentials, args ...string) ([]byte, error) {
	gitEnv := append(os.Environ(), []string{
		"GIT_TERMINAL_PROMPT=0",
		"GIT_CONFIG_NOSYSTEM=1",       // ignore system config
		"GIT_CONFIG_GLOBAL=/dev/null", // ignore global config
		fmt.Sprintf("GIT_USER=%s", auth.Username),
		fmt.Sprintf("GIT_PASS=%s", auth.Password),
	}...)

	credHelper := "credential.helper=!f() { echo username=$GIT_USER; echo password=$GIT_PASS; }; f"
	args = append([]string{"-c", credHelper}, args...)

	defer func() {
		cmd := exec.CommandContext(ctx, "git", "config", "--local", "--unset-all", "credential.helper")
		cmd.Dir = dir
		cmd.Run()
	}()

	return RunGitCommand(ctx, dir, gitEnv, args...)
}

func RunGitCommand(ctx context.Context, dir string, env []string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	return cmd.CombinedOutput()
}

// RunGitLFSCommand executes a git-lfs command with credentials and returns its output and error
func RunGitLFSCommand(ctx context.Context, dir string, args ...string) ([]byte, error) {
	lfsArgs := append([]string{"lfs"}, args...)
	return RunGitCommand(ctx, dir, []string{}, lfsArgs...)
}

// HasLFSObjects checks if the repository has any Git LFS objects and returns the count
func HasLFSObjects(ctx context.Context, dir string) (int, error) {
	output, err := RunGitLFSCommand(ctx, dir, "ls-files")
	if err != nil {
		return 0, err
	}

	return bytes.Count(output, []byte{'\n'}), nil
}

func FetchLFSObjects(ctx context.Context, dir string) error {
	out, err := RunGitLFSCommand(ctx, dir, "fetch", "--all")
	if err != nil {
		return fmt.Errorf("failed to fetch LFS objects for repo %s, output: %s, err: %w", dir, out, err)
	}

	out, err = RunGitLFSCommand(ctx, dir, "pull")
	if err != nil {
		return fmt.Errorf("failed to checkout LFS objects for repo %s, output: %s, err: %w", dir, out, err)
	}

	return nil
}
