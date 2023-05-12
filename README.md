**harness-migrate** is a command line utility to help convert and migrate
continuous integration pipelines from other providers to Harness CI.

**Please review [known conversion and migration issues](KNOWN_ISSUES_CONVERT.md
) before using this tool.**

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

Import a drone namespace:

```term
harness-migrate drone import \
  --harness-account $HARNESS_ACCOUNT \
  --harness-org example \
  --docker-connector account.harnessImage \
  --github-token $GITHUB_TOKEN \
  export.json
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
```