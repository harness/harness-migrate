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

package github

import (
	"context"
	"fmt"

	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/types"
)

const (
	unknownUserId      = -1
	CheckpointKeyUsers = "users"
)

func (e *Export) FindEmailByUsername(ctx context.Context, username string) (string, error) {
	user, ok := e.userMap[username]
	if ok {
		return user.Email, nil
	}

	u, _, err := e.GetUserByUserName(ctx, username)
	if err != nil {
		return "", err
	}

	if u == nil {
		u = &types.User{
			ID:    unknownUserId,
			Email: getDefaultEmail(username),
		}

		if err := e.fileLogger.Log("no info found for user %s using %s as fallback email", username, u.Email); err != nil {
			return "", fmt.Errorf("cannot log file for unknown user, error: %w", err)
		}
	}

	if u.Email == "" {
		u.Email = getDefaultEmail(username)
		if err := e.fileLogger.Log("no primary email found for user %s using %s as fallback email", username, u.Email); err != nil {
			return "", fmt.Errorf("cannot log file for unknown email, error: %w", err)
		}
	}

	e.userMap[username] = *u
	if err := e.checkpointManager.SaveCheckpoint(CheckpointKeyUsers, e.userMap); err != nil {
		return "", fmt.Errorf("cannot get checkpoint, error: %w", err)
	}

	return u.Email, nil
}

func getDefaultEmail(username string) string {
	return username + gitexporter.UnknownEmailSuffix
}
