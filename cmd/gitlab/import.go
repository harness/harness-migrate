// Copyright 2023 Harness Inc. All rights reserved.

package gitlab

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"golang.org/x/exp/slog"

	"github.com/alecthomas/kingpin/v2"
	scmgitlab "github.com/drone/go-scm/scm/driver/gitlab"
	"github.com/drone/go-scm/scm/transport"

	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/migrate/gitlab"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
)

type importCommand struct {
	debug bool
	file  string

	harnessEndpoint string
	harnessToken    string
	harnessAccount  string
	harnessOrg      string

	gitlabToken    string
	gitlabEndpoint string
}

func (c *importCommand) run(*kingpin.ParseContext) error {

	// create the logger
	log := createLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	// read the data file
	data, err := ioutil.ReadFile(c.file)
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

	// create the circle client (url, token, org)
	client := scmgitlab.NewDefault()
	client.Client = &http.Client{
		Transport: &transport.PrivateToken{
			Token: c.gitlabToken,
		},
	}

	// get the current user id.
	user, _, err := client.Users.Find(ctx)
	if err != nil {
		log.Error("cannot retrieve git user", nil)
		return err
	}

	log.Debug("verified user and token",
		slog.String("user", user.Login),
	)

	// create the importer
	importer := &gitlab.Importer{
		Harness: harness.New(c.harnessAccount, c.harnessToken,
			harness.WithAddress(c.harnessEndpoint), harness.WithTracing(c.debug)),
		HarnessOrg:   c.harnessOrg,
		HarnessToken: c.harnessToken,
		ScmType:      "gitlab",
		ScmToken:     c.gitlabToken,
		ScmLogin:     user.Login,
		Tracer:       tracer_,
	}

	// // execute the import routine.
	return importer.Import(ctx, org)
}

// helper function registers the import command.
func registerImport(app *kingpin.CmdClause) {
	c := new(importCommand)

	cmd := app.Command("import", "import gitlab data").
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

	cmd.Flag("harness-endpoint", "harness endpoint").
		Envar("HARNESS_ENDPOINT").
		StringVar(&c.harnessEndpoint)

	cmd.Flag("gitlab-token", "gitlab token").
		Envar("GITLAB_TOKEN").
		StringVar(&c.gitlabToken)

	cmd.Flag("gitlab-endpoint", "gitlab endpoint for on-prem installs").
		Envar("GITLAB_ENDPONT").
		StringVar(&c.gitlabEndpoint)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)
}
