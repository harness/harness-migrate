# Git migrator for Gitlab
We support migrating these entities from Gitlab:
- Code Repository
- LFS objects *(requires [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) and [git-lfs](https://git-lfs.com/) to be installed)*
- Repository Public/Private status
- Merge requests
- Merge requests comments
- Webhooks
- Branch Protection Rules

Items that would not imported or imported differently:
- Labels
- Emoji reactions
- Merge request reviewers and approvers
- Any attachment
- Webhooks: Some webhook events are not supported. You can check supported triggers [here](https://apidocs.harness.io/tag/webhook#operation/createWebhook)

### Estimating export duration
Export will depend on the size of repo and its merge request. A repo which has more merge request but less comments will take more time than one which has more comments and lesser merge requests.

## Prerequisites
To export projects from Gitlab, you must have admin write access in for the group/project to successfully export all the supported entities. 

If your repository has Git Large File Storage (LFS) objects which you want to migrate, you must have [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) and [git-lfs](https://git-lfs.com/) to be installed where you run the migrator (or [run the migrator in Docker](../../README.md#using-docker)).

### Users
All the users encountered anywhere are stored by email and can be found in users.json in the exported zip file.

### Installing
You can install the migrator via github releases or run `make build` with latest go version present in your system.

## Webhooks
When they are exported, supported Gitlab webhooks are stored in zip file, which later during import to harness code are mapped according to:

### Branch Protection and Webhooks
When they are exported, supported Gitlab branch protection rules and webhooks are stored in zip file, which later during import to harness code are mapped according to:

### Webhooks
| Gitlab events | Harness Code events
|---|---|
| Push events |	Branch Created, Branch Updated, Branch Deleted, PR Branch Updated |
| Tag Push events |	Tag Created, Tag Updated, Tag Deleted |
| Comments | PR comment created |
| Merge request events	| PR created, PR updated, PR closed, PR reopened, PR merged  |

### Branch protection rules 
| Gitlab branch protection rule | Harness Code rule
|---|---|
| Allowed to push and merge | Require pull request |
| Allowed to merge | Block branch update |
| Allow Force Push   | Dont Block force push | 
| All threads must be resolved |  Require comment resolution |
| Enable "Delete source branch" option by default |  Auto delete branch on merge |
| Require approval from code owners |  Require review from code owners |

Here's how Merge requests settings applies as branch rule:
| Gitlab merge method | Harness Code merge strategies (allowed)
|---|---|
| Merge, Squash is not required | Merge, Squash and merge|
| Merge, Squash commits "Required" when merging | Squash and merge |
| Merge commit with semi-linear history | Merge, Rebase, Squash and merge |
| Fast-forward merge | Rebase, Fast-forward, Squash and merge |

GitLab allows users to bypass branch protection rules based on their roles or specific user/group settings. After migration, users will be transferred, but roles won’t be. 

Note Gitlab branch protection rules follow the "most permissive" rule if multiple rules can apply to a single branch. Harness Code follows the "Apply all" rule.

Group-level merge request approvals need to be set manually after the migration. Refer to `Require a minimum number of reviewers` rule in the [Docs](https://developer.harness.io/docs/code-repository/config-repos/rules/#available-rules).

#### Merge Request Comments
Code comments attached to specific lines will be visible in the "Changes" tab of the pull request, not in the "Conversation" tab.

## Commands 
As a quick start you can run 
```
./migrator gitlab git-export --group <group name/including subgroups> --project <project-name> --host <host-url> --username <gitlab-username> --token <token> <zip-folder-path> 
```
where you have to replace all values enclosed in brackets `<>`. You can pass Gitlab Personal Access Token or Group/Project Access token given your use cases. Please include subgroups for `--group` arg if you are exporting an individual project otherwise only include the group name (w/o subgroups).

If exporting from Gitlab deployed on premise, the `--host` flag with the Gitlab host URL is required.

You can also provide more advanced options. You can look at those via help: 
```
./migrator gitlab git-export --help
```

Application also supports advanced option like `resume` which can help you resume run from last successful run and avoid overhead of re-running the same commands.

## Troubleshooting
### General
#### Export fails due to reach the Gitlab rate limit
If project export fails due to reaching the Gitlab API rate limit, you could wait for an hour and re-run the migrator or exclude exporting metadata (options available are `--no-pr, --no-comment, --no-webhook, and --no-rule`)

#### Missing Git LFS objects
Make sure the Git Large File Storage (LFS) is enabled on your Gitlab project in Settings -> General settings -> Repository.

#### Missing webhooks
If you see missing items for webhooks you can refer `ExporterLogs.log` file in root of zip folder.

#### Webhooks don't have all the events
As of now all webhook events are not supported and you can check `ExporterLogs.log` file to get warning logs. 