package gitimporter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/types"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func (m *Importer) Push(
	ctx context.Context,
	repoPath string,
	repo *harness.Repository,
	tracer tracer.Tracer,
) error {
	tracer.Start(common.MsgStartImportGit, repo.GitURL)
	gitPath := filepath.Join(repoPath, types.GitDir)

	gitRepo, err := git.PlainOpen(gitPath)
	if err != nil {
		tracer.Stop("failed to open git dir from %q", repoPath)
		return fmt.Errorf("failed to open the exported repository from %q: %w", repoPath, err)
	}

	const gitRemoteHarness = "harnessCode"
	_, err = gitRepo.CreateRemote(&config.RemoteConfig{
		Name: gitRemoteHarness,
		URLs: []string{repo.GitURL},
	})
	if err != nil && !errors.Is(err, git.ErrRemoteExists) {
		return fmt.Errorf("failed to set remote to %q: %w", repo.GitURL, err)
	}

	var output bytes.Buffer

	err = gitRepo.PushContext(ctx, &git.PushOptions{
		RemoteName: gitRemoteHarness,
		Auth: &http.BasicAuth{
			Username: "git-importer",
			Password: m.HarnessToken,
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
		tracer.Stop(common.ErrGitPush, repo.GitURL, err, output.String())
		return fmt.Errorf(common.ErrGitPush, repo.GitURL, err, output.String())
	}

	tracer.Stop(common.MsgCompleteImportGit, repo.GitURL)
	return nil
}
