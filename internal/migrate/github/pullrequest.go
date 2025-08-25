// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package github

import (
	"context"
	"fmt"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

// IMPORT STEP 1
//const skipUpTo = 60646

// PR range limits
const maxPR = 11364 // 32817 //32818 //30999
const minPR = 11201 // 32102 //31000 // 1

func (e *Export) ListPullRequests(
	ctx context.Context,
	repoSlug string,
	params types.PullRequestListOptions,
) ([]types.PRResponse, error) {
	e.tracer.Start(common.MsgStartExportPRs, repoSlug)
	// IMPORT STEP 1
	//params.Open = false
	opts := scm.PullRequestListOptions{
		Page:   params.Page,
		Size:   params.Size,
		Open:   true, // params.Open, //params.Open,
		Closed: true, // params.Closed,
	}
	var allPrs []types.PRResponse
	msgPrExport := common.MsgCompleteExportPRs
	defer func() {
		e.tracer.Stop(msgPrExport, len(allPrs), repoSlug)
	}()

	checkpointDataKey := fmt.Sprintf(common.PullRequestCheckpointData, repoSlug)
	val, ok, err := checkpoint.GetCheckpointData[[]types.PRResponse](e.checkpointManager, checkpointDataKey)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointDataRead, err)
		return nil, fmt.Errorf(common.PanicCheckpointSaveErr)
	}
	if ok && val != nil {
		allPrs = append(allPrs, val...)
	}

	checkpointPageKey := fmt.Sprintf(common.PullRequestCheckpointPage, repoSlug)
	checkpointPageIntfc, ok := e.checkpointManager.GetCheckpoint(checkpointPageKey)
	var checkpointPage int
	if ok && checkpointPageIntfc != nil {
		checkpointPage = int(checkpointPageIntfc.(float64))
		opts.Page = checkpointPage
	}

	// all pages done
	if checkpointPage == -1 {
		msgPrExport = common.MsgCheckpointLoadPr
		return allPrs, nil
	}

	for {
		prs, resp, err := e.github.PullRequests.List(ctx, repoSlug, opts)
		if err != nil {
			e.tracer.LogError(common.ErrListPr, err)
			return nil, fmt.Errorf("cannot list prs: %w", err)
		}
		mappedPrs := common.MapPullRequest(prs)
		// Filter PRs to only include those in our desired range (1-30999)
		mappedPrs = filterPRsInRange(mappedPrs, minPR, maxPR)

		mappedPrsWithAuthor, err := e.addEmailToPRAuthor(ctx, mappedPrs)
		if err != nil {
			return nil, fmt.Errorf("cannot add email to author: %w", err)
		}
		allPrs = append(allPrs, mappedPrsWithAuthor...)

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allPrs)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrDataSave, err)
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, resp.Page.Next)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrPageSave, err)
		}

		// Break if no more pages
		if resp.Page.Next == 0 {
			break
		}

		// If all PRs in this page have numbers less than our minimum range, we can stop
		// This assumes GitHub returns PRs in descending order by default
		if len(prs) > 0 && prs[len(prs)-1].Number <= minPR {
			break
		}
		opts.Page = resp.Page.Next
	}

	err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointPrPageSave, err)
	}

	return allPrs, nil
}

func (e *Export) addEmailToPRAuthor(ctx context.Context, prs []types.PRResponse) ([]types.PRResponse, error) {
	for i, pr := range prs {
		email, err := e.FindEmailByUsername(ctx, pr.Author.Login)
		if err != nil {
			return nil, fmt.Errorf("cannot find email for author %s: %w", pr.Author.Login, err)
		}
		pr.Author.Email = email
		prs[i] = pr
	}
	return prs, nil
}

// filterPRsInRange filters PRs to only include those with numbers in the specified range
func filterPRsInRange(prs []types.PRResponse, minPRNumber, maxPRNumber int) []types.PRResponse {
	var filteredPRs []types.PRResponse
	for _, pr := range prs {
		if pr.Number >= minPRNumber && pr.Number <= maxPRNumber {
			filteredPRs = append(filteredPRs, pr)
		}
	}
	return filteredPRs
}
