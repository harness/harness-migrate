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
	Yaml   []byte `json:"yaml"`

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
