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
	"io"
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

	u.tracer.Log("Extracting the zip file")

	if err := util.Unzip(u.zipFilePath, tempDir); err != nil {
		return fmt.Errorf("failed to extract zip file: %w", err)
	}

	if err := u.processDirectory(tempDir, mapping); err != nil {
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
	file, err := os.Open(u.userMappingPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var mapping UserMapping
	if err := json.Unmarshal(data, &mapping); err != nil {
		return nil, err
	}

	return mapping, nil
}

// processDirectory walks through the extracted files and updates user emails
func (u *Updater) processDirectory(rootDir string, mapping UserMapping) error {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to read root directory: %w", err)
	}

	var orgDirs []string
	// find the organization directory
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		potentialOrgDir := filepath.Join(rootDir, entry.Name())

		repoEntries, err := os.ReadDir(potentialOrgDir)
		if err != nil {
			u.tracer.Log("Error reading directory %s: %v", entry.Name(), err)
			continue
		}

		hasRepos := false
		for _, repoEntry := range repoEntries {
			if !repoEntry.IsDir() {
				continue
			}

			infoJsonPath := filepath.Join(potentialOrgDir, repoEntry.Name(), types.InfoFileName)
			if _, err := os.Stat(infoJsonPath); err == nil {
				hasRepos = true
				break
			}
		}

		if hasRepos {
			orgDirs = append(orgDirs, potentialOrgDir)
		}
	}

	if len(orgDirs) == 0 {
		return fmt.Errorf("no valid organization directories found in %s", rootDir)
	}

	if len(orgDirs) > 1 {
		return fmt.Errorf("multiple organization directories found in %s. it should be only one", rootDir)
	}

	orgDir := orgDirs[0]
	repoEntries, err := os.ReadDir(orgDir)
	if err != nil {
		return fmt.Errorf("failed to read org directory %s: %w", orgDir, err)
	}

	for _, repoEntry := range repoEntries {
		if !repoEntry.IsDir() || strings.HasPrefix(repoEntry.Name(), ".") {
			continue
		}

		repoDir := filepath.Join(orgDir, repoEntry.Name())
		if err := u.processPullRequests(repoDir, mapping); err != nil {
			return err
		}

		if err := u.processRules(repoDir, mapping); err != nil {
			return err
		}
	}

	return nil
}
