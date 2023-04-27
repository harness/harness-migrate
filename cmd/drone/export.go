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

	"github.com/harness/harness-migrate/cmd/util"
	"github.com/harness/harness-migrate/internal/migrate/drone"
	"github.com/harness/harness-migrate/internal/migrate/drone/repo"
	"github.com/harness/harness-migrate/internal/tracer"

	"github.com/alecthomas/kingpin/v2"
	"github.com/jmoiron/sqlx"
	"golang.org/x/exp/slog"
)

type exportCommand struct {
	debug     bool
	downgrade bool
	trace     bool
	file      string

	Driver     string
	Datasource string
	namespace  string

	githubToken    string
	gitlabToken    string
	bitbucketToken string

	proj       string
	org        string
	repoConn   string
	kubeName   string
	kubeConn   string
	dockerConn string
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
	tracer_ := tracer.New()
	defer tracer_.Close()

	client := util.CreateClient(
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
	)

	if c.githubToken == "" && c.gitlabToken == "" && c.bitbucketToken == "" {
		return errors.New("no scm token provided")
	}

	user, _, err := client.Users.Find(ctx)
	if err != nil {
		log.Error("cannot retrieve git user", nil)
		return err
	}

	// extract the data
	exporter := &drone.Exporter{
		Downgrade:  c.downgrade,
		Repository: droneRepo,
		Namespace:  c.namespace,
		Tracer:     tracer_,
		ScmClient:  client,
		ScmLogin:   user.Login,
		DockerConn: c.dockerConn,
		KubeName:   c.kubeName,
		KubeConn:   c.kubeConn,
		Org:        c.org,
		RepoConn:   c.repoConn,
	}
	data, err := exporter.Export(ctx)
	if err != nil {
		return err
	}

	//if no file path is provided, write the data export
	//to stdout.
	if c.file == "" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	// else write the data export to the file.
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.file, file, 0644)
}

// helper function registers the export command
func registerExport(app *kingpin.CmdClause) {
	c := new(exportCommand)

	cmd := app.Command("export", "export drone data").
		Action(c.run)

	cmd.Arg("save", "save the output to a file").
		StringVar(&c.file)

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

	cmd.Flag("github-token", "github token").
		Envar("GITHUB_TOKEN").
		StringVar(&c.githubToken)

	cmd.Flag("gitlab-token", "gitlab token").
		Envar("GITLAB_TOKEN").
		StringVar(&c.gitlabToken)

	cmd.Flag("bitbucket-token", "bitbucket token").
		Envar("BITBUCKET_TOKEN").
		StringVar(&c.bitbucketToken)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)

	cmd.Flag("org", "harness organization").
		Default("default").
		StringVar(&c.org)

	cmd.Flag("project", "harness project").
		Default("default").
		StringVar(&c.proj)

	cmd.Flag("repo-connector", "repository connector").
		Default("").
		StringVar(&c.repoConn)

	cmd.Flag("kube-connector", "kubernetes connector").
		Default("").
		StringVar(&c.kubeConn)

	cmd.Flag("kube-namespace", "kubernetes namespace").
		Default("").
		StringVar(&c.kubeName)

	cmd.Flag("docker-connector", "dockerhub connector").
		Default("").
		StringVar(&c.kubeName)
}
