
# Compile

```
go build
```

# Usage

Export data from circle:

```sh
./harness-migrate cirlce export \
  --org=${CIRCLE_ORG} \
  --token=${CIRCLE_TOKEN} \
  --out=/tmp/circle.json
```

Import data from circle:

```sh
./harness-migrate cirlce import /tmp/circle.json \
  --harness-account=${HARNESS_ACCOUNT} \
  --harness-org=${HARNESS_ORG} \
  --harness-token=${HARNESS_TOKEN} \
  --github-token=${GITHUB_TOKEN}
```

Convert a circle pipeline:

```sh
./harness-migrate circle convert /path/to/.circle/config.yml
```