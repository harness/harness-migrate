package stash

import (
	"context"
	"fmt"
	"strings"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/types/enum"

	"github.com/drone/go-scm/scm"
)

func (e *Export) ListWebhooks(
	ctx context.Context,
	repoSlug string,
	logger gitexporter.Logger,
	_ types.WebhookListOptions,
) (types.WebhookData, error) {
	e.tracer.Start(common.MsgStartExportWebhook, repoSlug)

	var allWebhooks []*scm.Hook
	defer func() {
		e.tracer.Stop(common.MsgCompleteExportWebhooks, len(allWebhooks), repoSlug)
	}()

	opts := scm.ListOptions{Size: 25, Page: 1}

	for {
		webhooks, resp, err := e.stash.Repositories.ListHooks(ctx, repoSlug, opts)
		if err != nil {
			e.tracer.LogError(common.ErrWebhookList, repoSlug, err)
			return types.WebhookData{}, err
		}
		allWebhooks = append(allWebhooks, webhooks...)

		if resp.Page.Next == 0 {
			break
		}
		opts.Page = resp.Page.Next
	}

	convertedHooks, notSupportedHooks := mapWebhooks(allWebhooks)
	// logs the not supported hooks
	if len(notSupportedHooks) != 0 {
		var logs []string
		var warningMsg string
		for _, hook := range notSupportedHooks {
			warningMsg = fmt.Sprintf("[%s] Skipped repo %q's webhook with ID %q Name %q for Target URL %q on events %v \n",
				enum.LogLevelWarning, repoSlug, hook.ID, hook.Name, hook.Target, hook.Events)
			logs = append(logs, warningMsg)
		}
		if err := logger.Log([]byte(strings.Join(logs, ""))); err != nil {
			return types.WebhookData{}, fmt.Errorf("failed to log the not supported webhooks for repo %q: %w",
				repoSlug, err)
		}
	}

	return types.WebhookData{
		ConvertedHooks: convertedHooks,
	}, nil
}

func mapWebhooks(hooks []*scm.Hook) ([]*scm.Hook, []*scm.Hook) {
	var convertedHooks []*scm.Hook
	var notSupportedHooks []*scm.Hook
	for _, hook := range hooks {
		convertedHook, notSupportedHook := mapWebhook(hook)
		if convertedHook != nil {
			convertedHooks = append(convertedHooks, convertedHook)
		}
		if notSupportedHook != nil {
			notSupportedHooks = append(notSupportedHooks, notSupportedHook)
		}
	}

	return convertedHooks, notSupportedHooks
}

// mapWebhook converts the supported events and returns not supported events in a separate webhook
func mapWebhook(hook *scm.Hook) (*scm.Hook, *scm.Hook) {
	events, notSupportedEvents := mapEvents(hook.Events)

	var convertedHook *scm.Hook
	var notSupportedHook *scm.Hook

	if len(events) != 0 {
		convertedHook = &scm.Hook{
			ID:     hook.ID,
			Name:   hook.Name,
			Target: hook.Target,
			Active: hook.Active,
			Events: enum.ToStringSlice(events),
		}
	}

	if len(notSupportedEvents) != 0 {
		notSupportedHook = &scm.Hook{
			ID:     hook.ID,
			Name:   hook.Name,
			Target: hook.Target,
			Active: hook.Active,
			Events: notSupportedEvents,
		}
	}
	return convertedHook, notSupportedHook
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
