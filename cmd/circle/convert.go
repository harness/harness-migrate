// Copyright 2023 Harness Inc. All rights reserved.

package circle

import (
	"os"

	"github.com/drone/go-convert/convert/circle"

	"github.com/alecthomas/kingpin/v2"
)

type convertCommand struct {
	path string
}

func (c *convertCommand) run(*kingpin.ParseContext) error {
	a, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}
	converter := circle.New()
	yaml, err := converter.ConvertBytes(a)
	if err != nil {
		return err
	}
	os.Stdout.Write(yaml)
	return nil
}

// helper function registers the convert command
func registerConvert(app *kingpin.CmdClause) {
	c := new(convertCommand)

	cmd := app.Command("convert", "convert a circle yaml").
		Action(c.run)

	cmd.Arg("path", "path to circle yaml").
		Default(".circle/config.yml").
		StringVar(&c.path)
}
