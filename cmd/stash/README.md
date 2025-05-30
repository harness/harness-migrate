# Git Migrator for Bitbucket onprem

## Support
Stash or bitbucket on-prem has multiple entities for which we support migration which are:
- Repository
- LFS objects *(requires [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) and [git-lfs](https://git-lfs.com/) to be installed)*
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
- Webhooks: Some webhook events are not supported. You can check supported triggers [here](https://apidocs.harness.io/tag/webhook#operation/createWebhook)

### Estimating export duration
Export will depend on the size of repo and its pull request. A repo which has more pull request but less comments will take more time than one which has more comments and lesser pull requests.

## Prerequisites
To export projects from stash, you must have admin access in for the project to successfully export all the supported entities. 


If your repository has Git Large File Storage (LFS) objects which you want to migrate, you must have [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) and [git-lfs](https://git-lfs.com/) to be installed where you run the migrator (or [run the migrator in Docker](../../README.md#using-docker)).

### Users
All the users encountered anywhere are stored by email and can be found in users.json in the exported zip file.

### Installing
You can install the migrator via github releases or run `make build` with latest go version present in your system.

## Branch Protection and Webhooks
When they are exported, supported Stash branch protection rules and webhooks are stored in zip file, which later during import to harness code are mapped according to:

### Webhooks
| Bitbucket Server events | Harness Code events
|---|---|
| Repository Push | Branch Created, Branch Updated, Branch Deleted, Tag Created, Tag updated, Tag Deleted |
| Pull Request Opened |  PR Created, PR Reopened |
| Pull Request Merged |  PR Merged |
| Pull Request Modified | PR Updated |
| Pull Request Declined | PR Closed |
| Pull Request Source branch updated | PR Branch Updated |
| Pull Request Comment Added | PR Comment Created |

### Branch protection rules 
| Bitbucket Server rule | Harness Code rule
|---|---|
| Prevent all changes | Block branch creation, Block branch deletion, Require pull request|
| Prevent deletion |Block branch deletion |
| Prevent rewriting history, Prevent changes without a pull request | Require pull request |

## Commands 
As a quick start you can run 
```
./harness-migrate stash git-export --project <project name> --repository <repo-name> --host <host-url> --username <stash-username> --token <token> <zip-folder-path> 
```
where you have to replace all values enclosed in brackets `<>`.

You can also provide more advanced options. You can look at those via help: 
```
./harness-migrate stash git-export --help
```

Application also supports advanced option like `resume` which can help you resume run from last successful run and avoid overhead of re-running the same commands.

## Troubleshooting
### General
#### Export fails due to unresolved host
If project export fails due to unresolved host make sure bitbucket server is reachable from the machine which is running the migrator.

#### Missing Git LFS objects
Make sure the Large File Storage (LFS) is allowed on your Bitbucket Server repository in Repository settings -> Repository details -> Allow LFS.

#### Missing webhooks or branch rules
If you see missing items for any webhooks or branch rules you can refer `ExporterLogs.log` file in root of zip folder.

#### Webhooks don't have all the events
As of now all webhook events are not supported and you can check `ExporterLogs.log` file to get error logs. 
