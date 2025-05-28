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
	"os"
	"path/filepath"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/types"
)

func (m *Importer) ImportLabels(
	repoRef string,
	repoFolder string,
) error {
	m.Tracer.Start(common.MsgStartImportLabels, repoRef)
	in, err := m.readLabels(repoFolder)
	if err != nil {
		m.Tracer.Stop(common.ErrImportLabels, repoRef, err)
		return fmt.Errorf("failed to read labels from %q: %w", repoFolder, err)
	}

	if len(in) == 0 {
		m.Report[repoRef].ReportMetric(report.ReportTypeLabels, len(in))
		m.Tracer.Stop(common.MsgCompleteImportLabels, len(in), repoRef)
		return nil
	}

	if err := m.Harness.ImportLabels(repoRef, &types.LabelsInput{Labels: in}); err != nil {
		m.Tracer.Stop(common.ErrImportLabels, repoRef, err)
		return fmt.Errorf("failed to import labels for '%s' : %w",
			repoRef, err)
	}

	m.Report[repoRef].ReportMetric(report.ReportTypeLabels, len(in))
	m.Tracer.Stop(common.MsgCompleteImportLabels, len(in), repoRef)
	return nil
}

func (m *Importer) readLabels(repoFolder string) ([]*types.Label, error) {
	labelsFile := filepath.Join(repoFolder, types.LabelsFileName)
	var labels []*types.Label

	if _, err := os.Stat(labelsFile); os.IsNotExist(err) {
		return labels, nil
	}

	data, err := os.ReadFile(labelsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q content from %q: %w", types.LabelsFileName, labelsFile, err)
	}

	if err := json.Unmarshal(data, &labels); err != nil {
		return nil, fmt.Errorf("error parsing repo webhooks json: %w", err)
	}

	return labels, nil
}
