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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/util"
	"github.com/harness/harness-migrate/types"
)

// Importer imports data from gitlab to Harness.
type Importer struct {
	Harness harness.Client

	HarnessSpace string
	HarnessToken string
	Endpoint     string

	ZipFileLocation string
	SkipUsers       bool
	Gitness         bool

	Tracer tracer.Tracer

	RequestId string
}

func NewImporter(baseURL, space, token, location, requestId string, skipUsers, gitness, trace bool, tracer tracer.Tracer) *Importer {
	spaceSplit := strings.Split(space, "/")

	client := harness.New(spaceSplit[0], token, harness.WithAddress(baseURL), harness.WithTracing(trace))

	if gitness {
		client = harness.NewGitness(token, baseURL, harness.WithTracing(trace))
	}

	return &Importer{
		Harness:         client,
		HarnessSpace:    space,
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
	folders, err := getRepoBaseFolders(unzipLocation)
	if err != nil {
		return fmt.Errorf("cannot get repo folders in unzip: %w", err)
	}

	m.Tracer.Log("importing folders: %v", folders)

	// call git importer and other importers after this.
	var dupRepos []types.Repository
	for _, f := range folders {
		repo, err := m.ReadRepoInfo(f)
		if err != nil {
			return fmt.Errorf("failed to read repo infos: %w", err)
		}

		hRepo, err := m.CreateRepo(repo, m.HarnessSpace, m.Tracer)
		if err != nil && util.IsErrConflict(err) {
			m.Tracer.LogError("found a duplicate repo for %q: %w", repo.Slug, err)
			dupRepos = append(dupRepos, repo)
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to create repo %q: %w", repo.Slug, err)
		}

		err = m.Push(ctx, f, hRepo, m.Tracer)
		if err != nil {
			return fmt.Errorf("failed to push to repo: %w", err)
		}
	}

	return nil
}

func getRepoBaseFolders(directory string) ([]string, error) {
	var folders []string

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
