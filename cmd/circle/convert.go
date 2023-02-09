// Copyright 2023 Harness Inc. All rights reserved.

package circle

import (
	"io/ioutil"
	"os"

	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/commons"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter"

	"github.com/alecthomas/kingpin/v2"
)

type convertCommand struct {
	path string
}

func (c *convertCommand) run(*kingpin.ParseContext) error {
	a, err := ioutil.ReadFile(c.path)
	if err != nil {
		return err
	}
	opts := commons.Opts{}
	b, err := converter.Convert(opts, a)
	if err != nil {
		return err
	}
	os.Stdout.Write(b)
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
