package stash

import (
	"context"
	"fmt"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"
)

const (
	prCommentCheckpointPage = "%s/%d/comment"
	prCommentCheckpointData = "%s/%d/comment/data"
)

func (e *Export) ListPullRequestComments(
	ctx context.Context,
	repoSlug string, prNumber int,
	opts types.ListOptions,
) ([]*types.PRComment, error) {
	e.tracer.Start(common.MsgStartPrCommentsExport, repoSlug, prNumber)
	allComments := []*types.PRComment{}
	msgCommentsExport := common.MsgCompletePrCommentsExport
	defer func() {
		e.tracer.Stop(msgCommentsExport, repoSlug, prNumber, len(allComments))
	}()

	checkpointDataKey := fmt.Sprintf(prCommentCheckpointData, repoSlug, prNumber)
	val, ok, err := checkpoint.GetCheckpointData[[]*types.PRComment](e.checkpointManager, checkpointDataKey)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointDataRead, err)
		panic(common.PanicCheckpointSaveErr)
	}
	if ok && val != nil {
		allComments = append(allComments, val...)
	}

	checkpointPageKey := fmt.Sprintf(prCommentCheckpointPage, repoSlug, prNumber)
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

	for {
		comments, res, err := e.stash.ListPRComments(ctx, repoSlug, prNumber, opts, e.tracer)
		if err != nil {
			e.tracer.LogError(common.ErrCommentsList, repoSlug, prNumber, err)
			return nil, fmt.Errorf(common.ErrCommentsList, repoSlug, prNumber, err)
		}
		allComments = append(allComments, comments...)

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allComments)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrCommentsDataSave)
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, res.Page.Next)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrCommentsPageSave)
		}

		if res.Page.Next == 0 {
			break
		}
		opts.Page = res.Page.Next
	}

	err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointPrCommentsPageSave)
	}

	return allComments, nil
}
