// Copyright 2023 Harness Inc. All rights reserved.

package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

type client struct {
	address string
	account string
	token   string
	tracing bool
}

// New returns a new Client.
func New(account, token string, opts ...Option) Client {
	client_ := &client{
		account: account,
		token:   token,
	}
	// set optional parameters.
	for _, opt := range opts {
		opt(client_)
	}
	// set defualt address if not provided.
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

//
// http request helper functions
//

// helper function for making an http GET request.
func (c *client) get(rawurl string, out interface{}) error {
	return c.do(rawurl, "GET", nil, out)
}

// helper function for making an http POST request.
func (c *client) post(rawurl string, in, out interface{}) error {
	return c.do(rawurl, "POST", in, out)
}

// helper function for making an http PATCH request.
func (c *client) patch(rawurl string, in, out interface{}) error {
	return c.do(rawurl, "PATCH", in, out)
}

// helper function for making an http DELETE request.
func (c *client) delete(rawurl string) error {
	return c.do(rawurl, "DELETE", nil, nil)
}

// helper function to make an http request
func (c *client) do(rawurl, method string, in, out interface{}) error {
	body, err := c.open(rawurl, method, in, out)
	if err != nil {
		return err
	}
	defer body.Close()
	if out != nil {
		return json.NewDecoder(body).Decode(out)
	}
	return nil
}

// helper function to open an http request
func (c *client) open(rawurl, method string, in, out interface{}) (io.ReadCloser, error) {
	uri, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, uri.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.token)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "curl/7.79.1")
	if in != nil {
		if buf, ok := in.(*bytes.Buffer); ok {
			req.Body = ioutil.NopCloser(buf)
			req.ContentLength = int64(buf.Len())
		} else {
			decoded, derr := json.Marshal(in)
			if derr != nil {
				return nil, derr
			}
			buf := bytes.NewBuffer(decoded)
			req.Body = ioutil.NopCloser(buf)
			req.ContentLength = int64(len(decoded))
			req.Header.Set("Content-Length", strconv.Itoa(len(decoded)))
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// if tracing enabled, dump the request body.
	if c.tracing {
		dump, _ := httputil.DumpRequest(req, true)
		os.Stdout.Write(dump)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// if tracing enabled, dump the response body.
	if c.tracing {
		dump, _ := httputil.DumpResponse(resp, true)
		os.Stdout.Write(dump)
	}

	if resp.StatusCode > 299 {
		defer resp.Body.Close()
		out, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("client error %d: %s", resp.StatusCode, string(out))
	}
	return resp.Body, nil
}
