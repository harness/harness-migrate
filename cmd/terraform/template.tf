// Locals
locals {
  projects = {
{{- range .Org.Projects }}
{{- $yaml := fromYaml .Yaml }}
    {{ .Name }} = {
      yaml_properties = <<-EOT
        properties:
{{ indent (toYaml $yaml.pipeline.properties) 10 -}}
      EOT
      yaml_stages = <<-EOT
        stages:
{{- /* Escape variable references which conflict with terraform */}}
{{ indent (replace (toYaml $yaml.pipeline.stages) "${" "$${") 10 -}}
      EOT
      branch = "{{ .Branch }}"
      {{- $repo := split .Repo "/" }}
      namespace = "{{ index $repo 3 }}"
      repo = "{{ trimSuffix (index $repo 4) ".git" }}"
{{- if .Secrets }}
      secrets = {
{{- range .Secrets }}
        {{ .Name }} = {
          value = "{{ .Value }}"
        }
{{- end }}
      }
{{- end }}
    }
{{- end }}
  }
{{- if .Org.Secrets }}
  secrets = {
{{- range .Org.Secrets }}
    {{ .Name }} = {
      value = "{{ .Value }}"
    }
{{- end }}
  }
{{- end }}
}

terraform {
  required_providers {
    harness = {
      source  = "{{ .Provider.Source }}"
      version = "= {{ .Provider.Version }}"
    }
  }
}

provider "harness" {
  endpoint = "{{ .Auth.Endpoint }}"
{{- if .Account.ID }}
  account_id = "{{ .Account.ID }}"
{{- end }}
}

// Organization
module "organization" {
  source  = "harness-community/structure/harness//modules/organizations"
  version = "~> 0.1"

  name = "{{ .Account.Organization }}"
}

// Organization secrets
{{ if .Org.Secrets -}}
module "organization_secrets" {
  for_each = local.secrets

  source  = "harness-community/structure/harness//modules/secrets/text"
  version = "~> 0.1"

  name            = each.key
  organization_id = module.organization.details.id
  value           = each.value.value
}
{{- end }}

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
{{ range .Org.Projects -}}
{{ if .Secrets -}}
module "project_{{ slugify .Name }}_secrets" {
  for_each = local.projects["{{ .Name }}"].secrets

  source  = "harness-community/structure/harness//modules/secrets/text"
  version = "~> 0.1"

  name            = each.key
  organization_id = module.organization.details.id
  project_id      = module.projects["{{ .Name }}"].details.id
  value           = each.value.value
}
{{ end -}}
{{ end -}}

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
        connectorRef: {{ .Connectors.Repo }}
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
        connectorRef: {{ .Connectors.Repo }}
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
        connectorRef: {{ .Connectors.Repo }}
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
