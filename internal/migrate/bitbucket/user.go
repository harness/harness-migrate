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
	"errors"
	"fmt"

	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/harness"
)

const CheckpointKeyUsers = "users"

func (e *Export) FindEmailByAccountID(ctx context.Context, accID string) (string, error) {
	userData, ok := e.userMap[accID]
	if ok {
		return userData.Email, nil
	}

	u, _, err := e.GetUserByAccountID(ctx, accID)
	if errors.Is(err, harness.ErrForbidden) {
		return "",
			fmt.Errorf("token must have 'user' scope to fetch workspace members emails: %w", harness.ErrForbidden)
	}

	if err != nil && !errors.Is(err, harness.ErrNotFound) {
		return "", err
	}

	if errors.Is(err, harness.ErrNotFound) || u == nil {
		return "", fmt.Errorf("couldn't find the user with account id %s: %w", accID, harness.ErrNotFound)
	}

	if u.Email == "" {
		userData.Email = getDefaultEmail(accID)
		if err := e.fileLogger.Log("no public email found for user %s using %s as fallback email", accID, userData.Email); err != nil {
			return "", fmt.Errorf("cannot log file for unknown email, error: %w", err)
		}
	}

	e.userMap[accID] = *u
	if err := e.checkpointManager.SaveCheckpoint(CheckpointKeyUsers, e.userMap); err != nil {
		return "", fmt.Errorf("cannot get checkpoint, error: %w", err)
	}

	return u.Email, nil
}

func getDefaultEmail(username string) string {
	return username + gitexporter.UnknownEmailSuffix
}
