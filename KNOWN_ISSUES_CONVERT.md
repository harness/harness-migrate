This document describes known limitations for CI pipeline yaml conversion.

NOTE: In this document, "v0 yaml" refers to current Harness CI yaml, "v1 yaml" 
refers to the new "simplified" Harness CI yaml.

`harness-migrate` supports converting to v0 yaml with the `--downgrade` flag,
for example:

```
harness-migrate github convert --downgrade example.yml
```

To see all supported conversion flags for a provider, pass the `--help` flag:

```
harness-migrate github convert --help
```

## GitHub Actions

### [env](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#env)

`env` is a `map` of variables that are available to the steps of all jobs in the
workflow.

#### v0 yaml

**Problem**

GitHub Actions allows variables to contain hyphen characters, for example:
```yaml
env:
  my-environment: production
```

Example converted v0 yaml:
```yaml
  variables:
  - name: my-environment # this is invalid
    type: String
    value: production
```

**Fix**

Replace `-` characters in the variable with `_`, or remove them, then update 
all references to the variable in the pipeline.

Example valid yaml:
```yaml
  variables:
  - name: myenvironment # this is valid
    type: String
    value: production
```

### [jobs.<job_id>.timeout-minutes](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idtimeout-minutes)

`jobs.<job_id>.timeout-minutes` is the maximum number of minutes to let a job 
run before GitHub automatically cancels it.

NOTE: `timeout-minutes` conversion is supported for [steps](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepstimeout-minutes).

**Problem**

Harness CI does not currently support timeouts at the stage level.

`timeout-minutes` at the job level in a GitHub Action workflow will not convert, for example:

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    timeout-minutes: 10 # this will be lost during conversion
```

**Fix**

Add a timeout for the overall pipeline, for example:

```yaml
pipeline:
  timeout: 10m # this is the best alternative
```
