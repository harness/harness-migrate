// Copyright 2023 Harness Inc. All rights reserved.

package circle

import (
	"context"

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

	githubToken    string
	gitlabToken    string
	bitbucketToken string
}

func (c *migrateCommand) run(*kingpin.ParseContext) error {

	// create the logger
	log := createLogger(c.debug)

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
	importer := createImporter(
		c.harnessAccount,
		c.harnessOrg,
		c.harnessToken,
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
	)
	importer.Tracer = tracer_

	// create an scm cient to verify the token
	// and retrieve the user id.
	scmClient := createClient(
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
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

	cmd.Flag("github-token", "github token").
		Envar("GITHUB_TOKEN").
		StringVar(&c.githubToken)

	cmd.Flag("gitlab-token", "gitlab token").
		Envar("GITLAB_TOKEN").
		StringVar(&c.gitlabToken)

	cmd.Flag("bitbucket-token", "bitbucket token").
		Envar("BITBUCKET_TOKEN").
		StringVar(&c.bitbucketToken)

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
