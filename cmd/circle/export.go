// Copyright 2023 Harness Inc. All rights reserved.

package circle

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	"golang.org/x/exp/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/harness/harness-migrate/internal/migrate/circle"
	"github.com/harness/harness-migrate/internal/migrate/circle/client"
)

type exportCommand struct {
	debug bool
	trace bool
	file  string

	circleToken string
	circleOrg   string
}

func (c *exportCommand) run(*kingpin.ParseContext) error {

	// create the logger
	log := createLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	// create the circle client (url, token, org)
	client := client.New(c.circleToken,
		client.WithTracing(true),
	)

	// extract the data
	exporter := circle.Exporter{
		Circle:    client,
		CircleOrg: c.circleOrg,
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

	cmd := app.Command("export", "export circle data").
		Action(c.run)

	cmd.Flag("org", "circle organization").
		Required().
		Envar("CIRCLE_ORG").
		StringVar(&c.circleOrg)

	cmd.Flag("token", "circle token").
		Required().
		Envar("CIRCLE_TOKEN").
		StringVar(&c.circleToken)

	cmd.Flag("out", "save the output to a file").
		StringVar(&c.file)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)
}
