package util

import (
	"strings"

	"github.com/gotidy/ptr"
	"github.com/harness/harness-migrate/internal/harness"
)

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

// CreateGitlabConnector helper function to create a github connector
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

// IsErrConflict helper function return true if the error message
// indicate the resource already exists.
func IsErrConflict(err error) bool {
	return strings.Contains(err.Error(), "already present") ||
		strings.Contains(err.Error(), "already exists")
}
