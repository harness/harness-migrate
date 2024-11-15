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
	"net/url"
	"strconv"
	"strings"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

// IMPORT STEP 1
const skipUpTo = 60646

// IMPORT STEP 2
const maxPR = 60647

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
		Open:   params.Open, //params.Open,
		Closed: params.Closed,
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
		// 		TEMPORARY FOR IMPORT STEP 2
		// filter out already fetched prs
		//
		// totalPages := 1
		// linkHeader := resp.Header.Get("Link")
		// if linkHeader != "" {
		// 	totalPages = parseLastPageFromLink(linkHeader)
		// }

		// skipToPage := (maxPR / params.Size) + 1
		// if skipToPage < totalPages {
		// 	opts.Page = skipToPage
		// }

		mappedPrs := common.MapPullRequest(prs)

		// TEMPORARY FOR IMPORT STEP 1
		mappedPrs = filterPRs(mappedPrs, skipUpTo)

		// TEMPORARY FOR IMPORT STEP 2
		//mappedPrs = filterPRs(mappedPrs, maxPR)

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

		if resp.Page.Next == 0 || prs[len(prs)-1].Number <= maxPR {
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

// filterPRs filters PRs that have a number greater than the last processed PR number
func filterPRs(prs []types.PRResponse, lastPRNumber int) []types.PRResponse {
	var newPRs []types.PRResponse
	for _, pr := range prs {
		if pr.Number > lastPRNumber {
			newPRs = append(newPRs, pr)
		}
	}
	return newPRs
}

func parseLastPageFromLink(linkHeader string) int {
	// Find the link with rel="last"
	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		parts := strings.Split(strings.TrimSpace(link), ";")
		if len(parts) < 2 {
			continue
		}
		// Look for `rel="last"`
		if strings.Contains(parts[1], `rel="last"`) {
			// Extract the page number from the URL
			u, err := url.Parse(parts[0][1 : len(parts[0])-1]) // Remove angle brackets
			if err != nil {
				return 1
			}
			page := u.Query().Get("page")
			totalPages, err := strconv.Atoi(page)
			if err != nil {
				return 1
			}
			return totalPages
		}
	}
	return 1
}
