// Copyright 2023 Harness Inc. All rights reserved.

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
