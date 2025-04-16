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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/harness/harness-migrate/internal/command"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/util"
	"github.com/harness/harness-migrate/types"

	"github.com/drone/go-scm/scm"
	"github.com/go-git/go-git/v5/config"
)

func (e *Exporter) CloneRepository(
	ctx context.Context,
	repoData scm.Repository,
	repoPath string,
	repoSlug string,
	pullreqRef []config.RefSpec,
	noGitLFS bool,
	tracer tracer.Tracer,
) (bool, error) {
	var isEmpty bool
	tracer.Start(common.MsgStartGitClone, repoSlug)

	if err := command.CheckGitInstallation(); err != nil {
		return isEmpty, err
	}

	if !noGitLFS {
		if err := command.CheckGitLFSInstallation(); err != nil {
			return isEmpty, err
		}
	}

	gitPath := filepath.Join(repoPath, types.GitDir)
	if err := util.CreateFolder(gitPath); err != nil {
		return isEmpty, err
	}

	if _, err := os.Stat(filepath.Join(gitPath, ".git")); err == nil {
		tracer.Log(common.MsgRepoAlreadyExists, repoSlug)
		return isEmpty, nil
	}

	// prepare git authentication
	gitEnv := []string{
		"GIT_TERMINAL_PROMPT=0",
		fmt.Sprintf("GIT_USERNAME=%s", e.ScmLogin),
		fmt.Sprintf("GIT_PASSWORD=%s", e.ScmToken),
	}

	// initialize git-lfs
	if !noGitLFS {
		output, err := command.RunGitLFSCommand(ctx, gitPath, gitEnv, "install", "--force")
		if err != nil {
			tracer.LogError(common.ErrGitClone, repoSlug, err, string(output))
			return isEmpty, fmt.Errorf("failed to initialize git-lfs: %w", err)
		}
	}

	// clone repository with all refs
	cloneArgs := []string{
		"clone",
		"--mirror", // includes all refs
		"--no-checkout",
		repoData.Clone,
		".",
	}

	output, err := command.RunGitCommand(ctx, gitPath, gitEnv, cloneArgs...)
	if err != nil {
		if strings.Contains(string(output), "You appear to have cloned an empty repository.") {
			isEmpty = true
			tracer.Stop(common.MsgGitCloneEmptyRepo, repoSlug)
			return isEmpty, nil
		}

		tracer.LogError(common.ErrGitClone, repoSlug, err, string(output))
		return isEmpty, fmt.Errorf("failed to clone repo %s from %q: %w", repoSlug, repoData.Clone, err)
	}

	// fetch all refs
	output, err = command.RunGitCommand(ctx, gitPath, gitEnv, "fetch", "--all", "--force", "--prune")
	if err != nil {
		tracer.LogError(common.ErrGitFetch, repoSlug, err, string(output))
		return isEmpty, fmt.Errorf("failed to sync repo %s from %q: %w", repoSlug, repoData.Clone, err)
	}

	if !noGitLFS {
		// Check for LFS objects with credentials
		hasLFS, err := command.HasLFSObjects(ctx, gitPath, gitEnv)
		if err != nil {
			tracer.LogError("git-lfs-check", repoSlug, err, string(output))
			return isEmpty, fmt.Errorf("failed to check LFS objects for repo %s: %w", repoSlug, err)
		}

		if hasLFS {
			if err := command.PullLFSObjects(ctx, gitPath, gitEnv); err != nil {
				tracer.LogError("git-lfs-pull", repoSlug, err, string(output))
				return isEmpty, fmt.Errorf("failed to pull LFS objects for repo %s: %w", repoSlug, err)
			}
		}
	}

	tracer.Stop(common.MsgCompleteGitClone, repoSlug)
	return isEmpty, nil
}
