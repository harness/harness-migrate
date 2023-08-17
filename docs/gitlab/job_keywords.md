Level of conversion support for GitLab [job keywords](https://docs.gitlab.com/ee/ci/yaml/#job-keywords) to Harness CI YAML.

| | Support level |
|-|-----------|
| 🟢 | Full |
| 🟡 | Partial |
| 🔴 | Unsupported |

## 🔴 [`after_script`](https://docs.gitlab.com/ee/ci/yaml/#after_script)

## 🔴 [`allow_failure`](https://docs.gitlab.com/ee/ci/yaml/#allow_failure)

### 🔴 [`allow_failure:exit_codes`](https://docs.gitlab.com/ee/ci/yaml/#allow_failureexit_codes)

## 🔴 [`artifacts`](https://docs.gitlab.com/ee/ci/yaml/#artifacts)

### 🔴 [`artifacts:paths`](https://docs.gitlab.com/ee/ci/yaml/#artifactspaths)

### 🔴 [`artifacts:exclude`](https://docs.gitlab.com/ee/ci/yaml/#artifactsexclude)

### 🔴 [`artifacts:expire_in`](https://docs.gitlab.com/ee/ci/yaml/#artifactsexpire_in)

### 🔴 [`artifacts:expose_as`](https://docs.gitlab.com/ee/ci/yaml/#artifactsexpose_as)

### 🔴 [`artifacts:name`](https://docs.gitlab.com/ee/ci/yaml/#artifactsname)

### 🔴 [`artifacts:public`](https://docs.gitlab.com/ee/ci/yaml/#artifactspublic)

### 🔴 [`artifacts:reports`](https://docs.gitlab.com/ee/ci/yaml/#artifactsreports)

### 🔴 [`artifacts:untracked`](https://docs.gitlab.com/ee/ci/yaml/#artifactsuntracked)

### 🔴 [`artifacts:when`](https://docs.gitlab.com/ee/ci/yaml/#artifactswhen)

## 🟢 [`before_script`](https://docs.gitlab.com/ee/ci/yaml/#before_script)

<details>
  <summary>Example</summary>

Source
```yaml
job:
  before_script:
    - echo "Execute this command before any 'script:' commands."
  script:
    - echo "This command executes after the job's 'before_script' commands."
```

Converted
```yaml
pipeline:
  identifier: default
  name: default
  orgIdentifier: default
  projectIdentifier: default
  properties:
    ci:
      codebase:
        build: <+input>
  stages:
  - stage:
      identifier: test
      name: test
      spec:
        cloneCodebase: true
        execution:
          steps:
          - step:
              identifier: job
              name: job
              spec:
                command: |-
                  echo "Execute this command before any 'script:' commands."
                  echo "This command executes after the job's 'before_script' commands."
              timeout: ""
              type: Run
        platform:
          arch: Amd64
          os: Linux
        runtime:
          spec: {}
          type: Cloud
      type: CI
```

</details>

## 🟡 [`cache`](https://docs.gitlab.com/ee/ci/yaml/#cache)

### 🟢 [`cache:paths`](https://docs.gitlab.com/ee/ci/yaml/#cachepaths)

<details>
  <summary>Example</summary>

Source
```yaml
rspec:
  script:
    - echo "This job uses a cache."
  cache:
    key: binaries-cache
    paths:
      - binaries/*.apk
      - .config
```

Converted
```yaml
pipeline:
  identifier: default
  name: default
  orgIdentifier: default
  projectIdentifier: default
  properties:
    ci:
      codebase:
        build: <+input>
  stages:
  - stage:
      identifier: test
      name: test
      spec:
        cache:
          enabled: true
          key: binaries-cache
          paths:
          - binaries/*.apk
          - .config
        cloneCodebase: true
        execution:
          steps:
          - step:
              identifier: rspec
              name: rspec
              spec:
                command: echo "This job uses a cache."
              timeout: ""
              type: Run
        platform:
          arch: Amd64
          os: Linux
        runtime:
          spec: {}
          type: Cloud
      type: CI
```

</details>

### 🟡 [`cache:key`](https://docs.gitlab.com/ee/ci/yaml/#cachekey)

Notes:
- Multiple cache keys are not supported

<details>
  <summary>Example</summary>

Source
```yaml
rspec:
  script:
    - echo "This job uses a cache."
  cache:
    key: binaries-cache
    paths:
      - binaries/*.apk
      - .config
```

Converted
```yaml
pipeline:
  identifier: default
  name: default
  orgIdentifier: default
  projectIdentifier: default
  properties:
    ci:
      codebase:
        build: <+input>
  stages:
  - stage:
      identifier: test
      name: test
      spec:
        cache:
          enabled: true
          key: binaries-cache
          paths:
          - binaries/*.apk
          - .config
        cloneCodebase: true
        execution:
          steps:
          - step:
              identifier: rspec
              name: rspec
              spec:
                command: echo "This job uses a cache."
              timeout: ""
              type: Run
        platform:
          arch: Amd64
          os: Linux
        runtime:
          spec: {}
          type: Cloud
      type: CI
```

</details>

#### 🔴 [`cache:key:files`](https://docs.gitlab.com/ee/ci/yaml/#cachekeyfiles)

#### 🔴 [`cache:key:prefix`](https://docs.gitlab.com/ee/ci/yaml/#cachekeyprefix)

### 🔴 [`cache:untracked`](https://docs.gitlab.com/ee/ci/yaml/#cacheuntracked)

### 🔴 [`cache:unprotect`](https://docs.gitlab.com/ee/ci/yaml/#cacheunprotect)

### 🔴 [`cache:when`](https://docs.gitlab.com/ee/ci/yaml/#cachewhen)

### 🔴 [`cache:policy`](https://docs.gitlab.com/ee/ci/yaml/#cachepolicy)

### 🔴 [`cache:fallback_keys`](https://docs.gitlab.com/ee/ci/yaml/#cachefallback_keys)

## 🔴 [`coverage`](https://docs.gitlab.com/ee/ci/yaml/#coverage)

## 🔴 [`dast_configuration`](https://docs.gitlab.com/ee/ci/yaml/#dast_configuration)

## 🔴 [`dependencies`](https://docs.gitlab.com/ee/ci/yaml/#dependencies)

## 🔴 [`environment`](https://docs.gitlab.com/ee/ci/yaml/#environment)

### 🔴 [`environment:name`](https://docs.gitlab.com/ee/ci/yaml/#environmentname)

### 🔴 [`environment:url`](https://docs.gitlab.com/ee/ci/yaml/#environmenturl)

### 🔴 [`environment:on_stop`](https://docs.gitlab.com/ee/ci/yaml/#environmenton_stop)

### 🔴 [`environment:action`](https://docs.gitlab.com/ee/ci/yaml/#environmentaction)

### 🔴 [`environment:auto_stop_in`](https://docs.gitlab.com/ee/ci/yaml/#environmentauto_stop_in)

### 🔴 [`environment:kubernetes`](https://docs.gitlab.com/ee/ci/yaml/#environmentkubernetes)

### 🔴 [`environment:deployment_tier`](https://docs.gitlab.com/ee/ci/yaml/#environmentdeployment_tier)

## 🟢 [`extends`](https://docs.gitlab.com/ee/ci/yaml/#extends)

<details>
  <summary>Example</summary>

Source
```yaml
.example:
  script: rake test
  before_script: echo run this first

rspec:
  extends: .example
  script: rake rspec
```

Converted
```yaml
pipeline:
  identifier: default
  name: default
  orgIdentifier: default
  projectIdentifier: default
  properties:
    ci:
      codebase:
        build: <+input>
  stages:
  - stage:
      identifier: test
      name: test
      spec:
        cloneCodebase: true
        execution:
          steps:
          - step:
              identifier: rspec
              name: rspec
              spec:
                command: |-
                  echo run this first
                  rake rspec
              timeout: ""
              type: Run
        platform:
          arch: Amd64
          os: Linux
        runtime:
          spec: {}
          type: Cloud
      type: CI
```

</details>

## 🔴 [`hooks`](https://docs.gitlab.com/ee/ci/yaml/#hooks)

### 🔴 [`hooks:pre_get_sources_script`](https://docs.gitlab.com/ee/ci/yaml/#hookspre_get_sources_script)

## 🔴 [`id_tokens`](https://docs.gitlab.com/ee/ci/yaml/#id_tokens)

## 🟢 [`image`](https://docs.gitlab.com/ee/ci/yaml/#image)

<details>
  <summary>Example</summary>

Source
```yaml
default:
  image: ruby:3.0

rspec:
  script: bundle exec rspec

rspec 2.7:
  image: registry.example.com/my-group/my-project/ruby:2.7
  script: bundle exec rspec
```

Converted
```yaml
pipeline:
  identifier: default
  name: default
  orgIdentifier: default
  projectIdentifier: default
  properties:
    ci:
      codebase:
        build: <+input>
  stages:
  - stage:
      identifier: test
      name: test
      spec:
        cloneCodebase: true
        execution:
          steps:
          - parallel:
            - step:
                identifier: rspec
                name: rspec
                spec:
                  command: bundle exec rspec
                  image: ruby:3.0
                timeout: ""
                type: Run
            - step:
                identifier: rspec27
                name: rspec 27
                spec:
                  command: bundle exec rspec
                  image: registry.example.com/my-group/my-project/ruby:2.7
                timeout: ""
                type: Run
        platform:
          arch: Amd64
          os: Linux
        runtime:
          spec: {}
          type: Cloud
      type: CI
```

</details>

### 🟢 [`image:name`](https://docs.gitlab.com/ee/ci/yaml/#imagename)

<details>
  <summary>Example</summary>

Source
```yaml
rspec 2.7:
  image:
    name: registry.example.com/my-group/my-project/ruby:2.7
  script: bundle exec rspec
```

Converted
```yaml
pipeline:
  identifier: default
  name: default
  orgIdentifier: default
  projectIdentifier: default
  properties:
    ci:
      codebase:
        build: <+input>
  stages:
  - stage:
      identifier: test
      name: test
      spec:
        cloneCodebase: true
        execution:
          steps:
          - step:
              identifier: rspec27
              name: rspec 27
              spec:
                command: bundle exec rspec
                image: registry.example.com/my-group/my-project/ruby:2.7
              timeout: ""
              type: Run
        platform:
          arch: Amd64
          os: Linux
        runtime:
          spec: {}
          type: Cloud
      type: CI
```

</details>

### 🔴 [`image:entrypoint`](https://docs.gitlab.com/ee/ci/yaml/#imageentrypoint)

### 🟡 [`image:pull_policy`](https://docs.gitlab.com/ee/ci/yaml/#imagepull_policy)

Notes:
- Must be `always`, `if-not-present` or `never`, a list is not supported

<details>
  <summary>Example</summary>

Source
```yaml
job1:
  script: echo "A single pull policy."
  image:
    name: ruby:3.0
    pull_policy: if-not-present

job2:
  script: echo "Multiple pull policies."
  image:
    name: ruby:3.0
    pull_policy: [always, if-not-present]
```

Converted
```yaml
pipeline:
  identifier: default
  name: default
  orgIdentifier: default
  projectIdentifier: default
  properties:
    ci:
      codebase:
        build: <+input>
  stages:
  - stage:
      identifier: test
      name: test
      spec:
        cloneCodebase: true
        execution:
          steps:
          - parallel:
            - step:
                identifier: job1
                name: job1
                spec:
                  command: echo "A single pull policy."
                  image: ruby:3.0
                  imagePullPolicy: IfNotPresent
                timeout: ""
                type: Run
            - step:
                identifier: job2
                name: job2
                spec:
                  command: echo "Multiple pull policies."
                  image: ruby:3.0
                timeout: ""
                type: Run
        platform:
          arch: Amd64
          os: Linux
        runtime:
          spec: {}
          type: Cloud
      type: CI
```

</details>

## 🔴 [`inherit`](https://docs.gitlab.com/ee/ci/yaml/#inherit)

### 🔴 [`inherit:default`](https://docs.gitlab.com/ee/ci/yaml/#inheritdefault)

### 🔴 [`inherit:variables`](https://docs.gitlab.com/ee/ci/yaml/#inheritvariables)

## 🔴 [`interruptible`](https://docs.gitlab.com/ee/ci/yaml/#interruptible)

## 🔴 [`needs`](https://docs.gitlab.com/ee/ci/yaml/#needs)

### 🔴 [`needs:artifacts`](https://docs.gitlab.com/ee/ci/yaml/#needsartifacts)

### 🔴 [`needs:project`](https://docs.gitlab.com/ee/ci/yaml/#needsproject)

#### 🔴 [`needs:pipeline:job`](https://docs.gitlab.com/ee/ci/yaml/#needspipelinejob)

### 🔴 [`needs:optional`](https://docs.gitlab.com/ee/ci/yaml/#needsoptional)

### 🔴 [`needs:pipeline`](https://docs.gitlab.com/ee/ci/yaml/#needspipeline)

#### 🔴 [`needs:parallel:matrix`](https://docs.gitlab.com/ee/ci/yaml/#needsparallelmatrix)

## 🔴 [`only / except`](https://docs.gitlab.com/ee/ci/yaml/#only--except)

### 🔴 [`only:refs / except:refs`](https://docs.gitlab.com/ee/ci/yaml/#onlyrefs--exceptrefs)

### 🔴 [`only:variables / except:variables`](https://docs.gitlab.com/ee/ci/yaml/#onlyvariables--exceptvariables)

### 🔴 [`only:changes / except:changes`](https://docs.gitlab.com/ee/ci/yaml/#onlychanges--exceptchanges)

### 🔴 [`only:kubernetes / except:kubernetes`](https://docs.gitlab.com/ee/ci/yaml/#onlykubernetes--exceptkubernetes)

## 🔴 [`pages`](https://docs.gitlab.com/ee/ci/yaml/#pages)

### 🔴 [`pages:publish`](https://docs.gitlab.com/ee/ci/yaml/#pagespublish)

## 🔴 [`parallel`](https://docs.gitlab.com/ee/ci/yaml/#parallel)

### 🔴 [`parallel:matrix`](https://docs.gitlab.com/ee/ci/yaml/#parallelmatrix)

## 🔴 [`release`](https://docs.gitlab.com/ee/ci/yaml/#release)

### 🔴 [`release:tag_name`](https://docs.gitlab.com/ee/ci/yaml/#releasetag_name)

### 🔴 [`release:tag_message`](https://docs.gitlab.com/ee/ci/yaml/#releasetag_message)

### 🔴 [`release:name`](https://docs.gitlab.com/ee/ci/yaml/#releasename)

### 🔴 [`release:description`](https://docs.gitlab.com/ee/ci/yaml/#releasedescription)

### 🔴 [`release:ref`](https://docs.gitlab.com/ee/ci/yaml/#releaseref)

### 🔴 [`release:milestones`](https://docs.gitlab.com/ee/ci/yaml/#releasemilestones)

### 🔴 [`release:released_at`](https://docs.gitlab.com/ee/ci/yaml/#releasereleased_at)

### 🔴 [`release:assets:links`](https://docs.gitlab.com/ee/ci/yaml/#releaseassetslinks)

## 🔴 [`resource_group`](https://docs.gitlab.com/ee/ci/yaml/#resource_group)

## 🔴 [`retry`](https://docs.gitlab.com/ee/ci/yaml/#retry)

### 🔴 [`retry:when`](https://docs.gitlab.com/ee/ci/yaml/#retrywhen)

## 🔴 [`rules`](https://docs.gitlab.com/ee/ci/yaml/#rules)

### 🔴 [`rules:if`](https://docs.gitlab.com/ee/ci/yaml/#rulesif)

### 🔴 [`rules:changes`](https://docs.gitlab.com/ee/ci/yaml/#ruleschanges)

#### 🔴 [`rules:changes:paths`](https://docs.gitlab.com/ee/ci/yaml/#ruleschangespaths)

#### 🔴 [`rules:changes:compare_to`](https://docs.gitlab.com/ee/ci/yaml/#ruleschangescompare_to)

### 🔴 [`rules:exists`](https://docs.gitlab.com/ee/ci/yaml/#rulesexists)

### 🔴 [`rules:allow_failure`](https://docs.gitlab.com/ee/ci/yaml/#rulesallow_failure)

### 🔴 [`rules:needs`](https://docs.gitlab.com/ee/ci/yaml/#rulesneeds)

### 🔴 [`rules:variables`](https://docs.gitlab.com/ee/ci/yaml/#rulesvariables)

## 🟢 [`script`](https://docs.gitlab.com/ee/ci/yaml/#script)

<details>
  <summary>Example</summary>

Source
```yaml
job1:
  script: "bundle exec rspec"

job2:
  script:
    - uname -a
    - bundle exec rspec
```

Converted
```yaml
pipeline:
  identifier: default
  name: default
  orgIdentifier: default
  projectIdentifier: default
  properties:
    ci:
      codebase:
        build: <+input>
  stages:
  - stage:
      identifier: test
      name: test
      spec:
        cloneCodebase: true
        execution:
          steps:
          - parallel:
            - step:
                identifier: job1
                name: job1
                spec:
                  command: bundle exec rspec
                timeout: ""
                type: Run
            - step:
                identifier: job2
                name: job2
                spec:
                  command: |-
                    uname -a
                    bundle exec rspec
                timeout: ""
                type: Run
        platform:
          arch: Amd64
          os: Linux
        runtime:
          spec: {}
          type: Cloud
      type: CI
```

</details>

## 🟡 [`secrets`](https://docs.gitlab.com/ee/ci/yaml/#secrets)

Notes:
- Secrets are converted to environment variable placeholders, secrets must still be created in the Harness project

<details>
  <summary>Example</summary>

Source
```yaml
job:
  script: echo "reading secrets"
  secrets:
    FIRST_SECRET:
      vault:
        engine:
          name: kv-v2
          path: ops
        path: production/db
        field: password
    SECOND_SECRET:
      vault: production/db/password
    VAULT_SECRET:
      vault: gitlab/production/db
      token: $VAULT_TOKEN
```

Converted
```yaml
pipeline:
  identifier: default
  name: default
  orgIdentifier: default
  projectIdentifier: default
  properties:
    ci:
      codebase:
        build: <+input>
  stages:
  - stage:
      identifier: test
      name: test
      spec:
        cloneCodebase: true
        execution:
          steps:
          - step:
              identifier: job
              name: job
              spec:
                command: echo "reading secrets"
                envVariables:
                  FIRST_SECRET: <+secrets.getValue("FIRST_SECRET")>
                  SECOND_SECRET: <+secrets.getValue("SECOND_SECRET")>
                  VAULT_SECRET: <+secrets.getValue("VAULT_SECRET")>
              timeout: ""
              type: Run
        platform:
          arch: Amd64
          os: Linux
        runtime:
          spec: {}
          type: Cloud
      type: CI
```

</details>

### 🔴 [`secrets:vault`](https://docs.gitlab.com/ee/ci/yaml/#secretsvault)

### 🔴 [`secrets:azure_key_vault`](https://docs.gitlab.com/ee/ci/yaml/#secretsazure_key_vault)

### 🔴 [`secrets:file`](https://docs.gitlab.com/ee/ci/yaml/#secretsfile)

### 🔴 [`secrets:token`](https://docs.gitlab.com/ee/ci/yaml/#secretstoken)

## 🔴 [`services`](https://docs.gitlab.com/ee/ci/yaml/#services)

### 🔴 [`service:pull_policy`](https://docs.gitlab.com/ee/ci/yaml/#servicepull_policy)

## 🟢 [`stage`](https://docs.gitlab.com/ee/ci/yaml/#stage)

<details>
  <summary>Example</summary>

Source
```yaml
stages:
  - build
  - test

job1:
  stage: build
  script:
    - echo "This job compiles code."

job2:
  stage: test
  script:
    - echo "This job tests the compiled code. It runs when the build stage completes."

job3:
  script:
    - echo "This job also runs in the test stage".
```

Converted
```yaml
pipeline:
  identifier: default
  name: default
  orgIdentifier: default
  projectIdentifier: default
  properties:
    ci:
      codebase:
        build: <+input>
  stages:
  - stage:
      identifier: test
      name: test
      spec:
        cloneCodebase: true
        execution:
          steps:
          - step:
              identifier: job1
              name: job1
              spec:
                command: echo "This job compiles code."
              timeout: ""
              type: Run
          - parallel:
            - step:
                identifier: job2
                name: job2
                spec:
                  command: echo "This job tests the compiled code. It runs when the
                    build stage completes."
                timeout: ""
                type: Run
            - step:
                identifier: job3
                name: job3
                spec:
                  command: echo "This job also runs in the test stage".
                timeout: ""
                type: Run
        platform:
          arch: Amd64
          os: Linux
        runtime:
          spec: {}
          type: Cloud
      type: CI
```

</details>

### 🔴 [`stage: .pre`](https://docs.gitlab.com/ee/ci/yaml/#stage-pre)

### 🔴 [`stage: .post`](https://docs.gitlab.com/ee/ci/yaml/#stage-post)

## 🔴 [`tags`](https://docs.gitlab.com/ee/ci/yaml/#tags)

## 🔴 [`timeout`](https://docs.gitlab.com/ee/ci/yaml/#timeout)

## 🔴 [`trigger`](https://docs.gitlab.com/ee/ci/yaml/#trigger)

### 🔴 [`trigger:include`](https://docs.gitlab.com/ee/ci/yaml/#triggerinclude)

### 🔴 [`trigger:project`](https://docs.gitlab.com/ee/ci/yaml/#triggerproject)

### 🔴 [`trigger:strategy`](https://docs.gitlab.com/ee/ci/yaml/#triggerstrategy)

### 🔴 [`trigger:forward`](https://docs.gitlab.com/ee/ci/yaml/#triggerforward)

## 🟢 [`variables`](https://docs.gitlab.com/ee/ci/yaml/#variables)

<details>
  <summary>Example</summary>

Source
```yaml
variables:
  DEPLOY_SITE: "https://example.com"

deploy_job:
  stage: deploy
  script:
    - deploy-script --url $DEPLOY_SITE --path "/"
  environment: production

deploy_review_job:
  stage: deploy
  variables:
    REVIEW_PATH: "/review"
  script:
    - deploy-review-script --url $DEPLOY_SITE --path $REVIEW_PATH
  environment: production
```

Converted
```yaml
pipeline:
  identifier: default
  name: default
  orgIdentifier: default
  projectIdentifier: default
  properties:
    ci:
      codebase:
        build: <+input>
  stages:
  - stage:
      identifier: test
      name: test
      spec:
        cloneCodebase: true
        execution:
          steps:
          - parallel:
            - step:
                identifier: deployjob
                name: deploy_job
                spec:
                  command: deploy-script --url $DEPLOY_SITE --path "/"
                timeout: ""
                type: Run
            - step:
                identifier: deployreviewjob
                name: deploy_review_job
                spec:
                  command: deploy-review-script --url $DEPLOY_SITE --path $REVIEW_PATH
                timeout: ""
                type: Run
        platform:
          arch: Amd64
          os: Linux
        runtime:
          spec: {}
          type: Cloud
      type: CI
      variables:
      - name: DEPLOY_SITE
        type: String
        value: https://example.com
```

</details>

### 🔴 [`variables:description`](https://docs.gitlab.com/ee/ci/yaml/#variablesdescription)

### 🟢 [`variables:value`](https://docs.gitlab.com/ee/ci/yaml/#variablesvalue)

<details>
  <summary>Example</summary>

Source
```yaml
variables:
  DEPLOY_SITE: "https://example.com"
  DEPLOY_ENVIRONMENT:
    value: "staging"

deploy_job:
  script:
    - deploy-script --url $DEPLOY_SITE/$DEPLOY_ENVIRONMENT --path "/"

deploy_review_job:
  variables:
    REVIEW_PATH: "/review"
  script:
    - deploy-review-script --url $DEPLOY_SITE/$DEPLOY_ENVIRONMENT --path $REVIEW_PATH
```

Converted
```yaml
pipeline:
  identifier: default
  name: default
  orgIdentifier: default
  projectIdentifier: default
  properties:
    ci:
      codebase:
        build: <+input>
  stages:
  - stage:
      identifier: test
      name: test
      spec:
        cloneCodebase: true
        execution:
          steps:
          - parallel:
            - step:
                identifier: deployjob
                name: deploy_job
                spec:
                  command: deploy-script --url $DEPLOY_SITE/$DEPLOY_ENVIRONMENT --path
                    "/"
                timeout: ""
                type: Run
            - step:
                identifier: deployreviewjob
                name: deploy_review_job
                spec:
                  command: deploy-review-script --url $DEPLOY_SITE/$DEPLOY_ENVIRONMENT
                    --path $REVIEW_PATH
                timeout: ""
                type: Run
        platform:
          arch: Amd64
          os: Linux
        runtime:
          spec: {}
          type: Cloud
      type: CI
      variables:
      - name: DEPLOY_ENVIRONMENT
        type: String
        value: staging
      - name: DEPLOY_SITE
        type: String
        value: https://example.com
```

</details>

### 🔴 [`variables:options`](https://docs.gitlab.com/ee/ci/yaml/#variablesoptions)

### 🔴 [`variables:expand`](https://docs.gitlab.com/ee/ci/yaml/#variablesexpand)

## 🔴 [`when`](https://docs.gitlab.com/ee/ci/yaml/#when)