// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bitbucket

import (
	"time"

	"github.com/harness/harness-migrate/internal/types"
)

const (
	KindPush                              = "push"
	KindForce                             = "force"
	KindDelete                            = "delete"
	KindRestrictMerges                    = "restrict_merges"
	KindRequireTasksToBeCompleted         = "require_tasks_to_be_completed"
	KindRequireApprovalsToMerge           = "require_approvals_to_merge"
	KindRequireDefaultReviewerApprovals   = "require_default_reviewer_approvals_to_merge"
	KindRequireNoChangesRequested         = "require_no_changes_requested"
	KindRequireCommitsBehind              = "require_commits_behind"
	KindResetPullRequestApprovalsOnChange = "reset_pullrequest_approvals_on_change"
	KindSmartResetPullRequestApprovals    = "smart_reset_pullrequest_approvals"
	KindResetPullRequestChangesRequested  = "reset_pullrequest_changes_requested_on_change"
	KindEnforceMergeChecks                = "enforce_merge_checks"
	KindRequirePassingBuildsToMerge       = "require_passing_builds_to_merge"
	KindAllowAutoMergeWhenBuildsPass      = "allow_auto_merge_when_builds_pass"
)

type (
	Error struct {
		Message string `json:"message"`
		Detail  string `json:"detail"`
	}

	comments struct {
		pagination
		Values []codeComment `json:"values"`
	}

	codeComment struct {
		Type      string    `json:"type"`
		ID        int       `json:"id"`
		CreatedOn time.Time `json:"created_on"`
		UpdatedOn time.Time `json:"updated_on"`
		Content   struct {
			Raw string `json:"raw"`
		} `json:"content"`
		Parent struct {
			ID int `json:"id"`
		} `json:"parent"`
		User    user    `json:"user"`
		Inline  *inline `json:"inline"`
		Pending bool    `json:"pending"`
	}

	// user represents the user who made the comment.
	user struct {
		types.User
		Type        string `json:"type"`
		UUID        string `json:"uuid"`
		AccountID   string `json:"account_id"`
		DisplayName string `json:"display_name"`
	}

	// inline represents inline comment details.
	inline struct {
		From         *int   `json:"from"`
		To           *int   `json:"to"`
		Path         string `json:"path"`
		SrcRev       string `json:"src_rev"`
		DestRev      string `json:"dest_rev"`
		ContextLines string `json:"context_lines"`
		Outdated     bool   `json:"outdated"`
	}

	rules struct {
		Values []branchRule `json:"values"`
		pagination
	}

	branchRule struct {
		ID              int     `json:"id"`
		Kind            string  `json:"kind"`
		Value           *int    `json:"value"`
		BranchMatchKind string  `json:"branch_match_kind"`
		Type            string  `json:"type"`
		Pattern         string  `json:"pattern"`
		Users           []user  `json:"users"`
		Groups          []group `json:"groups"`
		BranchType      string  `json:"branch_type"`
	}

	group struct {
		Name string `json:"name"`
	}

	// pagination is Bitbucket pagination properties in list responses.
	pagination struct {
		PageLen int    `json:"pagelen"`
		Page    int    `json:"page"`
		Size    int    `json:"size"`
		Next    string `json:"next"`
	}
)

func (e *Error) Error() string {
	return e.Message
}
