package gitimporter

import (
	"fmt"

	"github.com/harness/harness-migrate/internal/common"
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
	tracer.Start(common.MsgStartImportCreateRepo, repo.Name)
	in := &harness.CreateRepositoryForMigrateInput{
		Identifier:    repo.Name,
		DefaultBranch: repo.Branch,
		IsPublic:      !repo.Private,
		ParentRef:     targetSpace,
	}

	repoOut, err := m.Harness.CreateRepositoryForMigration(in)
	if err != nil {
		tracer.Stop("failed to create repository %q", repo.Name)
		return nil, fmt.Errorf("failed to create a repo '%s' at %s: %w",
			repo.Name, targetSpace, err)
	}

	m.Tracer.Stop(common.MsgCompleteImportCreateRepo, repo.Name, targetSpace)
	return repoOut, nil
}
