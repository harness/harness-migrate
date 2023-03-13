// Copyright 2023 Harness Inc. All rights reserved.

package circle

import (
	"context"
	"encoding/json"
	"os"

	"github.com/harness/harness-migrate/cmd/util"

	"golang.org/x/exp/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/harness/harness-migrate/internal/migrate/circle"
	"github.com/harness/harness-migrate/internal/migrate/circle/client"
	"github.com/harness/harness-migrate/internal/tracer"
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

	return os.WriteFile(c.file, file, 0644)
}

// helper function registers the export command
func registerExport(app *kingpin.CmdClause) {
	c := new(exportCommand)

	cmd := app.Command("export", "export circle data").
		Action(c.run)

	cmd.Arg("save", "save the output to a file").
		StringVar(&c.file)

	cmd.Flag("org", "circle organization").
		Required().
		Envar("CIRCLE_ORG").
		StringVar(&c.circleOrg)

	cmd.Flag("token", "circle token").
		Required().
		Envar("CIRCLE_TOKEN").
		StringVar(&c.circleToken)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)
}
