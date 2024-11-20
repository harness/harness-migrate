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

package gitlab

import (
	"context"
	"errors"
	"fmt"

	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/harness"
)

const CheckpointKeyUsers = "users"

func (e *Export) FindEmailByUsername(ctx context.Context, username string) (string, error) {
	user, ok := e.userMap[username]
	if ok {
		return user.PublicEmail, nil
	}

	u, _, err := e.GetUserByUserName(ctx, username)
	if err != nil && !errors.Is(err, harness.ErrNotFound) {
		return "", err
	}

	if errors.Is(err, harness.ErrNotFound) {
		if err := e.fileLogger.Log("no info found for user %s using fallback email", username); err != nil {
			return "", fmt.Errorf("cannot log file for unknown user, error: %w", err)
		}
		return getDefaultEmail(username), nil
	}

	email, err := e.FindEmailByUserID(ctx, u.ID)
	if err != nil && !errors.Is(err, harness.ErrNotFound) {
		return "", err
	}

	if errors.Is(err, harness.ErrNotFound) {
		if err := e.fileLogger.Log("no info found for user %s with id %d using fallback email", username, u.ID); err != nil {
			return "", fmt.Errorf("cannot log file for unknown user, error: %w", err)
		}
		return getDefaultEmail(username), nil
	}

	return email, nil
}

func (e *Export) FindEmailByUserID(ctx context.Context, id int) (string, error) {
	userData, _, err := e.GetUserByID(ctx, id)
	if err != nil && !errors.Is(err, harness.ErrNotFound) {
		return "", err
	}

	if errors.Is(err, harness.ErrNotFound) || userData == nil {
		return "", fmt.Errorf("couldn't find the user with id %d: %w", id, harness.ErrNotFound)
	}

	if userData.PublicEmail == "" {
		userData.PublicEmail = getDefaultEmail(userData.UserName)
		if err := e.fileLogger.Log("no public email found for user %s using %s as fallback email", userData.UserName, userData.PublicEmail); err != nil {
			return "", fmt.Errorf("cannot log file for unknown email, error: %w", err)
		}
	}

	e.userMap[userData.UserName] = *userData
	if err := e.checkpointManager.SaveCheckpoint(CheckpointKeyUsers, e.userMap); err != nil {
		return "", fmt.Errorf("cannot get checkpoint, error: %w", err)
	}

	return userData.PublicEmail, nil
}

func getDefaultEmail(username string) string {
	return username + gitexporter.UnknownEmailSuffix
}
