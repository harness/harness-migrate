package stash

import (
	"context"
	"fmt"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

func (e *Export) ListPullRequests(
	ctx context.Context,
	repoSlug string,
	params types.PullRequestListOptions,
) ([]types.PRResponse, error) {
	e.tracer.Start(common.MsgStartExportPRs, repoSlug)
	opts := scm.PullRequestListOptions{
		Page:   params.Page,
		Size:   params.Size,
		Open:   params.Open,
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
		prs, resp, err := e.stash.PullRequests.List(ctx, repoSlug, opts)
		if err != nil {
			e.tracer.LogError(common.ErrListPr, err)
			return nil, fmt.Errorf("cannot list prs: %w", err)
		}
		mappedPrs := common.MapPullRequest(prs)
		allPrs = append(allPrs, mappedPrs...)

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allPrs)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrDataSave, err)
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, resp.Page.Next)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointPrPageSave, err)
		}

		if resp.Page.Next == 0 {
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
