package drone

import (
	"context"

	"github.com/drone/go-scm/scm"

	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/migrate/drone/yaml"
	"github.com/harness/harness-migrate/internal/slug"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/util"
)

type Importer struct {
	Harness    harness.Client
	ScmClient  *scm.Client
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
	// create if the secret does not already exist.
	if _, err = m.Harness.FindSecretOrg(org.ID, m.ScmType); err != nil {
		// create the scm secret as an inline secret using
		// the harness secret manager.
		secret := util.CreateSecretOrg(org.ID, m.ScmType, m.ScmToken)
		// save the secret to the organization
		if err := m.Harness.CreateSecretOrg(secret); err != nil {
			return err
		}
	}

	m.Tracer.Stop("create secret %s [done]", m.ScmType)

	m.Tracer.Start("create connector %s", m.ScmType)
	// create if the connector does not already exist.
	if _, err = m.Harness.FindConnectorOrg(org.ID, m.ScmType); err != nil {
		switch m.ScmType {
		case "gitlab":
			conn := util.CreateGitlabConnector(org.ID, m.ScmType, m.ScmLogin, "org."+m.ScmType)
			if err := m.Harness.CreateConnectorOrg(conn); err != nil {
				return err
			}
		default:
			conn := util.CreateGithubConnector(org.ID, m.ScmType, m.ScmLogin, "org."+m.ScmType)
			if err := m.Harness.CreateConnectorOrg(conn); err != nil {
				return err
			}
		}
	}

	m.Tracer.Stop("create connector %s [done]", m.ScmType)

	// convert each drone repo to a harness project.
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
			if !util.IsErrConflict(err) {
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

		// for each environment variable
		for _, srcEnv := range srcProject.Secrets {
			// convert the environment variable to an inline
			// secret, stored in the harness secret manager.
			secret := util.CreateSecret(org.ID, project.Identifier, slug.Create(srcEnv.Name), srcEnv.Desc, srcEnv.Value)
			// save the secret to harness.
			if err := m.Harness.CreateSecret(secret); err != nil {
				// if the error indicates the secret already
				// exists we can continue with the import,
				// else we should return the error and exit
				// the import.
				if !util.IsErrConflict(err) {
					return err
				}
			}
		}

		// convert the drone yaml to a harness yaml
		yamlFile, _, err := m.ScmClient.Contents.Find(ctx, srcProject.Name, srcProject.Yaml, srcProject.Branch)
		if err != nil {
			return err
		}

		conf, err := yaml.Convert(yamlFile.Data)
		if err != nil {
			return err
		}

		// create the harness pipeline with an inline yaml
		err = m.Harness.CreatePipeline(org.ID, project.Identifier, conf)
		if err != nil {
			// if the error indicates the pipeline already
			// exists we can continue with the import, else
			// we should return the error and exit the import.
			if !util.IsErrConflict(err) {
				return err
			}
		}

		m.Tracer.Stop("create project %s [done]", srcProject.Name)
	}
	return nil
}
