## Support
Import repository to Harness Code Repository with metadata including:
- Pull Requests and comments
- Webhooks
- Branch Protection Rules

### Installing
You can install the migrator via github releases or run `make build` with latest go version present in your system.

### NOTE 
You need to export the repository first using the same migrator tool. Supported SCM providers are Bitbucket Server (Stash), and Github (coming soon).

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

Importing a repository without mapping users and skip the meta data.
```sh
export harness_TOKEN={harness-pat}
./migrator git-import ./harness/harness.zip  --space "acc/MyOrg/Myproject" --endpoint "https://app.harness.io/"  --skip-users  --skip-pr --skip-webhook --skip-rule --file-size-limit 102000000
```

## Troubleshooting

### Import fails due to

| Error | Resolution |
|---|---|
| Resource not found | Target space must exist on Harness. You can import repositories into account, organization or projects |
| Unauthorized/Forbidden | Provided token must be able to create and edit a repository. |
| Forbidden | Contact support to enable FF `CODE_IMPORTER_V2_ENABLED` for your account. |
| Push contains files exceeding the size limit | Increase `--file-size-limit`, default is 100MB. |

### Users
All the users encountered anywhere are stored by email and can be found in users.json in the exported zip file. Users need to exist on the harness platform if you'd want to keep original activities or manually update them to a dummy email. Using `--skip-users` will map non-existing users to the migrator (from the token you passed)

### Webhooks and Branch Rules names
Imported repositories could have a diff name for webhooks and branch rules. (e.g., names start with `webhook_` or `rule_`)

