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

// Package harness provides a Harness http client.
package harness

import (
	"time"

	"github.com/harness/harness-migrate/types"
)

// Client is used to communicate with the Harness server.
type Client interface {
	// FindOrg returns an organization by identifier.
	FindOrg(id string) (*Org, error)

	// FindProject returns a project by organization and
	// identifier.
	FindProject(org, id string) (*Project, error)

	// FindPipeline returns a pipeline by organization,
	// project and identifer.
	FindPipeline(org, project, id string) (*Pipeline, error)

	// FindSecret returns a secret by organization, project
	// and identifer.
	FindSecret(org, project, id string) (*Secret, error)

	// FindSecretOrg returns a secret by organization and
	// identifer.
	FindSecretOrg(org, id string) (*Secret, error)

	// FindConnector returns a connector by organization,
	// project and identifer.
	FindConnector(org, project, id string) (*Connector, error)

	// FindConnectorOrg returns a connector by organization
	// and identifer.
	FindConnectorOrg(org, id string) (*Connector, error)

	// CreateOrg creates an organization.
	CreateOrg(org *Org) error

	// CreateProject creates a project.
	CreateProject(project *Project) error

	// CreateSecret creates a secret.
	CreateSecret(secret *Secret) error

	// CreateSecretOrg creates an organization secret.
	CreateSecretOrg(secret *Secret) error

	// CreateConnector creates a connector.
	CreateConnector(connector *Connector) error

	// CreateConnectorOrg creates an organization connector.
	CreateConnectorOrg(connector *Connector) error

	// CreatePipeline creates a pipeline for the
	// organization and pipeline identifier, with the
	// given identifier and name.
	CreatePipeline(org, project string, pipeline []byte) error

	// CreateRepository creates a repository.
	CreateRepository(parentRef string, repo *CreateRepositoryInput) (*Repository, error)

	// DeleteRepository deletes a repository
	DeleteRepository(repoRef string) error

	// CreateRepositoryForMigration creates an empty repository ready for migration.
	CreateRepositoryForMigration(in *CreateRepositoryForMigrateInput) (*Repository, error)

	// FindRepoSettings finds general settings of a repository.
	FindRepoSettings(repoRef string) (*RepoSettings, error)

	// UpdateRepoSettings updates general settings of a repository.
	UpdateRepoSettings(repoRef string, in *RepoSettings) (*RepoSettings, error)

	// UpdateRepositoryState updates a repository state (for different steps of the migration).
	UpdateRepositoryState(repoRef string, in *UpdateRepositoryStateInput) (*Repository, error)

	// ImportPRs imports pull requests of a repository.
	ImportPRs(repoRef string, in *types.PRsImportInput) error

	// ImportWebhooks imports webhooks of a repository
	ImportWebhooks(repoRef string, in *types.WebhookInput) error

	// ImportRules imports protection rules of a repository
	ImportRules(repoRef string, in *types.RulesInput) error

	// CheckUsers provides all email id to harness code of users which needs to be checked for existence.
	CheckUsers(in *types.CheckUsersInput) (*types.CheckUsersOutput, error)
}

// WaitHarnessSecretManager blocks until the harness
// secret manager is created for the project.
func WaitHarnessSecretManager(client Client, org, project string) error {
	for i := 0; ; i++ {
		if _, err := client.FindConnector(org, project, "harnessSecretManager"); err == nil {
			return nil
		} else if i == 30 {
			return err
		} else {
			time.Sleep(5 * time.Second)
		}
	}
}

// WaitHarnessSecretManagerOrg blocks until the harness
// secret manager is created for the organization.
func WaitHarnessSecretManagerOrg(client Client, org string) error {
	for i := 0; ; i++ {
		if _, err := client.FindConnectorOrg(org, "harnessSecretManager"); err == nil {
			return nil
		} else if i == 30 {
			return err
		} else {
			time.Sleep(5 * time.Second)
		}
	}
}
