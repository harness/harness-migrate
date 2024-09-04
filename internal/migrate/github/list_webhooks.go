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

	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/migrate"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/types/enum"
)

// ListWebhooks implements gitexporter.Interface.
func (e *Export) ListWebhooks(
	ctx context.Context,
	repoSlug string,
	options types.WebhookListOptions,
) (types.WebhookData, error) {
	e.tracer.Start(common.MsgStartExportWebhook, repoSlug)
	var allWebhooks []*scm.Hook
	opts := scm.ListOptions{
		Size: options.Size,
		Page: options.Page,
	}

	checkpointDataKey := fmt.Sprintf(common.WebhookCheckpointData, repoSlug)
	val, ok, err := checkpoint.GetCheckpointData[[]*scm.Hook](e.checkpointManager, checkpointDataKey)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointDataRead, err)
	}
	if ok && val != nil {
		allWebhooks = append(allWebhooks, val...)
	}

	checkpointPageKey := fmt.Sprintf(common.WebhookCheckpointPage, repoSlug)
	checkpointPageIntfc, ok := e.checkpointManager.GetCheckpoint(checkpointPageKey)
	var checkpointPage int
	if ok && checkpointPageIntfc != nil {
		checkpointPage = int(checkpointPageIntfc.(float64))
		opts.Page = checkpointPage
	}

	// all pages are done
	if checkpointPage == -1 {
		e.tracer.Stop(common.MsgCompleteExportWebhooks, len(allWebhooks), repoSlug)
		return e.processWebhooks(repoSlug, allWebhooks)
	}

	for {
		webhooks, resp, err := e.github.Repositories.ListHooks(ctx, repoSlug, opts)
		if err != nil {
			e.tracer.LogError(common.ErrListWebhook, repoSlug, err)
			e.tracer.Stop(common.ErrListWebhooks, repoSlug, err)
			return types.WebhookData{}, err
		}
		allWebhooks = append(allWebhooks, webhooks...)

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allWebhooks)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointWebhooksDataSave, err)
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, resp.Page.Next)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointWebhooksPageSave, repoSlug, err)
		}

		if resp.Page.Next == 0 {
			break
		}
		opts.Page = resp.Page.Next
	}

	err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointWebhooksPageSave, repoSlug, err)
	}

	e.tracer.Stop(common.MsgCompleteExportWebhooks, len(allWebhooks), repoSlug)
	return e.processWebhooks(repoSlug, allWebhooks)
}

func (e *Export) processWebhooks(repoSlug string, webhooks []*scm.Hook) (types.WebhookData, error) {
	convertedHooks, notSupportedHooks := migrate.MapWebhooks(webhooks, mapEvents)
	// logs the not supported hooks
	err := migrate.LogNotSupportedWebhookEvents(repoSlug, notSupportedHooks, e.fileLogger)
	if err != nil {
		e.tracer.Stop(common.ErrListWebhooks, repoSlug, err)
		return types.WebhookData{}, fmt.Errorf("failed to log the not supported webhooks for repo %q: %w",
			repoSlug, err)
	}

	return types.WebhookData{
		ConvertedHooks: convertedHooks,
	}, nil
}

func mapEvents(triggers []string) ([]enum.WebhookTrigger, []string) {
	var events []enum.WebhookTrigger
	var notSupportedEvents []string

	for _, v := range triggers {
		switch v {
		case "create":
			events = append(events, enum.WebhookTriggerBranchCreated, enum.WebhookTriggerTagCreated)
		case "delete":
			events = append(events, enum.WebhookTriggerBranchDeleted, enum.WebhookTriggerTagDeleted)
		case "pull_request":
			events = append(events, enum.WebhookTriggerPullReqCreated, enum.WebhookTriggerPullReqReopened,
				enum.WebhookTriggerPullReqClosed, enum.WebhookTriggerPullReqUpdated, enum.WebhookTriggerPullReqMerged)
		case "pull_request_review_comment", "commit_comment":
			events = append(events, enum.WebhookTriggerPullReqCommentCreated)
		case "push":
			events = append(events, enum.WebhookTriggerPullReqBranchUpdated, enum.WebhookTriggerBranchUpdated,
				enum.WebhookTriggerTagUpdated)
		default:
			notSupportedEvents = append(notSupportedEvents, v)
		}
	}

	return events, notSupportedEvents
}
