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

	convert "github.com/drone/go-convert/convert/drone"
	migrate "github.com/harness/harness-migrate/internal/migrate/drone"
	"github.com/harness/harness-migrate/internal/slug"

	"github.com/drone/go-convert/convert/harness/downgrader"
	"github.com/harness/harness-migrate/cmd/util"
	"github.com/harness/harness-migrate/internal/migrate/drone/repo"
	"github.com/harness/harness-migrate/internal/tracer"

	"github.com/alecthomas/kingpin/v2"
	"github.com/jmoiron/sqlx"
	"golang.org/x/exp/slog"
)

type migrateCommand struct {
	debug bool
	trace bool

	Driver         string
	Datasource     string
	namespace      string
	repositoryList string

	downgrade bool
	KubeName  string
	KubeConn  string

	harnessToken   string
	harnessAccount string
	harnessOrg     string
	harnessAddress string

	githubToken    string
	githubURL      string
	gitlabToken    string
	gitlabURL      string
	bitbucketToken string
	bitbucketURL   string

	repoConn   string
	kubeName   string
	kubeConn   string
	dockerConn string
}

func (c *migrateCommand) run(*kingpin.ParseContext) error {
	// create the logger
	log := util.CreateLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	var db *sqlx.DB
	droneRepo, err := repo.NewRepository(c.Driver, c.Datasource, db)
	if err != nil {
		return err
	}

	// create scm client to verify the token
	// and retrieve the user id.
	client := util.CreateClient(
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
		c.githubURL,
		c.gitlabURL,
		c.bitbucketURL,
	)

	// get the current user id.
	user, _, err := client.Users.Find(ctx)
	if err != nil {
		log.Error("cannot retrieve git user", nil)
		return err
	}

	var repository []string
	if c.repositoryList != "" {
		repository = strings.Split(c.repositoryList, ",")
	}

	tracer_ := tracer.New()
	defer tracer_.Close()

	// extract the data
	exporter := &migrate.Exporter{
		Repository:     droneRepo,
		Namespace:      c.namespace,
		Tracer:         tracer_,
		ScmClient:      client,
		RepositoryList: repository,
	}
	data, err := exporter.Export(ctx)
	if err != nil {
		return err
	}

	// convert all yaml into v1 or v0 format
	converter := convert.New(
		convert.WithDockerhub(c.dockerConn),
		convert.WithKubernetes(c.kubeName, c.kubeConn),
	)
	for _, project := range data.Projects {
		// convert to v1
		convertedYaml, err := converter.ConvertBytes(project.Yaml)
		if err != nil {
			return err
		}
		// downgrade to v0 if needed
		if c.downgrade {
			d := downgrader.New(
				downgrader.WithCodebase(project.Name, c.repoConn),
				downgrader.WithDockerhub(c.dockerConn),
				downgrader.WithKubernetes(c.kubeName, c.kubeConn),
				downgrader.WithName(project.Name),
				downgrader.WithOrganization(c.harnessOrg),
				downgrader.WithProject(slug.Create(project.Name)),
			)
			convertedYaml, err = d.Downgrade(convertedYaml)
			if err != nil {
				return nil
			}
		}
		project.Yaml = convertedYaml
	}

	importer := util.CreateImporter(
		c.harnessAccount,
		c.harnessOrg,
		c.harnessToken,
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
		c.harnessAddress,
	)

	// map the kube namespace and kube connector
	if c.KubeName == "" && c.KubeConn == "" {
		importer.KubeName = c.KubeName
		importer.KubeConn = c.KubeConn
	}

	importer.Tracer = tracer_

	// provide the user id to the importer. the user id
	// is required by the connector despite the fact that
	// it can be retrieved using the token itself (like we just did)
	importer.ScmLogin = user.Login

	log.Debug("verified user and token",
		slog.String("user", user.Login),
	)

	importer.ScmClient = client

	// execute the import routine.
	return importer.Import(ctx, data)
}

func registerMigrate(app *kingpin.CmdClause) {
	c := new(migrateCommand)

	cmd := app.Command("migrate", "migrate drone data to harness").
		Action(c.run)

	cmd.Flag("harness-account", "harness account").
		Required().
		Envar("HARNESS_ACCOUNT").
		StringVar(&c.harnessAccount)

	cmd.Flag("harness-org", "harness organization").
		Required().
		Envar("HARNESS_ORG").
		StringVar(&c.harnessOrg)

	cmd.Flag("harness-token", "harness token").
		Required().
		Envar("HARNESS_TOKEN").
		StringVar(&c.harnessToken)

	cmd.Flag("harness-address", "harness address").
		Envar("HARNESS_ADDRESS").
		Default("https://app.harness.io").
		StringVar(&c.harnessAddress)

	cmd.Flag("repository-list", "optional list of repositories to export").
		Envar("REPOSITORY_LIST").
		StringVar(&c.repositoryList)

	cmd.Flag("github-token", "github token").
		Envar("GITHUB_TOKEN").
		StringVar(&c.githubToken)

	cmd.Flag("github-url", "github url").
		Envar("GITHUB_URL").
		StringVar(&c.githubURL)

	cmd.Flag("gitlab-token", "gitlab token").
		Envar("GITLAB_TOKEN").
		StringVar(&c.gitlabToken)

	cmd.Flag("gitlab-url", "gitlab url").
		Envar("GITLAB_URL").
		StringVar(&c.gitlabURL)

	cmd.Flag("bitbucket-token", "bitbucket token").
		Envar("BITBUCKET_TOKEN").
		StringVar(&c.bitbucketToken)

	cmd.Flag("bitbucket-url", "bitbucket url").
		Envar("BITBUCKET_URL").
		StringVar(&c.bitbucketURL)

	cmd.Flag("downgrade", "downgrade to the legacy yaml format").
		Default("true").
		BoolVar(&c.downgrade)

	cmd.Flag("namespace", "drone namespace").
		Required().
		Envar("DRONE_NAMESPACE").
		StringVar(&c.namespace)

	cmd.Flag("driver", "drone db type").
		Default("sqlite3").
		StringVar(&c.Driver)

	cmd.Flag("datasource", "drone database datasource").
		Envar("DRONE_DATABASE_DATASOURCE").
		Default("database.sqlite3").
		StringVar(&c.Datasource)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)

	cmd.Flag("kube-connector", "kubernetes connector").
		Envar("KUBE_CONN").
		StringVar(&c.kubeConn)

	cmd.Flag("kube-namespace", "kubernetes namespace").
		Envar("KUBE_NAMESPACE").
		StringVar(&c.kubeName)

	cmd.Flag("docker-connector", "dockerhub connector").
		Default("").
		StringVar(&c.dockerConn)

	cmd.Flag("repo-connector", "repository connector").
		Default("").
		StringVar(&c.repoConn)
}
