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
	"github.com/harness/harness-migrate/internal/report"
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
	lfsObjectCount int
	token          string
}

type nativeGitPusher struct {
	params pushParams
	tracer tracer.Tracer
}

type goGitPusher struct {
	params pushParams
	tracer tracer.Tracer
}

func (m *Importer) Push(
	ctx context.Context,
	repoRef string,
	repoPath string,
	repo *harness.Repository,
	gitLFSDisabled bool,
	lfsObjectCount int,
	tracer tracer.Tracer,
) error {
	tracer.Start(common.MsgStartImportGit, repo.GitURL)

	if !gitLFSDisabled {
		if err := command.CheckGitDependencies(); err != nil {
			tracer.LogError(common.ErrSkipGitLFS, err)
			gitLFSDisabled = true
		}
	}

	gitPath := filepath.Join(repoPath, types.GitDir)
	const remoteName = "harnessRemote"

	output, err := command.RunGitCommand(ctx, gitPath, []string{}, "remote", "add", remoteName, repo.GitURL)
	if err != nil {
		tracer.Stop(common.ErrGitRemoteAdd, err, string(output))
		return fmt.Errorf("failed to add remote %q: %w", repo.GitURL, err)
	}

	if !gitLFSDisabled {
		if err := m.handleLFSPush(ctx, gitPath, repo, remoteName, lfsObjectCount, tracer); err != nil {
			tracer.Stop(common.ErrGitLFSPush, err)
			return err
		}
		m.Report[repoRef].ReportMetric(report.ReportTypeGitLFSObjects, lfsObjectCount)
	}

	params := pushParams{
		gitPath:        gitPath,
		repo:           repo,
		remoteName:     remoteName,
		lfsObjectCount: lfsObjectCount,
		token:          m.HarnessToken,
	}

	pusher := m.selectGitPusher(params, gitLFSDisabled)
	if err := pusher.push(ctx, repo, gitPath, remoteName); err != nil {
		tracer.Stop(common.ErrGitPush, err)
		return fmt.Errorf("failed to push git repository: %w", err)
	}

	tracer.Stop(common.MsgCompleteImportGit, repo.GitURL)
	return nil
}

func (m *Importer) selectGitPusher(params pushParams, gitLFSDisabled bool) gitPusher {
	if gitLFSDisabled {
		return &goGitPusher{params: params, tracer: m.Tracer}
	}
	return &nativeGitPusher{params: params, tracer: m.Tracer}
}

func (m *Importer) handleLFSPush(
	ctx context.Context,
	gitPath string,
	repo *harness.Repository,
	remoteName string,
	lfsObjectCount int,
	tracer tracer.Tracer,
) error {
	if lfsObjectCount == 0 {
		return nil
	}

	output, err := command.RunGitCommandWithAuth(ctx, gitPath,
		command.Credentials{Username: "git-importer", Password: m.HarnessToken},
		"lfs", "push", "--all", remoteName, repo.DefaultBranch)
	if err != nil {
		tracer.LogError(common.ErrGitLFSPush, err, string(output))
		return fmt.Errorf("failed to push LFS objects to %q: %w", repo.GitURL, err)
	}

	return nil
}

func (p *nativeGitPusher) push(ctx context.Context, repo *harness.Repository, gitPath, remoteName string) error {
	output, err := command.RunGitCommandWithAuth(ctx, gitPath,
		command.Credentials{Username: "git-importer", Password: p.params.token},
		"push", remoteName,
		"refs/heads/*:refs/heads/*",
		"refs/tags/*:refs/tags/*",
		"refs/pullreq/*/head:refs/pullreq/*/head")
	if err != nil {
		p.tracer.LogError(common.ErrGitPush, repo.GitURL, err, string(output))
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
		p.tracer.Stop("failed to open git dir from %q", gitPath)
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
			Username: "git-importer",
			Password: p.params.token,
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
		p.tracer.LogError(common.ErrGitPush, repo.GitURL, err, output.String())
		return fmt.Errorf("failed to push refs to %q: %w", repo.GitURL, err)
	}

	return nil
}
