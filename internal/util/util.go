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

package util

import (
	"strings"

	"github.com/gotidy/ptr"
	"github.com/harness/harness-migrate/internal/harness"
)

type DockerConnectorOptions struct {
	Username string
	Password string
	URL      string
}

func CreateSecret(org, project, identifier, desc, data string) *harness.Secret {
	return &harness.Secret{
		Name:              identifier,
		Identifier:        identifier,
		Orgidentifier:     org,
		Projectidentifier: project,
		Description:       desc,
		Type:              "SecretText",
		Spec: &harness.SecretText{
			Value:   ptr.String(data),
			Type:    "Inline",
			Manager: "harnessSecretManager",
		},
	}
}

// CreateSecretOrg helper function to create an org secret.
func CreateSecretOrg(org, identifier, data string) *harness.Secret {
	return CreateSecret(org, "", identifier, "", data)
}

// CreateGithubConnector helper function to create a github connector
func CreateGithubConnector(org, id, username, token string) *harness.Connector {
	return &harness.Connector{
		Name:          id,
		Identifier:    id,
		Orgidentifier: org,
		Type:          "Github",
		Spec: &harness.ConnectorGithub{
			Type: "Account",
			URL:  "https://github.com",
			Authentication: &harness.Resource{
				Type: "Http",
				Spec: &harness.Resource{
					Type: "UsernameToken",
					Spec: &harness.ConnectorToken{
						Username: username,
						Tokenref: token,
					},
				},
			},
			Apiaccess: &harness.Resource{
				Type: "Token",
				Spec: &harness.ConnectorToken{
					Tokenref: token,
				},
			},
		},
	}
}

// CreateGitlabConnector helper function to create a GitHub connector
func CreateGitlabConnector(org, id, username, token string) *harness.Connector {
	return &harness.Connector{
		Name:          id,
		Identifier:    id,
		Orgidentifier: org,
		Type:          "Gitlab",
		Spec: &harness.ConnectorGitlab{
			Type: "Account",
			URL:  "https://gitlab.com",
			Authentication: &harness.Resource{
				Type: "Http",
				Spec: &harness.Resource{
					Type: "UsernameToken",
					Spec: &harness.ConnectorToken{
						Username: username,
						Tokenref: token,
					},
				},
			},
			Apiaccess: &harness.Resource{
				Type: "Token",
				Spec: &harness.ConnectorToken{
					Tokenref: token,
				},
			},
		},
	}
}

// CreateDockerConnector helper function to create a docker connector
func CreateDockerConnector(org, id string, args ...interface{}) *harness.Connector {
	var authentication *harness.Resource
	var options DockerConnectorOptions
	if len(args) > 0 {
		if arg, ok := args[0].(DockerConnectorOptions); ok {
			options = arg
		}
	}
	// Check if username and password are provided
	if options.Username != "" && options.Password != "" {
		authentication = &harness.Resource{
			Type: "UsernamePassword",
			Spec: map[string]string{
				"username":    options.Username,
				"passwordRef": options.Password,
			},
		}
	} else {
		authentication = &harness.Resource{
			Type: "Anonymous",
		}
	}
	if options.URL == "" {
		options.URL = "https://registry.hub.docker.com/v2/"
	}
	return &harness.Connector{
		Name:          id,
		Identifier:    id,
		Orgidentifier: org,
		Type:          "DockerRegistry",
		Spec: &harness.ConnectorDocker{
			ExecuteOnDelegate: false,
			DockerRegistryURL: options.URL,
			ProviderType:      "DockerHub",
			Authentication:    authentication,
		},
	}
}

// IsErrConflict helper function return true if the codeerror message
// indicate the resource already exists.
func IsErrConflict(err error) bool {
	return strings.Contains(err.Error(), "already present") ||
		strings.Contains(err.Error(), "already exists")
}
