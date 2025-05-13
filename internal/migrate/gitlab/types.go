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

package gitlab

import (
	"time"

	"github.com/harness/harness-migrate/internal/types"
)

type (
	squashOption string
	mergeMethod  string
	level        int
)

const (
	squashOptionOn     squashOption = "default_on"
	squashOptionOff    squashOption = "default_off"
	squashOptionNever  squashOption = "never"
	squashOptionAlways squashOption = "always"

	mergeMethodMerge       mergeMethod = "merge"
	mergeMethodRebase      mergeMethod = "rebase_merge"
	mergeMethodFastForward mergeMethod = "ff"

	levelNoAccess level = 0
	levelAdmin    level = 60
	// skipping Access Level 30: Developer and 40 Maintainer as we dont support migrating roles
)

type (
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	author struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Username  string `json:"username"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
	}

	user struct {
		types.User
		UserName    string `json:"username"`
		PublicEmail string `json:"public_email"`
	}

	repoInfo struct {
		LFSEnabled bool `json:"lfs_enabled"`
	}

	mergeRequest struct {
		Number         int    `json:"iid"`
		Sha            string `json:"sha"`
		Title          string `json:"title"`
		Desc           string `json:"description"`
		State          string `json:"state"`
		WorkInProgress bool   `json:"work_in_progress"`
		Link           string `json:"web_url"`
		Author         struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Name     string `json:"name"`
			Avatar   string `json:"avatar_url"`
		}
		SourceBranch string    `json:"source_branch"`
		TargetBranch string    `json:"target_branch"`
		Created      time.Time `json:"created_at"`
		Updated      time.Time `json:"updated_at"`
		Closed       time.Time
		Labels       []string `json:"labels"`
		DiffRefs     struct {
			BaseSha  string `json:"base_sha"`
			HeadSha  string `json:"head_sha"`
			StartSha string `json:"start_sha"`
		} `json:"diff_refs"`
	}

	line struct {
		LineCode string `json:"line_code"`
		Type     string `json:"type"`
		OldLine  *int   `json:"old_line"`
		NewLine  *int   `json:"new_line"`
	}

	position struct {
		BaseSHA  string `json:"base_sha"`
		StartSHA string `json:"start_sha"`
		HeadSHA  string `json:"head_sha"`

		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`

		Type    string `json:"position_type"`
		OldLine int    `json:"old_line"`
		NewLine int    `json:"new_line"`

		LineRange struct {
			Start line `json:"start"`
			End   line `json:"end"`
		} `json:"line_range"`
	}

	codeComment struct {
		ID           int       `json:"id"`
		Type         string    `json:"type"`
		Body         string    `json:"body"`
		Author       author    `json:"author"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		System       bool      `json:"system"`
		NoteableID   int       `json:"noteable_id"`
		NoteableType string    `json:"noteable_type"`
		ProjectID    int       `json:"project_id"`
		CommitID     string    `json:"commit_id"`
		Position     *position `json:"position"`
		Resolved     bool      `json:"resolved"`
		Resolvable   bool      `json:"resolvable"`
		ResolvedBy   *author   `json:"resolved_by"`
		Suggestions  []string  `json:"suggestions"`
	}

	discussion struct {
		Id    string        `json:"id"`
		Notes []codeComment `json:"notes"`
	}

	accessLevel struct {
		Id           int    `json:"id"`
		AccessLevel  level  `json:"access_level"`
		Description  string `json:"access_level_description"`
		DeployeKeyID *int   `json:"deploy_key_id"`
		UserID       *int   `json:"user_id"`
		GroupID      *int   `json:"group_id"`
	}

	branchRule struct {
		Id                int            `json:"id"`
		Name              string         `json:"name"`
		PushAccess        []*accessLevel `json:"push_access_levels"`
		MergeAccess       []*accessLevel `json:"merge_access_levels"`
		UnprotectAccess   []*accessLevel `json:"unprotect_access_levels"`
		AllowForcePush    bool           `json:"allow_force_push"`
		CodeOwnerRequired bool           `json:"code_owner_approval_required"`
	}

	project struct {
		Id int `json:"id"`
		mergeRequestRules
	}

	mergeRequestRules struct {
		DeleteBranch           bool         `json:"remove_source_branch_after_merge"`
		RequireResolveComments bool         `json:"only_allow_merge_if_all_discussions_are_resolved"`
		SquashOption           squashOption `json:"squash_option"`
		MergeMethod            mergeMethod  `json:"merge_method"`
	}
)

func (e *Error) Error() string {
	return e.Message
}
