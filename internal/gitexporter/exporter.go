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

package gitexporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/harness/harness-migrate/internal/codeerror"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/util"

	externalTypes "github.com/harness/harness-migrate/types"
)

const (
	maxChunkSize = 25 * 1024 * 1024 // 25 MB
	prFileName   = "pr%d.json"
)

type Exporter struct {
	exporter    Interface
	zipLocation string
}

func NewExporter(exporter Interface, location string) Exporter {
	return Exporter{exporter: exporter, zipLocation: location}
}

// Export calls exporter methods in order and serialize an object for import.
func (e *Exporter) Export(ctx context.Context) {
	path := filepath.Join(".", e.zipLocation)
	err := util.CreateFolder(path)
	if err != nil {
		panic(fmt.Sprintf(common.PanicCannotCreateFolder, err))
	}
	data, _ := e.getData(ctx)
	for _, repo := range data {
		err = e.writeJsonForRepo(mapRepoData(repo))
		if err != nil {
			panic(fmt.Sprintf(common.PanicWritingFileData, err))
		}
	}
}

// Calculate the size of the struct in bytes
func calculateSize(s *externalTypes.PullRequestData) int {
	data, err := json.Marshal(s)
	// will never happen
	if err != nil {
		panic(err)
	}
	return len(data)
}

// Split the array into smaller chunks if the size exceeds the maxChunkSize
func splitArray(arr []*externalTypes.PullRequestData) [][]*externalTypes.PullRequestData {
	var chunks [][]*externalTypes.PullRequestData
	var currentChunk []*externalTypes.PullRequestData
	currentSize := 0

	for _, item := range arr {
		itemSize := calculateSize(item)
		if currentSize+itemSize > maxChunkSize {
			chunks = append(chunks, currentChunk)
			currentChunk = []*externalTypes.PullRequestData{}
			currentSize = 0
		}
		currentChunk = append(currentChunk, item)
		currentSize += itemSize
	}
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}
	return chunks
}

func (e *Exporter) writeJsonForRepo(repo *externalTypes.RepositoryData) error {
	repoJson, _ := util.GetJson(repo.Repository)
	pathRepo := filepath.Join(".", e.zipLocation, repo.Repository.Slug)
	err := util.CreateFolder(pathRepo)
	if err != nil {
		return fmt.Errorf("cannot create folder")
	}

	err = util.WriteFile(filepath.Join(pathRepo, externalTypes.InfoFileName), repoJson)
	if err != nil {
		return err
	}

	if len(repo.PullRequestData) == 0 {
		return nil
	}

	prDataInSize := splitArray(repo.PullRequestData)
	pathPR := filepath.Join(pathRepo, externalTypes.PRFolderName)
	err = util.CreateFolder(pathPR)
	if err != nil {
		return err
	}

	for i, data := range prDataInSize {
		prJson, err := util.GetJson(data)
		if err != nil {
			// todo: fix this
			log.Printf("cannot serialize into json: %v", err)
		}
		prFilePath := fmt.Sprintf(prFileName, i)

		err = util.WriteFile(filepath.Join(pathPR, prFilePath), prJson)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Exporter) getData(ctx context.Context) ([]*types.RepoData, error) {
	repoData := make([]*types.RepoData, 0)
	// 1. list all the repos for the given org
	repositories, err := e.exporter.ListRepositories(ctx, types.ListRepoOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot list repositories: %w", err)
	}

	for _, repository := range repositories {
		data := &types.RepoData{Repository: repository}
		repoData = append(repoData, data)
	}

	// 2. list pr per repo
	for i, repo := range repositories {
		prs, err := e.exporter.ListPullRequest(ctx, repo.RepoSlug, types.PullRequestListOptions{})
		var notSupportedErr *codeerror.OpNotSupportedError
		if errors.As(err, &notSupportedErr) {
			return repoData, nil
		}
		if err != nil {
			log.Default().Printf("encountered error in getting pr: %q", err)
		}

		// 3. get all data for each pr
		prData := make([]*types.PullRequestData, len(prs))

		for j, pr := range prs {
			// todo: implement comment
			prData[j] = mapPrData(pr, nil)
		}

		repoData[i].PullRequestData = prData
	}

	return repoData, nil
}

func mapPrData(pr types.PRResponse, comments []types.PRComments) *types.PullRequestData {
	return &types.PullRequestData{
		PullRequest: pr,
		Comments:    comments,
	}
}

func mapRepoData(repoData *types.RepoData) *externalTypes.RepositoryData {
	d := new(externalTypes.RepositoryData)
	d.Repository.Slug = repoData.Repository.RepoSlug
	d.Repository.Repository = repoData.Repository.Repository

	d.PullRequestData = make([]*externalTypes.PullRequestData, len(repoData.PullRequestData))
	for i, prData := range repoData.PullRequestData {
		d.PullRequestData[i] = &externalTypes.PullRequestData{}
		d.PullRequestData[i].PullRequest.PullRequest = prData.PullRequest.PullRequest
		// todo: map comment data
	}
	return d
}
