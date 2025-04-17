package gitimporter

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/harness/harness-migrate/internal/command"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/types"
)

func (m *Importer) Push(
	ctx context.Context,
	repoPath string,
	repo *harness.Repository,
	lfsObjectCount int64,
	tracer tracer.Tracer,
) error {
	tracer.Start(common.MsgStartImportGit, repo.GitURL)
	gitPath := filepath.Join(repoPath, types.GitDir)

	// check git installation
	if err := command.CheckGitInstallation(); err != nil {
		return err
	}

	// check LFS if needed
	if !m.flags.NoLFS || lfsObjectCount > 0 {
		if err := command.CheckGitLFSInstallation(); err != nil {
			return err
		}
	}

	gitEnv := []string{
		"GIT_TERMINAL_PROMPT=0",
		fmt.Sprintf("GIT_USERNAME=%s", "git-importer"),
		fmt.Sprintf("GIT_PASSWORD=%s", m.HarnessToken),
	}

	// set remote for push
	const gitRemoteHarness = "origin"
	output, err := command.RunGitCommand(ctx, gitPath, gitEnv, "remote", "set-url", gitRemoteHarness, repo.GitURL)
	if err != nil {
		tracer.LogError("git-remote", repo.GitURL, err, string(output))
		return fmt.Errorf("failed to set remote to %q: %w", repo.GitURL, err)
	}

	if !m.flags.NoLFS || lfsObjectCount > 0 {
		output, err = command.RunGitLFSCommand(ctx, gitPath, gitEnv, "install", "--force")
		if err != nil {
			tracer.LogError("git-lfs-install", repo.GitURL, err, string(output))
			return fmt.Errorf("failed to initialize git-lfs: %w", err)
		}
	}

	output, err = command.RunGitCommand(ctx, gitPath, gitEnv, "push", "--mirror", gitRemoteHarness)
	if err != nil {
		tracer.LogError(common.ErrGitPush, repo.GitURL, err, string(output))
		return fmt.Errorf("failed to push refs to %q: %w", repo.GitURL, err)
	}

	// push LFS objects if needed
	if !m.flags.NoLFS || lfsObjectCount > 0 {
		hasLFS, err := command.HasLFSObjects(ctx, gitPath, gitEnv)
		if err != nil {
			tracer.LogError("git-lfs-check", repo.GitURL, err, string(output))
			return fmt.Errorf("failed to check LFS objects: %w", err)
		}

		if hasLFS > 0 {
			output, err = command.RunGitLFSCommand(ctx, gitPath, gitEnv, "push", "--all", gitRemoteHarness, repo.DefaultBranch)
			if err != nil {
				tracer.LogError("git-lfs-push", repo.GitURL, err, string(output))
				return fmt.Errorf("failed to push LFS objects to %q: %w", repo.GitURL, err)
			}
		}
	}

	tracer.Stop(common.MsgCompleteImportGit, repo.GitURL)
	return nil
}
