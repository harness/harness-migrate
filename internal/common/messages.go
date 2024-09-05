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

package common

const (
	MsgStartExport               = "Starting export repo(s)."
	MsgCompleteExport            = "Finished export repo(s). Total repos: %d"
	MsgStartExportPRs            = "Starting export pull requests for repository %s."
	MsgCompleteExportPRs         = "Finished export %d pull requests for repository %s."
	MsgCheckpointLoadPr          = "Finished export %d pull requests for repository %s from checkpoint."
	MsgStartRepoList             = "Starting repositories listing for %s org %s."
	MsgCompleteRepoList          = "Finished repositories listing. Total repos: %d"
	MsgStartGitClone             = "Cloning git repository for %s."
	MsgRepoAlreadyExists         = "Cloned git repository for %s already exists."
	MsgCompleteGitClone          = "Finished clone for %s."
	MsgGitCloneEmptyRepo         = "Empty repo %s, skipping clone and other metadata."
	MsgStartExportWebhook        = "Starting webhooks export for repository %s."
	MsgCompleteExportWebhooks    = "Finished export %d webhooks for repository %s."
	MsgStartExportPrComments     = "Starting export comments for repository %s pull request number %d."
	MsgCompleteExportPrComments  = "Finished export %d comments for repository %s pull request number %d."
	MsgStartCommentsFetch        = "Starting fetching comments for PRs in repo %s"
	MsgCompleteCommentsFetch     = "Finished fetching comments for PRs in repo %s"
	MsgCheckpointLoadPRComments  = "Finished export %d comments for repository %s pull request number %d from checkpoint."
	MsgStartExportBranchRules    = "Starting export branch rules for repository %s."
	MsgFailedExportBranchRules   = "Could not export branch rules for repository %s."
	MsgCompleteExportBranchRules = "Finished export %d branch rules for repository %s."

	MsgStartImportFromFolders    = "Starting import repositories from folders: %v"
	MsgCompleteImport            = "Finished import repositories. Total repos: %d."
	MsgStartGetRepoSetting       = "Starting get repository settings for %s"
	MsgCompleteGetRepoSetting    = "Finished get repository settings for %s, push size limit is %d"
	MsgStartUpdateRepoSize       = "Starting update push size limit of repository %s to %d"
	MsgCompleteUpdateRepoSetting = "Finished update push size limit of repository %s to %d"
	MsgStartImportCreateRepo     = "Starting create repository %s."
	MsgCompleteImportCreateRepo  = "Finished create repository %s on %s."
	MsgStartImportGit            = "Starting git push to '%s'."
	MsgCompleteImportGit         = "Finished git push to '%s'."
	MsgStartImportBranchRules    = "Starting importing branch rules for repository %s."
	MsgCompleteImportBranchRules = "Finished import %d branch rules for repository %s."
	MsgStartImportPRs            = "Starting importing pull requests and comments for repository %s."
	MsgCompleteImportPRs         = "Finished import %d pull requests with comments for repository %s."
	MsgStartImportWebhooks       = "Starting importing webhooks for repository %s."
	MsgCompleteImportWebhooks    = "Finished import %d webhooks for repository %s."
	MsgStartRepoCleanup          = "Starting repo cleanup due to an incomplete import of %s"
	MsgCompleteRepoCleanup       = "Finished repo cleanup due to an incomplete import of %s"

	ErrGitClone                     = "cannot clone the git repository %q due to error: %w. output: %s"
	ErrGitFetch                     = "cannot fetch repository references for %s: %w. output: %s"
	ErrCreateRepo                   = "failed to create repository %q at %s due to: %w"
	ErrListWebhook                  = "cannot list webhooks for repository %s: %w"
	ErrListRepo                     = "cannot fetch repositories due to error :%w"
	ErrCheckpointDataRead           = "cannot read checkpoint data: %w"
	ErrCheckpointRepoDataSave       = "cannot save checkpoint repository data for %s: %w"
	ErrCheckpointRepoPageSave       = "cannot save checkpoint repository page for %s: %w"
	ErrCheckpointPrDataSave         = "cannot save checkpoint pr data: %w"
	ErrCheckpointPrPageSave         = "cannot save checkpoint pr page: %w"
	ErrCheckpointWebhooksDataSave   = "cannot save checkpoint webhook data for '%s': %w"
	ErrCheckpointWebhooksPageSave   = "cannot save checkpoint webhook page for '%s': %w"
	ErrCheckpointRulesDataSave      = "cannot save checkpoint rules data for '%s': %w"
	ErrCheckpointRulesPageSave      = "cannot save checkpoint rules page for '%s': %w"
	ErrCheckpointPrCommentsDataSave = "cannot save checkpoint pr comments data: %w"
	ErrCheckpointPrCommentsPageSave = "cannot save checkpoint pr comments page: %w"
	ErrListPr                       = "cannot list pr due to error: %w"
	ErrListComments                 = "cannot list comments for repository %s pull request %d: %w"
	ErrListBranchRules              = "cannot list branch rules for repository %s: %w"
	ErrListBranchRulesets           = "cannot list branch rulesets for repo %s: %w"
	ErrFetchBranchRuleset           = "cannot fetch branch ruleset %d for repo %s: %w"
	ErrListWebhooks                 = "cannot fetch webhooks for repo %s: %w"
	ErrGitPush                      = "cannot git push to '%s' due to %w. output:%s"
	ErrImportBranchRules            = "cannot import branch rules for repository %s: %w"
	ErrImportPRs                    = "cannot import pull requests and comments for repository %s: %w"
	ErrImportWebhooks               = "cannot import webhooks for repository %s: %w"
	ErrCleanupRepo                  = "cannot clean up the repo on server: %w"

	PanicCheckpointSaveErr = "error occurred in reading checkpoint data"
	ErrCannotCreateFolder  = "cannot create folder: %w"
	ErrFetchingFileData    = "error fetching data for file: %w"
	ErrWritingFileData     = "error writing data in file: %w"
	ErrGettingComments     = "encountered error in fetching comments: %v"
)
