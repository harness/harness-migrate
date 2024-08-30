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

package gitimporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/types"
)

func (m *Importer) ImportBranchRules(
	repoRef string,
	repoFolder string,
	tracer tracer.Tracer,
) error {
	tracer.Start(common.MsgStartImportBranchRules, repoRef)
	in, err := m.readBranchRules(repoFolder)
	if err != nil {
		tracer.Stop(common.ErrImportBranchRules, repoRef, err)
		return fmt.Errorf("failed to read branch rules from %q: %w", repoFolder, err)
	}

	if len(in) == 0 {
		m.Tracer.Stop(common.MsgCompleteImportBranchRules, 0, repoRef)
		return nil
	}

	rules, err := convertBranchRulesToRules(in)
	if err != nil {
		tracer.Stop(common.ErrImportBranchRules, repoRef, err)
		return fmt.Errorf("failed to convert branch rules for import: %w", err)
	}

	err = m.Harness.ImportRules(repoRef, &types.RulesInput{Rules: rules, Type: types.RuleTypeBranch})
	if err != nil {
		tracer.Stop(common.ErrImportBranchRules, repoRef, err)
		return fmt.Errorf("failed to import branch rules for repo '%s' : %w",
			repoRef, err)
	}
	m.Tracer.Stop(common.MsgCompleteImportBranchRules, len(rules), repoRef)

	return nil
}

func (m *Importer) readBranchRules(repoFolder string) ([]*types.BranchRule, error) {
	rules := make([]*types.BranchRule, 0)

	branchRulesFile := filepath.Join(repoFolder, types.BranchRulesFileName)
	if _, err := os.Stat(branchRulesFile); os.IsNotExist(err) {
		return rules, nil
	}

	data, err := ioutil.ReadFile(branchRulesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read content from %q: %w", branchRulesFile, err)
	}

	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("error parsing repo rules json: %w", err)
	}

	return rules, nil
}

func convertBranchRulesToRules(branchRules []*types.BranchRule) ([]*types.Rule, error) {
	rules := make([]*types.Rule, len(branchRules))

	for i, br := range branchRules {
		definitionJSON, err := json.Marshal(br.Definition)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal branch rule definition: %w", err)
		}

		patternJSON, err := json.Marshal(br.Pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal branch rule pattern: %w", err)
		}

		rules[i] = &types.Rule{
			ID:         br.ID,
			Identifier: br.Identifier,
			State:      br.State,
			Definition: definitionJSON,
			Pattern:    patternJSON,
			Created:    br.Created,
			Updated:    br.Updated,
		}
	}

	return rules, nil
}
