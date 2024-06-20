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

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/codeerror"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/util"
	externalTypes "github.com/harness/harness-migrate/types"
)

const (
	maxChunkSize = 25 * 1024 * 1024 // 25 MB
	prFileName   = "pr%d.json"
	zipFileName  = "harness.zip"
)

type Exporter struct {
	exporter    Interface
	zipLocation string
	ScmLogin    string
	ScmToken    string

	Tracer tracer.Tracer
}

func NewExporter(
	exporter Interface,
	location string,
	scmLogin string,
	scmToken string,
	tracer tracer.Tracer,
) Exporter {
	return Exporter{
		exporter:    exporter,
		zipLocation: location,
		ScmLogin:    scmLogin,
		ScmToken:    scmToken,
		Tracer:      tracer,
	}
}

// Export calls exporter methods in order and serialize an object for import.
func (e *Exporter) Export(ctx context.Context) {
	path := filepath.Join(".", e.zipLocation)
	err := util.CreateFolder(path)
	if err != nil {
		panic(fmt.Sprintf(common.PanicCannotCreateFolder, err))
	}
	data, err := e.getData(ctx, path)
	if err != nil {
		panic(fmt.Sprintf(common.PanicFetchingFileData, err))
	}
	for _, repo := range data {
		err = e.writeJsonForRepo(mapRepoData(repo), path)
		if err != nil {
			panic(fmt.Sprintf(common.PanicWritingFileData, err))
		}
	}

	err = checkpoint.CleanupCheckpoint(path)
	if err != nil {
		log.Printf("error cleaning checkpoint: %v", err)
	}

	err = zipFolder(path)
	if err != nil {
		log.Printf("zipping error: %v", err)
	}

	err = deleteFolders(path)
	if err != nil {
		log.Printf("error cleaning up folder: %v", err)
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

func (e *Exporter) writeJsonForRepo(repo *externalTypes.RepositoryData, path string) error {
	repoJson, _ := util.GetJson(repo.Repository)

	pathRepo := filepath.Join(path, repo.Repository.Slug)
	err := util.WriteFile(filepath.Join(pathRepo, externalTypes.InfoFileName), repoJson)
	if err != nil {
		return fmt.Errorf("error writing info file: %w", err)
	}

	err = e.writeWebhooks(repo, pathRepo)
	if err != nil {
		return fmt.Errorf("unable to write webhook: %w", err)
	}

	err = e.writeBranchRules(repo, err, pathRepo)
	if err != nil {
		return fmt.Errorf("cannot write branch rules: %w", err)
	}

	err = e.writePRs(repo, pathRepo)
	if err != nil {
		return fmt.Errorf("cannot write PRs: %w", err)
	}

	return nil
}

func (e *Exporter) writePRs(repo *externalTypes.RepositoryData, pathRepo string) error {
	if len(repo.PullRequestData) == 0 {
		return nil
	}

	prDataInSize := splitArray(repo.PullRequestData)
	pathPR := filepath.Join(pathRepo, externalTypes.PRDir)
	err := util.CreateFolder(pathPR)
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

func (e *Exporter) writeBranchRules(repo *externalTypes.RepositoryData, err error, pathRepo string) error {
	rulesJson, err := util.GetJson(repo.BranchRules)
	if err != nil {
		return fmt.Errorf("cannot serialize branch rules into json: %w", err)
	}
	if len(repo.BranchRules) != 0 {
		err = util.WriteFile(filepath.Join(pathRepo, externalTypes.BranchRulesFileName), rulesJson)
		if err != nil {
			return fmt.Errorf("couldn't write branch rules into a file: %w", err)
		}
	}
	return nil
}

func (e *Exporter) writeWebhooks(repo *externalTypes.RepositoryData, pathRepo string) error {
	if len(repo.Webhooks.Hooks) != 0 {
		hooksJson, err := util.GetJson(repo.Webhooks)
		if err != nil {
			log.Printf("cannot serialize into json: %v", err)
		}
		err = util.WriteFile(filepath.Join(pathRepo, externalTypes.WebhookFileName), hooksJson)
		if err != nil {
			return fmt.Errorf("error writing webhook json: %w", err)
		}
	}
	return nil
}

func (e *Exporter) getData(ctx context.Context, path string) ([]*types.RepoData, error) {
	repoData := make([]*types.RepoData, 0)
	var notSupportedErr *codeerror.OpNotSupportedError

	// 1. list all the repos for the given org
	repositories, err := e.exporter.ListRepositories(ctx, types.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot list repositories: %w", err)
	}

	for _, repository := range repositories {
		data := &types.RepoData{Repository: repository}
		repoData = append(repoData, data)
	}

	// 2. list pr per repo
	for i, repo := range repositories {
		repoPath := filepath.Join(path, repo.RepoSlug)
		err := util.CreateFolder(repoPath)
		if err != nil {
			panic(fmt.Sprintf(common.PanicWritingFileData, err))
		}

		gitRepo, err := e.CloneRepository(ctx, repo.Repository, repoPath, repo.RepoSlug, e.Tracer)
		if err != nil {
			return nil, fmt.Errorf("cannot clone the git repo for %s: %w", repo.RepoSlug, err)
		}

		err = e.exporter.FetchPullRequestRefs(ctx, gitRepo, repo.RepoSlug, e.ScmLogin, e.ScmToken)
		if err != nil {
			return nil, fmt.Errorf("cannot fetch the repo pull request references for %s: %w", repo.RepoSlug, err)
		}

		// 3. get all webhooks for each repo
		webhooks, err := e.exporter.ListWebhooks(ctx, repo.RepoSlug, e, types.WebhookListOptions{})
		if errors.As(err, &notSupportedErr) {
			return repoData, nil
		}
		if err != nil {
			log.Default().Printf("encountered error in getting webhooks: %w", err)
		}
		repoData[i].Webhooks = webhooks

		// 4. get all branch rules for each repo
		branchRules, err := e.exporter.ListBranchRules(ctx, repo.RepoSlug, types.ListOptions{Page: 1})
		if err != nil {
			return nil, fmt.Errorf("encountered error in getting branch rules: %w", err)
		}
		repoData[i].BranchRules = branchRules

		// 5. get all data for each pr
		prs, err := e.exporter.ListPullRequests(ctx, repo.RepoSlug, types.PullRequestListOptions{})
		if errors.As(err, &notSupportedErr) {
			return repoData, nil
		}
		if err != nil {
			return nil, fmt.Errorf("encountered error in getting pr: %w", err)
		}

		prData := make([]*types.PullRequestData, len(prs))
		for j, pr := range prs {
			comments, err := e.exporter.ListPullRequestComments(ctx, repo.RepoSlug, pr.Number, types.ListOptions{Page: 1})
			if err != nil {
				return nil, fmt.Errorf("encountered error in getting comments: %w", err)
			}
			prData[j] = mapPRData(pr, comments)
		}
		repoData[i].PullRequestData = prData
	}

	return repoData, nil
}

func mapPRData(pr types.PRResponse, comments []*types.PRComment) *types.PullRequestData {
	return &types.PullRequestData{
		PullRequest: pr,
		Comments:    comments,
	}
}

func mapRepoData(repoData *types.RepoData) *externalTypes.RepositoryData {
	d := new(externalTypes.RepositoryData)
	d.Repository.Slug = repoData.Repository.RepoSlug
	d.Repository.Repository = repoData.Repository.Repository
	d.BranchRules = MapBranchRules(repoData.BranchRules)

	d.PullRequestData = make([]*externalTypes.PullRequestData, len(repoData.PullRequestData))
	for i, prData := range repoData.PullRequestData {
		d.PullRequestData[i] = new(externalTypes.PullRequestData)
		d.PullRequestData[i].PullRequest = externalTypes.PR{
			PullRequest: prData.PullRequest.PullRequest,
		}
		d.PullRequestData[i].Comments = MapPRComment(prData.Comments)
	}

	d.Webhooks.Hooks = repoData.Webhooks.ConvertedHooks

	return d
}

func zipFolder(path string) error {
	err := util.ZipFolder(path, getZipFilePath(path))
	if err != nil {
		return fmt.Errorf("error zipping folder: %w", err)
	}
	return nil
}

func getZipFilePath(path string) string {
	return filepath.Join(path, zipFileName)
}

func deleteFolders(path string) error {
	err := util.DeleteDirsExcept(path, getZipFilePath(path))
	if err != nil {
		return fmt.Errorf("error cleaning up already zipped folders: %w", err)
	}

	return nil
}
