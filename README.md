`harness-migrate` is a command line utility to help convert and migrate
continuous integration pipelines from other providers to Harness CI.

__Install__

Download the appropriate binary from 
[releases](https://github.com/harness/harness-migrate/releases).

__Build__

```term
$ git clone https://github.com/harness/harness-migrate.git
$ cd harness-migrate
$ go build
```

__Usage__

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
