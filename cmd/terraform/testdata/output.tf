// Locals
locals {
  projects = {
    hello-world = {
      yaml_properties = <<-EOT
        properties:
          null
          EOT
      yaml_stages = <<-EOT
        stages:
          null
          EOT
      branch = "main"
      namespace = "octocat"
      repo = "hello-world"
      secrets = {
        PROJECT_SECRET = {
          value = "topsecret"
        }
      }
    }
  }
  secrets = {
    ORG_SECRET = {
      value = "confidential"
    }
  }
}

terraform {
  required_providers {
    harness = {
      source  = ""
      version = "= "
    }
  }
}

provider "harness" {
  endpoint = ""
}

// Organization
module "organization" {
  source  = "harness-community/structure/harness//modules/organizations"
  version = "~> 0.1"

  name = ""
}

// Organization secrets
module "organization_secrets" {
  for_each = local.secrets

  source  = "harness-community/structure/harness//modules/secrets/text"
  version = "~> 0.1"

  name            = each.key
  organization_id = module.organization.details.id
  value           = each.value.value
}

// Projects
module "projects" {
  for_each = local.projects

  source  = "harness-community/structure/harness//modules/projects"
  version = "~> 0.1"

  name            = each.key
  organization_id = module.organization.details.id
}

// Pipelines
module "pipelines" {
  for_each = local.projects
  
  source  = "harness-community/content/harness//modules/pipelines"
  version = "~> 0.1"

  name            = each.key
  organization_id = module.organization.details.id
  project_id      = module.projects[each.key].details.id
  yaml_data       = <<-EOT
${each.value.yaml_properties}
${each.value.yaml_stages}
EOT
}

// Project secrets
module "project_helloworld_secrets" {
  for_each = local.projects["hello-world"].secrets

  source  = "harness-community/structure/harness//modules/secrets/text"
  version = "~> 0.1"

  name            = each.key
  organization_id = module.organization.details.id
  project_id      = module.projects["hello-world"].details.id
  value           = each.value.value
}
// Pull request trigger
module "trigger_pr" {
  for_each = local.projects
  
  source  = "harness-community/content/harness//modules/triggers"
  version = "~> 0.1"

  name = "Pull Request"
  organization_id = module.organization.details.id
  project_id      = module.projects[each.key].details.id
  pipeline_id     = module.pipelines[each.key].details.id
  yaml_data       = <<-EOT
source:
  type: Webhook
  spec:
    # TODO: support other SCM types
    type: Github
    spec:
      type: PullRequest
      spec:
        connectorRef: 
        autoAbortPreviousExecutions: false
        payloadConditions:
          - key: targetBranch
            operator: Equals
            value: ${each.value.branch}
        headerConditions: []
        repoName: ${each.value.namespace}/${each.value.repo}
        actions:
          - Open
          - Reopen
          - Synchronize
inputYaml: |
  pipeline:
    identifier: ${module.pipelines[each.key].details.id}
    properties:
      ci:
        codebase:
          build:
            type: PR
            spec:
              number: <+trigger.prNumber>
  EOT
}

// Push trigger
module "trigger_push" {
  for_each = local.projects
  
  source  = "harness-community/content/harness//modules/triggers"
  version = "~> 0.1"

  name = "Push"
  organization_id = module.organization.details.id
  project_id      = module.projects[each.key].details.id
  pipeline_id     = module.pipelines[each.key].details.id
  yaml_data       = <<-EOT
source:
  type: Webhook
  spec:
    # TODO: support other SCM types
    type: Github
    spec:
      type: Push
      spec:
        connectorRef: 
        autoAbortPreviousExecutions: false
        payloadConditions:
          - key: targetBranch
            operator: Equals
            value: ${each.value.branch}
        headerConditions: []
        repoName: ${each.value.namespace}/${each.value.repo}
        actions: []
inputYaml: |
  pipeline:
    identifier: ${module.pipelines[each.key].details.id}
    properties:
      ci:
        codebase:
          build:
            type: branch
            spec:
              branch: <+trigger.branch>
  EOT  
}

// Tag trigger
module "trigger_tag" {
  for_each = local.projects
  
  source  = "harness-community/content/harness//modules/triggers"
  version = "~> 0.1"

  name = "Tag"
  organization_id = module.organization.details.id
  project_id      = module.projects[each.key].details.id
  pipeline_id     = module.pipelines[each.key].details.id
  yaml_data       = <<-EOT
source:
  type: Webhook
  spec:
    # TODO: support other SCM types
    type: Github
    spec:
      type: Push
      spec:
        connectorRef: 
        autoAbortPreviousExecutions: false
        payloadConditions:
          - key: <+trigger.payload.ref>
            operator: StartsWith
            value: refs/tags/
        headerConditions: []
        repoName: ${each.value.namespace}/${each.value.repo}
        actions: []
inputYaml: |
  pipeline:
    identifier: ${module.pipelines[each.key].details.id}
    properties:
      ci:
        codebase:
          build:
            type: branch
            spec:
              branch: <+trigger.branch>
  EOT
}
