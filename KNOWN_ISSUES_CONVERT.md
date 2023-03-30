This document attempts to describe known limitations and problems for yaml
conversion.

In this document, `v0 yaml` refers to current Harness CI yaml. `harness-migrate`
supports converting to `v0 yaml` with the `--downgrade` flag, for example:

```
harness-migrate github convert --downgrade example.yml
```

## GitHub Actions

### env

[env](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#env)
is a `map` of variables that are available to the steps of all jobs in the
workflow.

#### v0 yaml

**Problem**

GitHub Actions allows variables to contain hyphen characters, for example:
```yaml
env:
  my-environment: production
```

Which converts to this `v0 yaml` which contains hyphens in variable names, which
is invalid:
```yaml
  variables:
  - name: my-environment
    type: String
    value: production
```

**Fix**

Replace `-` characters in the variable with `_`, or remove them, then update 
all references to the variable in the pipeline.
