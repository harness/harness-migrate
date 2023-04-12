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

package repo

type Repo struct {
	ID            int64  `db:"repo_id"`
	UID           string `db:"repo_uid"`
	UserID        int64  `db:"repo_user_id"`
	Namespace     string `db:"repo_namespace"`
	Name          string `db:"repo_name"`
	Slug          string `db:"repo_slug"`
	SCM           string `db:"repo_scm"`
	CloneURL      string `db:"repo_clone_url"`
	SSHURL        string `db:"repo_ssh_url"`
	HTMLURL       string `db:"repo_html_url"`
	Branch        string `db:"repo_branch"`
	Private       bool   `db:"repo_private"`
	Visibility    string `db:"repo_visibility"`
	Active        bool   `db:"repo_active"`
	Config        string `db:"repo_config"`
	Trusted       bool   `db:"repo_trusted"`
	Protected     bool   `db:"repo_protected"`
	NoForks       bool   `db:"repo_no_forks"`
	NoPulls       bool   `db:"repo_no_pulls"`
	CancelPulls   bool   `db:"repo_cancel_pulls"`
	CancelPush    bool   `db:"repo_cancel_push"`
	CancelRunning bool   `db:"repo_cancel_running"`
	Timeout       int    `db:"repo_timeout"`
	Throttle      int    `db:"repo_throttle"`
	Counter       int64  `db:"repo_counter"`
	Synced        int64  `db:"repo_synced"`
	Created       int64  `db:"repo_created"`
	Updated       int64  `db:"repo_updated"`
	Version       int    `db:"repo_version"`
	Signer        string `db:"repo_signer"`
	Secret        string `db:"repo_secret"`
	Build         *Build `db:"-"`
}

type Build struct {
	ID           int64  `db:"build_id"`
	RepoID       int64  `db:"build_repo_id"`
	Trigger      string `db:"build_trigger"`
	Number       int    `db:"build_number"`
	Parent       int    `db:"build_parent"`
	Status       string `db:"build_status"`
	Error        string `db:"build_error"`
	Event        string `db:"build_event"`
	Action       string `db:"build_action"`
	Link         string `db:"build_link"`
	Timestamp    int64  `db:"build_timestamp"`
	Title        string `db:"build_title"`
	Message      string `db:"build_message"`
	Before       string `db:"build_before"`
	After        string `db:"build_after"`
	Ref          string `db:"build_ref"`
	Fork         string `db:"build_source_repo"`
	Source       string `db:"build_source"`
	Target       string `db:"build_target"`
	Author       string `db:"build_author"`
	AuthorName   string `db:"build_author_name"`
	AuthorEmail  string `db:"build_author_email"`
	AuthorAvatar string `db:"build_author_avatar"`
	Sender       string `db:"build_sender"`
	Params       string `db:"build_params"`
	Cron         string `db:"build_cron"`
	Deploy       string `db:"build_deploy"`
	DeployID     int64  `db:"build_deploy_id"`
	Debug        bool   `db:"build_debug"`
	Started      int64  `db:"build_started"`
	Finished     int64  `db:"build_finished"`
	Created      int64  `db:"build_created"`
	Updated      int64  `db:"build_updated"`
	Version      int    `db:"build_version"`
}

type Secret struct {
	ID              int64  `db:"secret_id"`
	RepoID          int64  `db:"secret_repo_id"`
	Name            string `db:"secret_name"`
	Data            string `db:"secret_data"`
	PullRequest     bool   `db:"secret_pull_request"`
	PullRequestPush bool   `db:"secret_pull_request_push"`
	Created         int64  `db:"secret_created"`
	Updated         int64  `db:"secret_updated"`
}

type OrgSecret struct {
	ID              int64  `db:"secret_id"`
	Namespace       string `db:"secret_namespace"`
	Name            string `db:"secret_name"`
	Type            string `db:"secret_type"`
	Data            []byte `db:"secret_data"`
	PullRequest     bool   `db:"secret_pull_request"`
	PullRequestPush bool   `db:"secret_pull_request_push"`
}
