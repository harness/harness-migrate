// Copyright © 2026 Harness Inc.
// SPDX-License-Identifier: Apache-2.0

// Package migrateplugin adapts harness-migrate's engine to the Harness unified
// CLI plugin contract. Commands are declared in migrate.spec.yaml; the Go code
// here is only the workflow handlers those declarations dispatch to.
package migrateplugin

import (
	_ "embed"

	"github.com/harness/cli/v3/pkg/registry"
)

// ModuleName is the plugin name. It fixes the installed binary name to
// harness-<ModuleName>, per the plugin design's naming requirement.
const ModuleName = "migrate"

// SpecFileName must be "<ModuleName>.spec.yaml" — the loader derives the module
// name from it.
const SpecFileName = "migrate.spec.yaml"

// SpecYAML is the grammar this plugin registers, and is dumped verbatim by
// --spec so the host can capture it at install time.
//
//go:embed migrate.spec.yaml
var SpecYAML []byte

// defaultBundleDir is where a migration writes its scm bundle when --to is
// omitted, matching the standalone CLI's default output folder.
const defaultBundleDir = "harness"

// ModuleInit registers the migrate workflows. Commands are declared in
// migrate.spec.yaml.
func ModuleInit(reg registry.ModuleRegistrar) {
	reg.RegisterWorkflow(migrateGithubOrgToBundleID, migrateGithubOrgToBundle)
	reg.RegisterWorkflow(migrateGitlabGroupToBundleID, migrateGitlabGroupToBundle)
	reg.RegisterWorkflow(migrateBitbucketWorkspaceToBundleID, migrateBitbucketWorkspaceToBundle)
	reg.RegisterWorkflow(migrateStashProjectToBundleID, migrateStashProjectToBundle)
	reg.RegisterWorkflow(migrateBundleToRepositoryID, migrateBundleToRepository)
	reg.RegisterWorkflow(updateBundleUsersID, updateBundleUsers)
}
