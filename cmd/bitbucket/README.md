# Git migrator for Bitbucket
We support migrating these entities from Bitbucket:
- Repository
- LFS objects *(requires [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) and [git-lfs](https://git-lfs.com/) to be installed)*
- Repository Public/Private status
- Pull requests
- Pull request comments
- Webhooks
- Branch Rules

Items that would not imported or imported differently:
- Pull request reviewers and approvers
- Pending tasks/comments
- Any attachment
- Webhooks: Some webhook events are not supported. You can check supported triggers [here](https://apidocs.harness.io/tag/webhook#operation/createWebhook)

### Estimating export duration
Export will depend on the size of repo and its pull request. A repo which has more pull request but less comments will take more time than one which has more comments and lesser pull requests.

## Prerequisites
To export workspaces from Bitbucket, you must have admin write access in for the workspace to successfully export all the supported entities. 

If your repository has Git Large File Storage (LFS) objects which you want to migrate, you must have [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) and [git-lfs](https://git-lfs.com/) to be installed where you run the migrator (or [run the migrator in Docker](../../README.md#using-docker)).

### Users
Due to Bitbucket's GDPR compliance, email addresses might be missing for certain user activities. Here's what you need to know:

- Git commit data (author name and email) is preserved as-is from the repository
- For other user activities (PR authors, commenters, branch restrictions), we generate placeholder emails in the format:
  `display_name.account_id@unknownemail.harness.io`

The exported zip file contains:
- Original files with placeholder emails where needed (PRs, comments, branch restrictions)
- users.json: A reference file listing all encountered users and their email status

If you have workspace admin access and want to preserve original user activities, you can replace the placeholder emails in the original files before importing to Harness Code. The users.json file helps you identify which emails need replacement.

### Installing
You can install the migrator via github releases or run `make build` with latest go version present in your system.

## Branch Protection and Webhooks
When they are exported, supported Bitbucket branch protection rules and webhooks are stored in zip file, which later during import to harness code are mapped according to:

### Webhooks
| Bitbucket events | Harness Code events
|---|---|
| Repository Push |	Branch Created, Branch Updated, Branch Deleted, Tag Created, Tag updated, Tag Deleted, PR branch updated |
| Repository Commit comment created, Pull request Comment created |  PR comment created |
| Pull request Created | PR created |
| Pull request Updated | PR updated, PR branch updated|
| Pull request Merged | PR merged|

### Branch protection rules 
| Bitbucket rule(set) | Harness Code rule
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
./harness-migrate bitbucket git-export --workspace <workspace name> --repository <repo-name> --host <host-url> --username <bitbucket-username> --token <token> <zip-folder-path> 
```
where you have to replace all values enclosed in brackets `<>`.

You can also provide more advanced options. You can look at those via help: 
```
./harness-migrate bitbucket git-export --help
```

Application also supports advanced option like `resume` which can help you resume run from last successful run and avoid overhead of re-running the same commands.

## Troubleshooting
### General
#### Missing tasks or comments on the pull requests
Please make sure to `Finish Review` on open pull requests before starting the export. Pending comments and task will not be exported.

#### Export fails due to reach the Bitbucket rate limit
If project export fails due to reaching the Bitbucket API rate limit, you could wait for an hour and re-run the migrator or exclude exporting metadata (options available are `--no-pr, --no-comment, --no-webhook, and --no-rule`)

#### Export fails due to missing permission
If you see errors in listing webhooks, make sure the provided token has admin permission.

#### Missing webhooks or branch rules
If you see missing items for any webhooks or branch rules you can refer `ExporterLogs.log` file in root of zip folder.

#### Webhooks don't have all the events
As of now all webhook events are not supported and you can check `ExporterLogs.log` file to get error logs. 