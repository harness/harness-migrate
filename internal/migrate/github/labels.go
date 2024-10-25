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
	externalTypes "github.com/harness/harness-migrate/types"
)

func (e *Export) ListLabels(
	ctx context.Context,
	repoSlug string,
	opts types.ListOptions,
) (map[string]externalTypes.Label, error) {
	e.tracer.Start(common.MsgStartExportLabels, repoSlug)
	allLabels := make(map[string]externalTypes.Label)
	defer func() {
		e.tracer.Stop(common.MsgCompleteExportLabels, len(allLabels), repoSlug)
	}()

	checkpointDataKey := fmt.Sprintf(common.LabelCheckpointData, repoSlug)
	val, ok, err := checkpoint.GetCheckpointData[map[string]externalTypes.Label](e.checkpointManager, checkpointDataKey)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointDataRead, err)
		panic(common.PanicCheckpointSaveErr)
	}
	if ok && val != nil {
		for k, v := range val {
			allLabels[k] = v
		}
	}

	checkpointPageKey := fmt.Sprintf(common.LabelCheckpointPage, repoSlug)
	checkpointPageIntfc, ok := e.checkpointManager.GetCheckpoint(checkpointPageKey)
	var checkpointPage int
	if ok && checkpointPageIntfc != nil {
		checkpointPage = int(checkpointPageIntfc.(float64))
		opts.Page = checkpointPage
	}

	// all pages done
	if checkpointPage == -1 {
		return allLabels, nil
	}

	for {
		labels, res, err := e.ListRepoLabels(ctx, repoSlug, opts)
		if err != nil {
			e.tracer.LogError(common.ErrListLabels, repoSlug, err)
			return nil, fmt.Errorf(common.ErrListLabels, repoSlug, err)
		}
		if len(labels) == 0 {
			break
		}

		for _, label := range labels {
			allLabels[label.Name] = label
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allLabels)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointLabelsDataSave, err)
		}
		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, res.Page.Next)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointLabelsPageSave, err)
		}

		opts.Page += 1
	}

	err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointPrPageSave, err)
	}

	return allLabels, nil
}
