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

package migrate

import (
	"context"
	"strings"

	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/slug"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/util"
)

type Importer struct {
	Harness    harness.Client
	HarnessOrg string

	RepositoryList []string

	KubeName string
	KubeConn string

	Tracer tracer.Tracer
}

const dockerConnectorName = "docker"

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

	m.Tracer.Start("create organisation secrets if they exist")
	// create org secrets
	for _, secret := range data.Secrets {
		if _, err = m.Harness.FindSecretOrg(org.ID, secret.Name); err != nil {
			s := util.CreateSecretOrg(org.ID, secret.Name, secret.Value)
			// save the secret to the organization
			if err := m.Harness.CreateSecretOrg(s); err != nil {
				return err
			}
		}
	}

	m.Tracer.Stop("create organisation secrets [done]")

	m.Tracer.Start("create docker connector")

	dockerConnector, _ := m.Harness.FindConnectorOrg(org.ID, dockerConnectorName)
	if dockerConnector == nil {
		conn := util.CreateDockerConnector(org.ID, dockerConnectorName)
		if err := m.Harness.CreateConnectorOrg(conn); err != nil {
			return err
		}
	}

	m.Tracer.Stop("create docker connector [done]")

	// convert each drone repo to a harness project.
	for _, srcProject := range data.Projects {

		// Skip repositories that are not in the m.RepositoryList
		if len(m.RepositoryList) > 0 && !m.repositoryInList(srcProject.Name) {
			continue
		}

		m.Tracer.Start("create project %s", srcProject.Name)
		projectSlug := slug.Create(srcProject.Name)

		// convert the project to a harness project
		// structure and convert the project name to
		// a harness project identifier.
		project := &harness.Project{
			Orgidentifier: org.ID,
			Identifier:    projectSlug,
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
			m.Harness, m.HarnessOrg, projectSlug); err != nil {
			return err
		}

		// for each environment variable
		for _, srcEnv := range srcProject.Secrets {
			// convert the environment variable to an inline
			// secret, stored in the harness secret manager.
			secret := util.CreateSecret(org.ID, projectSlug, slug.Create(srcEnv.Name), srcEnv.Desc, srcEnv.Value)
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

		//create the harness pipeline with an inline yaml
		err = m.Harness.CreatePipeline(org.ID, projectSlug, srcProject.Yaml)
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

func (m *Importer) repositoryInList(repoName string) bool {
	lowerRepoName := strings.ToLower(repoName)
	for _, name := range m.RepositoryList {
		if strings.ToLower(name) == lowerRepoName {
			return true
		}
	}
	return false
}
