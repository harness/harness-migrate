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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/util"
	"github.com/harness/harness-migrate/types"
)

// UserMapping represents the mapping between old and new email addresses
type UserMapping map[string]string

// Updater handles the process of updating user emails in the exported zip file
type Updater struct {
	userMappingPath string
	zipFilePath     string
	tracer          tracer.Tracer
}

// NewUpdater creates a new instance of the Updater
func NewUpdater(userMappingPath, zipFilePath string, tracer tracer.Tracer) *Updater {
	return &Updater{
		userMappingPath: userMappingPath,
		zipFilePath:     zipFilePath,
		tracer:          tracer,
	}
}

// Update performs the user email update process
func (u *Updater) Update(ctx context.Context) error {
	u.tracer.Log("Loading user mapping from %s", u.userMappingPath)

	mapping, err := u.loadUserMapping()
	if err != nil {
		return fmt.Errorf("failed to load user mapping: %w", err)
	}

	u.tracer.Log("Found %d user mappings", len(mapping))

	tempDir := "harness-updated"
	err = util.CreateFolder(tempDir)
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	u.tracer.Log("Extracting the zip file from %s", u.zipFilePath)

	if err := util.Unzip(u.zipFilePath, tempDir); err != nil {
		return fmt.Errorf("failed to extract zip file: %w", err)
	}

	if err := u.processFiles(tempDir, mapping); err != nil {
		return fmt.Errorf("failed to process files: %w", err)
	}

	updatedZipPath := u.zipFilePath + ".updated"
	if err := util.ZipFolder(tempDir, updatedZipPath); err != nil {
		return fmt.Errorf("failed to create updated zip file: %w", err)
	}

	if err := os.Rename(updatedZipPath, u.zipFilePath); err != nil {
		return fmt.Errorf("failed to replace original zip file: %w", err)
	}

	u.tracer.Log("Successfully updated user emails in %s", u.zipFilePath)
	return nil
}

// loadUserMapping loads the user mapping from the JSON file
func (u *Updater) loadUserMapping() (UserMapping, error) {
	data, err := os.ReadFile(u.userMappingPath)
	if err != nil {
		return nil, err
	}

	var mapping UserMapping
	if err := json.Unmarshal(data, &mapping); err != nil {
		return nil, err
	}

	return mapping, nil
}

// processFiles walks through the extracted files and updates user emails
func (u *Updater) processFiles(rootDir string, mapping UserMapping) error {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to read root directory: %w", err)
	}

	var orgDir string
	// find the organization directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(rootDir, entry.Name())
		if u.isValidOrgDir(dirPath) {
			if orgDir != "" {
				return fmt.Errorf("multiple organization directories found in %s", rootDir)
			}
			orgDir = dirPath
		}

		if orgDir == "" {
			return fmt.Errorf("no valid organization directories found in %s", rootDir)
		}
	}

	if err := u.processRepositories(orgDir, mapping); err != nil {
		return fmt.Errorf("failed to process repositories: %w", err)
	}

	// update the users.json file after processing all repositories
	if err := u.updateUsersJson(rootDir, mapping); err != nil {
		return fmt.Errorf("failed to update users.json: %w", err)
	}

	return nil
}

// isValidOrgDir checks if a directory contains at least one repo with an info.json
func (u *Updater) isValidOrgDir(path string) bool {
	repoEntries, err := os.ReadDir(path)
	if err != nil {
		u.tracer.Log("Error reading directory %s: %v", path, err)
		return false
	}

	for _, repo := range repoEntries {
		if repo.IsDir() {
			infoPath := filepath.Join(path, repo.Name(), types.InfoFileName)
			if _, err := os.Stat(infoPath); err == nil {
				return true
			}
		}
	}
	return false
}

// processRepositories handles pull requests and rules for each repo in the org
func (u *Updater) processRepositories(orgDir string, mapping UserMapping) error {
	repos, err := os.ReadDir(orgDir)
	if err != nil {
		return fmt.Errorf("failed to read org directory %s: %w", orgDir, err)
	}

	for _, repo := range repos {
		if !repo.IsDir() || strings.HasPrefix(repo.Name(), ".") {
			continue
		}

		repoDir := filepath.Join(orgDir, repo.Name())
		if err := u.processPullRequests(repoDir, mapping); err != nil {
			return err
		}
		if err := u.processRules(repoDir, mapping); err != nil {
			return err
		}
	}
	return nil
}

// updateUsersJson updates emails in users.json based on the provided mapping
func (u *Updater) updateUsersJson(rootDir string, mapping UserMapping) error {
	usersJsonPath := filepath.Join(rootDir, types.UsersFileName)
	if _, err := os.Stat(usersJsonPath); err != nil {
		u.tracer.Log("No users.json file found at the root directory")
		return nil
	}

	u.tracer.Log("Updating users.json file at %s", usersJsonPath)

	usersFileData, err := os.ReadFile(usersJsonPath)
	if err != nil {
		return fmt.Errorf("error reading users.json file: %w", err)
	}

	var usersInput types.CheckUsersInput
	err = json.Unmarshal(usersFileData, &usersInput)
	if err != nil {
		return fmt.Errorf("error unmarshalling provided user mapping file : %w. \nExpected format is {\"emailoriginal\": \"emailnew\", ...}", err)
	}

	updatedEmails := make([]string, 0, len(usersInput.Emails))
	for _, email := range usersInput.Emails {
		if newEmail, exists := mapping[email]; exists {
			updatedEmails = append(updatedEmails, newEmail)
		} else {
			updatedEmails = append(updatedEmails, email)
		}
	}
	updatedUsersInput := types.CheckUsersInput{
		Emails: updatedEmails,
	}

	updatedUsersJson, err := json.MarshalIndent(updatedUsersInput, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling updated users.json: %w", err)
	}

	if err := os.WriteFile(usersJsonPath, updatedUsersJson, 0644); err != nil {
		return fmt.Errorf("error writing updated users.json: %w", err)
	}

	u.tracer.Log("Successfully updated users.json file")

	return nil
}
