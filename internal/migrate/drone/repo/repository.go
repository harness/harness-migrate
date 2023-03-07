package repo

import (
	"context"
)

// Repository provides access to the Drone database.
type Repository interface {
	// GetRepos returns the list of all repositories in the specified namespace.
	GetRepos(ctx context.Context, namespace string) ([]*Repo, error)

	// LatestBuild returns the last build for the specified repository
	LatestBuild(ctx context.Context, repoId int64) (*Build, error)

	// GetSecrets returns the list of secrets for the specified repository.
	GetSecrets(ctx context.Context, repoID int64) ([]*Secret, error)

	GetOrgSecrets(ctx context.Context, namespace string) ([]*OrgSecret, error)
}
