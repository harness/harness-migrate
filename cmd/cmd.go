// Copyright 2023 Harness Inc. All rights reserved.

package cmd

import (
	"context"
	"os"

	"github.com/harness/harness-migrate/cmd/bitbucket"
	"github.com/harness/harness-migrate/cmd/circle"
	"github.com/harness/harness-migrate/cmd/cloudbuild"
	"github.com/harness/harness-migrate/cmd/drone"
	"github.com/harness/harness-migrate/cmd/github"
	"github.com/harness/harness-migrate/cmd/gitlab"
	"github.com/harness/harness-migrate/cmd/travis"

	"github.com/alecthomas/kingpin/v2"
)

// empty context
var nocontext = context.Background()

// application name
const application = "harness-migrate"

// application description
const description = "import repositories and pipelines into harness"

// application version
var version string

// Command parses the command line arguments and then executes a
// subcommand program.
func Command() {
	app := kingpin.New(application, description)

	bitbucket.Register(app)
	cloudbuild.Register(app)
	circle.Register(app)
	drone.Register(app)
	gitlab.Register(app)
	github.Register(app)
	travis.Register(app)

	app.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
