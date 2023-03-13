// Copyright 2023 Harness Inc. All rights reserved.

package circle

import (
	"context"
	"encoding/json"
	"os"

	"github.com/harness/harness-migrate/cmd/util"

	"golang.org/x/exp/slog"

	"github.com/alecthomas/kingpin/v2"

	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
)

type importCommand struct {
	debug bool
	file  string

	harnessToken   string
	harnessAccount string
	harnessOrg     string
	harnessAddress string

	githubToken    string
	gitlabToken    string
	bitbucketToken string
}

func (c *importCommand) run(*kingpin.ParseContext) error {

	// create the logger
	log := util.CreateLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	// read the data file
	data, err := os.ReadFile(c.file)
	if err != nil {
		log.Error("cannot read data file", nil)
		return err
	}

	// unmarshal the data file
	org := new(types.Org)
	if err := json.Unmarshal(data, org); err != nil {
		log.Error("cannot unmarshal data file", nil)
		return err
	}

	// create the tracer
	tracer_ := tracer.New()
	defer tracer_.Close()

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

	// create a scm client to verify the token
	// and retrieve the user id.
	client := util.CreateClient(
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
	)

	// get the current user id.
	user, _, err := client.Users.Find(ctx)
	if err != nil {
		log.Error("cannot retrieve git user", nil)
		return err
	}

	// provide the user id to the importer. the user id
	// is required by the connector despite the fact that
	// it can be retrieved using the token itself (like we just did)
	importer.ScmLogin = user.Login

	log.Debug("verified user and token",
		slog.String("user", user.Login),
	)

	// execute the import routine.
	return importer.Import(ctx, org)
}

// helper function registers the import command.
func registerImport(app *kingpin.CmdClause) {
	c := new(importCommand)

	cmd := app.Command("import", "import circle data").
		Action(c.run)

	cmd.Arg("file", "data file to import").
		StringVar(&c.file)

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

	cmd.Flag("gitlab-token", "gitlab token").
		Envar("GITLAB_TOKEN").
		StringVar(&c.gitlabToken)

	cmd.Flag("bitbucket-token", "bitbucket token").
		Envar("BITBUCKET_TOKEN").
		StringVar(&c.bitbucketToken)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)
}
