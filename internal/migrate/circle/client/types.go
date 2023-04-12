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

package client

import "time"

type (
	// Pipeline defines a pipeline.
	Pipeline struct {
		ID      string    `json:"id"`
		Slug    string    `json:"project_slug"`
		Number  int       `json:"number"`
		State   string    `json:"state"`
		Created time.Time `json:"created_at"`
		Updated time.Time `json:"updated_at"`
		Params  struct {
			Gitlab struct {
				ID            string `json:"project_id"`
				Name          string `json:"repo_name"`
				Remote        string `json:"repo_url"`
				DeepLink      string `json:"web_url"`
				Branch        string `json:"branch"`
				BranchDefault string `json:"default_branch"`
			} `json:"gitlab"`
		} `json:"trigger_parameters"`
	}

	// PipelineList defines a pipeline list.
	PipelineList struct {
		Items []*Pipeline `json:"items"`
		Token *string     `json:"next_page_token"`
	}

	// Project defines a project.
	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
		Vcs  struct {
			Link     string `json:"vcs_url"`
			Branch   string `json:"default_branch"`
			Provider string `json:"provider"`
		} `json:"vcs_info"`
	}

	// ProjectList defines a project list.
	ProjectList struct {
		Items []*Project `json:"items"`
		Token *string    `json:"next_page_token"`
	}

	// Org defines an organization.
	Org struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	// Collab defines a collaboration.
	Collab struct {
		ID   string `json:"id"`
		Type string `json:"vcs_type"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	// Org defines a pipeline configuration.
	Config struct {
		Source   string `json:"source"`
		Compiled string `json:"compiled"`
	}

	// Env defines an environment variable.
	Env struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	// EnvList defines an environment variable.
	EnvList struct {
		Items []*Env  `json:"items"`
		Token *string `json:"next_page_token"`
	}
)
