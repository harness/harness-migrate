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

// Package github provides automatic migration tools from Github to Harness.
package github

import (
	"context"
	"fmt"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"
)

func (e *Export) ListPullRequestReviews(
	ctx context.Context,
	repoSlug string, prNumber int,
	opts types.ListOptions,
) ([]*types.PRReview, error) {
	e.tracer.Debug().Start(common.MsgStartExportPrReviewers, repoSlug, prNumber)
	var allReviewers []*types.PRReview
	msgReviewersExport := common.MsgCompleteExportPrReviewers
	defer func() {
		e.tracer.Debug().Stop(msgReviewersExport, len(allReviewers), repoSlug, prNumber)
	}()

	checkpointDataKey := fmt.Sprintf(common.PRReviewerCheckpointData, repoSlug, prNumber)
	val, ok, err := checkpoint.GetCheckpointData[[]*types.PRReview](e.checkpointManager, checkpointDataKey)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointDataRead, err)
		panic(common.PanicCheckpointSaveErr)
	}
	if ok && val != nil {
		allReviewers = append(allReviewers, val...)
	}

	checkpointPageKey := fmt.Sprintf(common.PRReviewerCheckpointPage, repoSlug, prNumber)
	checkpointPageIntfc, ok := e.checkpointManager.GetCheckpoint(checkpointPageKey)
	var checkpointPage int
	if ok && checkpointPageIntfc != nil {
		checkpointPage = int(checkpointPageIntfc.(float64))
		opts.Page = checkpointPage
	}

	// all pages done
	if checkpointPage == -1 {
		msgReviewersExport = common.MsgCheckpointLoadPRReviewers
		return allReviewers, nil
	}

	currentPage := opts.Page
	if currentPage == 0 {
		currentPage = 1
	}

	for {
		pageOpts := types.ListOptions{Page: currentPage, Size: opts.Size}
		reviews, res, err := e.ListPRReviews(ctx, repoSlug, prNumber, pageOpts)
		if err != nil {
			e.tracer.LogError(common.ErrListReviewers, repoSlug, prNumber, err)
			return nil, fmt.Errorf(common.ErrListReviewers, repoSlug, prNumber, err)
		}
		if len(reviews) == 0 {
			break
		}

		reviewersWithAuthor, err := e.addEmailToAuthorInReviewers(ctx, reviews)
		if err != nil {
			return nil, fmt.Errorf("error getting author email: %w", err)
		}
		allReviewers = append(allReviewers, reviewersWithAuthor...)

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allReviewers)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrReviewersDataSave, err)
		}
		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, res.Page.Next)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrReviewersPageSave, err)
		}

		if res.Page.Next == 0 || len(reviews) < opts.Size {
			break
		}
		currentPage = res.Page.Next
	}

	err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointPrReviewersPageSave)
	}

	return allReviewers, nil
}

func (e *Export) addEmailToAuthorInReviewers(ctx context.Context, reviewers []*types.PRReview) ([]*types.PRReview, error) {
	reviewersCopy := make([]*types.PRReview, len(reviewers))
	for i, reviewer := range reviewers {
		reviewerCopy := *reviewer
		email, err := e.FindEmailByUsername(ctx, reviewerCopy.Review.Author.Login)
		if err != nil {
			return nil, fmt.Errorf("cannot find email for author %s: %w", reviewerCopy.Review.Author.Login, err)
		}
		reviewerCopy.Review.Author.Email = email
		reviewersCopy[i] = &reviewerCopy
	}
	return reviewersCopy, nil
}

func (e *Export) addEmailToRequestedReviewers(ctx context.Context, reviewers []*types.PRReviewer) ([]*types.PRReviewer, error) {
	reviewersCopy := make([]*types.PRReviewer, len(reviewers))
	for i, reviewer := range reviewers {
		reviewerCopy := *reviewer
		email, err := e.FindEmailByUsername(ctx, reviewerCopy.User.Login)
		if err != nil {
			return nil, fmt.Errorf("cannot find email for requested reviewer %s: %w", reviewerCopy.User.Login, err)
		}
		reviewerCopy.User.Email = email

		reviewersCopy[i] = &reviewerCopy
	}
	return reviewersCopy, nil
}
