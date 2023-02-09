// Copyright 2023 Harness Inc. All rights reserved.

package circle

import (
	"context"
	"fmt"

	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/commons"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter"
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
		switch m.ScmType {
		case "gitlab":
			conn := createGitlabConnector(org.ID, m.ScmType, m.ScmLogin, "org."+m.ScmType)
			if err := m.Harness.CreateConnectorOrg(conn); err != nil {
				return err
			}
		default:
			conn := createGithubConnector(org.ID, m.ScmType, m.ScmLogin, "org."+m.ScmType)
			if err := m.Harness.CreateConnectorOrg(conn); err != nil {
				return err
			}
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
			return err
		}

		// wait for the harness secret manager to be created for the
		// project. It is created async and if we do not wait, it
		// could result in failure to add secrets in subsequent steps.
		if err := harness.WaitHarnessSecretManager(
			m.Harness, m.HarnessOrg, project.Identifier); err != nil {
			return err
		}

		// for each environment variable
		for _, srcEnv := range srcProject.Secrets {
			// convert the environment variable to an inline
			// secret, stored in the harness secret manager.
			secret := createSecret(org.ID, project.Identifier, slug.Create(srcEnv.Name), srcEnv.Desc, srcEnv.Value)
			// save the secret to harness.
			if err := m.Harness.CreateSecret(secret); err != nil {
				return err
			}
		}

		// convert the circle yaml to a harness yaml
		opts := commons.Opts{}
		conf, err := converter.Convert(opts, []byte(srcProject.Yaml))
		if err != nil {
			return err
		}

		// TODO remove this once conversion works as expected
		conf = []byte(fmt.Sprintf(
			string(dummy),
			project.Identifier,
			project.Identifier,
			project.Identifier,
			project.Orgidentifier,
		))

		// create the harness pipeline with an inline yaml
		err = m.Harness.CreatePipeline(org.ID, project.Identifier, conf)
		if err != nil {
			return err
		}

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
func createGithubConnector(org, id, username, token string) *harness.Connector {
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

var dummy = `
pipeline:
  name: %s
  identifier: %s
  projectIdentifier: %s
  orgIdentifier: %s
  tags: {}
  stages:
    - stage:
        name: test
        identifier: test
        description: ""
        type: CI
        spec:
          cloneCodebase: false
          platform:
            os: Linux
            arch: Amd64
          runtime:
            type: Cloud
            spec: {}
          execution:
            steps:
              - step:
                  type: Run
                  name: Run_1
                  identifier: Run_1
                  spec:
                    shell: Sh
                    command: echo hello world
`
