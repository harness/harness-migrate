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
	"fmt"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"
)

func (e *Export) ListPullRequestComments(
	ctx context.Context,
	repoSlug string,
	prNumber int,
	opts types.ListOptions,
) ([]*types.PRComment, error) {
	e.tracer.Start(common.MsgStartExportPrComments, repoSlug, prNumber)
	var allComments []*types.PRComment
	msgCommentsExport := common.MsgCompleteExportPrComments
	defer func() {
		e.tracer.Stop(msgCommentsExport, len(allComments), repoSlug, prNumber)
	}()

	checkpointDataKey := fmt.Sprintf(common.PRCommentCheckpointData, repoSlug, prNumber)
	val, ok, err := checkpoint.GetCheckpointData[[]*types.PRComment](e.checkpointManager, checkpointDataKey)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointDataRead, err)
		panic(common.PanicCheckpointSaveErr)
	}
	if ok && val != nil {
		allComments = append(allComments, val...)
	}

	checkpointPageKey := fmt.Sprintf(common.PRCommentCheckpointPage, repoSlug, prNumber)
	checkpointPageIntfc, ok := e.checkpointManager.GetCheckpoint(checkpointPageKey)
	var checkpointPage int
	if ok && checkpointPageIntfc != nil {
		checkpointPage = int(checkpointPageIntfc.(float64))
		opts.Page = checkpointPage
	}

	// all pages done
	if checkpointPage == -1 {
		msgCommentsExport = common.MsgCheckpointLoadPRComments
		return allComments, nil
	}

	// for fetching merge requests comments
	for {
		comments, res, err := e.ListPRComments(ctx, repoSlug, prNumber, opts)
		if err != nil {
			e.tracer.LogError(common.ErrListComments, repoSlug, prNumber, err)
			return nil, fmt.Errorf(common.ErrListComments, repoSlug, prNumber, err)
		}
		if len(comments) == 0 {
			break
		}

		commentsWithAuthor, err := e.addEmailToAuthorInComments(ctx, comments)
		if err != nil {
			return nil, fmt.Errorf("error getting author email: %w", err)
		}
		allComments = append(allComments, commentsWithAuthor...)

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allComments)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrCommentsDataSave, err)
		}
		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, res.Page.Next)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrCommentsPageSave, err)
		}

		opts.Page += 1
	}

	err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointPrCommentsPageSave)
	}

	return allComments, nil
}

func (e *Export) addEmailToAuthorInComments(ctx context.Context, comments []*types.PRComment) ([]*types.PRComment, error) {
	commentsCopy := make([]*types.PRComment, len(comments))
	for i, comment := range comments {
		commentCopy := *comment
		email, err := e.FindEmailByUsername(ctx, commentCopy.Author.Login)
		if err != nil {
			return nil, fmt.Errorf("cannot find email for author %s: %w", commentCopy.Author.Login, err)
		}
		commentCopy.Author.Email = email
		commentsCopy[i] = &commentCopy
	}
	return commentsCopy, nil
}
