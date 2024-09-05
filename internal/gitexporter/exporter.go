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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/util"
	externalTypes "github.com/harness/harness-migrate/types"
	"golang.org/x/sync/errgroup"
)

const (
	maxChunkSize       = 25 * 1024 * 1024 // 25 MB
	prFileName         = "pr%d.json"
	zipFileName        = "harness.zip"
	maxParallelism     = 20 // TODO: make this configurable by the user
	UnknownEmailSuffix = "@unknownemail.harness.io"
)

type (
	Exporter struct {
		exporter    Interface
		zipLocation string
		ScmLogin    string
		ScmToken    string

		Tracer tracer.Tracer
		Report map[string]*report.Report

		flags Flags
	}

	Flags struct {
		NoPR      bool // to not export pull requests and comments
		NoWebhook bool // to not export webhooks
		NoRule    bool // to not export branch protection rules
		NoComment bool // to not export pull request comments
	}
)

func NewExporter(
	exporter Interface,
	location string,
	scmLogin string,
	scmToken string,
	tracer tracer.Tracer,
	report map[string]*report.Report,
	flags Flags,
) Exporter {
	return Exporter{
		exporter:    exporter,
		zipLocation: location,
		ScmLogin:    scmLogin,
		ScmToken:    scmToken,
		Tracer:      tracer,
		Report:      report,
		flags:       flags,
	}
}

// Export calls exporter methods in order and serialize an object for import.
func (e *Exporter) Export(ctx context.Context) error {
	path := filepath.Join(".", e.zipLocation)
	err := util.CreateFolder(path)
	if err != nil {
		return fmt.Errorf(common.ErrCannotCreateFolder, err)
	}

	e.Tracer.Log(common.MsgStartExport)

	data, err := e.getData(ctx, path)
	if err != nil {
		return fmt.Errorf(common.ErrFetchingFileData, err)
	}

	users := make(map[string]bool)
	for _, repo := range data {
		err = e.writeJsonForRepo(mapRepoData(repo), path)
		if err != nil {
			return fmt.Errorf(common.ErrWritingFileData, err)
		}
		usersForRepo := extractUsers(repo, users)
		e.reportUserMetrics(repo.Repository.RepoSlug, usersForRepo)
	}

	err = e.writeUsersJson(users)
	if err != nil {
		return fmt.Errorf("error writing users json: %w", err)
	}

	err = checkpoint.CleanupCheckpoint(path)
	if err != nil {
		log.Printf("error cleaning checkpoint: %v", err)
	}

	err = zipFolder(path)
	if err != nil {
		return fmt.Errorf("zipping error: %v", err)
	}

	e.Tracer.Log(common.MsgCompleteExport, len(data))

	err = deleteFolders(path)
	if err != nil {
		log.Printf("error cleaning up folder: %v", err)
	}

	err = deleteFiles(path)
	if err != nil {
		log.Printf("error cleaning up files: %v", err)
	}

	publishReport(e.Report)
	return nil
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

	err = e.writeBranchRules(repo, pathRepo)
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
	pathPR := filepath.Join(pathRepo, externalTypes.PullRequestDir)
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

func (e *Exporter) writeBranchRules(repo *externalTypes.RepositoryData, pathRepo string) error {
	if len(repo.BranchRules) == 0 {
		return nil
	}
	rulesJson, err := util.GetJson(repo.BranchRules)
	if err != nil {
		return fmt.Errorf("cannot serialize branch rules into json: %w", err)
	}
	err = util.WriteFile(filepath.Join(pathRepo, externalTypes.BranchRulesFileName), rulesJson)
	if err != nil {
		return fmt.Errorf("couldn't write branch rules into a file: %w", err)
	}
	return nil
}

func (e *Exporter) writeWebhooks(repo *externalTypes.RepositoryData, pathRepo string) error {
	if len(repo.Webhooks.Hooks) == 0 {
		return nil
	}
	hooksJson, err := util.GetJson(repo.Webhooks)
	if err != nil {
		log.Printf("cannot serialize into json: %v", err)
	}
	err = util.WriteFile(filepath.Join(pathRepo, externalTypes.WebhookFileName), hooksJson)
	if err != nil {
		return fmt.Errorf("error writing webhook json: %w", err)
	}
	return nil
}

func (e *Exporter) writeUsersJson(usersMap map[string]bool) error {
	if len(usersMap) == 0 {
		return nil
	}
	var users []string

	for user := range usersMap {
		users = append(users, user)
	}
	usersInput := externalTypes.CheckUsersInput{
		Emails: users,
	}
	usersJson, err := util.GetJson(usersInput)
	if err != nil {
		log.Printf("cannot serialize into json: %v", err)
	}
	err = util.WriteFile(filepath.Join(e.zipLocation, externalTypes.UsersFileName), usersJson)
	if err != nil {
		return fmt.Errorf("couldn't write users into a file: %w", err)
	}
	return nil
}

func (e *Exporter) getData(ctx context.Context, path string) ([]*types.RepoData, error) {
	repoData := make([]*types.RepoData, 0)

	// 1. list all the repos for the given org
	repositories, err := e.exporter.ListRepositories(ctx, types.ListOptions{Page: 1, Size: 25})
	if err != nil {
		return nil, fmt.Errorf("cannot list repositories: %w", err)
	}

	for _, repository := range repositories {
		data := &types.RepoData{Repository: repository}
		repoData = append(repoData, data)
		e.Report[repository.RepoSlug] = report.Init(repository.RepoSlug)
	}

	// 2. list pr per repo
	for i, repo := range repositories {
		repoPath := filepath.Join(path, repo.RepoSlug)
		err := util.CreateFolder(repoPath)
		if err != nil {
			return nil, fmt.Errorf(common.ErrWritingFileData, err)
		}

		isEmpty, err := e.CloneRepository(ctx, repo.Repository, repoPath, repo.RepoSlug, e.exporter.PullRequestRefs(), e.Tracer)
		if err != nil {
			return nil, fmt.Errorf("cannot clone the git repo for %s: %w", repo.RepoSlug, err)
		}

		if isEmpty {
			repoData[i].Repository.IsEmpty = true
			continue
		}

		// 3. get all webhooks for each repo
		if !e.flags.NoWebhook {
			webhooks, err := e.exporter.ListWebhooks(ctx, repo.RepoSlug, types.WebhookListOptions{})
			if err != nil {
				return nil, fmt.Errorf("encountered error in getting webhooks: %v", err)
			}
			repoData[i].Webhooks = webhooks
			e.Report[repo.RepoSlug].ReportMetric(ReportTypeWebhooks, len(webhooks.ConvertedHooks))
		}

		// 4. get all branch rules for each repo
		if !e.flags.NoRule {
			branchRules, err := e.exporter.ListBranchRules(ctx, repo.RepoSlug, types.ListOptions{Page: 1, Size: 25})
			if err != nil {
				return nil, fmt.Errorf("encountered error in getting branch rules: %w", err)
			}
			repoData[i].BranchRules = branchRules
			e.Report[repo.RepoSlug].ReportMetric(ReportTypeBranchRules, len(branchRules))
		}

		// 5. get all data for each pr
		if !e.flags.NoPR {
			prs, err := e.exporter.ListPullRequests(ctx, repo.RepoSlug,
				types.PullRequestListOptions{Page: 1, Size: 25, Open: true, Closed: true})
			if err != nil {
				return nil, fmt.Errorf("encountered error in getting pr: %w", err)
			}
			e.Report[repo.RepoSlug].ReportMetric(ReportTypePRs, len(prs))

			if e.flags.NoComment {
				pullreqData := make([]*types.PullRequestData, len(prs))
				for j := range prs {
					pullreqData[j] = &types.PullRequestData{
						PullRequest: prs[j],
					}
				}
				repoData[i].PullRequestData = pullreqData
				continue
			}

			prData, err := e.exportCommentsForPRs(ctx, prs, repo, e.Tracer)
			if err != nil {
				return nil, fmt.Errorf("error getting comments for pr: %w", err)
			}
			repoData[i].PullRequestData = prData
		}
	}

	return repoData, nil
}

func (e *Exporter) exportCommentsForPRs(
	ctx context.Context,
	prs []types.PRResponse,
	repo types.RepoResponse,
	t tracer.Tracer,
) ([]*types.PullRequestData, error) {
	e.Tracer.Start(common.MsgStartCommentsFetch, repo.RepoSlug)
	defer e.Tracer.Stop(common.MsgCompleteCommentsFetch, repo.RepoSlug)
	taskPool := util.NewTaskPool(ctx, maxParallelism)
	err := taskPool.Start()
	if err != nil {
		return nil, fmt.Errorf("error starting thread pool: %w", err)
	}

	prData := make([]*types.PullRequestData, len(prs))
	g := e.startResultChannel(ctx, taskPool, prData, t)

	for j, pr := range prs {
		task := e.createPRTask(j, pr, repo)
		taskPool.Submit(task)
	}

	taskPool.Shutdown()
	err = g.Wait()
	if err != nil {
		return nil, err
	}

	return prData, nil
}

func (e *Exporter) createPRTask(j int, pr types.PRResponse, repo types.RepoResponse) *util.Task {
	return &util.Task{
		ID: j,
		Execute: func(ctx context.Context) (any, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				comments, err := e.exporter.ListPullRequestComments(ctx, repo.RepoSlug, pr.Number,
					types.ListOptions{Page: 1, Size: 25})
				if err != nil {
					return nil, fmt.Errorf("encountered error in getting comments for PR %d: %w",
						pr.Number, err)
				}
				return mapPRData(pr, comments), nil
			}
		},
	}
}

func (e *Exporter) startResultChannel(
	ctx context.Context,
	taskPool *util.TaskPool,
	prData []*types.PullRequestData,
	tracer tracer.Tracer,
) *errgroup.Group {
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		for result := range taskPool.ResultCh {
			if result.Err != nil {
				tracer.LogError(common.ErrGettingComments, result.Err)
				taskPool.ForceShutdown()
				//todo: handle error better
				os.Exit(1)
			}
			if result.Data != nil {
				prData[result.ID] = result.Data.(*types.PullRequestData)
			}
			taskPool.MarkResultRead()
		}
		return nil
	})
	return g
}

func (e *Exporter) reportUserMetrics(repo string, users map[string]bool) {
	unknownEmailsCount := 0
	for user := range users {
		if strings.HasSuffix(user, UnknownEmailSuffix) {
			e.Report[repo].ReportError(ReportTypeUsers, user, "User mapped to new email")
			unknownEmailsCount++
		}
	}
	e.Report[repo].ReportMetric(ReportTypeUsers, len(users)-unknownEmailsCount)
}

func extractUsers(repo *types.RepoData, users map[string]bool) map[string]bool {
	repoUsers := make(map[string]bool)

	for _, prData := range repo.PullRequestData {
		users[prData.PullRequest.PullRequest.Author.Email] = true
		repoUsers[prData.PullRequest.PullRequest.Author.Email] = true
		for _, comment := range prData.Comments {
			if comment.Author.Email != "" {
				users[comment.Author.Email] = true
				repoUsers[comment.Author.Email] = true
			}
		}
		if prData.PullRequest.Author.Email != "" {
			users[prData.PullRequest.Author.Email] = true
			repoUsers[prData.PullRequest.Author.Email] = true
		}
	}

	for _, rule := range repo.BranchRules {
		for _, email := range rule.Definition.Bypass.UserEmails {
			users[email] = true
			repoUsers[email] = true
		}
	}
	return repoUsers
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

func mapPRData(pr types.PRResponse, comments []*types.PRComment) *types.PullRequestData {
	return &types.PullRequestData{
		PullRequest: pr,
		Comments:    comments,
	}
}

func mapRepoData(repoData *types.RepoData) *externalTypes.RepositoryData {
	d := new(externalTypes.RepositoryData)
	d.Repository.Slug = repoData.Repository.RepoSlug
	d.Repository = mapRepository(repoData.Repository)
	d.BranchRules = mapBranchRules(repoData.BranchRules)

	d.PullRequestData = make([]*externalTypes.PullRequestData, len(repoData.PullRequestData))
	for i, prData := range repoData.PullRequestData {
		d.PullRequestData[i] = new(externalTypes.PullRequestData)
		d.PullRequestData[i].PullRequest = mapPR(prData.PullRequest.PullRequest)
		d.PullRequestData[i].Comments = mapPRComment(prData.Comments)
	}

	d.Webhooks.Hooks = mapHooks(repoData.Webhooks.ConvertedHooks)

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

func deleteFiles(path string) error {
	// delete users.json skipping exporter json for now
	return os.Remove(filepath.Join(path, externalTypes.UsersFileName))
}
