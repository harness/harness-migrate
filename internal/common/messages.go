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
	MsgStartPrImport           = "Starting Pull request import for repo %s"
	MsgCompletePrImport        = "Pull request import finished for repo %s with total %d PRs"
	MsgCheckpointLoadPr        = "Pull request import finished for repo %s with total %d PRs from checkpoint"
	MsgStartRepoList           = "starting repositories listing for %s org %s"
	MsgCompleteRepoList        = "repositories listing complete. Total repos: %d"
	MsgStartGitClone           = "Starting clone git repo for %q"
	MsgCompleteGitClone        = "Git clone finished for %q"
	MsgGitFetchRef             = "fetching pull request references for %s"
	MsgCompleteGitFetchRef     = "fetched pull request references for %s"
	MsgRepoListErr             = "cannot fetch repositories due to error :%w"
	MsgCheckpointDataReadErr   = "cannot read checkpoint data: %s"
	MsgCheckpointPrDataSaveErr = "cannot save checkpoint pr data"
	MsgCheckpointPrPageSaveErr = "cannot save checkpoint pr page"
	MsgPrListErr               = "cannot list pr due to error :%w"
	ErrGitCloneMsg             = "cannot clone the git repo %q due to error: %w"
	ErrGitFetch                = "cannot fetch repo pull request references for %s: %w"

	PanicCheckpointSaveErr = "error occured in reading checkpoint data"
)
