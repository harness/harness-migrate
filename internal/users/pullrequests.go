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

package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/harness/harness-migrate/internal/util"
	"github.com/harness/harness-migrate/types"
)

// processPullRequests updates user emails in pull request files
func (u *Updater) processPullRequests(repoDir string, mapping UserMapping) error {
	prDir := filepath.Join(repoDir, types.PullRequestDir)
	if _, err := os.Stat(prDir); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	u.tracer.Log("Processing PRs for %s", repoDir)

	// List all PR JSON files in the pr directory
	prFiles, err := filepath.Glob(filepath.Join(prDir, "pr[0-9]*.json"))
	if err != nil {
		u.tracer.Log("Error listing PR files in %s: %v", prDir, err)
		return fmt.Errorf("error listing PR files in %s: %w", prDir, err)
	}

	for _, prFile := range prFiles {
		if err := u.processPRFile(prFile, mapping); err != nil {
			u.tracer.Log("Error processing PR file %s: %v", prFile, err)
			continue
		}
	}

	return nil
}

// processPRFile processes a single PR JSON file and updates user emails
func (u *Updater) processPRFile(filePath string, mapping UserMapping) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open PR file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var prEntries []types.PullRequestData
	if err := decoder.Decode(&prEntries); err != nil {
		return fmt.Errorf("failed to decode PR JSON: %w", err)
	}

	updated := false
	var updatedPRCount int

	for i := range prEntries {
		entryUpdated := false

		if newEmail, exists := mapping[prEntries[i].PullRequest.Author.Email]; exists {
			prEntries[i].PullRequest.Author.Email = newEmail
			entryUpdated = true
		}

		for j := range prEntries[i].Comments {
			if newEmail, exists := mapping[prEntries[i].Comments[j].Author.Email]; exists {
				prEntries[i].Comments[j].Author.Email = newEmail
				entryUpdated = true
			}
		}

		// count each PR only once, regardless of how many emails were updated in it
		if entryUpdated {
			updatedPRCount++
		}

		updated = updated || entryUpdated
	}

	if updated {
		prJson, err := util.GetJson(prEntries)
		if err != nil {
			return fmt.Errorf("cannot serialize pull requests data into json: %w", err)
		}

		err = util.WriteFile(filePath, prJson)
		if err != nil {
			u.tracer.Log("Error updating the PR file %s: %v", filePath, err)
			return fmt.Errorf("failed to write updated PR file: %w", err)
		}

		u.tracer.Log("Updated user emails in %d PR entries.", updatedPRCount)
	}

	return nil
}
