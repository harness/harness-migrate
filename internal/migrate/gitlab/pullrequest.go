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

package gitlab

import (
	"context"
	"errors"
	"fmt"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/types"
)

func (e *Export) ListPullRequests(
	ctx context.Context,
	repoSlug string,
	opt types.PullRequestListOptions,
) ([]types.PRResponse, error) {
	e.tracer.Start(common.MsgStartExportPRs, repoSlug)
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

	prNumber := 1

	checkpointPRNumKey := fmt.Sprintf(common.PullRequestCheckpointPage, repoSlug)
	checkpointPageIntfc, ok := e.checkpointManager.GetCheckpoint(checkpointPRNumKey)

	if ok && checkpointPageIntfc != nil {
		prNumber = int(checkpointPageIntfc.(float64))
	}

	// all pages done
	if prNumber == -1 {
		msgPrExport = common.MsgCheckpointLoadPr
		return allPrs, nil
	}

	maxMRNumber, err := e.GetHighestMRNumber(ctx, repoSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count of merge requests in project %s: %w", repoSlug, err)
	}

	for prNumber <= maxMRNumber {
		pr, _, err := e.FindPR(ctx, repoSlug, prNumber)
		if err != nil && !errors.Is(err, harness.ErrNotFound) {
			e.tracer.LogError(common.ErrListPr, err)
			return nil, fmt.Errorf("cannot find pr %d: %w", prNumber, err)
		}

		if pr == nil {
			prNumber++
			continue
		}

		mappedPrWithAuthor, err := e.addEmailToPRAuthor(ctx, pr)
		if err != nil {
			return nil, fmt.Errorf("cannot add email to author: %w", err)
		}
		allPrs = append(allPrs, mappedPrWithAuthor)

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allPrs)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrDataSave, err)
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointPRNumKey, prNumber)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrPageSave, err)
		}

		prNumber++
	}

	err = e.checkpointManager.SaveCheckpoint(checkpointPRNumKey, -1)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointPrPageSave, err)
	}

	return allPrs, nil
}

func (e *Export) addEmailToPRAuthor(ctx context.Context, pr *types.PRResponse) (types.PRResponse, error) {
	email, err := e.FindEmailByUsername(ctx, pr.Author.Login)
	if err != nil {
		return types.PRResponse{}, fmt.Errorf("cannot find email for author %s: %w", pr.Author.Login, err)
	}
	pr.Author.Email = email

	return *pr, nil
}
