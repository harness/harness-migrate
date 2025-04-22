package gitimporter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/harness/harness-migrate/internal/command"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/types"
)

type gitPusher interface {
	push(ctx context.Context, repo *harness.Repository, gitPath, remoteName string) error
}

type pushParams struct {
	gitPath        string
	repo           *harness.Repository
	remoteName     string
	lfsObjectCount int64
	tracer         tracer.Tracer
	auth           credentials
}

type nativeGitPusher struct {
	params pushParams
}

type goGitPusher struct {
	params pushParams
}

type credentials struct {
	username string
	token    string
}

func (m *Importer) Push(
	ctx context.Context,
	repoPath string,
	repo *harness.Repository,
	lfsObjectCount int64,
	tracer tracer.Tracer,
) error {
	tracer.Start(common.MsgStartImportGit, repo.GitURL)

	gitPath := filepath.Join(repoPath, types.GitDir)
	const remoteName = "origin"
	gitEnv := []string{
		"GIT_TERMINAL_PROMPT=1",
		fmt.Sprintf("GIT_USERNAME=%s", "git-importer"),
		fmt.Sprintf("GIT_PASSWORD=%s", m.HarnessToken),
	}

	output, err := command.RunGitCommand(ctx, gitPath, gitEnv, "remote", "set-url", remoteName, repo.GitURL)
	if err != nil {
		tracer.LogError("git-remote", repo.GitURL, err, string(output))
		return fmt.Errorf("failed to set remote to %q: %w", repo.GitURL, err)
	}

	if err := m.handleLFSPush(ctx, gitPath, repo, remoteName, lfsObjectCount, tracer, gitEnv); err != nil {
		return err
	}

	params := pushParams{
		gitPath:        gitPath,
		repo:           repo,
		remoteName:     remoteName,
		lfsObjectCount: lfsObjectCount,
		tracer:         tracer,
		auth: credentials{
			username: "git-importer",
			token:    m.HarnessToken,
		},
	}

	pusher := m.selectGitPusher(params)

	if err := pusher.push(ctx, repo, gitPath, remoteName); err != nil {
		tracer.LogError("git-push", repo.GitURL, err)
		return fmt.Errorf("failed to push git repository: %w", err)
	}

	tracer.Stop(common.MsgCompleteImportGit, repo.GitURL)
	return nil
}

func (m *Importer) selectGitPusher(params pushParams) gitPusher {
	if err := command.CheckGitInstallation(); err == nil {
		return &nativeGitPusher{params: params}
	}
	return &goGitPusher{params: params}
}

func (m *Importer) handleLFSPush(
	ctx context.Context,
	gitPath string,
	repo *harness.Repository,
	remoteName string,
	lfsObjectCount int64,
	tracer tracer.Tracer,
	gitEnv []string,
) error {
	if m.flags.NoLFS || lfsObjectCount == 0 {
		tracer.LogError("git-lfs-push leaving due to 0 LFS objects", repo.GitURL)
		return nil
	}

	if err := command.CheckGitLFSInstallation(); err != nil {
		tracer.LogError("git-lfs-push leaving due to missing LFS installation", repo.GitURL)
		return err
	}

	setupCmds := [][]string{
		{"config", "--local", "lfs.url", repo.GitURL + "/info/lfs"},
		{"config", "--local", "lfs." + repo.GitURL + "/info/lfs.access", "basic"},
		{"config", "--local", "lfs." + repo.GitURL + ".locksverify", "false"},
		{"config", "--local", "lfs." + repo.GitURL + "/info/lfs.username", "git-importer"},
		{"config", "--local", "lfs." + repo.GitURL + "/info/lfs.password", m.HarnessToken},
	}

	for _, cmd := range setupCmds {
		if output, err := command.RunGitCommand(ctx, gitPath, gitEnv, cmd...); err != nil {
			tracer.LogError("git-lfs-config", repo.GitURL, err, string(output))
			return fmt.Errorf("failed to configure LFS for %q: %w", repo.GitURL, err)
		}
	}

	output, err := command.RunGitLFSCommand(ctx, gitPath, gitEnv, "push", "--all", remoteName, repo.DefaultBranch)
	tracer.LogError("git-lfs-push", repo.GitURL, string(output))
	if err != nil {
		tracer.LogError("git-lfs-push", repo.GitURL, err, string(output))
		return fmt.Errorf("failed to push LFS objects to %q: %w", repo.GitURL, err)
	}

	return nil
}

func (p *nativeGitPusher) push(ctx context.Context, repo *harness.Repository, gitPath, remoteName string) error {
	gitEnv := []string{
		"GIT_TERMINAL_PROMPT=0",
		fmt.Sprintf("GIT_USERNAME=%s", p.params.auth.username),
		fmt.Sprintf("GIT_PASSWORD=%s", p.params.auth.token),
	}

	output, err := command.RunGitCommand(ctx, gitPath, gitEnv, "push", "--mirror", remoteName)
	if err != nil {
		p.params.tracer.LogError(common.ErrGitPush, repo.GitURL, err, string(output))
		return fmt.Errorf("failed to push refs to %q: %w", repo.GitURL, err)
	}

	return nil
}

func (p *goGitPusher) push(
	ctx context.Context,
	repo *harness.Repository,
	gitPath,
	remoteName string,
) error {
	gitRepo, err := git.PlainOpen(gitPath)
	if err != nil {
		p.params.tracer.Stop("failed to open git dir from %q", gitPath)
		return fmt.Errorf("failed to open the exported repository from %q: %w", gitPath, err)
	}

	_, err = gitRepo.CreateRemote(&config.RemoteConfig{
		Name: remoteName,
		URLs: []string{repo.GitURL},
	})
	if err != nil && !errors.Is(err, git.ErrRemoteExists) {
		return fmt.Errorf("failed to set remote to %q: %w", repo.GitURL, err)
	}

	var output bytes.Buffer
	err = gitRepo.PushContext(ctx, &git.PushOptions{
		RemoteName: remoteName,
		Auth: &http.BasicAuth{
			Username: p.params.auth.username,
			Password: p.params.auth.token,
		},
		RefSpecs: []config.RefSpec{
			"refs/heads/*:refs/heads/*",
			"refs/tags/*:refs/tags/*",
			"refs/pullreq/*/head:refs/pullreq/*/head",
		},
		RemoteURL:       repo.GitURL,
		Force:           true,
		InsecureSkipTLS: true,
		Progress:        &output,
	})

	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		p.params.tracer.Stop(common.ErrGitPush, repo.GitURL, err, output.String())
		return fmt.Errorf(common.ErrGitPush, repo.GitURL, err, output.String())
	}

	return nil
}
