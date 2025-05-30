// Copyright 2024 Harness, Inc.
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

package gitlab

import (
	"context"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/harness/harness-migrate/cmd/util"
	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/migrate/gitlab"
	report "github.com/harness/harness-migrate/internal/report"

	"github.com/alecthomas/kingpin/v2"
	"github.com/drone/go-scm/scm"
	scmgitlab "github.com/drone/go-scm/scm/driver/gitlab"
	"github.com/drone/go-scm/scm/transport/oauth2"
)

type exportGitCommand struct {
	debug      bool
	trace      bool
	noProgress bool

	file string

	group   string
	project string
	user    string
	token   string
	url     string

	checkpoint bool
	flags      gitexporter.Flags
}

func (c *exportGitCommand) run(*kingpin.ParseContext) error {
	// create the logger
	log := util.CreateLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = util.WithLogger(ctx, log)

	// create the gitlab client
	var client *scm.Client
	var err error
	if c.url != "" {
		client, err = scmgitlab.New(c.url)
		if err != nil {
			return err
		}
	} else {
		client = scmgitlab.NewDefault()
	}

	if c.trace {
		client.DumpResponse = httputil.DumpResponse
	}

	// provide a custom http.Client with a transport
	// that injects the private gitlab token through
	// the PRIVATE_TOKEN header variable.
	t := &oauth2.Transport{
		Scheme: oauth2.SchemeBearer,
		Source: oauth2.StaticTokenSource(&scm.Token{Token: c.token}),
	}

	client.Client = &http.Client{
		Transport: t,
	}

	// create the tracer
	tracer_ := util.CreateTracerWithLevelAndType(c.debug, c.noProgress)
	defer tracer_.Close()

	checkpointManager := checkpoint.NewCheckpointManager(c.file)

	if c.checkpoint {
		err := checkpointManager.LoadCheckpoint()
		if err != nil {
			tracer_.LogError("unable to load checkpoint %v", err)
			panic("unable to load checkpoint")
		}
	}

	var repository string
	if c.project != "" {
		repository = strings.Trim(c.project, "/")
	}

	c.group = strings.Trim(c.group, "/")

	fileLogger := &gitexporter.FileLogger{Location: c.file}
	reporter := make(map[string]*report.Report)

	flags := gitexporter.Flags{
		NoPR:      c.flags.NoPR,
		NoComment: c.flags.NoComment,
		NoWebhook: c.flags.NoWebhook,
		NoRule:    c.flags.NoRule,
		NoLabel:   c.flags.NoLabel,
		NoLFS:     c.flags.NoLFS,
	}

	e := gitlab.New(client, c.group, repository, checkpointManager, fileLogger, tracer_, reporter)

	exporter := gitexporter.NewExporter(e, c.file, c.user, c.token, tracer_, reporter, flags)
	return exporter.Export(ctx)
}

// helper function registers the export command
func registerGit(app *kingpin.CmdClause) {
	c := new(exportGitCommand)

	cmd := app.Command("git-export", "export gitlab git data").
		Hidden().
		Action(c.run)

	cmd.Arg("save", "save the output to a folder").
		Default("harness").
		StringVar(&c.file)

	cmd.Flag("host", "gitlab host url").
		Envar("gitlab_HOST").
		StringVar(&c.url)

	cmd.Flag("group", "gitlab group followed by subgroups").
		Required().
		Envar("gitlab_GROUP").
		StringVar(&c.group)

	cmd.Flag("project", "optional name of the project to export").
		Envar("gitlab_PROJECT").
		StringVar(&c.project)

	cmd.Flag("username", "gitlab username").
		Envar("gitlab_USERNAME").
		StringVar(&c.user)

	cmd.Flag("token", "gitlab token").
		Required().
		Envar("gitlab_TOKEN").
		StringVar(&c.token)

	cmd.Flag("resume", "resume from last checkpoint").
		Default("false").
		BoolVar(&c.checkpoint)

	cmd.Flag("no-pr", "do NOT export pull requests and comments").
		Default("false").
		BoolVar(&c.flags.NoPR)

	cmd.Flag("no-comment", "do NOT export pull request comments").
		Default("false").
		BoolVar(&c.flags.NoComment)

	cmd.Flag("no-webhook", "do NOT export webhooks").
		Default("false").
		BoolVar(&c.flags.NoWebhook)

	cmd.Flag("no-rule", "do NOT export branch protection rules").
		Default("false"). // revert when supported
		BoolVar(&c.flags.NoRule)

	cmd.Flag("no-label", "do NOT export labels").
		Default("false").
		BoolVar(&c.flags.NoLabel)

	cmd.Flag("no-lfs", "do NOT export LFS objects").
		Default("false").
		BoolVar(&c.flags.NoLFS)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)

	cmd.Flag("no-progress", "disable progress bar logger").
		Default("false").
		BoolVar(&c.noProgress)
}
