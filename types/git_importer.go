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

// Package types provides common types for data export and import.

package types

import (
	"encoding/json"
	"time"
)

type (
	Response struct {
		Success      bool   `json:"success"`
		ErrorMessage string `json:"error_message"`
	}

	// CheckUsersInput is object for checking users existence during repo(s) import.
	CheckUsersInput struct {
		Emails []string `json:"emails"`
	}

	// CheckUsersOutput is output object for checking users existence.
	CheckUsersOutput struct {
		UnknownEmails []string `json:"unknown_emails"`
	}

	PRsImportInput struct {
		PullRequestData []*PullRequestData `json:"pull_request_data"`
	}

	WebhookInput struct {
		WebhookData
	}

	Rule struct {
		ID         int             `json:"id"`
		Identifier string          `json:"identifier"`
		State      string          `json:"state"`
		Definition json.RawMessage `json:"definition"`
		Pattern    json.RawMessage `json:"pattern"`
		Created    time.Time       `json:"created"`
		Updated    time.Time       `json:"updated"`
	}

	RulesInput struct {
		Type  RuleType `json:"type"`
		Rules []*Rule  `json:"rules"`
	}
)
