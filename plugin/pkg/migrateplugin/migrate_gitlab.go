// Copyright © 2026 Harness Inc.
// SPDX-License-Identifier: Apache-2.0

package migrateplugin

import (
	"fmt"
	"strings"

	"github.com/drone/go-scm/scm"
	scmgitlab "github.com/drone/go-scm/scm/driver/gitlab"

	"github.com/harness/cli/v3/pkg/cmdctx"
	"github.com/harness/cli/v3/pkg/hlog"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/migrate/gitlab"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/plugin/pkg/bridge"
)

const migrateGitlabGroupToBundleID = "migrate_gitlab_group_scm_bundle"

// migrateGitlabGroupToBundle writes a GitLab group's git data into a local scm
// bundle. It calls the same engine constructors as the standalone
// `harness-migrate gitlab git-export` command; only the front-end differs.
func migrateGitlabGroupToBundle(ctx *cmdctx.Ctx) error {
	// --from is declared presence: required in the spec, so it is already
	// non-empty; --to has no spec-level default, hence the fallback here.
	group := strings.Trim(ctx.MigrateFrom, "/")
	dir := ctx.MigrateTo
	if dir == "" {
		dir = defaultBundleDir
	}
	// gitlab-token is declared required in the spec, so cobra has already
	// rejected the run if it is missing.
	token := cmdctx.GetString(ctx.FlagValues, "gitlab-token")
	project := strings.Trim(cmdctx.GetString(ctx.FlagValues, "repo"), "/")
	user := cmdctx.GetString(ctx.FlagValues, "gitlab-user")
	host := cmdctx.GetString(ctx.FlagValues, "gitlab-host")
	includeSubgroups := cmdctx.GetBool(ctx.FlagValues, "include-subgroups")

	flags := gitexporter.Flags{
		NoPR:         cmdctx.GetBool(ctx.FlagValues, "no-pr"),
		NoComment:    cmdctx.GetBool(ctx.FlagValues, "no-comment"),
		NoPRMetadata: cmdctx.GetBool(ctx.FlagValues, "no-pr-metadata"),
		NoWebhook:    cmdctx.GetBool(ctx.FlagValues, "no-webhook"),
		NoRule:       cmdctx.GetBool(ctx.FlagValues, "no-rule"),
		NoLabel:      cmdctx.GetBool(ctx.FlagValues, "no-label"),
		NoLFS:        cmdctx.GetBool(ctx.FlagValues, "no-lfs"),
	}

	client, err := newGitlabClient(host, token)
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

	e := gitlab.New(client, group, project, checkpointManager, fileLogger, tracer, reporter, includeSubgroups)
	exporter := gitexporter.NewExporter(e, dir, user, token, tracer, reporter, flags)

	hlog.Debug("starting gitlab_group:scm_bundle migration", "group", group, "project", project, "dir", dir)
	return exporter.Export(ctx.Context)
}

// newGitlabClient builds an scm client that injects the token as a bearer
// token, matching the standalone command's transport.
func newGitlabClient(host, token string) (*scm.Client, error) {
	var client *scm.Client
	var err error
	if host != "" {
		client, err = scmgitlab.New(host)
		if err != nil {
			return nil, err
		}
	} else {
		client = scmgitlab.NewDefault()
	}
	client.Client = oauth2BearerClient(token)
	return client, nil
}
