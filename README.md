
# Compile

```
go build
```

# Usage

Export Circle data

```sh
./harness-migrate cirlce export \
  --org=${CIRCLE_ORG_UUID} \
  --token=${CIRCLE_TOKEN} \
  --out=/tmp/circle.json
```

Import Circle data into Harness

```sh
./harness-migrate cirlce import /tmp/circle.json
```

Convert a Circle yaml to Harness yaml

```sh
./harness-migrate circle convert /path/to/.circle/config.yml
```