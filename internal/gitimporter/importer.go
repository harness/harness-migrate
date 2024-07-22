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
	"os"
	filepath "path/filepath"
	"strings"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types/enum"
	"github.com/harness/harness-migrate/internal/util"
	"github.com/harness/harness-migrate/types"
)

// Importer imports data from gitlab to Harness.
type Importer struct {
	Harness harness.Client

	HarnessSpace string
	HarnessRepo  string // optional for single repo import
	HarnessToken string
	Endpoint     string

	ZipFileLocation string
	SkipUsers       bool
	Gitness         bool

	Tracer tracer.Tracer

	RequestId string
}

func NewImporter(baseURL, space, repo, token, location, requestId string, skipUsers, gitness, trace bool, tracer tracer.Tracer) *Importer {
	spaceSplit := strings.Split(space, "/")

	client := harness.New(spaceSplit[0], token, harness.WithAddress(baseURL), harness.WithTracing(trace))

	if gitness {
		client = harness.NewGitness(token, baseURL, harness.WithTracing(trace))
	}

	return &Importer{
		Harness:         client,
		HarnessSpace:    space,
		HarnessRepo:     repo,
		HarnessToken:    token,
		Tracer:          tracer,
		RequestId:       requestId,
		Endpoint:        baseURL,
		Gitness:         gitness,
		ZipFileLocation: location,
		SkipUsers:       skipUsers,
	}
}

func (m *Importer) Import(ctx context.Context) error {
	unzipLocation := filepath.Dir(m.ZipFileLocation)
	err := util.Unzip(m.ZipFileLocation, unzipLocation)
	if err != nil {
		return fmt.Errorf("error unzipping: %w", err)
	}

	folders, err := getRepoBaseFolders(unzipLocation, m.HarnessRepo)
	if err != nil {
		return fmt.Errorf("cannot get repo folders in unzip: %w", err)
	}

	if len(folders) == 0 {
		m.Tracer.Log("no folder found for importing in zip")
		return nil
	}

	m.Tracer.Log(common.MsgStartImportFromFolders, folders)

	// call git importer and other importers after this.
	err = m.checkUsers(unzipLocation)
	if err != nil {
		return err
	}

	for _, f := range folders {
		repo, err := m.createRepoAndDoPush(ctx, f)
		if err != nil {
			return fmt.Errorf("failed to create or push git data: %w", err)
		}

		repoRef := util.JoinPaths(m.HarnessSpace, repo.Name)
		// update the repo state to migrate data import
		_, err = m.Harness.UpdateRepositoryState(
			repoRef,
			&harness.UpdateRepositoryStateInput{State: enum.RepoStateMigrateDataImport},
		)
		if err != nil {
			return fmt.Errorf("failed to update the repo state to %s: %w", enum.RepoStateMigrateDataImport, err)
		}

		if err := m.importRepoMetaData(ctx, repoRef, f); err != nil {
			return fmt.Errorf("failed to import repo metadata: %w", err)
		}

		// update the repo state to active
		_, err = m.Harness.UpdateRepositoryState(
			repoRef,
			&harness.UpdateRepositoryStateInput{State: enum.RepoStateActive},
		)
		if err != nil {
			return fmt.Errorf("failed to update the repo state to %s: %w", enum.RepoStateActive, err)
		}
	}

	m.Tracer.Log(common.MsgCompleteImport, len(folders))
	return nil
}

func (m *Importer) checkUsers(unzipLocation string) error {
	if m.SkipUsers {
		return nil
	}

	usersFile := filepath.Join(unzipLocation, types.UsersFileName)
	if _, err := os.Stat(usersFile); os.IsNotExist(err) {
		return nil
	}
	usersFileData, err := os.ReadFile(usersFile)
	if err != nil {
		return fmt.Errorf("error reading users file: %w", err)
	}

	in := types.CheckUsersInput{}
	err = json.Unmarshal(usersFileData, &in)
	if err != nil {
		return fmt.Errorf("error unmarshalling: %w", err)
	}

	unknownUsers, err := m.CheckUsers(in.Emails)
	if err != nil {
		return fmt.Errorf("error checking users: %w", err)
	}
	if len(unknownUsers) == 0 {
		return fmt.Errorf("users not present in system: %v", unknownUsers)
	}

	return nil
}

func (m *Importer) createRepoAndDoPush(ctx context.Context, repoFolder string) (*types.Repository, error) {
	repo, err := m.ReadRepoInfo(repoFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to read repo infos: %w", err)
	}

	hRepo, err := m.CreateRepo(repo, m.HarnessSpace, m.Tracer)
	if err != nil {
		return nil, fmt.Errorf("failed to create repo %q: %w", repo.Slug, err)
	}

	err = m.Push(ctx, repoFolder, hRepo, m.Tracer)
	if err != nil {
		return nil, fmt.Errorf("failed to push to repo: %w", err)
	}

	return &repo, nil
}

func (m *Importer) importRepoMetaData(_ context.Context, repoRef, repoFolder string) error {
	if err := m.ImportPullRequests(repoRef, repoFolder, m.Tracer); err != nil {
		return fmt.Errorf("failed to import pull requests and comments for repo '%s': %w", repoRef, err)
	}

	if err := m.ImportWebhooks(repoRef, repoFolder, m.Tracer); err != nil {
		return fmt.Errorf("failed to import webhooks for repo '%s': %w", repoRef, err)
	}

	if err := m.ImportBranchRules(repoRef, repoFolder, m.Tracer); err != nil {
		return fmt.Errorf("failed to import branch rules for repo '%s': %w", repoRef, err)
	}

	return nil
}

func getRepoBaseFolders(directory string, singleRepo string) ([]string, error) {
	var folders []string

	if singleRepo != "" {
		repoFolder := filepath.Join(directory, singleRepo)
		if _, err := os.Stat(repoFolder); os.IsNotExist(err) {
			return folders, fmt.Errorf("cannot find the folder for repo %q: %w", singleRepo, err)
		}
		return []string{repoFolder}, nil
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("cannot get folders: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs, err := os.ReadDir(filepath.Join(directory, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("cannot get folders inside org: %w", err)
			}
			for _, dir := range dirs {
				folders = append(folders, filepath.Join(directory, entry.Name(), dir.Name()))
			}
		}
	}

	return folders, nil
}
