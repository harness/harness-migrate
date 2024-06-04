package common

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/drone/go-scm/scm"
	git "github.com/go-git/go-git/v5"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/util"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func (e *Exporter) CloneRepository(
	ctx context.Context,
	repoData scm.Repository,
	repoPath string,
	repoSlug string, //for logging
	tracer tracer.Tracer,
) error {
	tracer.Start(MsgStartGitClone, repoSlug)
	gitPath := filepath.Join(repoPath, "git")
	if err := util.CreateFolder(gitPath); err != nil {
		return err
	}

	_, err := git.PlainCloneContext(ctx, gitPath, true, &git.CloneOptions{
		URL: repoData.Clone,
		Auth: &http.BasicAuth{
			Username: e.ScmLogin,
			Password: e.ScmToken,
		},
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", repoData.Branch)),
		SingleBranch:  false,
		Tags:          git.AllTags,
		NoCheckout:    true,
	})
	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
		tracer.Log(MsgCloneEmptyRepo, repoData.Clone)
		return nil
	}
	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		tracer.Log(MsgRepoAlreadyExists, repoSlug)
		return nil
	}
	if err != nil {
		tracer.LogError(ErrGitCloneMsg, repoSlug, err)
		return fmt.Errorf("failed to clone repo %s from %q: %w", repoSlug, repoData.Clone, err)
	}
	tracer.Stop(MsgCompleteGitClone, repoSlug)

	return nil
}
