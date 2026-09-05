// Copyright © 2026 Harness Inc.
// SPDX-License-Identifier: Apache-2.0

package migrateplugin

import (
	"github.com/harness/cli/v3/pkg/cmdctx"
	"github.com/harness/cli/v3/pkg/hlog"

	"github.com/harness/harness-migrate/internal/users"
	"github.com/harness/harness-migrate/plugin/pkg/bridge"
)

const updateBundleUsersID = "update_scm_bundle_users"

// updateBundleUsers rewrites the email addresses inside an scm bundle from a
// mapping file, so an import can find its authors in Harness. It calls the same
// engine constructor as the standalone `harness-migrate update-users` command;
// the bundle arrives as the command's id rather than as --zipFilePath.
func updateBundleUsers(ctx *cmdctx.Ctx) error {
	// update takes a mandatory positional, so ctx.Id is non-empty here. It goes
	// through the same folder-or-zip rule as the import's --from: the bundle is
	// the same artifact either way, so it should be nameable the same way.
	zipPath, err := resolveBundleZip(ctx.Id)
	if err != nil {
		return err
	}
	// user-mapping is declared required in the spec, so cobra has already
	// rejected the run if it is missing.
	mapping := cmdctx.GetString(ctx.FlagValues, "user-mapping")

	tracer := bridge.NewTracer(cmdctx.GetBool(ctx.FlagValues, "no-progress"))
	defer tracer.Close()

	hlog.Debug("starting scm_bundle:users update", "zip", zipPath, "mapping", mapping)
	return users.NewUpdater(mapping, zipPath, tracer).Update(ctx.Context)
}
