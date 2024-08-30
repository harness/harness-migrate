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

func (m *Importer) ImportWebhooks(
	repoRef string,
	repoFolder string,
	tracer tracer.Tracer,
) error {
	tracer.Start(common.MsgStartImportWebhooks, repoRef)
	in, err := m.readWebhooks(repoFolder)
	if err != nil {
		tracer.Stop(common.ErrImportWebhooks, repoRef, err)
		return fmt.Errorf("failed to read webhooks from %q: %w", repoFolder, err)
	}

	if in == nil {
		m.Tracer.Stop(common.MsgCompleteImportWebhooks, 0, repoRef)
		return nil
	}

	if err := m.Harness.ImportWebhooks(repoRef, &types.WebhookInput{*in}); err != nil {
		tracer.Stop(common.ErrImportWebhooks, repoRef, err)
		return fmt.Errorf("failed to import webhooks for repo '%s' : %w",
			repoRef, err)
	}
	m.Tracer.Stop(common.MsgCompleteImportWebhooks, len(in.Hooks), repoRef)

	return nil
}

func (m *Importer) readWebhooks(repoFolder string) (*types.WebhookData, error) {
	webhookFile := filepath.Join(repoFolder, types.WebhookFileName)
	var hooks *types.WebhookData

	if _, err := os.Stat(webhookFile); os.IsNotExist(err) {
		return hooks, nil
	}

	data, err := ioutil.ReadFile(webhookFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q content from %q: %w", types.WebhookFileName, webhookFile, err)
	}

	if err := json.Unmarshal(data, &hooks); err != nil {
		return nil, fmt.Errorf("error parsing repo webhooks json: %w", err)
	}

	return hooks, nil
}
