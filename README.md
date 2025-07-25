# Harness Migrator

**harness-migrate** is a command line utility to help convert and migrate
continuous integration pipelines from other providers to Harness CI. You can use this tool to migrate repositories following guidlines [below](#migrate-repositories).

**Please review [known conversion and migration issues](KNOWN_ISSUES_CONVERT.md
) before using this tool.**

## Migrate Repositories

Migrating repositories is a two-step process. 

1. **Export**: Use `git-export` to export repositories with metadata from your current SCM provider. Guidlines for [Bitbucket On-perm](cmd/stash/README.md), [GitHub](cmd/github/README.md), [Gitlab](cmd/gitlab/README.md), and [Bitbucket](cmd/bitbucket/README.md). The exported data will be saved in a zip file.

   **[OPTIONAL]** Use the `update-users` command to map user emails in the exported data to their corresponding Harness emails. See [documentation](cmd/users/README.md) for details. Without this step, unmatched emails will default to the migrator user.

2. **Import**: Import the exported data into Harness CODE using `git-import` command as explained [here](cmd/gitimporter/README.md).

### Using Docker

The migrator is available as a Docker image with all required dependencies pre-installed. Follow these steps:

1. Build and run the container:
```sh
docker build -t harness-migrator .
docker volume create migrator-data
docker run -it --name migrator -v migrator-data:/data harness-migrator
```

2. Inside the container, run your export/import commands:
```sh
# Export example
migrator [SCM-Provider] git-export [flags]

# Import example
migrator git-import [flags]
```

Note: All data is stored in the Docker volume `migrator-data` which persists even if you stop or remove the container. To clean up after you're done:
```sh
docker rm migrator
docker volume rm migrator-data
```

## Migrate Continuous Integration Pipelines 

## Install on Mac

Intel CPU

```sh
curl -L https://github.com/harness/harness-migrate/releases/latest/download/harness-migrate-darwin-amd64.tar.gz | tar zx
```

Apple silicon (M1 or M2) CPU

```sh
curl -L https://github.com/harness/harness-migrate/releases/latest/download/harness-migrate-darwin-arm64.tar.gz | tar zx
```

Copy the binary into place

```sh
sudo cp harness-migrate /usr/local/bin
```

Verify the install

```sh
harness-migrate --help
```

## Build

```term
$ git clone https://github.com/harness/harness-migrate.git
$ cd harness-migrate
$ go build
```

## Usage

### Drone

Convert a drone pipeline:

```term
harness-migrate drone convert /path/to/.drone.yml
```

Export a github namespace from drone:

```term
harness-migrate drone export \
  --namespace example \
  --github-token $GITHUB_TOKEN \
  export.json
```

❗ To avoid pipelines triggering in both your Drone instance and in Harness CI, you must first deactivate the pipelines in your Drone instance.

This script uses [jq](https://jqlang.github.io/jq/) and the [Drone CLI](https://docs.drone.io/cli/install/) to disable all pipelines defined in your `export.json`:

```bash
#!/bin/bash

DRONE_NAMESPACE=$(jq -r .name export.json)
for REPO_NAME in $(jq -r .project[].name export.json); do
  drone repo disable $DRONE_NAMESPACE/$REPO_NAME
done
```

Import a drone namespace:

```term
harness-migrate drone import \
  --harness-account $HARNESS_ACCOUNT \
  --harness-org example \
  --docker-connector account.harnessImage \
  --github-token $GITHUB_TOKEN \
  export.json
```

### BitBucket

Convert a bitbucket pipeline:

```term
harness-migrate bitbucket convert /path/to/bitbucket-pipelines.yml
```

### CircleCI

Convert a circle pipeline:

```term
harness-migrate circle convert /path/to/.circle/config.yml
```

### GitHub Actions

Convert a github pipeline:

```term
harness-migrate github convert /path/to/.github/workflows/main.yml
```

### GitLab

Convert a gitlab pipeline:

```term
harness-migrate gitlab convert /path/to/.gitlab-ci.yml
```

### Terraform

Generate terraform configuration from an export, and apply it to your Harness account:

```term
$ harness-migrate terraform \
  --account $HARNESS_ACCOUNT \
  --docker-connector account.harnessImage \
  --org $HARNESS_ORG \
  --repo-connector $HARNESS_REPO_CONNECTOR \
  export.json \
  output.tf
$ terraform init
$ terraform apply
