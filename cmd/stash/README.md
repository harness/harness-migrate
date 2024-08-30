## Support
Stash or bitbucket on-prem has multiple entities for which we support migration which are:
- Repository
- Repository Public/Private status
- Pull requests
- Pull request comments
- Pull request review comments
- Webhooks
- Branch Rules

Items that would not imported or imported differently:
- Task lists: Task lists are imported as normal comments
- Emoji reactions
- Pull request reviewers and approvers
- Any attachment
- LFS objects
- Webhooks: Some webhook events are not supported. You can check supported triggers [here](https://apidocs.harness.io/tag/webhook#operation/createWebhook)

### Estimating export duration
Export will depend on the size of repo and its pull request. A repo which has more pull request but less comments will take more time than one which has more comments and lesser pull requests.

## Prerequisites
To export projects from stash, you must have admin access in for the project to successfully export all the supported entities. 

### Users
All the users encountered anywhere are stored by email and can be found in users.json in the exported zip file.

### Installing
You can install the migrator via github releases or run `make build` with latest go version present in your system.

## Commands 
As a quick start you can run 
```
./migrator stash git-export --project <project name> --repository <repo-name> --host <host-url> --username <stash-username> --token <token> <zip-folder-path> 
```
where you have to replace all values enclosed in brackets `<>`.

You can also provide more advanced options. You can look at those via help: 
```
./migrator stash git-export --help
```

Application also supports advanced option like `resume` which can help you resume run from last successful run and avoid overhead of re-running the same commands.

## Troubleshooting
### General
#### Export fails due to unresolved host
If project export fails due to unresolved host make sure bitbucket server is reachable from the machine which is running the migrator.

#### Missing webhooks or branch rules
If you see missing items for any webhooks or branch rules you can refer `ExporterLogs.log` file in root of zip folder.

#### Webhooks don't have all the events
As of now all webhook events are not supported and you can check `ExporterLogs.log` file to get error logs.


