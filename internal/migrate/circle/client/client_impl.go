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
	"strings"
)

type client struct {
	tracing bool
	address string
	token   string
}

// New returns a new Client.
func New(token string, opts ...Option) Client {
	client_ := &client{
		token: token,
	}
	// set optional parameters.
	for _, opt := range opts {
		opt(client_)
	}
	// set default address if not provided.
	if client_.address == "" {
		client_.address = "https://circleci.com"
	}
	return client_
}

// FindOrg returns an organization by slug.
func (c *client) FindOrg(slug string) (*Org, error) {
	slug = strings.ReplaceAll(slug, "/", "%2f")
	out := new(Org)
	uri := fmt.Sprintf("%s/api/private/organization/%s", c.address, slug)
	err := c.get(uri, &out)
	return out, err
}

// FindOrgID returns an organization by ID.
func (c *client) FindOrgID(id string) (*Org, error) {
	// Circle does not provide an endpoint for this,
	// so we fetch the list of organization membership
	// and find an item in the list with a matching ID.
	orgs, err := c.ListOrgs()
	if err != nil {
		return nil, err
	}
	for _, org := range orgs {
		if org.ID == id {
			return org, nil
		}
	}
	return nil, errors.New("circleci: organization not found")
}

// FindProject returns a project by slug.
func (c *client) FindProject(slug string) (*Project, error) {
	out := new(Project)
	uri := fmt.Sprintf("%s/api/v2/project/%s", c.address, slug)
	err := c.get(uri, &out)
	return out, err
}

// FindPipeline returns a pipeline by slug.
func (c *client) FindPipeline(slug string) (*Pipeline, error) {
	return nil, errors.New("circleci: not implemented")
}

// FindConfig returns a pipeline configuration by
// pipeline id.
func (c *client) FindConfig(id string) (*Config, error) {
	out := new(Config)
	uri := fmt.Sprintf("%s/api/v2/pipeline/%s/config", c.address, id)
	err := c.get(uri, &out)
	return out, err
}

// ListOrgs returns the user organization list.
func (c *client) ListOrgs() ([]*Org, error) {
	var tmp []*Collab
	var out []*Org
	uri := fmt.Sprintf("%s/api/v2/me/collaborations", c.address)
	err := c.get(uri, &tmp)
	for _, src := range tmp {
		dst := &Org{
			ID:   src.ID,
			Slug: src.Slug,
			Name: src.Name,
			Type: src.Type,
		}
		out = append(out, dst)
	}
	return out, err
}

// ListProjects returns a project list by organization id.
func (c *client) ListProjects(id string) ([]*Project, error) {
	out := new(ProjectList)
	uri := fmt.Sprintf("%s/api/private/project?organization-id=%s", c.address, id)
	err := c.get(uri, &out)
	return out.Items, err
}

// ListPipelines returns a pipeline list by pipeline slug.
func (c *client) ListPipelines(slug string) ([]*Pipeline, error) {
	out := new(PipelineList)
	uri := fmt.Sprintf("%s/api/v2/project/%s/pipeline", c.address, slug)
	err := c.get(uri, &out)
	return out.Items, err
}

// ListEnvs returns an environment variable list by
// pipeline slug.
func (c *client) ListEnvs(slug string) ([]*Env, error) {
	out := new(EnvList)
	uri := fmt.Sprintf("%s/api/v2/project/%s/envvar", c.address, slug)
	err := c.get(uri, &out)
	return out.Items, err
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
	req.Header.Set("Circle-Token", c.token)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "curl/7.79.1")
	if in != nil {
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
	if c.tracing {
		dump, _ := httputil.DumpRequest(req, true)
		os.Stdout.Write(dump)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
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

// helper function converts a map to url.Values
func mapValues(params map[string]string) url.Values {
	values := url.Values{}
	for key, val := range params {
		values.Add(key, val)
	}
	return values
}
