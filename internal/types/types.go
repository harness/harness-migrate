// Copyright 2023 Harness Inc. All rights reserved.

// Package types provides common types for data export and import.
package types

// Org defines an organization.
type Org struct {
	Name string `json:"name"`

	Projects []*Project `json:"project,omitempty"`
	Secrets  []*Secret  `json:"secrets,omitempty"`
}

// Project defines a project.
type Project struct {
	Name   string `json:"name"`
	Desc   string `json:"desc,omitempty"`
	Repo   string `json:"repo,omitempty"`
	Branch string `json:"branch,omitempty"`
	Type   string `json:"type"` // github, gitlab, bitbucket
	Yaml   string `json:"yaml"`

	Secrets   []*Secret   `json:"secrets,omitempty"`
	Pipelines []*Pipeline `json:"pipelines,omitempty"`
}

// Pipeline defines a pipeline.
type Pipeline struct {
	Name   string `json:"name"`
	Repo   string `json:"repo,omitempty"`
	Branch string `json:"branch,omitempty"`
	Type   string `json:"type"` // github, gitlab, bitbucket
	Yaml   string `json:"yaml"`
}

// Secret defines a secret.
type Secret struct {
	Name  string `json:"name"`
	Desc  string `json:"desc,omitempty"`
	Value string `json:"value,omitempty"`
}
