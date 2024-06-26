package gitimporter

import (
	"fmt"

	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/types"

	"github.com/harness/harness-migrate/internal/tracer"
)

func (m *Importer) CreateRepo(
	repo types.Repository,
	targetSpace string,
	tracer tracer.Tracer,
) (*harness.Repository, error) {
	// TODO: check license for size limit
	tracer.Start("create repository %q", repo.Name)
	repoCreate := &harness.RepositoryCreateRequest{
		UID:           repo.Name,
		DefaultBranch: repo.Branch,
		Description:   "Imported by the migrator", //TODO: get the original repo description
		IsPublic:      !repo.Private,
	}

	repoOut, err := m.Harness.CreateRepository(targetSpace, repoCreate)
	if err != nil {
		tracer.Stop("failed to create repository %q", repo.Name)
		return nil, fmt.Errorf("failed to create a repo '%s' at %s: %w",
			repo.Name, targetSpace, err)
	}

	m.Tracer.Stop("create repo %q at %s [done]", repoOut.UID, targetSpace)
	return repoOut, nil
}
