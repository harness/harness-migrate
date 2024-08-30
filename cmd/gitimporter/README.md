## Support
Import repositories to Harness Code Repository with metadata including:
- Pull Requests and comments
- Webhooks
- Branch Protection Rules

### Installing
You can install the migrator via github releases or run `make build` with latest go version present in your system. You can skip this step if you have the migrator installed already.

### NOTE 
You need to export the repository first using the same migrator tool. Supported SCM providers are [Bitbucket Server (Stash)](../stash/README.md), and Github (coming soon).

## Commands 
As a quick start you can run 

```sh
./migrator git-import <zip-folder-path>  --space <target acc/org/project> --endpoint <harness-url> --token <token>
```

where you have to replace all values enclosed in brackets `<>`.

You can also provide more advanced options. You can look at those via help: 
```
./migrator git-import --help
```
## Examples

Importing a repository without mapping users, skip the meta data, and increase the file-size-limit on server to 102MB.
```sh
export harness_TOKEN={harness-pat}
./migrator git-import ./harness/harness.zip  --space "acc/MyOrg/Myproject" --endpoint "https://app.harness.io/"  --skip-users  --skip-pr --skip-webhook --skip-rule --file-size-limit 102000000
```

## Troubleshooting

### Import fails due to

| Error | Resolution |
|---|---|
| Resource not found | Target space must exist on Harness. You can import repositories into account, organization or projects. |
| Unauthorized/Forbidden | Provided token must be able to create and edit a repository. |
| Forbidden | Contact support to enable `CODE_IMPORTER_V2_ENABLED` FF for your account. |
| Push contains files exceeding the size limit | Increase `--file-size-limit`, default is 100MB. |

### Users
When exporting repositories, we collect information about users who have interacted with them. This data is saved in a `users.json` file within the exported zip. During the import process, if a user exists on the Harness platform with the same email, their activities will be preserved. However, if the user is missing, you have two options: using `--skip-users` which skips mapping their activities to the migrator (determined by the token you provided) or manually create the user on the target platform first before the import.

### Webhooks and Branch Rules names
Imported repositories could have a diff name for webhooks and branch rules. (e.g., names start with `webhook_` or `rule_`)

