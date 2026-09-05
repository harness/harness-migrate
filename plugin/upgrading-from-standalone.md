# harness-migrate → harness: implemented commands

```sh
# Export a GitHub org's git data (repos, PRs, comments, webhooks, rules, labels, LFS) to a local bundle
harness-migrate github git-export --org my-org --token $GH_TOKEN harness
harness migrate github_organization:scm_bundle --from my-org --to ./harness --github-token $GH_TOKEN

# Export a GitLab group's (and subgroups') git data to a local bundle
harness-migrate gitlab git-export --group my-group --include-subgroups --token $GITLAB_TOKEN harness
harness migrate gitlab_group:scm_bundle --from my-group --to ./harness --include-subgroups --gitlab-token $GITLAB_TOKEN

# Export a Bitbucket Cloud workspace's git data to a local bundle
harness-migrate bitbucket git-export --workspace my-workspace --token $BITBUCKET_TOKEN harness
harness migrate bitbucket_workspace:scm_bundle --from my-workspace --to ./harness --bitbucket-token $BITBUCKET_TOKEN

# Export a Bitbucket Server (Stash) project's git data to a local bundle
harness-migrate stash git-export --host https://stash.example.com --project MYPROJ --username me --token $STASH_TOKEN harness
harness migrate stash_project:scm_bundle --from MYPROJ --to ./harness --stash-host https://stash.example.com --stash-user me --stash-token $STASH_TOKEN

# Rewrite email addresses inside a bundle from a mapping file, before importing
harness-migrate update-users user-mapping.json --zipFilePath harness/harness.zip
harness update scm_bundle:users ./harness --user-mapping user-mapping.json

# Import a bundle into Harness Code — the only step that mutates Harness
harness-migrate git-import ./harness/harness.zip --token $HARNESS_TOKEN --space account/org/project
harness migrate scm_bundle:repository --from ./harness --org my-org --project my-project
```

## What changed, mechanically

- **Positional folder → `--from`/`--to`.** Every export took a trailing positional
  (`... harness`, default `"harness"`) for the output folder. The new form is `--to ./harness`
  (also defaulting to `./harness`), and the source id moves from a named flag (`--org`, `--group`,
  `--workspace`, `--project`) to `--from`.
- **Provider tokens keep their prefix.** `--token` becomes `--github-token` / `--gitlab-token` /
  `--bitbucket-token` / `--stash-token` — disambiguates since `--token` alone would be ambiguous
  once auth flags exist at the top level.
- **`--zipFilePath` → positional id.** `update-users` named the bundle via a flag; `update
  scm_bundle:users` takes it as the command's positional id, same slot the import's `--from` uses.
  Either a folder (`./harness`) or the zip inside it (`./harness/harness.zip`) works, in both spots.
- **`--space` is gone.** The import no longer takes an explicit `account/org/project` string.
  Destination scope comes from the active profile plus the global `--org`/`--project` flags
  (`harness auth setscope`, or pass them per-invocation).
- **`--token` on import is gone too.** Auth is the profile's token (`harness auth login`), not a
  flag — the import now requires an API-token profile, since it authenticates with `x-api-key`.
- **Zip filename is fixed.** The bundle is always `harness.zip` inside whatever folder `--to`
  names — never derived from the folder's own name.

## Not yet covered by this doc

Drone, CircleCI, Terraform, and all `convert` (pipeline-YAML) subcommands have no plugin
equivalent yet.
