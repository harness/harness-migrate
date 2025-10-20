## Support
Import repositories to Harness Code Repository with metadata including:
- Pull Requests and comments
- Webhooks
- Branch Protection Rules

### Installing
You can install the migrator via github releases or run `make build` with latest go version present in your system. You can skip this step if you have the migrator installed already.

### NOTE 
You need to export the repository first using the same migrator tool. Supported SCM providers are [Github](../github/README.md), [Gitlab](../gitlab/README.md), [Bitbucket](../bitbucket/README.md), and [Bitbucket Server (Stash)](../stash/README.md).

## Commands 
As a quick start you can run 

```sh
./harness-migrate git-import <zip-folder-path>  --space <target acc/org/project> --endpoint <harness-url> --token <token>
```

where you have to replace all values enclosed in brackets `<>`.

You can also provide more advanced options. You can look at those via help: 
```
./harness-migrate git-import --help
```

### Examples

#### Standard Migration
Importing a repository without mapping users, skip the meta data, and increase the file-size-limit on server to 102MB.
```sh
export harness_TOKEN={harness-pat}
./harness-migrate git-import ./harness/harness.zip  --space "acc/MyOrg/Myproject" --endpoint "https://app.harness.io/"  --skip-users  --skip-pr --skip-webhook --skip-rule --file-size-limit 102000000
```

## Incremental Migration

The `--no-git` flag enables incremental migration for repositories that **already exist on Harness Code**. This feature allows you to migrate additional pull request metadata from your source SCM without re-importing the git repository itself.

### When to Use Incremental Migration

Use incremental migration when:
- You have already imported a repository to Harness Code using the standard migration process
- You want to import additional historical pull requests and metadata from the source SCM

### How Incremental Migration Works

1. **Repository Validation**: The tool verifies that the target repository already exists on Harness Code
2. **Offset Calculation**: Automatically determines the current highest PR number on Harness Code to avoid conflicts
3. **PR Number Adjustment**: Source PR numbers are increased by the calculated offset before importing
4. **Metadata Import**: Only pull request data, comments, and metadata are imported - git repository content is skipped

### Important Considerations

⚠️ **Repository Unavailability**: During incremental migration, the target repository will be temporarily unavailable on Harness Code.

⚠️ **Backup Recommended**: Create a backup of your Harness Code repository before proceeding with incremental migration.

### PR Number Mapping Example

If your Harness Code repository has 100 existing PRs and you're migrating 50 historical PRs from GitHub:

- **Existing Harness PRs**: #1 - #100 (remain unchanged)
- **Migrated GitHub PRs**: GitHub PR #1 becomes Harness PR #101, GitHub PR #50 becomes Harness PR #150
- **Future PRs**: Will continue from #151 onwards

This ensures no conflicts between existing and migrated pull request numbers.

### Usage

```sh
./harness-migrate git-import <zip-folder-path> --space <target-space> --endpoint <harness-url> --token <token> --no-git
```

### Prerequisites
- Target repository must already exist on Harness Code
- You must have write permissions to the target repository

### Example
Migrating additional pull requests to an existing repository on Harness Code:
```sh
export harness_TOKEN={harness-pat}
./harness-migrate git-import ./additional-prs.zip --space "acc/MyOrg/Myproject" --endpoint "https://app.harness.io/" --no-git
```

This will:
- Skip git repository import (since it already exists)
- Automatically calculate PR number offset to avoid conflicts
- Import only pull request metadata with adjusted PR numbers

## Troubleshooting

### Import fails due to

| Error | Resolution |
|---|---|
| Resource not found | Target space must exist on Harness. You can import repositories into account, organization or projects. |
| Unauthorized/Forbidden | Provided token must be able to create and edit a repository. |
| Forbidden | Contact support to enable `CODE_IMPORTER_V2_ENABLED` FF for your account. |
| Push contains files exceeding the size limit | Increase `--file-size-limit`, default is 100MB. |
| `failed to import pull requests and comments for repo 'TARGET-SPACE/REPO' : client error 413:` | The pull requests payload is too large. Use `--batch-size` with a smaller value (e.g., `--batch-size 50`). The default is 100. |

### Users
When exporting repositories, we collect information about users who have interacted with them. This data is saved in a `users.json` file within the exported zip. During the import process, if a user exists on the Harness platform with the same email, their activities will be preserved. However, if the user is missing, you have two options: using `--skip-users` which skips mapping their activities to the migrator (determined by the token you provided) or manually create the user on the target platform first before the import.

### Webhooks and Branch Rules names
Imported repositories could have a diff name for webhooks and branch rules. (e.g., names start with `webhook_` or `rule_`)

