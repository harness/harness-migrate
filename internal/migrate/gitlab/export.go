// Copyright 2023 Harness Inc. All rights reserved.

package gitlab

import (
	"context"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/traverse"

	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
)

// Exporter exports data from Circle.
type Exporter struct {
	Gitlab    *scm.Client
	GitlabOrg string

	Tracer tracer.Tracer
}

// Export exports Circle data.
func (m *Exporter) Export(ctx context.Context) (*types.Org, error) {

	m.Tracer.Start("export organization")

	// convert the gitlab org to the common format.
	dstOrg := &types.Org{
		Name: m.GitlabOrg,
	}

	// find the gitlab org
	srcOrg, _, err := m.Gitlab.Organizations.Find(ctx, m.GitlabOrg)
	if err != nil {
		// the organization may be a user account
		_, _, userErr := m.Gitlab.Users.FindLogin(ctx, m.GitlabOrg)
		if userErr != nil {
			return nil, err
		}
	}

	m.Tracer.Stop("export organization %s [done]", srcOrg.Name)

	// retrieve a list of all circle projects in the organization.
	// use the "traverse" helper to paginate and return the full list.
	srcRepos, err := traverse.Repos(ctx, m.Gitlab)
	if err != nil {
		return nil, err
	}

	// convert each gitlab repository to a harness project.
	for _, srcRepo := range srcRepos {

		// skip if the repository does not match
		if srcRepo.Namespace != dstOrg.Name {
			continue
		}

		m.Tracer.Start("export repository %s", srcRepo.Name)

		// convert the circle project to a common format.
		dstProject := &types.Project{
			Name: srcRepo.Name,
			Repo: srcRepo.Clone,
		}

		// append projects to the org
		dstOrg.Projects = append(dstOrg.Projects, dstProject)

		m.Tracer.Stop("export repository %s [done]", srcRepo.Name)
	}

	return dstOrg, nil
}
