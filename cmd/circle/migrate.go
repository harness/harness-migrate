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

package circle

import (
	"context"

	"github.com/harness/harness-migrate/cmd/util"

	"golang.org/x/exp/slog"

	"github.com/harness/harness-migrate/internal/migrate/circle"
	"github.com/harness/harness-migrate/internal/migrate/circle/client"
	"github.com/harness/harness-migrate/internal/tracer"

	"github.com/alecthomas/kingpin/v2"
)

type migrateCommand struct {
	debug bool
	trace bool

	circleToken string
	circleOrg   string

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
	skipVerify     bool
}

func (c *migrateCommand) run(*kingpin.ParseContext) error {

	// create the logger
	log := util.CreateLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	// create the circle client (url, token, org)
	client := client.New(c.circleToken,
		client.WithTracing(c.trace),
	)

	// create the tracer
	tracer_ := tracer.New()
	defer tracer_.Close()

	// extract the data
	exporter := circle.Exporter{
		Circle:    client,
		CircleOrg: c.circleOrg,
		Tracer:    tracer_,
	}
	org, err := exporter.Export(ctx)
	if err != nil {
		return err
	}

	// create the importer
	importer := util.CreateImporter(
		c.harnessAccount,
		c.harnessOrg,
		c.harnessToken,
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
		c.harnessAddress,
	)
	importer.Tracer = tracer_

	// create an scm cient to verify the token
	// and retrieve the user id.
	scmClient := util.CreateClient(
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
		c.githubURL,
		c.gitlabURL,
		c.bitbucketURL,
		c.skipVerify,
	)

	// get the current user id.
	user, _, err := scmClient.Users.Find(ctx)
	if err != nil {
		log.Error("cannot retrieve git user", nil)
		return err
	}

	// provide the user id to the importer. the user id
	// is required by the connector despite the fact that
	// it can be retrieve using the token itself (like we just did)
	importer.ScmLogin = user.Login

	log.Debug("verified user and token",
		slog.String("user", user.Login),
	)

	// execute the import routine.
	return importer.Import(ctx, org)
}

// helper function registers the migrate command
func registerMigrate(app *kingpin.CmdClause) {
	c := new(migrateCommand)

	cmd := app.Command("migrate", "migrate circle data to harness").
		Hidden().
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

	cmd.Flag("circle-org", "circle organization").
		Required().
		Envar("CIRCLE_ORG").
		StringVar(&c.circleOrg)

	cmd.Flag("circle-token", "circle token").
		Required().
		Envar("CIRCLE_TOKEN").
		StringVar(&c.circleToken)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)
}
