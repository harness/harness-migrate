package gitimporter

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

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
	tracer.Start("start git push to %q", repo.GitURL)
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

	err = gitRepo.PushContext(ctx, &git.PushOptions{
		RemoteName: gitRemoteHarness,
		Auth: &http.BasicAuth{
			Username: "git-importer",
			Password: m.HarnessToken,
		},
		RefSpecs: []config.RefSpec{
			"refs/heads/*:refs/heads/*",
			"refs/tags/*:refs/tags/*",
			// "refs/pullreq/*/head:refs/pullreq/*/head", TODO: uncomment after proper bypass pre-receive hooks for importing pr refs
			// "refs/pullreq/*/merge:refs/pullreq/*/merge",
		},
		RemoteURL:       repo.GitURL,
		Force:           true,
		InsecureSkipTLS: true,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		tracer.Stop("failed to push to '%s'", repo.GitURL)
		return fmt.Errorf("failed to push repo to '%s': %w", repo.GitURL, err)
	}

	tracer.Stop("finished git push to '%s'", repo.GitURL)
	return nil
}
