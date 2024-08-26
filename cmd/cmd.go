// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"

	"github.com/harness/harness-migrate/cmd/bitbucket"
	"github.com/harness/harness-migrate/cmd/circle"
	"github.com/harness/harness-migrate/cmd/cloudbuild"
	"github.com/harness/harness-migrate/cmd/drone"
	"github.com/harness/harness-migrate/cmd/github"
	"github.com/harness/harness-migrate/cmd/gitimporter"
	"github.com/harness/harness-migrate/cmd/gitlab"
	"github.com/harness/harness-migrate/cmd/jenkinsxml"
	"github.com/harness/harness-migrate/cmd/stash"
	"github.com/harness/harness-migrate/cmd/terraform"
	"github.com/harness/harness-migrate/cmd/travis"

	"github.com/alecthomas/kingpin/v2"
)

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
	jenkinsxml.Register(app)
	travis.Register(app)
	terraform.Register(app)
	stash.Register(app)

	gitimporter.Register(app)

	app.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
