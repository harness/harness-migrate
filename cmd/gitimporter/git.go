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

package gitimporter

import (
	"context"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/google/uuid"
	"github.com/harness/harness-migrate/cmd/util"
	"github.com/harness/harness-migrate/internal/gitimporter"
	"github.com/harness/harness-migrate/internal/tracer"
	"golang.org/x/exp/slog"
)

type gitImport struct {
	debug bool
	trace bool

	harnessEndpoint string
	harnessToken    string
	harnessSpace    string
	//harnessRepo     string

	skipUsers bool

	filePath string
}

type UserInvite bool

func (c *gitImport) run(*kingpin.ParseContext) error {
	// create the logger
	log := util.CreateLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	tracer_ := tracer.New()
	defer tracer_.Close()

	importUuid := uuid.New().String()
	c.harnessEndpoint, _ = strings.CutSuffix(c.harnessEndpoint, "/")
	importer := gitimporter.NewImporter(c.harnessEndpoint, c.harnessSpace, c.harnessToken, c.filePath, importUuid, c.skipUsers, tracer_)

	tracer_.Log("starting operation with id: %s", importUuid)
	return importer.Import()
}

func registerGitImporter(app *kingpin.CmdClause) {
	c := new(gitImport)

	cmd := app.Action(c.run)

	cmd.Arg("filePath", "location of the zip file").
		Required().
		StringVar(&c.filePath)

	cmd.Flag("harnessEndpoint", "url of harness code host").
		Default("https://app.harness.io/gateway/code").
		Envar("harness_HOST").
		StringVar(&c.harnessEndpoint)

	cmd.Flag("token", "harness api token").
		Required().
		Envar("harness_TOKEN").
		StringVar(&c.harnessToken)

	cmd.Flag("space", "harness path where import should take place. Example: account/org/project").
		Required().
		Envar("harness_SPACE").
		StringVar(&c.harnessSpace)

	cmd.Flag("skip-users", "skip unknown user and map to token uuid (Default:true)").
		Default("true").
		Envar("harness_SKIP_USERS").
		BoolVar(&c.skipUsers)

	// cmd.Flag("repo", "Required in case of single repo import which already exists.").
	//	Envar("HARNESS_REPO").
	//	StringVar(&c.harnessRepo)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)
}
