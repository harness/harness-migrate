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
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/harness/harness-migrate/types"
)

type client struct {
	*gitnessClient
	account string
}

// New returns a new Client.
func New(account, token string, opts ...Option) Client {
	client_ := &client{
		gitnessClient: &gitnessClient{
			token: token,
		},
		account: account,
	}
	// set optional parameters.
	for _, opt := range opts {
		opt(client_.gitnessClient)
	}
	// set default address if not provided.
	if client_.address == "" {
		client_.address = "https://app.harness.io"
	}
	return client_
}

// FindOrg returns an organization by identifier.
func (c *client) FindOrg(org string) (*Org, error) {
	out := new(orgEnvelope)
	uri := fmt.Sprintf("%s/gateway/ng/api/organizations/%s?accountIdentifier=%s", c.address, org, c.account)
	if err := c.get(uri, &out); err != nil {
		return nil, err
	} else if out.Data == nil || out.Data.Organization == nil {
		return nil, errors.New("not found")
	}
	return out.Data.Organization, nil
}

// FindProject returns a project by organization and
// identifier.
func (c *client) FindProject(org, project string) (*Project, error) {
	out := new(projectEnvelope)
	uri := fmt.Sprintf("%s/gateway/ng/api/projects/%s?accountIdentifier=%s&orgIdentifier=%s", c.address, project, c.account, org)
	if err := c.get(uri, &out); err != nil {
		return nil, err
	} else if out.Data == nil || out.Data.Project == nil {
		return nil, errors.New("not found")
	}
	return out.Data.Project, nil
}

// FindPipeline returns a pipeline by organization,
// project and identifer.
func (c *client) FindPipeline(org, project, pipeline string) (*Pipeline, error) {
	out := new(pipelineEnvelope)
	uri := fmt.Sprintf("%s/gateway/pipeline/api/pipelines/summary/%s?accountIdentifier=%s&orgIdentifier=%s&projectIdentifier=%s", c.address, pipeline, c.account, org, project)
	if err := c.get(uri, &out); err != nil {
		return nil, err
	} else if out.Data == nil {
		return nil, errors.New("not found")
	}
	return out.Data, nil
}

// FindSecret returns a secret by organization, project
// and identifer.
func (c *client) FindSecret(org, project, id string) (*Secret, error) {
	out := new(secretEnvelope)
	uri := fmt.Sprintf("%s/gateway/ng/api/v2/secrets/%s?accountIdentifier=%s&orgIdentifier=%s&projectIdentifier=%s", c.address, id, c.account, org, project)
	if err := c.get(uri, &out); err != nil {
		return nil, err
	} else if out.Data == nil || out.Data.Secret == nil {
		return nil, errors.New("not found")
	}
	return out.Data.Secret, nil
}

// FindSecretOrg returns a secret by organization and
// identifer.
func (c *client) FindSecretOrg(org, id string) (*Secret, error) {
	out := new(secretEnvelope)
	uri := fmt.Sprintf("%s/gateway/ng/api/v2/secrets/%s?accountIdentifier=%s&orgIdentifier=%s", c.address, id, c.account, org)
	if err := c.get(uri, &out); err != nil {
		return nil, err
	} else if out.Data == nil || out.Data.Secret == nil {
		return nil, errors.New("not found")
	}
	return out.Data.Secret, nil
}

// FindConnector returns a connector by organization,
// project and identifer.
func (c *client) FindConnector(org, project, conn string) (*Connector, error) {
	out := new(connectorEnvelope)
	uri := fmt.Sprintf("%s/gateway/ng/api/connectors/%s?accountIdentifier=%s&orgIdentifier=%s&projectIdentifier=%s", c.address, conn, c.account, org, project)
	if err := c.get(uri, &out); err != nil {
		return nil, err
	} else if out.Data == nil || out.Data.Connector == nil {
		return nil, errors.New("not found")
	}
	return out.Data.Connector, nil
}

// FindConnectorOrg returns a connector by organization
// and identifer.
func (c *client) FindConnectorOrg(org, conn string) (*Connector, error) {
	out := new(connectorEnvelope)
	uri := fmt.Sprintf("%s/gateway/ng/api/connectors/%s?accountIdentifier=%s&orgIdentifier=%s", c.address, conn, c.account, org)
	if err := c.get(uri, &out); err != nil {
		return nil, err
	} else if out.Data == nil {
		return nil, errors.New("not found")
	} else if out.Data.Connector == nil {
		return nil, errors.New("not found")
	}
	return out.Data.Connector, nil
}

// CreateOrg creates an organization.
func (c *client) CreateOrg(org *Org) error {
	in := new(orgCreateEnvelope)
	in.Org = org
	out := new(orgEnvelope)
	uri := fmt.Sprintf("%s/gateway/ng/api/organizations?accountIdentifier=%s", c.address, c.account)
	return c.post(uri, in, out)
}

// CreateProject creates a project.
func (c *client) CreateProject(project *Project) error {
	in := new(projectCreateEnvelope)
	in.Project = project
	out := new(projectEnvelope)
	uri := fmt.Sprintf("%s/gateway/ng/api/projects?accountIdentifier=%s&orgIdentifier=%s", c.address, c.account, project.Orgidentifier)
	return c.post(uri, in, out)
}

// CreateSecret creates a secret.
func (c *client) CreateSecret(secret *Secret) error {
	in := new(secretCreateEnvelope)
	in.Secret = secret
	out := new(secretEnvelope)
	uri := fmt.Sprintf("%s/gateway/ng/api/v2/secrets?accountIdentifier=%s&orgIdentifier=%s&projectIdentifier=%s",
		c.address,
		c.account,
		secret.Orgidentifier,
		secret.Projectidentifier,
	)
	return c.post(uri, in, out)
}

// CreateSecret creates an organization secret.
func (c *client) CreateSecretOrg(secret *Secret) error {
	in := new(secretCreateEnvelope)
	in.Secret = secret
	out := new(secretEnvelope)
	uri := fmt.Sprintf("%s/gateway/ng/api/v2/secrets?accountIdentifier=%s&orgIdentifier=%s",
		c.address,
		c.account,
		secret.Orgidentifier,
	)
	if err := c.post(uri, in, out); err != nil {
		return err
	} else if out.Data == nil {
		return errors.New("not found")
	}
	return nil
}

// CreateConnector creates a connector.
func (c *client) CreateConnector(connector *Connector) error {
	in := new(connectorCreateEnvelope)
	in.Connector = connector
	out := new(connectorEnvelope)
	uri := fmt.Sprintf("%s/gateway/ng/api/connectors?accountIdentifier=%s&orgIdentifier=%s&projectIdentifier=%s",
		c.address,
		c.account,
		connector.Orgidentifier,
		connector.Projectidentifier,
	)
	err := c.post(uri, in, out)
	return err
}

// CreateConnectorOrg creates an organization connector.
func (c *client) CreateConnectorOrg(connector *Connector) error {
	in := new(connectorCreateEnvelope)
	in.Connector = connector
	out := new(connectorEnvelope)
	uri := fmt.Sprintf("%s/gateway/ng/api/connectors?accountIdentifier=%s&orgIdentifier=%s",
		c.address,
		c.account,
		connector.Orgidentifier,
	)
	if err := c.post(uri, in, out); err != nil {
		return err
	} else if out.Data == nil {
		return errors.New("not found")
	}
	return nil
}

// CreatePipeline creates a pipeline for the
// organization and pipeline identifier, with the
// given identifier and name.
func (c *client) CreatePipeline(org, project string, pipeline []byte) error {
	buf := bytes.NewBuffer(pipeline)
	out := new(pipelineEnvelope)
	uri := fmt.Sprintf("%s/gateway/pipeline/api/pipelines/v2?accountIdentifier=%s&orgIdentifier=%s&projectIdentifier=%s&storeType=INLINE",
		c.address,
		c.account,
		org,
		project,
	)
	return c.post(uri, buf, out)
}

// CreateRepository creates a repository for the parentRef, if none provide repo will be at the acc level
func (c *client) CreateRepository(parentRef string, repo *CreateRepositoryInput) (*Repository, error) {
	out := new(Repository)
	pathParts := strings.Split(parentRef, "/")
	var org string
	var prj string
	if len(pathParts) >= 1 {
		org = pathParts[0]
	}
	if len(pathParts) >= 2 {
		prj = pathParts[1]
	}

	uri := fmt.Sprintf("%s/gateway/code/api/v1/accounts/%s/orgs/%s/projects/%s/repos",
		c.address,
		c.account,
		org, //org
		prj, //project
	)

	if err := c.post(uri, repo, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FindRepoSettings finds general settings of a repository.
func (c *client) FindRepoSettings(repoRef string) (*RepoSettings, error) {
	out := new(RepoSettings)
	queryParams, err := getQueryParamsFromRepoRef(repoRef)
	if err != nil {
		return nil, err
	}

	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/gateway/code/api/v1/repos/%s/settings/general?%s",
		c.address,
		repoRef,
		queryParams,
	)

	if err := c.get(uri, out); err != nil {
		return nil, err
	}

	return out, nil
}

// UpdateRepoSettings updates general settings of a repository.
func (c *client) UpdateRepoSettings(repoRef string, in *RepoSettings) (*RepoSettings, error) {
	out := new(RepoSettings)
	queryParams, err := getQueryParamsFromRepoRef(repoRef)
	if err != nil {
		return nil, err
	}

	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/gateway/code/api/v1/repos/%s/settings/general?%s",
		c.address,
		repoRef,
		queryParams,
	)

	if err := c.patch(uri, in, out); err != nil {
		return nil, err
	}

	return out, nil
}

func (c *client) DeleteRepository(repoRef string) error {
	queryParams, err := getQueryParamsFromRepoRef(repoRef)
	if err != nil {
		return err
	}

	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/gateway/code/api/v1/repos/%s?%s",
		c.address,
		repoRef,
		queryParams,
	)
	if err := c.delete(uri); err != nil {
		return err
	}

	return nil
}

func (c *client) CreateRepositoryForMigration(in *CreateRepositoryForMigrateInput) (*Repository, error) {
	out := new(Repository)
	queryParams, err := getQueryParamsFromRepoRef(path.Join(in.ParentRef, in.Identifier))
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s/gateway/code/api/v1/migrate/repos?%s", c.address, queryParams)
	if err := c.post(uri, in, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *client) UpdateRepositoryState(repoRef string, in *UpdateRepositoryStateInput) (*Repository, error) {
	out := new(Repository)
	queryParams, err := getQueryParamsFromRepoRef(repoRef)
	if err != nil {
		return nil, err
	}

	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/gateway/code/api/v1/migrate/repos/%s/update-state?%s",
		c.address,
		repoRef,
		queryParams,
	)

	if err := c.patch(uri, in, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *client) ImportPRs(repoRef string, in *types.PRsImportInput) error {
	queryParams, err := getQueryParamsFromRepoRef(repoRef)
	if err != nil {
		return err
	}

	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/gateway/code/api/v1/migrate/repos/%s/pullreqs?%s",
		c.address,
		repoRef,
		queryParams,
	)

	if err := c.post(uri, in, nil); err != nil {
		return err
	}
	return nil
}

func (c *client) ImportWebhooks(repoRef string, in *types.WebhookInput) error {
	queryParams, err := getQueryParamsFromRepoRef(repoRef)
	if err != nil {
		return err
	}

	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/gateway/code/api/v1/migrate/repos/%s/webhooks?%s",
		c.address,
		repoRef,
		queryParams,
	)
	if err := c.post(uri, in, nil); err != nil {
		return err
	}
	return nil
}

func (c *client) ImportRules(repoRef string, in *types.RulesInput) error {
	queryParams, err := getQueryParamsFromRepoRef(repoRef)
	if err != nil {
		return err
	}

	repoRef = strings.ReplaceAll(repoRef, pathSeparator, encodedPathSeparator)
	uri := fmt.Sprintf("%s/gateway/code/api/v1/migrate/repos/%s/rules?%s",
		c.address,
		repoRef,
		queryParams,
	)
	if err := c.post(uri, in, nil); err != nil {
		return err
	}
	return nil
}

func (c *client) CheckUsers(in *types.CheckUsersInput) (*types.CheckUsersOutput, error) {
	out := new(types.CheckUsersOutput)
	uri := fmt.Sprintf("%s/gateway/code/api/v1/principals/check-emails?routingId=%s&accountIdentifier=%s", c.address, c.account, c.account)

	if err := c.post(uri, in, out); err != nil {
		return nil, err
	}
	return out, nil
}

// http request helper functions
func (c *client) setAuthHeader() func(h *http.Header) {
	return func(h *http.Header) { h.Set("x-api-key", c.token) }
}

// helper function for making an http GET request.
func (c *client) get(rawurl string, out interface{}) error {
	return Do(rawurl, "GET", c.setAuthHeader(), nil, out, c.tracing)
}

// helper function for making an http POST request.
func (c *client) post(rawurl string, in, out interface{}) error {
	return Do(rawurl, "POST", c.setAuthHeader(), in, out, c.tracing)
}

// helper function for making an http PATCH request.
func (c *client) patch(rawurl string, in, out interface{}) error {
	return Do(rawurl, "PATCH", c.setAuthHeader(), in, out, c.tracing)
}

// helper function for making an http DELETE request.
func (c *client) delete(rawurl string) error {
	return Do(rawurl, "DELETE", c.setAuthHeader(), nil, nil, c.tracing)
}
