// Copyright 2023 Harness Inc. All rights reserved.

package gitlab

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/exp/slog"

	"github.com/harness/harness-migrate/internal/migrate/gitlab"
	"github.com/harness/harness-migrate/internal/tracer"

	scmgitlab "github.com/drone/go-scm/scm/driver/gitlab"
	"github.com/drone/go-scm/scm/transport"

	"github.com/alecthomas/kingpin/v2"
)

type exportCommand struct {
	debug bool
	trace bool
	file  string

	gitlabToken string
	gitlabOrg   string
}

func (c *exportCommand) run(*kingpin.ParseContext) error {

	// create the logger
	log := createLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	// create the gitlab client (url, token, org)
	client := scmgitlab.NewDefault()

	// provide a custom http.Client with a transport
	// that injects the private GitLab token through
	// the PRIVATE_TOKEN header variable.
	client.Client = &http.Client{
		Transport: &transport.PrivateToken{
			Token: c.gitlabToken,
		},
	}

	// create the tracer
	tracer_ := tracer.New()
	defer tracer_.Close()

	// extract the data
	exporter := gitlab.Exporter{
		Gitlab:    client,
		GitlabOrg: c.gitlabOrg,
		Tracer:    tracer_,
	}
	data, err := exporter.Export(ctx)
	if err != nil {
		return err
	}

	// if no file path is provided, write the data export
	// to stdout.
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

	return ioutil.WriteFile(c.file, file, 0644)
}

// helper function registers the export command
func registerExport(app *kingpin.CmdClause) {
	c := new(exportCommand)

	cmd := app.Command("export", "export gitlab data").
		Action(c.run)

	cmd.Arg("save", "save the output to a file").
		StringVar(&c.file)

	cmd.Flag("org", "gitlab organization").
		Required().
		Envar("GITLAB_ORG").
		StringVar(&c.gitlabOrg)

	cmd.Flag("token", "gitlab token").
		Required().
		Envar("GITLAB_TOKEN").
		StringVar(&c.gitlabToken)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)
}
