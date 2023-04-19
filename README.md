**harness-migrate** is a command line utility to help convert and migrate
continuous integration pipelines from other providers to Harness CI.

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

Convert a circle pipeline:

```term
$ harness-migrate circle convert /path/to/.circle/config.yml
```

Convert a github pipeline:

```term
$ harness-migrate github convert /path/to/.github/workflows/main.yml
```

Convert a drone pipeline:

```term
$ harness-migrate drone convert /path/to/.drone.yml
```
