# Git migrator for Github
We support migrating these entities from Github:
- Repository
- LFS objects *(requires [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) and [git-lfs](https://git-lfs.com/) to be installed)*
- Repository Public/Private status
- Pull requests
- Pull request comments
- Pull request review comments
- Labels
- Webhooks
- Branch Rules

Items that would not imported or imported differently:
- Task lists: Task lists are imported as normal comments
- Emoji reactions
- Pull request reviewers and approvers
- Any attachment
- Webhooks: Some webhook events are not supported. You can check supported triggers [here](https://apidocs.harness.io/tag/webhook#operation/createWebhook)

### Estimating export duration
Export will depend on the size of repo and its pull request. A repo which has more pull request but less comments will take more time than one which has more comments and lesser pull requests.

## Prerequisites
To export projects from Github, you must have admin write access in for the project to successfully export all the supported entities. 

If your repository has Git Large File Storage (LFS) objects which you want to migrate, you must have [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) and [git-lfs](https://git-lfs.com/) to be installed where you run the migrator (or [run the migrator in Docker](../../README.md#using-docker)).

### Users
All the users encountered anywhere are stored by email and can be found in users.json in the exported zip file.

### Installing
You can install the migrator via github releases or run `make build` with latest go version present in your system.

## Branch Protection and Webhooks
When they are exported, supported Github branch protection rules and webhooks are stored in zip file, which later during import to harness code are mapped according to:

### Webhooks
| Github events | Harness Code events
|---|---|
| Branch or tag creation |	Branch Created, Tag Created |
| Branch or tag deletion |	Branch Deleted, Tag Deleted |
| Pull requests | PR created, PR updated, PR closed, PR reopened, PR merged |
| Pull request review comments, Commit comments	| PR comment created |
| Pushes | Branch updated, Tag updated, PR branch updated |


### Branch protection rules 
| Github rule(set) | Harness Code rule
|---|---|
| Restrict creations | Block branch creation |
| Restrict deletions | Block branch deletion |
| Restrict updates   | Block branch update | 
| Require linear history |  Require pull request |
| Require a pull request before merging |  Require pull request |
| Dismiss stale pull request approvals when new commits are pushed |  Require approval of new changes |
| Require approval of the most recent reviewable push | Require approval of new changes |
| Require review from Code Owners | Require review from code owners |
| Require conversation resolution before merging | Require comment resolution |
| Block force pushes | Block force push |

## Commands 
As a quick start you can run 
```
./migrator github git-export --org <organization name> --repository <repo-name> --host <host-url> --username <github-username> --token <token> <zip-folder-path> 
```
where you have to replace all values enclosed in brackets `<>`.

You can also provide more advanced options. You can look at those via help: 
```
./migrator github git-export --help
```

Application also supports advanced option like `resume` which can help you resume run from last successful run and avoid overhead of re-running the same commands.

## Troubleshooting
### General
#### Export fails due to reach the Github rate limit
If project export fails due to reaching the Github API rate limit, you could wait for an hour and re-run the migrator or exclude exporting metadata (options available are `--no-pr, --no-comment, --no-webhook, and --no-rule`)

#### Export fails due to missing permission
If you see errors in listing webhooks, make sure the provided token has admin permission.

#### Missing webhooks or branch rules
If you see missing items for any webhooks or branch rules you can refer `ExporterLogs.log` file in root of zip folder.

#### Webhooks don't have all the events
As of now all webhook events are not supported and you can check `ExporterLogs.log` file to get error logs. 