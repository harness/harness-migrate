package stash

import (
	"context"
	"fmt"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/codeerror"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
)

const (
	pullRequestCheckpointPage = "%s/pr"
	pullRequestCheckpointData = "%s/pr/data"
)

type Export struct {
	stash    *scm.Client
	stashOrg string

	checkpointManager *checkpoint.CheckpointManager

	tracer tracer.Tracer
}

func New(
	client *scm.Client,
	org string,
	checkpointer *checkpoint.CheckpointManager,
	tracer tracer.Tracer,
) *Export {
	return &Export{
		stash:             client,
		stashOrg:          org,
		checkpointManager: checkpointer,
		tracer:            tracer,
	}
}

func (e *Export) ListRepositories(
	ctx context.Context,
	_ types.ListRepoOptions,
) ([]types.RepoResponse, error) {
	e.tracer.Start(common.MsgStartRepoList, "bitbucket", e.stashOrg)
	opts := scm.ListOptions{Size: 25}
	var allRepos []*scm.Repository

	for {
		repos, resp, err := e.stash.Repositories.ListNamespace(ctx, e.stashOrg, opts)
		if err != nil {
			e.tracer.LogError(common.MsgRepoListErr, err)
		}
		allRepos = append(allRepos, repos...)

		if resp.Page.Next == 0 {
			break
		}
		opts.Page = resp.Page.Next
	}
	e.tracer.Stop(common.MsgCompleteRepoList, len(allRepos))

	return mapRepository(allRepos), nil
}

func (e *Export) ListPullRequest(
	ctx context.Context,
	repoSlug string,
	_ types.PullRequestListOptions,
) ([]types.PRResponse, error) {
	e.tracer.Start(common.MsgStartPrImport, repoSlug)
	var allPrs []types.PRResponse
	opts := scm.PullRequestListOptions{Size: 25}

	checkpointDataKey := fmt.Sprintf(pullRequestCheckpointData, repoSlug)
	val, ok, err := checkpoint.GetCheckpointData[[]types.PRResponse](e.checkpointManager, checkpointDataKey)
	if err != nil {
		e.tracer.LogError(common.MsgCheckpointDataReadErr, err)
		panic(common.PanicCheckpointSaveErr)
	}
	if ok && val != nil {
		allPrs = append(allPrs, val...)
	}

	checkpointPageKey := fmt.Sprintf(pullRequestCheckpointPage, repoSlug)
	checkpointPageIntfc, ok := e.checkpointManager.GetCheckpoint(checkpointPageKey)
	var checkpointPage int
	if ok && checkpointPageIntfc != nil {
		checkpointPage = int(checkpointPageIntfc.(float64))
		opts.Page = checkpointPage
	}

	// all pages done
	if checkpointPage == -1 {
		e.tracer.Stop(common.MsgCheckpointLoadPr, repoSlug, len(allPrs))
		return allPrs, nil
	}

	for {
		prs, resp, err := e.stash.PullRequests.List(ctx, repoSlug, opts)
		if err != nil {
			e.tracer.LogError(common.MsgPrListErr, err)
			return nil, fmt.Errorf("cannot list pr: %w", err)
		}
		mappedPrs := common.MapPullRequest(prs)
		allPrs = append(allPrs, mappedPrs...)

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allPrs)
		if err != nil {
			e.tracer.LogError(common.MsgCheckpointPrDataSaveErr)
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, resp.Page.Next)
		if err != nil {
			e.tracer.LogError(common.MsgCheckpointPrPageSaveErr)
		}

		if resp.Page.Next == 0 {
			break
		}
		opts.Page = resp.Page.Next
	}

	e.tracer.Stop(common.MsgCompletePrImport, repoSlug, len(allPrs))

	err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
	if err != nil {
		e.tracer.LogError(common.MsgCheckpointPrPageSaveErr)
	}

	return allPrs, nil
}

func (e *Export) PullRequestReviewers(
	context.Context,
	int) error {
	return &codeerror.ErrorOpNotSupported{Name: "pullreqreview"}
}

func (e *Export) PullRequestComments(
	context.Context,
	int,
) error {
	return &codeerror.ErrorOpNotSupported{Name: "pullreqcomment"}
}

func mapRepository(repos []*scm.Repository) []types.RepoResponse {
	r := make([]types.RepoResponse, len(repos))
	for i, repo := range repos {
		repoSlug := make([]string, 0)

		if repo.Namespace != "" {
			repoSlug = append(repoSlug, repo.Namespace)
		}
		if repo.Name != "" {
			repoSlug = append(repoSlug, repo.Name)
		}

		r[i] = types.RepoResponse{Repository: *repo, RepoSlug: strings.Join(repoSlug, "/")}
	}
	return r
}
