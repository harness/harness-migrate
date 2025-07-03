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

// processRules updates user emails in branch rules files
func (u *Updater) processRules(repoDir string, mapping UserMapping) error {
	rulesFile := filepath.Join(repoDir, types.BranchRulesFileName)
	if _, err := os.Stat(rulesFile); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	updated, err := u.processRuleFile(rulesFile, mapping)
	if err != nil {
		u.tracer.Log("Error processing rules file %s: %v", rulesFile, err)
		return fmt.Errorf("error processing rules file %s: %w", rulesFile, err)
	}

	if updated {
		u.tracer.Log("Updated branch rules bypass user emails for repository: %s", repoDir)
	}

	return nil
}

// processRuleFile processes a single rule JSON file and updates user emails
func (u *Updater) processRuleFile(filePath string, mapping UserMapping) (bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read rule file: %w", err)
	}

	var rules []*types.BranchRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return false, fmt.Errorf("failed to parse rule JSON: %w", err)
	}

	updated := false
	for _, rule := range rules {
		if rule.Definition.Bypass.UserEmails == nil {
			continue
		}

		updatedEmails := make([]string, 0, len(rule.Definition.Bypass.UserEmails))

		for _, email := range rule.Definition.Bypass.UserEmails {
			if newEmail, exists := mapping[email]; exists {
				updated = true
				updatedEmails = append(updatedEmails, newEmail)
			} else {
				updatedEmails = append(updatedEmails, email)
			}
		}

		rule.Definition.Bypass.UserEmails = updatedEmails
	}

	if !updated {
		return false, nil
	}

	rulesJson, err := util.GetJson(rules)
	if err != nil {
		return false, fmt.Errorf("cannot serialize branch rules into json: %w", err)
	}
	if err := util.WriteFile(filePath, rulesJson); err != nil {
		return false, fmt.Errorf("failed to write updated rule file: %w", err)
	}

	return updated, nil
}
