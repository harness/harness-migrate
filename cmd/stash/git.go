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

package stash

import (
	"context"
	"net/http"
	"net/url"

	"github.com/harness/harness-migrate/cmd/util"
	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/migrate/stash"
	"github.com/harness/harness-migrate/internal/tracer"

	"github.com/alecthomas/kingpin/v2"
	scmstash "github.com/drone/go-scm/scm/driver/stash"
	"github.com/drone/go-scm/scm/transport"
	"golang.org/x/exp/slog"
)

type exportCommand struct {
	debug bool
	trace bool
	file  string

	stashOrg   string
	stashUser  string
	stashToken string
	stashUrl   string

	checkpoint bool
}

func (c *exportCommand) run(*kingpin.ParseContext) error {

	// create the logger
	log := util.CreateLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	// create the stash client (url, token, org)
	client, err := scmstash.New(c.stashUrl)
	if err != nil {
		return err
	}
	// provide a custom http.Client with a transport
	// that injects the private stash token through
	// the PRIVATE_TOKEN header variable.
	t := &transport.BasicAuth{
		Base:     nil,
		Username: c.stashUser,
		Password: c.stashToken,
	}

	client.Client = &http.Client{
		Transport: t,
	}

	// create the tracer
	tracer_ := tracer.New()
	defer tracer_.Close()

	checkpointManager := checkpoint.NewCheckpointManager(c.file)

	if c.checkpoint {
		err := checkpointManager.LoadCheckpoint()
		if err != nil {
			tracer_.LogError("unable to load checkpoint %v", err)
			panic("unable to load checkpoint")
		}
	}

	// extract the data
	e := stash.New(client, c.stashOrg, checkpointManager, tracer_)

	exporter := gitexporter.NewExporter(e, c.file)
	exporter.Export(ctx)
	return nil
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
		StringVar(&c.stashUrl)

	cmd.Flag("org", "stash organization").
		Required().
		Envar("stash_ORG").
		StringVar(&c.stashOrg)

	cmd.Flag("token", "stash token").
		Required().
		Envar("stash_TOKEN").
		StringVar(&c.stashToken)

	cmd.Flag("username", "stash username").
		Required().
		Envar("stash_USERNAME").
		StringVar(&c.stashUser)

	cmd.Flag("resume", "resume from last checkpoint").
		Default("false").
		BoolVar(&c.checkpoint)

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
