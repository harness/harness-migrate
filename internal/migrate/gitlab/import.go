// Copyright 2023 Harness Inc. All rights reserved.

package gitlab

import (
	"context"
	"strings"

	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/slug"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/gotidy/ptr"
)

// Importer imports data from Circle to Harness.
type Importer struct {
	Harness    harness.Client
	HarnessOrg string

	ScmType  string // github, gitlab, bitbucket
	ScmLogin string
	ScmToken string

	Tracer tracer.Tracer
}

func (m *Importer) Import(ctx context.Context, data *types.Org) error {

	m.Tracer.Start("create organization %s", m.HarnessOrg)

	// find the harness organization
	org, err := m.Harness.FindOrg(m.HarnessOrg)
	if err != nil {
		org = &harness.Org{
			ID:   m.HarnessOrg,
			Name: m.HarnessOrg,
		}
		// create the organization if not exists
		if err := m.Harness.CreateOrg(org); err != nil {
			return err
		}
	}

	// wait for the harness secret manager to be created for the
	// organization. It is created async and if we do not wait, it
	// could result in failure to add secrets in subsequent steps.
	if err := harness.WaitHarnessSecretManagerOrg(m.Harness, m.HarnessOrg); err != nil {
		return err
	}

	m.Tracer.Stop("create organization %s [done]", m.HarnessOrg)

	m.Tracer.Start("create secret %s", m.ScmType)

	// find the github, gitlab or bitbucket secret or
	// create if the secret does not already exist.
	if _, err = m.Harness.FindSecretOrg(org.ID, m.ScmType); err != nil {
		// create the scm secret as an inline secret using
		// the harness secret manager.
		secret := createSecretOrg(org.ID, m.ScmType, m.ScmToken)
		// save the secret to the organization
		if err := m.Harness.CreateSecretOrg(secret); err != nil {
			return err
		}
	}

	m.Tracer.Stop("create secret %s [done]", m.ScmType)

	m.Tracer.Start("create connector %s", m.ScmType)

	// find the github, gitlab or bitbucket connector or
	// create if the connector does not already exist.
	if _, err = m.Harness.FindConnectorOrg(org.ID, m.ScmType); err != nil {
		conn := createGitlabConnector(org.ID, m.ScmType, m.ScmLogin, "org."+m.ScmType)
		if err := m.Harness.CreateConnectorOrg(conn); err != nil {
			return err
		}
	}

	m.Tracer.Stop("create connector %s [done]", m.ScmType)

	// convert each circle project to a harness project.
	for _, srcProject := range data.Projects {

		m.Tracer.Start("create project %s", srcProject.Name)

		// convert the circle project to a harness project
		// structure and convert the circle project name to
		// a harness project identifier.
		project := &harness.Project{
			Orgidentifier: org.ID,
			Identifier:    slug.Create(srcProject.Name),
			Name:          srcProject.Name,
		}

		// create the harness project.
		if err := m.Harness.CreateProject(project); err != nil {
			// if the error indicates the project already exists
			// we can continue with the import, else we should return
			// the error and exit the import.
			if isErrConflict(err) == false {
				return err
			}
		}

		// wait for the harness secret manager to be created for the
		// project. It is created async and if we do not wait, it
		// could result in failure to add secrets in subsequent steps.
		if err := harness.WaitHarnessSecretManager(
			m.Harness, m.HarnessOrg, project.Identifier); err != nil {
			return err
		}

		// TODO(johannesHarness) create the repository in Harness

		// TODO(johannesHarness) clone the repository

		// TODO(johannesHarness) push the repository to harness

		// sample code to clone a repository using the native Go
		// git library.
		// https://gist.github.com/bradrydzewski/9cff9df374840cfee74eee25772dec49

		m.Tracer.Stop("create project %s [done]", srcProject.Name)
	}

	return nil
}

//
// helper functions to simplify complex resource creation.
//

// helper function to create a secret.
func createSecret(org, project, identifier, desc, data string) *harness.Secret {
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

// helper function to create an org secret.
func createSecretOrg(org, identifier, data string) *harness.Secret {
	return createSecret(org, "", identifier, "", data)
}

// helper function to create a github connector
func createGitlabConnector(org, id, username, token string) *harness.Connector {
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

// helper function return true if the error message
// indicate the resource already exists.
func isErrConflict(err error) bool {
	return strings.Contains(err.Error(), "already present") ||
		strings.Contains(err.Error(), "already exists")
}
