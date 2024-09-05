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

package stash

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	scmstash "github.com/drone/go-scm/scm/driver/stash"
	"github.com/drone/go-scm/scm/transport"
	"github.com/harness/harness-migrate/cmd/util"
	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/migrate/stash"
	"github.com/harness/harness-migrate/internal/report"
	"golang.org/x/exp/slog"
)

type exportCommand struct {
	debug bool
	trace bool
	file  string

	project       string
	srcRepository string
	user          string
	token         string
	url           string

	checkpoint bool

	flags gitexporter.Flags
}

func (c *exportCommand) run(*kingpin.ParseContext) error {

	// create the logger
	log := util.CreateLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	// create the stash client (url, token, org)
	client, err := scmstash.New(c.url)
	if err != nil {
		return err
	}
	// provide a custom http.Client with a transport
	// that injects the private stash token through
	// the PRIVATE_TOKEN header variable.
	t := &transport.BasicAuth{
		Base:     nil,
		Username: c.user,
		Password: c.token,
	}

	client.Client = &http.Client{
		Transport: t,
	}

	// create the tracer
	tracer_ := util.CreateTracerWithLevel(c.debug)
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
	if c.srcRepository != "" {
		repository = strings.Trim(c.srcRepository, " ")
	}

	fileLogger := &gitexporter.FileLogger{Location: c.file}
	reporter := make(map[string]*report.Report)

	flags := gitexporter.Flags{
		NoPR:      c.flags.NoPR,
		NoComment: c.flags.NoComment,
		NoWebhook: c.flags.NoWebhook,
		NoRule:    c.flags.NoRule,
	}
	// extract the data
	e := stash.New(client, c.project, repository, checkpointManager, fileLogger, tracer_, reporter)

	exporter := gitexporter.NewExporter(e, c.file, c.user, c.token, tracer_, reporter, flags)
	return exporter.Export(ctx)
}

// helper function registers the export command
func registerGit(app *kingpin.CmdClause) {
	c := new(exportCommand)

	cmd := app.Command("git-export", "export stash git data").
		Hidden().
		Action(c.run)

	cmd.Arg("save", "save the output to a folder").
		Default("harness").
		StringVar(&c.file)

	cmd.Flag("host", "stash host url").
		Required().
		Envar("stash_HOST").
		StringVar(&c.url)

	cmd.Flag("project", "stash project").
		Required().
		Envar("stash_PROJECT").
		StringVar(&c.project)

	cmd.Flag("repository", "optional name of the repository to export").
		Envar("stash_REPOSITORY").
		StringVar(&c.srcRepository)

	cmd.Flag("token", "stash token").
		Required().
		Envar("stash_TOKEN").
		StringVar(&c.token)

	cmd.Flag("username", "stash username").
		Required().
		Envar("stash_USERNAME").
		StringVar(&c.user)

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
		Default("false").
		BoolVar(&c.flags.NoRule)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)
}

// defaultTransport provides a default http.Transport.
// If skip verify is true, the transport will skip ssl verification.
// Otherwise, it will append all the certs from the provided path.
func defaultTransport(proxy string) http.RoundTripper {
	if proxy == "" {
		return &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}
	}

	proxyURL, _ := url.Parse(proxy)

	return &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
}
