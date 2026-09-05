# harness-migrate as a Harness unified CLI plugin

This is a second front-end for the same engine the standalone `harness-migrate`
binary drives. It is a [Harness unified CLI](https://github.com/harness/cli)
plugin: a separate binary the `harness` host execs, with its command grammar
declared in [migrate.spec.yaml](pkg/migrateplugin/migrate.spec.yaml).

No behavior under `../internal` or `../types` changes. The handlers here call the
same constructors the kingpin `run()` bodies in `../cmd` call — this adds a
caller, it does not replace one. The single edit outside this directory is
`gitexporter.getZipFilePath` becoming exported as `ZipFilePath`, so the export
and the import share one rule for where a bundle's zip lives.

## Layout

| Path | Role |
| --- | --- |
| `cmd/harness-migrate/main-harness-migrate.go` | plugin entrypoint (`--identity`, `--spec`, dispatch) |
| `pkg/migrateplugin/migrate.spec.yaml` | the grammar: nouns, commands, flags |
| `pkg/migrateplugin/migrateplugin.go` | `ModuleInit` — binds workflow IDs to handlers |
| `pkg/migrateplugin/migrate_github.go` | `github_organization:scm_bundle` handler |
| `pkg/migrateplugin/migrate_gitlab.go` | `gitlab_group:scm_bundle` handler |
| `pkg/migrateplugin/migrate_bitbucket.go` | `bitbucket_workspace:scm_bundle` handler |
| `pkg/migrateplugin/migrate_stash.go` | `stash_project:scm_bundle` handler |
| `pkg/migrateplugin/migrate_import.go` | `scm_bundle:repository` handler — the import half |
| `pkg/migrateplugin/update_users.go` | `update scm_bundle:users` handler — rewrite emails in a bundle |
| `pkg/migrateplugin/client.go` | shared bearer-token `scm.Client` transport |
| `pkg/bridge/` | shared glue: tracer, interrupt handling — reused by every handler |

Its own Go module, because the parent is on Go 1.23 and `github.com/harness/cli/v3`
requires 1.26. `github.com/harness/cli/v3` is a plain version requirement,
pulled from its published tags like any other dependency.

## Build and install

Plugin builds use [Task](https://taskfile.dev), from [../Taskfile.yml](../Taskfile.yml).
The root `Makefile` and `.drone.yml` build the standalone binary and are
deliberately untouched — that front-end is being deprecated, so the two build
systems share nothing.

```sh
task build      # -> plugin/bin/harness-migrate
task install    # harness install plugin plugin/bin/harness-migrate
harness migrate github_organization:scm_bundle --from myorg --github-token "$GITHUB_TOKEN"
harness update scm_bundle:users harness --user-mapping users.json   # optional
harness migrate scm_bundle:repository --from harness
```

The binary must be named `harness-<plugin-name>`, so it is also
`harness-migrate` — hence `plugin/bin/` rather than the repo root, which the
standalone build already occupies.

Useful while iterating (`task --list` for the rest):

```sh
task dev         # install into plugin/devhome, leaving ~/.harness alone
task check       # fmt, vet, spec validation, tests
task identity    # the JSON the host reads at install time
task spec        # the grammar YAML the host stores at install time
```

If the host CLI isn't on your PATH: `task install HARNESS=../cli/bin/harness`
(relative paths resolve from the repo root, not from `plugin/`).

Version stamping follows core's convention — a `migrate/v*` tag series separate
from the standalone binary's `v*` tags, with dev builds reporting the next patch
(`0.1.0-dev` until the first tag exists). Override with `task build PLUGIN_VERSION=0.1.0`.

## Status

Both phases are wired up under the `migrate <from>:<to>` pair verb — four
exports: `migrate github_organization:scm_bundle` (`../cmd/github/git.go`),
`migrate gitlab_group:scm_bundle` (`../cmd/gitlab/git.go`),
`migrate bitbucket_workspace:scm_bundle` (`../cmd/bitbucket/git.go`),
`migrate stash_project:scm_bundle` (`../cmd/stash/git.go`); and the import,
`migrate scm_bundle:repository` (`../cmd/gitimporter`). The one command that
edits a bundle between the two phases is ported too: `update scm_bundle:users`
(`../cmd/users`). That is every code-related command in `../cmd`; what is left
there is pipeline conversion (`circle`, `drone`, `jenkinsxml`, `travis`,
`terraform`, `cloudbuild`, and each provider's `convert`). Everything below is
known-open, not overlooked.

- **The import has no `--to`** (`presence: none`): its destination is a Harness
  *scope*, not a repo id, so it comes from the profile plus the global
  `--org`/`--project`, and is rendered into the `account/org/project` path the
  engine takes as `--space`. `repository` is `multi_level`, so an org- or
  account-scoped space is legitimate; only a project without an org is rejected.
- **Foreign noun aliases don't survive plugin dispatch.** The host resolves
  `migrate scm_bundle:repo` (via `code`'s `repo` alias) and execs this binary with
  that spelling, but this process loads only its own spec, so it has no `repository`
  noun to source aliases from and rejects it. Aliases on nouns declared *here*
  (`github_org`, `bitbucket_server_project`) work fine. Core would have to
  canonicalize the pair before exec'ing a plugin.
- **The import needs an API-token profile.** `internal/harness` authenticates with
  `x-api-key`, which an SSO profile's bearer token can't fill, so the handler
  checks `Auth.PATToken` and errors with a login hint. It checks the token rather
  than `Auth.AuthType`, which core leaves unset in `HARNESS_API_KEY` mode.
- **A bundle is named by a folder or a zip, everywhere.** A file is the zip, which
  is all the standalone CLI accepted; a folder is an export's output folder, and
  the zip inside is located with `gitexporter.ZipFilePath` — the same rule that
  wrote it. An export leaves nothing else in that folder, so `--from harness` is
  unambiguous and mirrors the export's `--to harness`. The resolved path is opened
  as a zip before the caller announces itself, so the four failure modes (missing
  path, folder without a `harness.zip`, unreadable zip, valid but empty) each
  report themselves by name instead of surfacing as `zip: not a valid zip file`
  from inside the engine. `resolveBundleZip` is shared by the import's `--from`
  and the id on `update scm_bundle:users`, so its messages name no flag.
- **`id_allow_slash: true` is mandatory on it.** The bundle is named by a path, and
  `validateIdParts` rejects any id containing `/` unless a command opts out. Without
  it `./out/harness.zip` fails before the handler runs.
- **The bundle is rewritten in place, with no `--out` alternative.** The host adds
  `--out` to every workflow command for output redirection, so it can't name a
  destination zip. That matches the standalone command, which also replaced its
  `--zipFilePath`. The engine renames over the original only after the rewrite
  succeeds, so an interrupt leaves the bundle intact — which is just as well,
  because `users.Updater.Update` accepts a `context.Context` and never reads it:
  `bridge.WithInterrupt` cannot stop a rewrite in flight.
- **The rewrite's scratch space is `./harness-updated` in the process CWD** —
  `internal/users` hardcodes it and `os.RemoveAll`s it on the way out, so a folder
  of that name in the caller's directory is clobbered and deleted. Left alone
  because the fix belongs in `internal/users`, which this port doesn't touch.
- **`--gitness` is dropped.** Importing into a Gitness instance needs an endpoint
  and a raw `Authorization` header that no Harness profile describes, so that
  target belongs to the standalone binary.
- **`--endpoint`, `--token` and `--space` are dropped** in favour of the resolved
  profile — that substitution is most of the point of being a plugin. The
  standalone `--skip-pr`/`--skip-label`/`--skip-webhook`/`--skip-rule` aliases are
  gone too: spec flags have no alias field, and `--no-*` matches the exports.
- **Numeric flags are strings.** `spec.Flag` is only string, bool or array, so
  `--file-size-limit` and `--batch-size` are declared with string `default:`s and
  parsed in the handler, which errors on a non-number rather than letting
  `cmdctx.GetInt` silently yield a batch size of zero.
- **Credentials and endpoints keep the provider prefix** (`--github-token`, not
  `--token`), because the noun says which *system* but not whose *credential* —
  and in this CLI a bare `--token` reads as a Harness token, which is exactly
  what `git-import --token` takes. Narrowing and behavior flags stay bare
  (`--repo`, `--resume`, `--no-pr`): they mean the same thing on all four
  providers. GitLab's `--repo` would be `--project` in GitLab's own vocabulary,
  but `--project` is global.
- **Globals to avoid when adding flags:** `--org`, `--project`, `--profile`,
  `--debug`, `--timeout` are global, and `--out`, `--format`, `--json`, `--yaml`,
  `--raw` are added to every workflow command.
- **`--to` has no spec-level default,** so each handler falls back to `harness`
  itself. `migrate_from`/`migrate_to` carry `label` and `presence`
  (`required`/`optional`/`none`) but not a default value.
- **Progress still goes through migrate's own `tracer`,** not `pkg/console`, so a
  migration doesn't look quite like the rest of `harness`. It keeps the animated
  progress bar, which core can't yet reproduce: `pkg/console` allows one animated
  stage per process and can't end one without printing a success/fail marker,
  while the engine nests `Start`/`Stop` across `errgroup` workers and reports
  failures through `Stop`. `pkg/bridge` overrides only `Debug()`, routing it to
  `hlog` so nothing needs to read the host's `--debug` flag.
- **No env-var fallbacks.** The standalone CLI accepts `github_ORG` and friends
  via kingpin's `Envar()`, but never documented them, so nothing published is
  lost. Core's spec schema has no equivalent; it can be added if asked for.
- **No Harness auth is used** (`no_auth: true`); an export only talks to the
  source provider. The import side is where `ctx.Auth` starts to matter.
