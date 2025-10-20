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
	"strconv"
	"strings"

	"github.com/harness/harness-migrate/cmd/util"
	"github.com/harness/harness-migrate/internal/gitimporter"
	"github.com/harness/harness-migrate/internal/report"

	"github.com/alecthomas/kingpin/v2"
	"github.com/google/uuid"
)

type gitImport struct {
	debug      bool
	trace      bool
	noProgress bool

	endpoint     string
	harnessToken string
	harnessSpace string
	harnessRepo  string // single repo import

	skipUsers     bool
	Gitness       bool
	fileSizeLimit int64

	filePath string

	// optional flags to skip import repo meta data
	noPR        bool
	noWebhook   bool
	noRule      bool
	noLabel     bool
	noGit       bool // for incremental migration - skip git operations
	prBatchSize int  // batch size for PR imports to avoid 413 errors
}

type UserInvite bool

func (c *gitImport) run(*kingpin.ParseContext) error {
	// create the logger
	log := util.CreateLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = util.WithLogger(ctx, log)

	tracer_ := util.CreateTracerWithLevelAndType(c.debug, c.noProgress)
	defer tracer_.Close()

	c.harnessRepo = strings.Trim(c.harnessRepo, "/")
	importUuid := uuid.New().String()
	c.endpoint, _ = strings.CutSuffix(c.endpoint, "/")
	reporter := make(map[string]*report.Report)
	importer := gitimporter.NewImporter(
		c.endpoint, c.harnessSpace, c.harnessRepo, c.harnessToken, c.filePath,
		importUuid, c.Gitness, c.trace,
		gitimporter.Flags{
			SkipUsers:     c.skipUsers,
			FileSizeLimit: c.fileSizeLimit,
			NoPR:          c.noPR,
			NoWebhook:     c.noWebhook,
			NoRule:        c.noRule,
			NoLabel:       c.noLabel,
			NoGit:         c.noGit,
			PRBatchSize:   c.prBatchSize,
		},
		tracer_,
		reporter)

	tracer_.Log("starting operation with id: %s", importUuid)
	return importer.Import(ctx)
}

func registerGitImporter(app *kingpin.CmdClause) {
	c := new(gitImport)

	cmd := app.Action(c.run)

	cmd.Arg("filePath", "location of the zip file").
		Required().
		StringVar(&c.filePath)

	cmd.Flag("endpoint", "url of target Harness Code/Gitness host").
		Default("https://app.harness.io/").
		Envar("target_HOST").
		StringVar(&c.endpoint)

	cmd.Flag("token", "harness api token").
		Required().
		Envar("harness_TOKEN").
		StringVar(&c.harnessToken)

	cmd.Flag("space", "harness path where import should take place. Example: account/org/project").
		Required().
		Envar("harness_SPACE").
		StringVar(&c.harnessSpace)

	cmd.Flag("skip-users", "skip unknown user and map to token uuid").
		Default("false").
		Envar("harness_SKIP_USERS").
		BoolVar(&c.skipUsers)

	cmd.Flag("repo-path", "optional path of a single repo to import (e.g, Org/repo).").
		Envar("HARNESS_REPO_PATH").
		StringVar(&c.harnessRepo)

	cmd.Flag("file-size-limit", "temporarily update git push file size limit for large repositories during migration. Default: 100MB").
		Default(strconv.FormatInt(int64(1e+8), 10)).
		Envar("FILE_SIZE_LIMIT").
		Int64Var(&c.fileSizeLimit)

	cmd.Flag("gitness", "import into a Gitness instance").
		Default("false").
		Envar("Gitness").
		BoolVar(&c.Gitness)

	cmd.Flag("no-pr", "").
		Hidden().
		Default("false").
		BoolVar(&c.noPR)

	cmd.Flag("skip-pr", "skip importing pull requests and comments (--no-pr is an alias)").
		Default("false").
		BoolVar(&c.noPR)

	cmd.Flag("no-label", "").
		Hidden().
		Default("false").
		BoolVar(&c.noLabel)

	cmd.Flag("skip-label", "skip importing labels (--no-label is an alias)").
		Default("false").
		BoolVar(&c.noLabel)

	cmd.Flag("no-webhook", "").
		Hidden().
		Default("false").
		BoolVar(&c.noWebhook)

	cmd.Flag("skip-webhook", "skip importing webhooks (--no-webhook is an alias)").
		Default("false").
		BoolVar(&c.noWebhook)

	cmd.Flag("no-rule", "").
		Hidden().
		Default("false").
		BoolVar(&c.noRule)

	cmd.Flag("skip-rule", "skip importing branch protection rules (--no-rule is an alias)").
		Default("false").
		BoolVar(&c.noRule)

	cmd.Flag("no-git", "perform incremental migration - skip git operations and import PR with offset").
		Default("false").
		BoolVar(&c.noGit)

	cmd.Flag("batch-size", "number of pull requests to import per batch (default: 100). Use a smaller value if encountering 413 errors.").
		Default("100").
		IntVar(&c.prBatchSize)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)

	cmd.Flag("no-progress", "disable progress bar logger").
		Default("false").
		BoolVar(&c.noProgress)
}
