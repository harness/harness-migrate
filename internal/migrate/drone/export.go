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

package drone

import (
	"context"
	"strings"

	"github.com/drone/go-scm/scm"

	"github.com/harness/harness-migrate/internal/migrate/drone/repo"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
)

// Exporter exports data from drone.
type Exporter struct {
	Repository     repo.Repository
	Namespace      string
	Downgrade      bool
	RepositoryList []string

	RepoConn   string
	DockerConn string
	KubeName   string
	KubeConn   string
	Org        string

	ScmClient *scm.Client
	ScmLogin  string

	Tracer tracer.Tracer
}

func (m *Exporter) Export(ctx context.Context) (*types.Org, error) {

	m.Tracer.Start("starting export organization")

	//retrieve the list of repositories and checking it exists
	repos, err := m.Repository.GetRepos(ctx, m.Namespace)
	if err != nil {
		m.Tracer.Stop("Failed to retrieve repositories: %s", err.Error())
		return nil, err
	}
	if len(repos) == 0 {
		m.Tracer.Stop("No repositories found for namespace: %s", m.Namespace)
		return nil, err
	}

	dstOrg := &types.Org{
		Name: repos[0].Namespace,
	}

	m.Tracer.Start("export organization secrets")
	orgSecrets, err := m.Repository.GetOrgSecrets(ctx, m.Namespace)
	if err != nil {
		m.Tracer.Stop("Failed to export organization secrets: %s", err.Error())
		return nil, err
	}
	// map org secrets to common format
	dstOrg.Secrets = convertOrgSecretsToSecrets(orgSecrets)

	m.Tracer.Stop("export organization secrets [done]")
	m.Tracer.Stop("export organization %s [done]", repos[0].Namespace)
	// convert each repository to a harness project
	for _, repo := range repos {
		// Skip repositories that are not in the specified namespace
		if !strings.HasPrefix(repo.Namespace, m.Namespace) {
			m.Tracer.Log("Skipping repository %s: not in specified namespace %s.", repo.Name, m.Namespace)
			continue
		}

		// Skip repositories that are not in the m.RepositoryList
		if len(m.RepositoryList) > 0 && !m.repositoryInList(repo.Name) {
			m.Tracer.Log("Skipping repository %s: not in the provided repository list.", repo.Name)
			continue
		}

		m.Tracer.Start("export project %s", repo.Name)

		//get the latest build for the default branch
		build, _ := m.Repository.LatestBuild(ctx, repo.ID)
		if build == nil {
			m.Tracer.Log("Skipping repository %s: no builds in pipeline.", repo.Name)
			continue
		}
		// convert the Drone repository to a common format
		dstProject := &types.Project{
			Name:   repo.Name,
			Type:   "drone",
			Repo:   repo.CloneURL,
			Branch: repo.Branch,
		}

		yamlFile, _, err := m.ScmClient.Contents.Find(ctx, repo.Slug, repo.Config, repo.Branch)
		if err != nil {
			m.Tracer.Log("Skipping repository %s: no .drone.yml file found.", repo.Name)
			continue
		}

		dstProject.Yaml = yamlFile.Data

		// find Drone secrets
		secrets, secretErr := m.Repository.GetSecrets(ctx, repo.ID)
		if secretErr != nil {
			m.Tracer.Log("Skipping repository %s: failed to retrieve secrets.", repo.Name)
			continue
		}

		// for each secret
		for _, secret := range secrets {
			dstSecret := &types.Secret{
				Name:  secret.Name,
				Desc:  secret.Name,
				Value: secret.Data,
			}
			// append the secret to the project
			dstProject.Secrets = append(dstProject.Secrets, dstSecret)
		}

		// append the project to the list of projects
		dstOrg.Projects = append(dstOrg.Projects, dstProject)

		m.Tracer.Stop("export project %s [done]", repo.Name)
	}
	return dstOrg, nil
}

func convertOrgSecretsToSecrets(orgSecrets []*repo.OrgSecret) []*types.Secret {
	secrets := make([]*types.Secret, len(orgSecrets))
	for i, orgSecret := range orgSecrets {
		secret := &types.Secret{
			Name:  orgSecret.Name,
			Value: string(orgSecret.Data),
		}
		secrets[i] = secret
	}
	return secrets
}

func (m *Exporter) repositoryInList(repoName string) bool {
	lowerRepoName := strings.ToLower(repoName)
	for _, name := range m.RepositoryList {
		if strings.ToLower(name) == lowerRepoName {
			return true
		}
	}
	return false
}
