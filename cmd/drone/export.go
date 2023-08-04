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
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/harness/harness-migrate/cmd/util"
	"github.com/harness/harness-migrate/internal/migrate/drone"
	"github.com/harness/harness-migrate/internal/migrate/drone/repo"
	"github.com/harness/harness-migrate/internal/tracer"

	"github.com/alecthomas/kingpin/v2"
	"github.com/jmoiron/sqlx"
	"golang.org/x/exp/slog"
)

type exportCommand struct {
	debug bool
	trace bool
	file  string

	Driver         string
	Datasource     string
	namespace      string
	repositoryList string

	githubToken    string
	githubURL      string
	gitlabToken    string
	gitlabURL      string
	bitbucketToken string
	bitbucketURL   string
	skipVerify     bool
}

func (c *exportCommand) run(*kingpin.ParseContext) error {
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
	// create the tracer
	log.Info("Creating new tracer...")
	tracer_ := tracer.New()
	defer tracer_.Close()

	log.Info("Creating new client...")
	client := util.CreateClient(
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
		c.githubURL,
		c.gitlabURL,
		c.bitbucketURL,
		c.skipVerify,
	)

	if c.githubToken == "" && c.gitlabToken == "" && c.bitbucketToken == "" {
		return errors.New("no scm token provided")
	}

	log.Info("Finding user...")
	user, _, err := client.Users.Find(ctx)
	if err != nil {
		log.Error("Cannot retrieve git user: ", err)
		return err
	}

	var repository []string
	if c.repositoryList != "" {
		repository = strings.Split(c.repositoryList, ",")
	}

	log.Info("Extracting data...")

	// extract the data
	exporter := &drone.Exporter{
		Repository:     droneRepo,
		Namespace:      c.namespace,
		Tracer:         tracer_,
		ScmClient:      client,
		ScmLogin:       user.Login,
		RepositoryList: repository,
	}
	data, err := exporter.Export(ctx)
	if err != nil {
		log.Error("Failed to extract data: ", err)
		return err
	}

	//if no file path is provided, write the data export
	//to stdout.
	if c.file == "" {
		log.Info("Writing data to stdout...")
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	log.Info("Writing data to file...")

	// else write the data export to the file.
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Error("Failed to write data to file: ", err)
		return err
	}

	log.Info("Data exported successfully.")
	return os.WriteFile(c.file, file, 0644)
}

// helper function registers the export command
func registerExport(app *kingpin.CmdClause) {
	c := new(exportCommand)

	cmd := app.Command("export", "export drone data").
		Action(c.run)

	cmd.Arg("save", "save the output to a file").
		StringVar(&c.file)

	cmd.Flag("namespace", "drone namespace").
		Required().
		Envar("DRONE_NAMESPACE").
		StringVar(&c.namespace)

	cmd.Flag("repository-list", "optional list of repositories to export").
		Envar("REPOSITORY_LIST").
		StringVar(&c.repositoryList)

	cmd.Flag("driver", "drone db type").
		Default("sqlite3").
		StringVar(&c.Driver)

	cmd.Flag("datasource", "drone database datasource").
		Envar("DRONE_DATABASE_DATASOURCE").
		Default("database.sqlite3").
		StringVar(&c.Datasource)

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

	cmd.Flag("skip-tls-verify", "skip TLS verification for SCM").
		Envar("SKIP_TLS_VERIFY").
		BoolVar(&c.skipVerify)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)
}
