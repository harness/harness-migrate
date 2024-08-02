package stash

import (
	"context"
	"fmt"
	"strings"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/migrate"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/types/enum"

	"github.com/drone/go-scm/scm"
)

func (e *Export) ListWebhooks(
	ctx context.Context,
	repoSlug string,
	_ types.WebhookListOptions,
) (types.WebhookData, error) {
	e.tracer.Start(common.MsgStartExportWebhook, repoSlug)
	var allWebhooks []*scm.Hook
	opts := scm.ListOptions{Size: 25, Page: 1}

	for {
		webhooks, resp, err := e.stash.Repositories.ListHooks(ctx, repoSlug, opts)
		if err != nil {
			e.tracer.LogError(common.ErrWebhookList, repoSlug, err)
			e.tracer.Stop(common.ErrListWebhooks, repoSlug, err)
			return types.WebhookData{}, err
		}
		allWebhooks = append(allWebhooks, webhooks...)

		if resp.Page.Next == 0 {
			break
		}
		opts.Page = resp.Page.Next
	}

	convertedHooks, notSupportedHooks := migrate.MapWebhooks(allWebhooks, mapEvents)
	// logs the not supported hooks
	if len(notSupportedHooks) != 0 {
		var logs []string
		var warningMsg string
		for _, hook := range notSupportedHooks {
			warningMsg = fmt.Sprintf("[%s] Skipped exporting some webhook events for repo %q with hook ID %q, Name %q, on events %v \n",
				enum.LogLevelWarning, repoSlug, hook.ID, hook.Name, hook.Events)
			logs = append(logs, warningMsg)
		}
		if err := e.fileLogger.Log(strings.Join(logs, "")); err != nil {
			e.tracer.Stop(common.ErrListWebhooks, repoSlug, err)
			return types.WebhookData{}, fmt.Errorf("failed to log the not supported webhooks for repo %q: %w",
				repoSlug, err)
		}
	}

	e.tracer.Stop(common.MsgCompleteExportWebhooks, len(allWebhooks), repoSlug)
	return types.WebhookData{
		ConvertedHooks: convertedHooks,
	}, nil
}

func mapEvents(triggers []string) ([]enum.WebhookTrigger, []string) {
	var events []enum.WebhookTrigger
	var notSupportedEvents []string

	for _, v := range triggers {
		switch v {
		case "repo:refs_changed":
			events = append(events, enum.WebhookTriggerBranchCreated, enum.WebhookTriggerBranchDeleted, enum.WebhookTriggerBranchUpdated, enum.WebhookTriggerTagCreated, enum.WebhookTriggerTagDeleted, enum.WebhookTriggerTagUpdated)
		case "pr:opened":
			events = append(events, enum.WebhookTriggerPullReqCreated, enum.WebhookTriggerPullReqReopened)
		case "pr:merged":
			events = append(events, enum.WebhookTriggerPullReqMerged)
		case "pr:modified":
			events = append(events, enum.WebhookTriggerPullReqUpdated)
		case "pr:declined":
			events = append(events, enum.WebhookTriggerPullReqClosed)
		case "pr:from_ref_updated":
			events = append(events, enum.WebhookTriggerPullReqBranchUpdated)
		case "pr:comment:added":
			events = append(events, enum.WebhookTriggerPullReqCommentCreated)
		default:
			notSupportedEvents = append(notSupportedEvents, v)
		}
	}

	return events, notSupportedEvents
}
