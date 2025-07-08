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

package users

import (
	"context"
	"path/filepath"

	"github.com/harness/harness-migrate/cmd/util"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/users"

	"github.com/alecthomas/kingpin/v2"
)

type updateUserCommand struct {
	usermapping string
	zipFilePath string
	debug       bool
	noProgress  bool
}

// run executes the update-users command
func (c *updateUserCommand) run(*kingpin.ParseContext) error {
	// create the logger
	log := util.CreateLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = util.WithLogger(ctx, log)

	tracer := util.CreateTracerWithLevelAndType(c.debug, c.noProgress)
	defer tracer.Close()

	tracer.Log("starting user email update operation")

	updater := users.NewUpdater(
		c.usermapping,
		c.zipFilePath,
		tracer,
	)

	return updater.Update(ctx)
}

func registerUpdateUsers(app *kingpin.CmdClause) {
	c := new(updateUserCommand)

	cmd := app.Action(c.run)

	cmd.Arg("users", "path to the JSON file containing user email mappings").
		Required().
		StringVar(&c.usermapping)

	cmd.Flag("zipFilePath", "path to the exported zip file containing SCM data").
		Default(filepath.Join("harness", gitexporter.ZipFileName)).
		StringVar(&c.zipFilePath)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("no-progress", "disable progress bar logger").
		Default("false").
		BoolVar(&c.noProgress)
}
