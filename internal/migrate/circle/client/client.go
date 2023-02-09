// Copyright 2023 Harness Inc. All rights reserved.

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
