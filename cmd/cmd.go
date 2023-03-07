// Copyright 2023 Harness Inc. All rights reserved.

package cmd

import (
	"context"
	"os"

	"github.com/harness/harness-migrate/cmd/drone"

	"github.com/harness/harness-migrate/cmd/circle"
	"github.com/harness/harness-migrate/cmd/gitlab"

	"github.com/alecthomas/kingpin/v2"
)

// empty context
var nocontext = context.Background()

// application name
const application = "harness-migrate"

// application description
const description = "import repositories and pipelines into harness"

// application version
const version = "1.0.0"

// Command parses the command line arguments and then executes a
// subcommand program.
func Command() {
	app := kingpin.New(application, description)

	circle.Register(app)
	gitlab.Register(app)
	drone.Register(app)

	kingpin.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
