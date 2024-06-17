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
	"errors"
	"fmt"
	"strings"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/codeerror"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/types/enum"

	"github.com/drone/go-scm/scm"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
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

	if e.stashRepository != "" {
		repoSlug := strings.Join([]string{e.stashOrg, e.stashRepository}, "/")
		repo, _, err := e.stash.Repositories.Find(ctx, repoSlug)
		if err != nil {
			e.tracer.LogError(common.ErrRepoList, err)
			return nil, fmt.Errorf("failed to get the repo %s: %w", repoSlug, err)
		}
		e.tracer.Stop(common.MsgCompleteRepoList, 1)
		return mapRepository([]*scm.Repository{repo}), nil
	}

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
	e.tracer.Start(common.MsgStartPrExport, repoSlug)
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

	e.tracer.Stop(common.MsgCompletePrExport, repoSlug, len(allPrs))

	err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointPrPageSave)
	}

	return allPrs, nil
}

func (e *Export) FetchPullRequestRefs(ctx context.Context, repo *git.Repository, repoSlug string, stashLogin string, stashToken string) error {
	e.tracer.Start(common.MsgGitFetchRef, repoSlug)
	// skip fetching references for empty repos.
	if repo == nil {
		e.tracer.Stop(common.MsgCompleteGitFetchRef, repoSlug)
		return nil
	}

	refSpecs := []config.RefSpec{"refs/pull-requests/*/from:refs/pullreq/*/head"}
	err := repo.Fetch(&git.FetchOptions{
		RefSpecs: refSpecs,
		Auth: &http.BasicAuth{
			Username: stashLogin,
			Password: stashToken,
		},
		Force: true,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		e.tracer.LogError(common.ErrGitFetch, repoSlug, err)
		e.tracer.Stop(common.ErrGitFetch, repoSlug, err)
		return fmt.Errorf("failed to fetch repo pull requests references %s: %w", repoSlug, err)
	}

	e.tracer.Stop(common.MsgCompleteGitFetchRef, repoSlug)
	return nil
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
	e.tracer.Start(common.MsgStartPrCommentsExport, repoSlug, prNumber)
	allComments := []*types.PRComment{}
	defer e.tracer.Stop(common.MsgCompletePrCommentsExport, repoSlug, prNumber, len(allComments))
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

func (e *Export) ListWebhooks(
	ctx context.Context,
	repoSlug string,
	logger gitexporter.Logger,
	_ types.WebhookListOptions,
) (types.WebhookData, error) {
	e.tracer.Start(common.MsgStartWebhookList, repoSlug)

	var allWebhooks []*scm.Hook
	defer func() {
		e.tracer.Stop("fetched repo %q webhooks. Total: %d", repoSlug, len(allWebhooks))
	}()

	opts := scm.ListOptions{Size: 25, Page: 1}

	for {
		webhooks, resp, err := e.stash.Repositories.ListHooks(ctx, repoSlug, opts)
		if err != nil {
			e.tracer.LogError(common.ErrWebhookList, repoSlug, err)
			return types.WebhookData{}, err
		}
		allWebhooks = append(allWebhooks, webhooks...)

		if resp.Page.Next == 0 {
			break
		}
		opts.Page = resp.Page.Next
	}

	convertedHooks, notSupportedHooks := mapWebhooks(allWebhooks)
	// logs the not supported hooks
	if len(notSupportedHooks) != 0 {
		var logs []string
		var warningMsg string
		for _, hook := range notSupportedHooks {
			warningMsg = fmt.Sprintf("[%s] Skipped repo %q's webhook with ID %q Name %q for Target URL %q on events %v \n",
				enum.LogLevelWarning, repoSlug, hook.ID, hook.Name, hook.Target, hook.Events)
			logs = append(logs, warningMsg)
		}
		if err := logger.Log([]byte(strings.Join(logs, ""))); err != nil {
			return types.WebhookData{}, fmt.Errorf("failed to log the not supported webhooks for repo %q: %w",
				repoSlug, err)
		}
	}

	return types.WebhookData{
		ConvertedHooks: convertedHooks,
	}, nil
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

func mapWebhooks(hooks []*scm.Hook) ([]*scm.Hook, []*scm.Hook) {
	var convertedHooks []*scm.Hook
	var notSupportedHooks []*scm.Hook
	for _, hook := range hooks {
		convertedHook, notSupportedHook := mapWebhook(hook)
		if convertedHook != nil {
			convertedHooks = append(convertedHooks, convertedHook)
		}
		if notSupportedHook != nil {
			notSupportedHooks = append(notSupportedHooks, notSupportedHook)
		}
	}

	return convertedHooks, notSupportedHooks
}

// mapWebhook converts the supported events and returns not supported events in a separate webhook
func mapWebhook(hook *scm.Hook) (*scm.Hook, *scm.Hook) {
	events, notSupportedEvents := mapEvents(hook.Events)

	var convertedHook *scm.Hook
	var notSupportedHook *scm.Hook

	if len(events) != 0 {
		convertedHook = &scm.Hook{
			ID:     hook.ID,
			Name:   hook.Name,
			Target: hook.Target,
			Active: hook.Active,
			Events: enum.ToStringSlice(events),
		}
	}

	if len(notSupportedEvents) != 0 {
		notSupportedHook = &scm.Hook{
			ID:     hook.ID,
			Name:   hook.Name,
			Target: hook.Target,
			Active: hook.Active,
			Events: notSupportedEvents,
		}
	}
	return convertedHook, notSupportedHook
}

func mapEvents(triggers []string) ([]enum.WebhookTrigger, []string) {
	var events []enum.WebhookTrigger
	var notSupportedEvents []string

	for _, v := range triggers {
		switch v {
		case "repo:refs_changed":
			events = append(events, enum.WebhookTriggerBranchCreated, enum.WebhookTriggerBranchDeleted, enum.WebhookTriggerBranchUpdated, enum.WebhookTriggerTagCreated, enum.WebhookTriggerTagDeleted, enum.WebhookTriggerTagUpdated)
		case "pr:opened":
			events = append(events, enum.WebhookTriggerPullReqCreated, enum.WebhookTriggerPullReqReopened)
		case "pr:merged":
			events = append(events, enum.WebhookTriggerPullReqMerged)
		case "pr:declined":
			events = append(events, enum.WebhookTriggerPullReqClosed)
		case "pr:from_ref_updated":
			events = append(events, enum.WebhookTriggerPullReqBranchUpdated)
		case "pr:comment:added":
			events = append(events, enum.WebhookTriggerPullReqCommentCreated)
		default:
			notSupportedEvents = append(notSupportedEvents, v)
		}
	}

	return events, notSupportedEvents
}
