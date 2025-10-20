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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/types"
)

const (
	// DefaultPRBatchSize is the default number of PRs to import in a single batch
	// This is to avoid 413 Payload Too Large errors
	DefaultPRBatchSize = 100
)

func (m *Importer) ImportPullRequests(
	repoRef string,
	repoFolder string,
) error {
	m.Tracer.Start(common.MsgStartImportPRs, repoRef)
	prDir := filepath.Join(repoFolder, types.PullRequestDir)
	in, err := m.readPRs(prDir)
	if err != nil {
		m.Tracer.Stop(common.ErrImportPRs, repoRef, err)
		return fmt.Errorf("failed to read pull requests and comments from %q: %w", prDir, err)
	}

	if len(in) == 0 {
		m.Report[repoRef].ReportMetric(report.ReportTypePRs, len(in))
		m.Tracer.Stop(common.MsgCompleteImportPRs, len(in), repoRef)
		return nil
	}

	batchSize := DefaultPRBatchSize
	if m.flags.PRBatchSize > 0 {
		batchSize = m.flags.PRBatchSize
	}

	if err := m.importPRsInBatches(context.Background(), repoRef, in, batchSize); err != nil {
		m.Tracer.Stop(common.ErrImportPRs, repoRef, err)
		return fmt.Errorf("failed to import pull requests and comments for repo '%s' : %w",
			repoRef, err)
	}

	m.Report[repoRef].ReportMetric(report.ReportTypePRs, len(in))
	m.Tracer.Stop(common.MsgCompleteImportPRs, len(in), repoRef)

	return nil
}

func (m *Importer) importPRsInBatches(
	ctx context.Context,
	repoRef string,
	prs []*types.PullRequestData,
	batchSize int,
) error {
	totalPRs := len(prs)

	if totalPRs <= batchSize {
		m.Tracer.Log("Importing %d pull requests for %s", totalPRs, repoRef)
		return m.Harness.ImportPRs(repoRef, &types.PRsImportInput{PullRequestData: prs})
	}

	var batches [][]*types.PullRequestData
	for i := 0; i < totalPRs; i += batchSize {
		end := i + batchSize
		if end > totalPRs {
			end = totalPRs
		}
		batches = append(batches, prs[i:end])
	}

	totalBatches := len(batches)
	m.Tracer.Log("Importing %d pull requests for %s in %d batches of %d",
		totalPRs, repoRef, totalBatches, batchSize)

	// Process batches sequentially
	for i, batch := range batches {
		select {
		case <-ctx.Done():
			return fmt.Errorf("import cancelled: %w", ctx.Err())
		default:
		}

		err := m.Harness.ImportPRs(repoRef, &types.PRsImportInput{PullRequestData: batch})
		if err != nil {
			m.Tracer.LogError("Failed to import PR batch %d/%d: %v", i+1, totalBatches, err)
			return fmt.Errorf("failed to import batch %d/%d: %w", i+1, totalBatches, err)
		}
	}

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
