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
	MsgStartExport               = "Starting export repositorie(s)."
	MsgCompleteExport            = "Finished export repositorie(s). Total repos: %d"
	MsgStartExportPRs            = "Starting export pull requests for repository %s."
	MsgCompleteExportPRs         = "Finished export %d pull requests for repository %s."
	MsgCheckpointLoadPr          = "Finished export %d pull requests for repository %s from checkpoint."
	MsgStartRepoList             = "Starting repositories listing for %s org %s."
	MsgCompleteRepoList          = "Finished repositories listing. Total repos: %d"
	MsgStartGitClone             = "Cloning git repository for %s."
	MsgRepoAlreadyExists         = "Cloned git repository for %s already exists."
	MsgCompleteGitClone          = "Finished clone for %s."
	MsgStartExportWebhook        = "Starting webhooks export for repository %s."
	MsgCompleteExportWebhooks    = "Finished export %d webhooks for repository %s."
	MsgStartExportPrComments     = "Starting export comments for repository %s pull request number %d."
	MsgCompleteExportPrComments  = "Finished export %d comments for repository %s pull request number %d."
	MsgCheckpointLoadPRComments  = "Finished export %d comments for repository %s pull request number %d from checkpoint."
	MsgStartExportBranchRules    = "Starting export branch rules for repository %s."
	MsgCompleteExportBranchRules = "Finished export %d branch rules for repository %s."

	MsgStartImportFromFolders    = "Starting import repositories from folders: %v"
	MsgCompleteImport            = "Finished import repositories. Total repos: %d."
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

	ErrGitClone                     = "cannot clone the git repository %q due to error: %w"
	ErrGitFetch                     = "cannot fetch repository references for %s: %w"
	ErrWebhookList                  = "cannot list webhooks for repository %s: %w"
	ErrRepoList                     = "cannot fetch repositories due to error :%v"
	ErrCheckpointDataRead           = "cannot read checkpoint data: %s"
	ErrCheckpointPrDataSave         = "cannot save checkpoint pr data"
	ErrCheckpointPrPageSave         = "cannot save checkpoint pr page"
	ErrCheckpointPrCommentsDataSave = "cannot save checkpoint pr comments data"
	ErrCheckpointPrCommentsPageSave = "cannot save checkpoint pr comments page"
	ErrPrList                       = "cannot list pr due to error :%v"
	ErrCommentsList                 = "cannot list comments for repository %s pull request %d: %v"
	ErrBranchRulesList              = "cannot list branch rules for repository %s: %v"
	ErrGitPush                      = "cannot git push to '%s': %w"
	ErrImportBranchRules            = "cannot import branch rules for repository %s"
	ErrImportPRs                    = "cannot import pull requests and comments for repository %s"
	ErrImportWebhooks               = "cannot import webhooks for repository %s"

	PanicCheckpointSaveErr = "error occurred in reading checkpoint data"
	ErrCannotCreateFolder  = "cannot create folder: %w"
	ErrFetchingFileData    = "error fetching data for file: %w"
	ErrWritingFileData     = "error writing data in file: %w"
)
