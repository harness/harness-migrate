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

package bitbucket

import (
	"context"
	"fmt"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"
)

func (e *Export) ListBranchRules(
	ctx context.Context,
	repoSlug string,
	opts types.ListOptions,
) ([]*types.BranchRule, error) {
	e.tracer.Start(common.MsgStartExportBranchRules, repoSlug)
	var allRules []*types.BranchRule

	checkpointDataKey := fmt.Sprintf(common.RuleCheckpointData, repoSlug)
	val, ok, err := checkpoint.GetCheckpointData[[]*types.BranchRule](e.checkpointManager, checkpointDataKey)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointDataRead, err)
	}
	if ok && val != nil {
		allRules = append(allRules, val...)
	}

	checkpointPageKey := fmt.Sprintf(common.RuleCheckpointPage, repoSlug)
	checkpointPageIntfc, ok := e.checkpointManager.GetCheckpoint(checkpointPageKey)
	var checkpointPage int
	if ok && checkpointPageIntfc != nil {
		checkpointPage = int(checkpointPageIntfc.(float64))
		opts.Page = checkpointPage
	}

	// all rules pages are done
	if checkpointPage == -1 {
		return allRules, nil
	}

	for {
		rules, resp, err := e.ListBranchRulesInternal(ctx, repoSlug, opts)
		if err != nil {
			e.tracer.LogError(common.ErrListBranchRules, repoSlug, err)
			e.tracer.Stop(common.MsgFailedExportBranchRules, repoSlug)
			return nil, fmt.Errorf(common.ErrListBranchRules, repoSlug, err)
		}
		allRules = append(allRules, rules...)

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allRules)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointRulesDataSave, repoSlug, err)
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, resp.Page.Next)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointRulesPageSave, repoSlug, err)
		}

		if resp.Page.Next == 0 {
			break
		}
		opts.Page = resp.Page.Next
	}

	err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointRulesPageSave, repoSlug, err)
	}

	e.tracer.Stop(common.MsgCompleteExportBranchRules, len(allRules), repoSlug)
	return allRules, nil
}
