// Copyright © 2026 Harness Inc.
// SPDX-License-Identifier: Apache-2.0

package migrateplugin

import (
	"archive/zip"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/harness/cli/v3/pkg/auth"
	"github.com/harness/cli/v3/pkg/client"
	"github.com/harness/cli/v3/pkg/cmdctx"
	"github.com/harness/cli/v3/pkg/hlog"

	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/gitimporter"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/plugin/pkg/bridge"
)

const migrateBundleToRepositoryID = "migrate_scm_bundle_repository"

// migrateBundleToRepository creates Harness Code repositories from a bundle an
// export wrote. It calls the same engine constructor as the standalone
// `harness-migrate git-import` command; the endpoint, token and destination
// space come from the host's resolved profile instead of from flags.
func migrateBundleToRepository(ctx *cmdctx.Ctx) error {
	// --from is declared presence: required in the spec, so it is non-empty
	// here. --to is presence: none — the destination is the profile's scope.
	zipPath, err := resolveBundleZip(ctx.MigrateFrom)
	if err != nil {
		return err
	}
	space, err := harnessSpace(ctx.Auth)
	if err != nil {
		return err
	}
	// SSO sessions have no PAT. Testing PATToken directly (rather than AuthType)
	// also covers HARNESS_API_KEY mode, where AuthType is left unset.
	if ctx.Auth.PATToken == "" {
		return errors.New("import doesn't support SSO sessions — run 'harness auth login' with an API token " +
			"(no --sso), or set HARNESS_API_KEY")
	}

	repository := strings.Trim(cmdctx.GetString(ctx.FlagValues, "repo"), "/")
	fileSizeLimit, err := int64Flag(ctx.FlagValues, "file-size-limit")
	if err != nil {
		return err
	}
	batchSize, err := int64Flag(ctx.FlagValues, "batch-size")
	if err != nil {
		return err
	}

	flags := gitimporter.Flags{
		SkipUsers:     cmdctx.GetBool(ctx.FlagValues, "skip-users"),
		FileSizeLimit: fileSizeLimit,
		NoPR:          cmdctx.GetBool(ctx.FlagValues, "no-pr"),
		NoWebhook:     cmdctx.GetBool(ctx.FlagValues, "no-webhook"),
		NoRule:        cmdctx.GetBool(ctx.FlagValues, "no-rule"),
		NoLabel:       cmdctx.GetBool(ctx.FlagValues, "no-label"),
		NoGit:         cmdctx.GetBool(ctx.FlagValues, "no-git"),
		PRBatchSize:   int(batchSize),
	}

	// Printed directly, ahead of the tracer: the progress bar renders a blank
	// spinner frame to stderr the moment it's constructed, with no trailing
	// newline, so whatever prints next (on either stream) lands glued to that
	// frame. Announcing identity before the tracer exists avoids the collision.
	identity, isSAT := resolveIdentity(ctx)
	kind := "user"
	if isSAT {
		kind = "service account"
	}
	accountName := resolveAccountName(ctx)
	fmt.Printf("Importing into %s\n", spaceDisplay(ctx.Auth, accountName))
	fmt.Printf("Authenticating as %s (%s)\n", identity, kind)
	fmt.Println("Commit history, including authors, remains intact.")
	if flags.SkipUsers {
		fmt.Println("Unmapped pull request, comment and branch-rule member authors will be attributed to this identity (--skip-users).")
	} else {
		fmt.Printf("The import stops on any pull request author, commenter or branch-rule member whose email isn't a Harness user.\n")
		fmt.Printf("    Pass --skip-users to attribute those to this identity instead.\n\n")
	}

	tracer := bridge.NewTracer(cmdctx.GetBool(ctx.FlagValues, "no-progress"))
	defer tracer.Close()

	requestID := uuid.New().String()
	endpoint, _ := strings.CutSuffix(ctx.Auth.APIUrl, "/")
	reporter := make(map[string]*report.Report)

	importer := gitimporter.NewImporter(
		endpoint, space, repository, ctx.Auth.PATToken, zipPath,
		requestID, false, cmdctx.GetBool(ctx.FlagValues, "trace"),
		flags, tracer, reporter,
	)

	hlog.Debug("starting scm_bundle:repository migration", "zip", zipPath, "space", space, "repository", repository)
	tracer.Log("importing %s with id: %s", zipPath, requestID)
	return importer.Import(ctx.Context)
}

// resolveBundleZip turns however a command named a bundle — the import's --from,
// the id on `update scm_bundle:users` — into the zip path the engine reads.
//
// A file is the zip itself, which is all the standalone CLI accepted. A folder is
// an export's output folder, and the zip inside it is found with the same rule the
// export used to write it — an export leaves nothing else behind, so
// `--from harness` mirrors the export's `--to harness`. Either way the result is
// opened before the caller announces itself, so a wrong path fails naming the
// path rather than as "not a valid zip file" from deep in the engine. Messages
// here name no flag, since the callers spell the bundle differently.
func resolveBundleZip(bundle string) (string, error) {
	path := filepath.Clean(bundle)
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("cannot read bundle %q: %w", bundle, err)
	}
	if info.IsDir() {
		path = gitexporter.ZipFilePath(path)
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("folder %q holds no %s; name the zip file itself if it goes by another name: %w",
				bundle, gitexporter.ZipFileName, err)
		}
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("%q is not a readable zip file: %w", path, err)
	}
	_ = zr.Close()
	return path, nil
}

// harnessSpace renders the resolved scope as the account/org/project path the
// importer creates repositories under — the value the standalone CLI takes as
// --space. Code repositories are multi-level, so an org or account scope is
// legitimate; only a project without an org is not.
func harnessSpace(a *auth.ResolvedAuth) (string, error) {
	if a.AccountID == "" {
		return "", errors.New("the resolved profile has no account; run 'harness auth login'")
	}
	if a.ProjectID != "" && a.OrgID == "" {
		return "", errors.New("a project scope needs an org too; pass --org or run 'harness auth setscope'")
	}
	parts := []string{a.AccountID}
	if a.OrgID != "" {
		parts = append(parts, a.OrgID)
	}
	if a.ProjectID != "" {
		parts = append(parts, a.ProjectID)
	}
	return strings.Join(parts, "/"), nil
}

// resolveIdentity looks up who the importing token authenticates as, so the
// import can announce whose account unmapped pull requests and comments will
// be attributed to before it creates anything. Best-effort: a failed lookup
// falls back to a generic label rather than blocking the import over it.
func resolveIdentity(ctx *cmdctx.Ctx) (identity string, isSAT bool) {
	cl := client.New(ctx)
	if auth.TokenType(ctx.Auth.PATToken) == auth.TokenKindSAT {
		result, _, err := cl.PostRaw("/ng/api/token/validate", nil, ctx.Auth.PATToken, "text/plain")
		if err != nil {
			return "service account", true
		}
		return identityDisplay(jsonStringAt(result, "data", "username"), jsonStringAt(result, "data", "email"), "service account"), true
	}
	result, _, err := cl.Get("/ng/api/user/currentUser", nil)
	if err != nil {
		return "current user", false
	}
	return identityDisplay(jsonStringAt(result, "data", "name"), jsonStringAt(result, "data", "email"), "current user"), false
}

// identityDisplay renders a resolved name/email pair for the header, quoting
// the name the same way spaceDisplay quotes the account name. Falls back to
// the bare email, then to fallback, as either field may be missing.
func identityDisplay(name, email, fallback string) string {
	switch {
	case name != "" && email != "":
		return fmt.Sprintf("%q (%s)", name, email)
	case email != "":
		return email
	case name != "":
		return fmt.Sprintf("%q", name)
	default:
		return fallback
	}
}

// resolveAccountName looks up the destination account's display name for the
// header. Best-effort: a failed lookup falls back to the account ID rather
// than blocking the import over it.
func resolveAccountName(ctx *cmdctx.Ctx) string {
	result, _, err := client.New(ctx).Get("/ng/api/accounts/"+ctx.Auth.AccountID, nil)
	if err != nil {
		return ctx.Auth.AccountID
	}
	if name := jsonStringAt(result, "data", "name"); name != "" {
		return name
	}
	return ctx.Auth.AccountID
}

// spaceDisplay renders the destination path for the header, substituting the
// resolved account name for the account ID it leads with.
func spaceDisplay(a *auth.ResolvedAuth, accountName string) string {
	first := a.AccountID
	if accountName != "" && accountName != a.AccountID {
		first = fmt.Sprintf("%q (%s)", accountName, a.AccountID)
	}
	parts := []string{first}
	if a.OrgID != "" {
		parts = append(parts, a.OrgID)
	}
	if a.ProjectID != "" {
		parts = append(parts, a.ProjectID)
	}
	return strings.Join(parts, "/")
}

func jsonStringAt(v any, keys ...string) string {
	for _, k := range keys {
		m, ok := v.(map[string]any)
		if !ok {
			return ""
		}
		v = m[k]
	}
	s, _ := v.(string)
	return s
}

// int64Flag parses a numeric flag. Spec flags are only ever strings or bools, so
// the conversion lives here rather than being silently swallowed by
// cmdctx.GetInt — a typo'd batch size should not turn into zero.
func int64Flag(fv map[string]any, name string) (int64, error) {
	raw := cmdctx.GetString(fv, name)
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("--%s must be a whole number, got %q", name, raw)
	}
	return n, nil
}
