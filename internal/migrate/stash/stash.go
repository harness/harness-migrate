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

package stash

import (
	"context"
	"fmt"
	"strings"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/codeerror"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

const (
	pullRequestCheckpointPage = "%s/pr"
	pullRequestCheckpointData = "%s/pr/data"
)

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
			e.tracer.LogError(common.ErrRepoList, err)
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

func (e *Export) ListPullRequests(
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
		e.tracer.LogError(common.ErrCheckpointDataRead, err)
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
			e.tracer.LogError(common.ErrPrList, err)
			return nil, fmt.Errorf("cannot list pr: %w", err)
		}
		mappedPrs := gitexporter.MapPullRequest(prs)
		allPrs = append(allPrs, mappedPrs...)

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allPrs)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrDataSave)
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, resp.Page.Next)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrPageSave)
		}

		if resp.Page.Next == 0 {
			break
		}
		opts.Page = resp.Page.Next
	}

	e.tracer.Stop(common.MsgCompletePrImport, repoSlug, len(allPrs))

	err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointPrPageSave)
	}

	return allPrs, nil
}

func (e *Export) PullRequestReviewers(
	context.Context,
	int) error {
	return &codeerror.OpNotSupportedError{Name: "pullreqreview"}
}

func (e *Export) ListPullRequestComments(
	ctx context.Context,
	repoSlug string, prNumber int,
	opts types.ListOptions,
) ([]*types.PRComment, error) {
	e.tracer.Start(common.MsgStartPrCommentsImport, repoSlug, prNumber)
	allComments := []*types.PRComment{}
	defer e.tracer.Stop(common.MsgCompletePrCommentsImport, repoSlug, prNumber, len(allComments))
	for {
		comments, res, err := e.stash.ListPRComments(ctx, repoSlug, prNumber, opts, e.tracer)
		if err != nil {
			e.tracer.LogError(common.ErrCommentsList, repoSlug, prNumber, err)
			return nil, fmt.Errorf(common.ErrCommentsList, repoSlug, prNumber, err)
		}
		allComments = append(allComments, comments...)
		if res.Page.Next == 0 {
			break
		}
		opts.Page = res.Page.Next
	}
	return allComments, nil
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
