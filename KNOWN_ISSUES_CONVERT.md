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

## All

Issues that are not specific to a provider.

### Resource limits

Kubernetes-based Harness CI pipeline steps run with default resource limits.

**Problem**

`harness-migrate` does not set resource limits during pipeline conversion.

**Solution**

Manually set the proper [resource limits](https://developer.harness.io/docs/continuous-integration/use-ci/run-ci-scripts/run-step-settings/#set-container-resources
) for your step(s) that run on Kubernetes infrastructure.

### JEXL syntax

Conversion will not always choose the correct [JEXL condition syntax](https://developer.harness.io/docs/platform/pipelines/w_pipeline-steps-reference/triggers-reference/#conditions-settings).

Example Drone pipeline

```yaml
kind: pipeline                                                                  
type: docker                                                                    
name: default                                                                   
                                                                                
steps:                                                                                                                       
  - name: publish                                                
    image: plugins/docker                                                       
    pull: if-not-exists                                                         
    settings:                                                                   
      repo: harness/example                                            
      auto_tag: true                                                            
      auto_tag_suffix: linux-amd64                                              
      dockerfile: docker/Dockerfile.linux.amd64                                 
      username:                                                                 
        from_secret: docker-username                                    
      password:                                                                 
        from_secret: docker-password                                    
    when:                                                                       
      ref:                                                                      
        - refs/heads/master                                                     
        - refs/tags/v*
```

The `when` conditions convert to this JEXL

```yaml
              when:
                condition: <+trigger.payload.ref> == "refs/heads/master" || <+trigger.payload.ref> == "refs/tags/v*"
```

The tag ref syntax needs to be manually changed to this

```yaml
              when:
                condition: <+trigger.payload.ref> == "refs/heads/master" || <+trigger.payload.ref> =~ "refs/tags/v.*"
```

### Expressions

See [Webhook Triggers Reference](https://developer.harness.io/docs/platform/pipelines/w_pipeline-steps-reference/triggers-reference/)
for available trigger expressions.

**Problem**

There might not be an expression for your use case. For example, the git [commit ref](https://git-scm.com/book/en/v2/Git-Internals-Git-References).

**Solution**

Retrieve the desired value from the webhook payload, which is available via `<+trigger.payload.*>`.
For example, for a GitHub repository, the git commit ref is available at `<+trigger.payload.ref>`.

### Triggers

All CI solutions have a way to control what events should trigger a pipeline execution.

**Problem**

Harness CI [webhook triggers](https://developer.harness.io/docs/platform/pipelines/w_pipeline-steps-reference/triggers-reference/) are managed separately from the pipeline yaml.

Harness CI also supports conditions at the [stage and step level](https://developer.harness.io/docs/platform/pipelines/w_pipeline-steps-reference/step-skip-condition-settings/) which are managed in the yaml.

Where supported, this tool converts stage and step-level conditions, but it does not convert webhook trigger conditions.

**Solution**

Webhook triggers can be modified to have logic that matches the stage-level conditions.

For example, here are [triggers](https://docs.drone.io/pipeline/triggers/) from a a Drone pipeline:

```yaml
trigger:
  branch:
    include:
      - develop
      - main
```

The above will convert to this stage-level condition:

```yaml
      when:
        condition: <+trigger.targetBranch> == "develop" || <+trigger.targetBranch> == "main"
```

Push and pull request webhook triggers in Harness CI can be updated to use this payload condition to maintain the same behavior:

```yaml
          payloadConditions:
            - key: targetBranch
              operator: In
              value: develop, main
```

## Cloud Build

### [artifacts](https://cloud.google.com/build/docs/build-config-file-schema#artifacts)

One or more non-container artifacts to be stored in Cloud Storage.

**Problem**

Cloud Build supports publishing artifacts to a GCP bucket in the YAML:

```yaml
steps:
- name: 'gcr.io/cloud-builders/go'
  args: ['build', 'my-package']
artifacts:
  objects:
    location: 'gs://mybucket/'
    paths: ['my-package']
```

**Fix**

Create a step that uses the built-in Upload Artifacts to GCS step.

### [availableSecrets](https://cloud.google.com/build/docs/build-config-file-schema#availablesecrets)

Adds the value of the secret to the environment and you can access this value via environment variable from scripts or processes.

**Problem**

Cloud Build supports reading secrets from Secrets Manager and assigning them to environment variables at the pipeline level.

```yaml
steps:
- name: python:slim
  entrypoint: python
  args: ['main.py']
  secretEnv: ['MYSECRET']
availableSecrets:
  secretManager:
  - versionName: projects/$PROJECT_ID/secrets/mySecret/versions/latest
    env: 'MYSECRET'
```

**Fix**

Create a GCP Secrets Manager connector to read the secrets into environment variables.

## Drone

### 

### [Variable Substitution](https://docs.drone.io/pipeline/environment/substitution/#string-operations)

Drone variables must be converted to the equivalent [Harness variable expression](https://developer.harness.io/docs/platform/variables-and-expressions/extracting-characters-from-harness-variable-expressions/).

**Problem**

The [go-convert](https://github.com/drone/go-convert/) library handles this conversion where possible, but not every Drone variable is supported. See https://github.com/drone/go-convert/issues/117 for what variables are currently supported.

**Solution**

Examine your converted pipeline yaml for any `DRONE_` variables that were not converted. Replace these variables with the equivalent Harness JEXL expression.

**Problem**

In Drone, substituted variables with no value will be empty. In Harness CI expressions, unset values will return a `null` string. **This may break logic that relies on an empty string for unset values**.

**Solution**

Determine if your pipelines rely on unset `DRONE_` variables that return an empty string, modify these references to allow for a string of `null` to be returned.

### [Variable String Operations](https://docs.drone.io/pipeline/environment/substitution/#string-operations)

Drone provides partial emulation for bash string operations, these must convert to an equivalent [Harness variable expression](https://developer.harness.io/docs/platform/variables-and-expressions/extracting-characters-from-harness-variable-expressions/).

**Problem**

The [go-convert](https://github.com/drone/go-convert/) library currently has limited support for these conversions. See https://github.com/drone/go-convert/issues/117 for what operations are currently supported.

**Solution**

Examine your converted pipeline yaml for any `DRONE_` variables with string operations that were not converted. Replace these with the equivalent Harness JEXL expression.

### [depends_on](https://docs.drone.io/pipeline/docker/syntax/parallelism/)

Steps can depend on the successful execution of previous steps.

**Problem**

Harness CI does not have equivalent support for dependencies at the step level.

**Fix**

TBD

### [image_pull_secrets](https://docs.drone.io/yaml/docker/#the-image_pull_secrets-attribute)

Docker credentials in the format of a Docker config.json file. This file may provide credentials for multiple repositories.

**Problem**

Harness CI does not support Docker config.json files in connectors.

**Fix**

Create Docker registry connectors for each private registry, and reference the necessary connector in each step.

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

### [jobs.<job_id>.needs](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idneeds)

Jobs in workflows **run in parallel by default**. To run jobs sequentially, you
can define dependencies on other jobs using the `jobs.<job_id>.needs` keyword.

**Problem**

Since the `jobs.<job_id>.needs` keyword is not currently supported, **stages
will likely not execute in the desired order**.

**Fix**

Review the workflow’s `jobs.<job_id>.needs` rules, and manually move the stages
into the correct order.

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

## GitLab

See [job keywords](docs/gitlab/job_keywords.md) docs
