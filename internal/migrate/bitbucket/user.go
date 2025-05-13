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

package bitbucket

import (
	"context"
	"fmt"
	"strings"

	"github.com/harness/harness-migrate/internal/gitexporter"
)

const CheckpointKeyUsers = "users"

// Bitbucket in accordance with GPDR does not provide PII for users
// https://developer.atlassian.com/cloud/bitbucket/bitbucket-api-changes-gdpr/
func (e *Export) GetDefaultEmail(ctx context.Context, accID, displayName string) (string, error) {
	userData, ok := e.userMap[accID]
	if ok {
		return userData.Email, nil
	}

	normalizedDisplayName := strings.ReplaceAll(displayName, " ", "_")
	if normalizedDisplayName == "" {
		normalizedDisplayName = "user"
	}

	userData.Email = normalizedDisplayName + "." + accID + gitexporter.UnknownEmailSuffix
	if err := e.fileLogger.Log("no email found for user with account ID %s and display name %s using '%s' as fallback email",
		accID, displayName, userData.Email); err != nil {
		return "", fmt.Errorf("cannot log file for unknown email, error: %w", err)
	}

	e.userMap[accID] = userData
	if err := e.checkpointManager.SaveCheckpoint(CheckpointKeyUsers, e.userMap); err != nil {
		return "", fmt.Errorf("cannot get checkpoint, error: %w", err)
	}

	return userData.Email, nil
}
