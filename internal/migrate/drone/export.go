package drone

import (
	"context"
	"fmt"
	"strings"

	"github.com/harness/harness-migrate/internal/migrate/drone/repo"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
)

// Exporter exports data from drone.
type Exporter struct {
	Repository repo.Repository
	Namespace  string

	Tracer tracer.Tracer
}

func (m *Exporter) Export(ctx context.Context) (*types.Org, error) {

	m.Tracer.Start("export organization")

	//retrieve the list of repositories and checking it exists
	repos, err := m.Repository.GetRepos(ctx, m.Namespace)
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories found for %s", m.Namespace)
	}

	dstOrg := &types.Org{
		Name: repos[0].Namespace,
	}

	m.Tracer.Start("export organization secrets")
	orgSecrets, err := m.Repository.GetOrgSecrets(ctx, m.Namespace)
	if err != nil {
		return nil, err
	}
	// map org secrets to common format
	dstOrg.Secrets = convertOrgSecretsToSecrets(orgSecrets)

	m.Tracer.Stop("export organization secrets [done]")
	m.Tracer.Stop("export organization %s [done]", repos[0].Name)

	// convert each repository to a harness project
	for _, repo := range repos {
		// Skip repositories that are not in the specified namespace
		if !strings.HasPrefix(repo.Namespace, m.Namespace) {
			continue
		}

		m.Tracer.Start("export project %s", repo.Name)

		//get the latest build for the default branch
		build, _ := m.Repository.LatestBuild(ctx, repo.ID)
		if build == nil {
			continue
		}
		// convert the Drone repository to a common format
		dstProject := &types.Project{
			Name:   repo.Name,
			Type:   "drone",
			Repo:   repo.CloneURL,
			Branch: repo.Branch,
			Yaml:   repo.Config,
		}

		// find Drone secrets
		secrets, secretErr := m.Repository.GetSecrets(ctx, repo.ID)
		if secretErr != nil {
			return nil, secretErr
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