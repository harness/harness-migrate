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
	"errors"
	"fmt"
	"os"
	filepath "path/filepath"
	"regexp"
	"strings"

	"github.com/harness/harness-migrate/internal/command"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types/enum"
	"github.com/harness/harness-migrate/internal/util"
	"github.com/harness/harness-migrate/types"
)

var (
	ErrAbortMigration = errors.New("aborting the migration. please checkout your command and try again")
	ErrInvalidRepoDir = errors.New("directory doesn't contain repo metadata")
)

// Importer imports data from gitlab to Harness.
type Importer struct {
	Harness harness.Client

	HarnessSpace    string
	HarnessRepo     string // optional for single repo import
	HarnessToken    string
	Endpoint        string
	ZipFileLocation string

	Gitness bool
	Tracer  tracer.Tracer
	Report  map[string]*report.Report

	RequestId string
	flags     Flags
}

type Flags struct {
	SkipUsers     bool
	FileSizeLimit int64
	NoPR          bool
	NoWebhook     bool
	NoRule        bool
	NoLabel       bool
	NoGit         bool // for incremental migration - skip git operations
	PRBatchSize   int  // batch size for PR imports to avoid 413 errors (default: 100)
}

func NewImporter(
	baseURL,
	space,
	repo,
	token,
	location,
	requestId string,
	gitness,
	trace bool,
	flags Flags,
	tracer tracer.Tracer,
	report map[string]*report.Report,
) *Importer {
	spaceParts := strings.Split(space, "/")

	client := harness.New(spaceParts[0], token, harness.WithAddress(baseURL), harness.WithTracing(trace))

	if gitness {
		client = harness.NewGitness(token, baseURL, harness.WithTracing(trace))
	}

	return &Importer{
		Harness:         client,
		HarnessSpace:    space,
		HarnessRepo:     repo,
		HarnessToken:    token,
		Tracer:          tracer,
		Report:          report,
		RequestId:       requestId,
		Endpoint:        baseURL,
		Gitness:         gitness,
		ZipFileLocation: location,
		flags:           flags,
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

	importedRepos := 0
	for _, f := range folders {
		repository, err := m.ReadRepoInfo(f)
		if errors.Is(err, ErrInvalidRepoDir) {
			continue
		}
		if err != nil {
			m.Tracer.LogError("failed to read repo info from %q: %s", f, err.Error())
			continue
		}

		repoRef := util.JoinPaths(m.HarnessSpace, repository.Name)

		m.Report[repoRef] = report.Init(repoRef)
		m.reportSkippedMetadata(m.Report[repoRef])

		if !m.flags.NoGit {
			if err := m.createRepoAndDoPush(ctx, f, &repository); err != nil {
				m.Tracer.LogError("failed to create or push git data for %q: %s", repoRef, err.Error())
				if !errors.Is(err, harness.ErrDuplicate) {
					m.cleanup(repoRef)
				}
				if notRecoverableError(err) {
					return ErrAbortMigration
				}

				continue
			}
			// update the repo state to migrate data import
			_, err = m.Harness.UpdateRepositoryState(
				repoRef,
				&harness.UpdateRepositoryStateInput{
					State: enum.RepoStateMigrateDataImport,
				},
			)
			if err != nil {
				return fmt.Errorf("failed to update the repo state to %s: %w", enum.RepoStateMigrateDataImport, err)
			}
		}

		if !repository.IsEmpty {
			err := m.importRepoMetaDataWithOffset(ctx, repoRef, f)
			if err != nil {
				m.Tracer.LogError("failed to import repo meta data for %q: %s", repoRef, err.Error())
				if !m.flags.NoGit {
					// For full migration, best effort delete the repo on server
					m.cleanup(repoRef)
				}

				if notRecoverableError(err) {
					return ErrAbortMigration
				}
				continue
			}
		}

		// update the repo state to active
		_, err = m.Harness.UpdateRepositoryState(
			repoRef,
			&harness.UpdateRepositoryStateInput{State: enum.RepoStateActive, Force: true},
		)
		if err != nil {
			return fmt.Errorf("failed to update the repo state to %s: %w", enum.RepoStateActive, err)
		}

		importedRepos++
	}

	report.PublishReports(m.Report)
	m.Tracer.Log(common.MsgCompleteImport, importedRepos)
	return nil
}

func (m *Importer) checkUsers(unzipLocation string) error {
	if m.flags.SkipUsers {
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
	if len(unknownUsers) != 0 {
		return fmt.Errorf("users not present in system: %v", unknownUsers)
	}

	return nil
}

func (m *Importer) createRepoAndDoPush(ctx context.Context, repoFolder string, repo *types.Repository) error {
	hRepo, err := m.CreateRepo(repo, m.HarnessSpace, m.Tracer)
	if err != nil {
		return fmt.Errorf("failed to create repo: %w", err)
	}

	if repo.IsEmpty {
		return nil
	}

	repoRef := util.JoinPaths(m.HarnessSpace, repo.Name)
	originalLimit, err := m.getFileSizeLimit(repoRef, m.Tracer)
	if err != nil {
		return fmt.Errorf("failed to get repo file size limit: %w", err)
	}

	// update the file-size-limit as push might get declined by the pre-receive hook on server due to large file sizes.
	if originalLimit < m.flags.FileSizeLimit || repo.GitLFSDisabled {
		err := m.updateRepoSetting(repoRef, m.flags.FileSizeLimit, !repo.GitLFSDisabled, m.Tracer)
		if err != nil {
			return fmt.Errorf("failed to set file size limit on repo: %w", err)
		}
	}

	if repo.GitLFSDisabled {
		m.Report[repoRef].ReportSkipped(report.ReportTypeGitLFSObjects)
	}

	err = m.Push(ctx, repoRef, repoFolder, hRepo, repo.GitLFSDisabled, repo.LfsObjectCount, m.Tracer)
	if err != nil {
		return fmt.Errorf("failed to push to repo: %w", err)
	}

	// revert the file-size-limit to it's original value
	if originalLimit < m.flags.FileSizeLimit {
		m.Tracer.Log("Reverting the file-size-limit from %d to its original value %d.", m.flags.FileSizeLimit, originalLimit)
		err := m.updateRepoSetting(repoRef, originalLimit, !repo.GitLFSDisabled, m.Tracer)
		if err != nil {
			return fmt.Errorf("failed to set file size limit on repo: %w", err)
		}
	}

	return nil
}

func (m *Importer) importRepoMetaData(_ context.Context, repoRef, repoFolder string) error {
	if !m.flags.NoLabel {
		if err := m.ImportLabels(repoRef, repoFolder); err != nil {
			return fmt.Errorf("failed to import labels for '%s': %w", repoRef, err)
		}
	}

	if !m.flags.NoPR {
		if err := m.ImportPullRequests(repoRef, repoFolder); err != nil {
			return fmt.Errorf("failed to import pull requests and comments for repo '%s': %w", repoRef, err)
		}
	}

	if !m.flags.NoWebhook {
		if err := m.ImportWebhooks(repoRef, repoFolder); err != nil {
			return fmt.Errorf("failed to import webhooks for repo '%s': %w", repoRef, err)
		}
	}

	if !m.flags.NoRule {
		if err := m.ImportBranchRules(repoRef, repoFolder); err != nil {
			return fmt.Errorf("failed to import branch rules for repo '%s': %w", repoRef, err)
		}
	}

	return nil
}

// Cleanup cleans up the repo best effort.
func (m *Importer) cleanup(repoRef string) {
	m.Tracer.Start(common.MsgStartRepoCleanup, repoRef)
	err := m.Harness.DeleteRepository(repoRef)
	if err != nil {
		m.Tracer.LogError(common.ErrCleanupRepo, err)
		return
	}

	m.Tracer.Stop(common.MsgCompleteRepoCleanup, repoRef)
}

// notRecoverableError checks if error is not recoverable, otherwise migration can continue
func notRecoverableError(err error) bool {
	if errors.Is(err, harness.ErrForbidden) ||
		errors.Is(err, harness.ErrUnauthorized) ||
		errors.Is(err, harness.ErrNotFound) ||
		errors.Is(err, harness.ErrInvalidRef) {
		return true
	}

	return false
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

func (m *Importer) reportSkippedMetadata(reporter *report.Report) {
	reportTypesMap := map[string]bool{
		report.ReportTypeWebhooks:    m.flags.NoWebhook,
		report.ReportTypePRs:         m.flags.NoPR,
		report.ReportTypeBranchRules: m.flags.NoRule,
		report.ReportTypeLabels:      m.flags.NoLabel,
	}

	for reportType, isSkipped := range reportTypesMap {
		if isSkipped {
			reporter.ReportSkipped(reportType)
		}
	}
}

func (m *Importer) getOffsetForIncrementalMigration(
	ctx context.Context,
	repoRef, repoFolder string,
) (int, error) {
	m.Tracer.Log("Starting incremental migration for repository: %s", repoRef)

	handler := NewIncrementalMigrationHandler(m.Harness, m.Tracer, repoRef)

	err := handler.CheckRepositoryExists(ctx)
	if err != nil {
		return 0, err
	}

	prOffset, err := handler.GetPROffset(ctx)
	if err != nil {
		return 0, err
	}

	return prOffset, nil
}

// importRepoMetaDataWithOffset imports repository metadata with optional PR offset applied.
func (m *Importer) importRepoMetaDataWithOffset(
	ctx context.Context,
	repoRef, repoFolder string,
) error {
	if !m.flags.NoGit {
		return m.importRepoMetaData(ctx, repoRef, repoFolder)
	}

	// for incremental migration, ensure repo state is restored to active on any error.
	var importErr error
	defer func() {
		if importErr != nil {
			m.Tracer.Log("Error occurred during incremental migration, restoring repository state to active")
			_, restoreErr := m.Harness.UpdateRepositoryState(
				repoRef,
				&harness.UpdateRepositoryStateInput{
					State: enum.RepoStateActive,
					Force: true,
				},
			)
			if restoreErr != nil {
				m.Tracer.LogError("Failed to restore repository %s to active state: %v", repoRef, restoreErr)
			}
		}
	}()

	prOffset, err := m.getOffsetForIncrementalMigration(ctx, repoRef, repoFolder)
	if err != nil {
		importErr = fmt.Errorf("failed to handle incremental migration for %q: %s", repoRef, err.Error())
		return importErr
	}

	m.Tracer.Log("Importing metadata with PR offset: %d", prOffset)

	if err := m.applyPROffsetToRepoData(repoFolder, prOffset); err != nil {
		importErr = fmt.Errorf("failed to apply PR offset to repository data: %w", err)
		return importErr
	}

	// For incremental migration, push the PR references AFTER offset is applied to JSON data
	// but BEFORE importing metadata so the references exist when metadata is processed
	if err := m.pushIncrementalPRReferences(ctx, repoRef, repoFolder, prOffset); err != nil {
		importErr = fmt.Errorf("failed to push incremental PR references: %w", err)
		return importErr
	}

	if err := m.importRepoMetaData(ctx, repoRef, repoFolder); err != nil {
		importErr = err
		return importErr
	}

	return nil
}

func (m *Importer) applyPROffsetToRepoData(repoFolder string, offset int) error {
	if offset == 0 {
		return nil
	}

	repository, err := m.ReadRepoInfo(repoFolder)
	if err != nil {
		return err
	}

	prDir := filepath.Join(repoFolder, types.PullRequestDir)
	err = m.applyOffsetToPRFiles(prDir, offset)
	if err != nil {
		return fmt.Errorf("failed to apply offset to PR files: %w", err)
	}

	m.Tracer.Log("Successfully applied PR offset %d to repository data for %s", offset, repository.Name)
	return nil
}

func (m *Importer) applyOffsetToPRFiles(prDir string, offset int) error {
	pattern := regexp.MustCompile(`^pr\d+\.json$`)

	if _, err := os.Stat(prDir); os.IsNotExist(err) {
		m.Tracer.Log("No PR directory found, skipping PR offset application")
		return nil
	}

	fileEntries, err := os.ReadDir(prDir)
	if err != nil {
		return fmt.Errorf("failed to read %s directory: %w", types.PullRequestDir, err)
	}

	for _, entry := range fileEntries {
		if entry.IsDir() || !pattern.MatchString(entry.Name()) {
			continue
		}

		prFilePath := filepath.Join(prDir, entry.Name())
		err := m.applyOffsetToPRFile(prFilePath, offset)
		if err != nil {
			return fmt.Errorf("failed to apply offset to file %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func (m *Importer) applyOffsetToPRFile(filePath string, offset int) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	var prs []*types.PullRequestData
	if err := json.Unmarshal(data, &prs); err != nil {
		return fmt.Errorf("error parsing PR JSON: %w", err)
	}

	// apply offset to PR numbers
	for _, pr := range prs {
		if pr == nil {
			continue
		}

		if pr.PullRequest.Number != 0 {
			pr.PullRequest.Number += offset
		}
	}

	modifiedData, err := json.MarshalIndent(prs, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling PR data: %w", err)
	}

	if err := util.WriteFile(filePath, modifiedData); err != nil {
		return fmt.Errorf("failed to write modified PR data: %w", err)
	}

	return nil
}

func (m *Importer) pushIncrementalPRReferences(
	ctx context.Context,
	repoRef, repoFolder string,
	prOffset int,
) error {
	_, err := m.Harness.UpdateRepositoryState(
		repoRef,
		&harness.UpdateRepositoryStateInput{
			State: enum.RepoStateMigrateGitPush,
			Force: true, // Force the state transition for incremental migration
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update repository state to migrate-git-push: %w", err)
	}

	repo, err := m.Harness.GetRepository(repoRef)
	if err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}

	gitDir := filepath.Join(repoFolder, types.GitDir)
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		m.Tracer.Log("No git directory found in export, skipping PR reference push")
	} else {
		handler := NewIncrementalMigrationHandler(m.Harness, m.Tracer, repoRef)

		err = handler.UpdatePRReferences(ctx, gitDir, prOffset)
		if err != nil {
			return fmt.Errorf("failed to update PR references with offset: %w", err)
		}

		err = m.pushPRReferencesOnly(ctx, repoRef, gitDir, repo.GitURL)
		if err != nil {
			return fmt.Errorf("failed to push PR references: %w", err)
		}
	}

	_, err = m.Harness.UpdateRepositoryState(
		repoRef,
		&harness.UpdateRepositoryStateInput{
			State: enum.RepoStateMigrateDataImport,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update repository state back to migrate-data-import: %w", err)
	}

	m.Tracer.Log("Successfully pushed incremental PR references for: %s", repoRef)
	return nil
}

func (m *Importer) pushPRReferencesOnly(
	ctx context.Context,
	repoRef, gitPath, gitURL string,
) error {
	const remoteName = "harnessRemote"

	output, err := command.RunGitCommand(ctx, gitPath, []string{}, "remote", "add", remoteName, gitURL)
	if err != nil {
		output, err = command.RunGitCommand(ctx, gitPath, []string{}, "remote", "set-url", remoteName, gitURL)
		if err != nil {
			return fmt.Errorf("failed to set remote %q: %w, output: %s", gitURL, err, string(output))
		}
	}

	listOutput, err := command.RunGitCommand(ctx, gitPath, []string{},
		"for-each-ref", "--format=%(refname)", "refs/pullreq/*/head")
	if err != nil {
		return fmt.Errorf("failed to list PR references: %w", err)
	}

	refs := strings.Split(strings.TrimSpace(string(listOutput)), "\n")
	if len(refs) == 1 && refs[0] == "" {
		return nil
	}

	for _, ref := range refs {
		if ref == "" {
			continue
		}

		refSpec := ref + ":" + ref
		output, err = command.RunGitCommandWithAuth(ctx, gitPath,
			command.Credentials{Username: "git-importer", Password: m.HarnessToken},
			"push", remoteName, refSpec)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to push %s: %v, output: %s", ref, err, string(output))
			m.Tracer.LogError(errorMsg)
		}
	}

	return nil
}
