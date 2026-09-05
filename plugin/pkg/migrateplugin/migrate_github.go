// Copyright © 2026 Harness Inc.
// SPDX-License-Identifier: Apache-2.0

package migrateplugin

import (
	"fmt"
	"strings"

	"github.com/drone/go-scm/scm"
	scmgithub "github.com/drone/go-scm/scm/driver/github"

	"github.com/harness/cli/v3/pkg/cmdctx"
	"github.com/harness/cli/v3/pkg/hlog"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/migrate/github"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/plugin/pkg/bridge"
)

const migrateGithubOrgToBundleID = "migrate_github_organization_scm_bundle"

// migrateGithubOrgToBundle writes a GitHub org's git data into a local scm
// bundle. It calls the same engine constructors as the standalone
// `harness-migrate github git-export` command; only the front-end differs.
func migrateGithubOrgToBundle(ctx *cmdctx.Ctx) error {
	// --from is declared presence: required in the spec, so it is already
	// non-empty; --to has no spec-level default, hence the fallback here.
	org := strings.Trim(ctx.MigrateFrom, "/")
	dir := ctx.MigrateTo
	if dir == "" {
		dir = defaultBundleDir
	}
	// github-token is declared required in the spec, so cobra has already
	// rejected the run if it is missing.
	token := cmdctx.GetString(ctx.FlagValues, "github-token")
	repository := strings.Trim(cmdctx.GetString(ctx.FlagValues, "repo"), "/")
	user := cmdctx.GetString(ctx.FlagValues, "github-user")
	host := cmdctx.GetString(ctx.FlagValues, "github-host")

	flags := gitexporter.Flags{
		NoPR:         cmdctx.GetBool(ctx.FlagValues, "no-pr"),
		NoComment:    cmdctx.GetBool(ctx.FlagValues, "no-comment"),
		NoPRMetadata: cmdctx.GetBool(ctx.FlagValues, "no-pr-metadata"),
		NoWebhook:    cmdctx.GetBool(ctx.FlagValues, "no-webhook"),
		NoRule:       cmdctx.GetBool(ctx.FlagValues, "no-rule"),
		NoLabel:      cmdctx.GetBool(ctx.FlagValues, "no-label"),
		NoLFS:        cmdctx.GetBool(ctx.FlagValues, "no-lfs"),
	}

	client, err := newGithubClient(host, token)
	if err != nil {
		return err
	}

	tracer := bridge.NewTracer(cmdctx.GetBool(ctx.FlagValues, "no-progress"))
	defer tracer.Close()

	checkpointManager := checkpoint.NewCheckpointManager(dir)
	if cmdctx.GetBool(ctx.FlagValues, "resume") {
		if err := checkpointManager.LoadCheckpoint(); err != nil {
			return fmt.Errorf("unable to load checkpoint from %s: %w", dir, err)
		}
	}

	fileLogger := &gitexporter.FileLogger{Location: dir}
	reporter := make(map[string]*report.Report)

	e := github.New(client, org, repository, checkpointManager, fileLogger, tracer, reporter)
	exporter := gitexporter.NewExporter(e, dir, user, token, tracer, reporter, flags)

	hlog.Debug("starting github_organization:scm_bundle migration", "org", org, "repository", repository, "dir", dir)
	return exporter.Export(ctx.Context)
}

// newGithubClient builds an scm client that injects the token as a bearer token,
// matching the standalone command's transport.
func newGithubClient(host, token string) (*scm.Client, error) {
	var client *scm.Client
	var err error
	if host != "" {
		client, err = scmgithub.New(host)
		if err != nil {
			return nil, err
		}
	} else {
		client = scmgithub.NewDefault()
	}
	client.Client = oauth2BearerClient(token)
	return client, nil
}
