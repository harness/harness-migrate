# Git migrator for Gitlab
We support migrating these entities from Gitlab:
- Code Repository
- Repository Public/Private status
- Merge requests
- Merge requests comments
- Webhooks

Items that would not imported or imported differently:
- Branch Protection Rules
- Labels
- Emoji reactions
- Merge request reviewers and approvers
- Any attachment
- LFS objects
- Webhooks: Some webhook events are not supported. You can check supported triggers [here](https://apidocs.harness.io/tag/webhook#operation/createWebhook)

### Estimating export duration
Export will depend on the size of repo and its merge request. A repo which has more merge request but less comments will take more time than one which has more comments and lesser merge requests.

## Prerequisites
To export projects from Gitlab, you must have admin write access in for the group/project to successfully export all the supported entities. 

### Users
All the users encountered anywhere are stored by email and can be found in users.json in the exported zip file.

### Installing
You can install the migrator via github releases or run `make build` with latest go version present in your system.

## Webhooks
When they are exported, supported Gitlab webhooks are stored in zip file, which later during import to harness code are mapped according to:

### Webhooks
| Gitlab events | Harness Code events
|---|---|
| Push events |	Branch Created, Branch Updated, Branch Deleted, PR Branch Updated |
| Tag Push events |	Tag Created, Tag Updated, Tag Deleted |
| Comments | PR comment created |
| Merge request events	| PR created, PR updated, PR closed, PR reopened, PR merged  |

#### Merge Request Comments
Code comments attached to specific lines will be visible in the "Changes" tab of the pull request, not in the "Conversation" tab.

## Commands 
As a quick start you can run 
```
./migrator gitlab git-export --group <group name/including subgroups> --project <project-name> --host <host-url> --username <gitlab-username> --token <token> <zip-folder-path> 
```
where you have to replace all values enclosed in brackets `<>`. You can pass Gitlab Personal Access Token or Group/Project Access token given your use cases.

You can also provide more advanced options. You can look at those via help: 
```
./migrator gitlab git-export --help
```

Application also supports advanced option like `resume` which can help you resume run from last successful run and avoid overhead of re-running the same commands.

## Troubleshooting
### General
#### Export fails due to reach the Gitlab rate limit
If project export fails due to reaching the Gitlab API rate limit, you could wait for an hour and re-run the migrator or exclude exporting metadata (options available are `--no-pr, --no-comment, --no-webhook, and --no-rule`)


#### Missing webhooks
If you see missing items for webhooks you can refer `ExporterLogs.log` file in root of zip folder.

#### Webhooks don't have all the events
As of now all webhook events are not supported and you can check `ExporterLogs.log` file to get warning logs. 