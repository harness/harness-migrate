// Copyright 2023 Harness Inc. All rights reserved.

package circle

import "github.com/alecthomas/kingpin/v2"

// Register the command.
func Register(app *kingpin.Application) {
	cmd := app.Command("circle", "migrate circle data")
	registerMigrate(cmd)
	registerExport(cmd)
	registerImport(cmd)
	registerConvert(cmd)
}
