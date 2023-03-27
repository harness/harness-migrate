__Install__

```term
$ git clone https://github.com/harness/harness-migrate.git
$ go install
```

__Usage__

Export data from circle:

```term
$ harness-migrate cirlce export \
  --org=${CIRCLE_ORG} \
  --token=${CIRCLE_TOKEN} \
  --out=/tmp/circle.json
```

Import data from circle:

```term
$ harness-migrate cirlce import /tmp/circle.json \
  --harness-account=${HARNESS_ACCOUNT} \
  --harness-org=${HARNESS_ORG} \
  --harness-token=${HARNESS_TOKEN} \
  --github-token=${GITHUB_TOKEN}
```

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