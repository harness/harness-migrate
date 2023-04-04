// Copyright 2023 Harness Inc. All rights reserved.

package gitlab

import "github.com/alecthomas/kingpin/v2"

// Register the command.
func Register(app *kingpin.Application) {
	cmd := app.Command("gitlab", "migrate gitlab data")
	registerExport(cmd)
	registerConvert(cmd)
}
