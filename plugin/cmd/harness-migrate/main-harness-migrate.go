// Copyright © 2026 Harness Inc.
// SPDX-License-Identifier: Apache-2.0

// Command harness-migrate is the Harness unified CLI plugin front-end for
// harness-migrate. It is not meant to be run directly — install it with
// `harness install plugin <path-to-binary>` and drive it through `harness`.
package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/harness/cli/v3/pkg/console"
	"github.com/harness/cli/v3/pkg/registry"
	"github.com/harness/cli/v3/pkg/rootcmd"
	"github.com/harness/cli/v3/pkg/specloader"

	"github.com/harness/harness-migrate/plugin/pkg/migrateplugin"
)

func main() {
	reg := registry.New()
	// Unlike in-tree modules, our spec lives in this repo rather than core's
	// embedded pkg/spec, so it is loaded from bytes we embed ourselves.
	if _, err := specloader.LoadSpecBytes(reg, migrateplugin.SpecFileName, migrateplugin.SpecYAML, true); err != nil {
		console.PrintError(err.Error())
		os.Exit(1)
	}
	migrateplugin.ModuleInit(reg.Module(migrateplugin.ModuleName))
	rootcmd.MaybeCheckSpecs(reg)
	root := &cobra.Command{
		Use:   "harness-" + migrateplugin.ModuleName,
		Short: "Harness migration CLI (import repositories and pull requests into Harness Code)",
	}
	rootcmd.SetupAndExecutePluginRootCmd(root, reg, migrateplugin.ModuleName, migrateplugin.SpecYAML)
}
