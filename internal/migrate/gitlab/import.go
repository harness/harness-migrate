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

package gitlab

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/slug"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/util"

	git "github.com/go-git/go-git/v5"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/gotidy/ptr"
)

// Importer imports data from gitlab to Harness.
type Importer struct {
	Harness      harness.Client
	HarnessOrg   string
	HarnessToken string

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
		conn := createGitlabConnector(org.ID, m.ScmType, m.ScmLogin, "org."+m.ScmType)
		if err := m.Harness.CreateConnectorOrg(conn); err != nil {
			return err
		}
	}

	m.Tracer.Stop("create connector %s [done]", m.ScmType)

	// create tmp dir for cloning repos
	tmpDir, err := os.MkdirTemp("", "harness-migrate-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir for org: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// convert each gitlab project to a harness project.
	for _, srcProject := range data.Projects {

		m.Tracer.Start("create project %s", srcProject.Name)

		// convert the gitlab project to a harness project
		// structure and convert the gitlab project name to
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
			if isErrConflict(err) == false {
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

		m.Tracer.Stop("create project %s [done]", srcProject.Name)

		m.Tracer.Start("create repository %s", project.Identifier)

		// create the harness repository in harness
		repoCreate := &harness.CreateRepositoryInput{
			Identifier:    project.Identifier,
			DefaultBranch: srcProject.Branch,
			IsPublic:      false, // TODO: Harness doesn't have private repos at the moment
		}

		repoRef := util.JoinPaths(project.Orgidentifier, project.Identifier)
		repo, err := m.Harness.CreateRepository(repoRef, repoCreate)
		if err != nil {
			// if the error indicates the project already exists, continue with next project.
			// This is a temporary workaround to avoid conflicts while pushing the git repo.
			// TODO: Handle conflicts properly, this only works because repo migration is the last step of a project migration.
			if isErrConflict(err) {
				m.Tracer.Stop("create repository %s [done]", project.Identifier)
				continue
			}
		}

		m.Tracer.Stop("create repository %s [done]", project.Identifier)

		m.Tracer.Start("clone git repository %s", project.Identifier)

		// create tmp dir for repo clone (use generated name to avoid issues with invalid chars)
		// NOTE: no extra clean-up required - tmpDir is already being cleaned-up.
		tmpRepoDir, err := os.MkdirTemp(tmpDir, "repo-*.git")
		if err != nil {
			return fmt.Errorf("failed to create tempo dir for repo: %w", err)
		}
		gitRepo, err := git.PlainCloneContext(ctx, tmpRepoDir, true, &git.CloneOptions{
			URL: srcProject.Repo,
			Auth: &http.BasicAuth{
				Username: m.ScmLogin,
				Password: m.ScmToken,
			},
			ReferenceName: plumbing.NewBranchReferenceName(srcProject.Branch),
			SingleBranch:  false,
			Tags:          git.AllTags,
			NoCheckout:    true,
		})
		if err != nil {
			return fmt.Errorf("failed to clone repo from '%s': %w", srcProject.Repo, err)
		}

		m.Tracer.Stop("clone git repository %s [done]", project.Identifier)

		m.Tracer.Start("push git repository %s", project.Identifier)

		// add empty harness repo as remote
		const gitRemoteHarness = "harness"
		gitRepo.CreateRemote(&config.RemoteConfig{
			Name: gitRemoteHarness,
			URLs: []string{repo.GitURL},
		})

		// push repo
		err = gitRepo.PushContext(ctx, &git.PushOptions{
			RemoteName: gitRemoteHarness,
			Auth: &http.BasicAuth{
				Username: "git",
				Password: m.HarnessToken,
			},
			RefSpecs: []config.RefSpec{
				config.RefSpec("refs/remotes/origin/*:refs/heads/*"),
				config.RefSpec("refs/tags/*:refs/tags/*")},
		})
		if err != nil {
			return fmt.Errorf("failed to push repo to '%s': %w", repo.GitURL, err)
		}

		m.Tracer.Stop("push git repository %s [done]", project.Identifier)
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

// helper function return true if the error message
// indicate the resource already exists.
func isErrConflict(err error) bool {
	return strings.Contains(err.Error(), "already present") ||
		strings.Contains(err.Error(), "already exists")
}
