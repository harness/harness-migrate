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
	"fmt"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/types"

	"github.com/harness/harness-migrate/internal/tracer"
)

func (m *Importer) CreateRepo(
	repo *types.Repository,
	targetSpace string,
	tracer tracer.Tracer,
) (*harness.Repository, error) {
	// TODO: check license for size limit
	tracer.Start(common.MsgStartImportCreateRepo, repo.Name)
	in := &harness.CreateRepositoryForMigrateInput{
		Identifier:    repo.Name,
		DefaultBranch: repo.Branch,
		IsPublic:      !repo.Private,
		ParentRef:     targetSpace,
	}

	repoOut, err := m.Harness.CreateRepositoryForMigration(in)
	if err != nil {
		tracer.LogError(common.ErrCreateRepo, repo.Name, targetSpace, err)
		return nil, fmt.Errorf(common.ErrCreateRepo, repo.Name, targetSpace, err)
	}

	m.Tracer.Stop(common.MsgCompleteImportCreateRepo, repo.Name, targetSpace)
	return repoOut, nil
}

func (m *Importer) getFileSizeLimit(
	repoRef string,
	tracer tracer.Tracer,
) (int64, error) {
	tracer.Start(common.MsgStartGetRepoSetting, repoRef)

	settings, err := m.Harness.FindRepoSettings(repoRef)
	if err != nil {
		tracer.Stop("failed to find repository settings for %s", repoRef)
		return 0, fmt.Errorf("failed to find repo settings for %s: %w", repoRef, err)
	}

	m.Tracer.Stop(common.MsgCompleteGetRepoSetting, repoRef, *settings.FileSizeLimit)
	return *settings.FileSizeLimit, nil
}

func (m *Importer) setFileSizeLimit(
	repoRef string,
	size int64,
	tracer tracer.Tracer,
) error {
	tracer.Start(common.MsgStartUpdateRepoSize, repoRef, size)

	in := &harness.RepoSettings{
		FileSizeLimit: &size,
	}
	settings, err := m.Harness.UpdateRepoSettings(repoRef, in)
	if err != nil {
		tracer.Stop("failed to update repository settings for %s", repoRef)
		return fmt.Errorf("failed to update repo settings for %s: %w", repoRef, err)
	}

	m.Tracer.Stop(common.MsgCompleteGetRepoSetting, repoRef, *settings.FileSizeLimit)
	return nil
}
