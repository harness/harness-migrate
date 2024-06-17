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
	MsgStartPrExport             = "Starting Pull request export for repo %s"
	MsgCompletePrExport          = "Pull request export finished for repo %s with total %d PRs"
	MsgCheckpointLoadPr          = "Pull request export finished for repo %s with total %d PRs from checkpoint"
	MsgStartRepoList             = "starting repositories listing for %s org %s"
	MsgCompleteRepoList          = "repositories listing complete. Total repos: %d"
	MsgStartGitClone             = "starting clone git repo for %q"
	MsgRepoAlreadyExists         = "cloned git repo for %s already exists"
	MsgCompleteGitClone          = "git clone finished for %q"
	MsgGitFetchRef               = "fetching pull request references for %s"
	MsgCompleteGitFetchRef       = "fetched pull request references for %s"
	MsgStartWebhookList          = "starting listing webhooks for repo %s"
	MsgStartPrCommentsExport     = "Starting PR comments export for repo %s pull request %d"
	MsgCompletePrCommentsExport  = "PR comments export finished for repo %s pull request %d with total %d comments"
	MsgCheckpointLoadPRComments  = "PR comments export finished for repo %s pull request %d with total %d comments from checkpoint"
	MsgStartBranchRulesExport    = "Starting branch rules export for repo %s"
	MsgCompleteBranchRulesExport = "Branch rules export finished for repo %s with total %d rules"

	ErrGitClone                     = "cannot clone the git repo %q due to error: %w"
	ErrGitFetch                     = "cannot fetch repo pull request references for %s: %w"
	ErrWebhookList                  = "cannot list webhooks for repo %s: %w"
	ErrRepoList                     = "cannot fetch repositories due to error :%v"
	ErrCheckpointDataRead           = "cannot read checkpoint data: %s"
	ErrCheckpointPrDataSave         = "cannot save checkpoint pr data"
	ErrCheckpointPrPageSave         = "cannot save checkpoint pr page"
	ErrCheckpointPrCommentsDataSave = "cannot save checkpoint pr comments data"
	ErrCheckpointPrCommentsPageSave = "cannot save checkpoint pr comments page"
	ErrPrList                       = "cannot list pr due to error :%v"
	ErrCommentsList                 = "cannot list comments for repo %s pull request %d: %v"
	ErrBranchRulesList              = "cannot list branch rules for repo %s: %v"

	PanicCheckpointSaveErr  = "error occurred in reading checkpoint data"
	PanicCannotCreateFolder = "cannot create folder: %v"
	PanicFetchingFileData   = "error fetching data for file: %v"
	PanicWritingFileData    = "error writing data in file: %v"
)
