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
	"regexp"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/types"
)

func (m *Importer) ImportPullRequests(
	repoRef string,
	repoFolder string,
	tracer tracer.Tracer,
) error {
	tracer.Start(common.MsgStartImportPRs, repoRef)
	prDir := filepath.Join(repoFolder, types.PullRequestDir)
	in, err := m.readPRs(prDir)
	if err != nil {
		tracer.Stop(common.ErrImportPRs, repoRef, err)
		return fmt.Errorf("failed to read pull requests and comments from %q: %w", prDir, err)
	}

	if len(in) == 0 {
		m.Tracer.Stop(common.MsgCompleteImportPRs, len(in), repoRef)
		return nil
	}

	if err := m.Harness.ImportPRs(repoRef, &types.PRsImportInput{in}); err != nil {
		tracer.Stop(common.ErrImportPRs, repoRef, err)
		return fmt.Errorf("failed to import pull requests and comments for repo '%s' : %w",
			repoRef, err)
	}
	m.Tracer.Stop(common.MsgCompleteImportPRs, len(in), repoRef)

	return nil
}

func (m *Importer) readPRs(prFolder string) ([]*types.PullRequestData, error) {
	pattern := regexp.MustCompile(`^pr\d+\.json$`)
	prOut := make([]*types.PullRequestData, 0)

	if _, err := os.Stat(prFolder); os.IsNotExist(err) {
		return prOut, nil
	}

	fileEntries, err := os.ReadDir(prFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s directory: %w", types.PullRequestDir, err)
	}

	for _, entry := range fileEntries {
		if entry.IsDir() || !pattern.MatchString(entry.Name()) {
			continue
		}

		prFile := entry.Name()
		data, err := ioutil.ReadFile(filepath.Join(prFolder, prFile))
		if err != nil {
			return nil, fmt.Errorf("failed to read %q content: %w", prFile, err)
		}

		var prs []*types.PullRequestData
		if err := json.Unmarshal(data, &prs); err != nil {
			return nil, fmt.Errorf("error parsing repo pull request json: %w", err)
		}

		prOut = append(prOut, prs...)
	}

	return prOut, nil
}
