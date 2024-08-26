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

package migrate

import (
	"fmt"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/types/enum"
)

func MapWebhooks(
	hooks []*scm.Hook,
	mapEvents func(events []string) ([]enum.WebhookTrigger, []string),
) ([]*scm.Hook, []*scm.Hook) {
	var convertedHooks []*scm.Hook
	var notSupportedHooks []*scm.Hook
	for _, hook := range hooks {
		convertedHook, notSupportedHook := mapWebhook(hook, mapEvents)
		if convertedHook != nil {
			convertedHooks = append(convertedHooks, convertedHook)
		}
		if notSupportedHook != nil {
			notSupportedHooks = append(notSupportedHooks, notSupportedHook)
		}
	}

	return convertedHooks, notSupportedHooks
}

func LogNotSupportedWebhookEvents(repoSlug string, notSupportedHooks []*scm.Hook, logger gitexporter.Logger) error {
	if len(notSupportedHooks) == 0 {
		return nil
	}

	var logs []string
	var warningMsg string
	for _, hook := range notSupportedHooks {
		warningMsg = fmt.Sprintf("[%s] Skipped exporting some webhook events for repo %q with hook ID %q, Name %q, on events %v \n",
			enum.LogLevelWarning, repoSlug, hook.ID, hook.Name, hook.Events)
		logs = append(logs, warningMsg)
	}
	if err := logger.Log(strings.Join(logs, "")); err != nil {
		return fmt.Errorf("failed to log the not supported webhooks for repo %q: %w", repoSlug, err)
	}

	return nil
}

// mapWebhook converts the supported events and returns not supported events in a separate webhook
func mapWebhook(
	hook *scm.Hook,
	mapEvents func(events []string) ([]enum.WebhookTrigger, []string),
) (*scm.Hook, *scm.Hook) {
	events, notSupportedEvents := mapEvents(hook.Events)

	var convertedHook *scm.Hook
	var notSupportedHook *scm.Hook

	if len(events) != 0 {
		convertedHook = &scm.Hook{
			ID:     hook.ID,
			Name:   DisplayNameToIdentifier(hook.Name, "webhook", hook.ID),
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
