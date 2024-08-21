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

package harness

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/harness/harness-migrate/types"
)

type gitnessClient struct {
	address string
	token   string
	tracing bool
}

func NewGitness(token, address string, opts ...Option) Client {
	client_ := &gitnessClient{
		token:   token,
		address: address,
	}
	// set optional parameters.
	for _, opt := range opts {
		opt(client_)
	}

	return client_
}

// FindOrg returns an organization by identifier.
func (c *gitnessClient) FindOrg(org string) (*Org, error) {
	return nil, fmt.Errorf("not implemented")
}

// FindProject returns a project by organization and
// identifier.
func (c *gitnessClient) FindProject(org, project string) (*Project, error) {
	return nil, fmt.Errorf("not implemented")
}

// FindPipeline returns a pipeline by organization,
// project and identifer.
func (c *gitnessClient) FindPipeline(org, project, pipeline string) (*Pipeline, error) {
	return nil, fmt.Errorf("not implemented")
}

// FindSecret returns a secret by organization, project
// and identifer.
func (c *gitnessClient) FindSecret(org, project, id string) (*Secret, error) {
	return nil, fmt.Errorf("not implemented")
}

// FindSecretOrg returns a secret by organization and
// identifer.
func (c *gitnessClient) FindSecretOrg(org, id string) (*Secret, error) {
	return nil, fmt.Errorf("not implemented")
}

// FindConnector returns a connector by organization,
// project and identifer.
func (c *gitnessClient) FindConnector(org, project, conn string) (*Connector, error) {
	return nil, fmt.Errorf("not implemented")
}

// FindConnectorOrg returns a connector by organization
// and identifer.
func (c *gitnessClient) FindConnectorOrg(org, conn string) (*Connector, error) {
	return nil, fmt.Errorf("not implemented")
}

// CreateOrg creates an organization.
func (c *gitnessClient) CreateOrg(org *Org) error {
	return fmt.Errorf("not implemented")
}

// CreateProject creates a project.
func (c *gitnessClient) CreateProject(project *Project) error {
	return fmt.Errorf("not implemented")
}

// CreateSecret creates a secret.
func (c *gitnessClient) CreateSecret(secret *Secret) error {
	return fmt.Errorf("not implemented")
}

// CreateSecret creates an organization secret.
func (c *gitnessClient) CreateSecretOrg(secret *Secret) error {
	return fmt.Errorf("not implemented")
}

// CreateConnector creates a connector.
func (c *gitnessClient) CreateConnector(connector *Connector) error {
	return fmt.Errorf("not implemented")
}

// CreateConnectorOrg creates an organization connector.
func (c *gitnessClient) CreateConnectorOrg(connector *Connector) error {
	return fmt.Errorf("not implemented")
}

// CreatePipeline creates a pipeline for the
// organization and pipeline identifier, with the
// given identifier and name.
func (c *gitnessClient) CreatePipeline(org, project string, pipeline []byte) error {
	return fmt.Errorf("not implemented")
}

func (c *gitnessClient) CreateRepository(parentRef string, repo *CreateRepositoryInput) (*Repository, error) {
	out := new(Repository)
	in := &CreateGitnessRepositoryInput{
		CreateRepositoryInput: *repo,
		ParentRef:             parentRef,
	}
	uri := fmt.Sprintf("%s/api/v1/repos",
		c.address,
	)

	if err := c.post(uri, in, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FindRepoSettings finds general settings of a repository.
func (c *gitnessClient) FindRepoSettings(repoRef string) (*RepoSettings, error) {
	out := new(RepoSettings)
	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/api/v1/repos/%s/settings/general",
		c.address,
		repoRef,
	)

	if err := c.get(uri, out); err != nil {
		return nil, err
	}

	return out, nil
}

// UpdateRepoSettings updates general settings of a repository.
func (c *gitnessClient) UpdateRepoSettings(repoRef string, in *RepoSettings) (*RepoSettings, error) {
	out := new(RepoSettings)
	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/api/v1/repos/%s/settings/general",
		c.address,
		repoRef,
	)

	if err := c.patch(uri, in, out); err != nil {
		return nil, err
	}

	return out, nil
}

func (c *gitnessClient) DeleteRepository(repoRef string) error {
	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/api/v1/repos/%s",
		c.address,
		repoRef,
	)
	if err := c.delete(uri); err != nil {
		return err
	}
	return nil
}

func (c *gitnessClient) CreateRepositoryForMigration(in *CreateRepositoryForMigrateInput) (*Repository, error) {
	out := new(Repository)
	uri := fmt.Sprintf("%s/api/v1/migrate/repos",
		c.address,
	)

	if err := c.post(uri, in, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitnessClient) UpdateRepositoryState(repoRef string, in *UpdateRepositoryStateInput) (*Repository, error) {
	out := new(Repository)
	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/api/v1/migrate/repos/%s/update-state",
		c.address,
		repoRef,
	)

	if err := c.patch(uri, in, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitnessClient) ImportPRs(repoRef string, in *types.PRsImportInput) error {
	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/api/v1/migrate/repos/%s/pullreqs",
		c.address,
		repoRef,
	)

	if err := c.post(uri, in, nil); err != nil {
		return err
	}
	return nil
}

func (c *gitnessClient) ImportWebhooks(repoRef string, in *types.WebhookInput) error {
	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/api/v1/migrate/repos/%s/webhooks",
		c.address,
		repoRef,
	)
	if err := c.post(uri, in, nil); err != nil {
		return err
	}
	return nil
}

func (c *gitnessClient) ImportRules(repoRef string, in *types.RulesInput) error {
	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/api/v1/migrate/repos/%s/rules",
		c.address,
		repoRef,
	)
	if err := c.post(uri, in, nil); err != nil {
		return err
	}
	return nil
}

func (c *gitnessClient) CheckUsers(in *types.CheckUsersInput) (*types.CheckUsersOutput, error) {
	out := new(types.CheckUsersOutput)
	uri := fmt.Sprintf("%s/api/v1/principals/check-emails", c.address)

	if err := c.post(uri, in, out); err != nil {
		return nil, err
	}
	return out, nil
}

// http request helper functions
func (c *gitnessClient) setAuthHeader() func(h *http.Header) {
	return func(h *http.Header) { h.Set("Authorization", c.token) }
}

// helper function for making an http GET request.
func (c *gitnessClient) get(rawurl string, out interface{}) error {
	return Do(rawurl, "GET", c.setAuthHeader(), nil, out, c.tracing)
}

// helper function for making an http POST request.
func (c *gitnessClient) post(rawurl string, in, out interface{}) error {
	return Do(rawurl, "POST", c.setAuthHeader(), in, out, c.tracing)
}

// helper function for making an http PATCH request.
func (c *gitnessClient) patch(rawurl string, in, out interface{}) error {
	return Do(rawurl, "PATCH", c.setAuthHeader(), in, out, c.tracing)
}

// helper function for making an http DELETE request.
func (c *gitnessClient) delete(rawurl string) error {
	return Do(rawurl, "DELETE", c.setAuthHeader(), nil, nil, c.tracing)
}
