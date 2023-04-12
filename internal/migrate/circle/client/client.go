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

// Package client provides a Circle http client.
package client

// Client is used to communicate with the Circle server.
type Client interface {
	// FindOrg returns an organization by slug.
	FindOrg(slug string) (*Org, error)

	// FindOrgID returns an organization by ID.
	FindOrgID(id string) (*Org, error)

	// FindProject returns a project by slug.
	FindProject(slug string) (*Project, error)

	// FindPipeline returns a pipeline by slug.
	FindPipeline(slug string) (*Pipeline, error)

	// FindConfig returns a pipeline configuration by
	// pipeline id.
	FindConfig(id string) (*Config, error)

	// ListOrgs returns the user organization list.
	ListOrgs() ([]*Org, error)

	// ListProjects returns a project list by organization id.
	ListProjects(id string) ([]*Project, error)

	// ListPipelines returns a pipeline list by pipeline slug.
	ListPipelines(slug string) ([]*Pipeline, error)

	// ListEnvs returns an environment variable list by
	// pipeline slug.
	ListEnvs(slug string) ([]*Env, error)
}
