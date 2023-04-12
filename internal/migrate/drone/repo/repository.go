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
