terraform {
  required_providers {
    harness = {
      source  = "{{ $.Provider.Source }}"
      version = "= {{ $.Provider.Version }}"
    }
  }
}

provider "harness" {
  endpoint = "{{ $.Auth.Endpoint }}"
{{- if .Account.ID }}
  account_id = "{{ $.Account.ID }}"
{{- end }}
}

module "organization" {
  source  = "harness-community/structure/harness//modules/organizations"
  version = "~> 0.1"

  name = "{{ $.Account.Organization }}"
}

{{- /* Create organization secrets */}}
{{- range .Org.Secrets -}}
resource "harness_platform_secret_text" "organization_{{ slugify .Name }}" {
  identifier  = "{{ slugify .Name }}"
  name        = "{{ .Name }}"
  org_id      = module.organization.details.id
  description = "{{ .Desc }}"
  value_type  = "Inline"
  value       = "{{ .Value }}"

  secret_manager_identifier = "harnessSecretManager"
}
{{- end -}}

{{- /* Create projects */}}
{{- range .Org.Projects -}}
{{- /* Read in pipeline yaml so its values can be referenced */}}
{{ $yaml := fromYaml .Yaml }}
{{ $projectSlug := (slugify .Name) -}}
module "project_{{ $projectSlug }}" {
  source  = "harness-community/structure/harness//modules/projects"
  version = "~> 0.1"

  name            = "{{- printf "%s" $yaml.pipeline.name -}}"
  organization_id = module.organization.details.id
}

{{/* Create project secrets */}}
{{- range .Secrets -}}
resource "harness_platform_secret_text" "project_{{ slugify .Name }}" {
  identifier  = "{{ slugify .Name }}"
  name        = "{{ .Name }}"
  org_id      = module.organization.details.id
  project_id  = module.project_{{ $projectSlug }}.details.id
  description = "{{ .Desc }}"
  value_type  = "Inline"
  value       = "{{ .Value }}"

  secret_manager_identifier = "harnessSecretManager"  
}
{{ end }}

{{- /* Create project pipeline */}}
module "pipeline_{{ slugify .Name }}" {
  source  = "harness-community/content/harness//modules/pipelines"
  version = "~> 0.1"

  name            = "{{- printf "%s" $yaml.pipeline.name -}}"
  organization_id = module.organization.details.id
  project_id      = module.project_{{ slugify .Name }}.details.id
  {{- /* TODO: could values other than 'properties' or 'stages' be needed? */}}
  yaml_data       = <<-EOT
{{ if $yaml.pipeline.properties }}
properties:
{{ indent (toYaml $yaml.pipeline.properties) 2 }}
{{- end }}
{{- if $yaml.pipeline.stages }}
stages:
{{ indent (toYaml $yaml.pipeline.stages) 2 -}}
{{- end }}
EOT
}

{{- /* Create pipeline triggers */}}
module "trigger_push_{{ slugify .Name }}" {
  source  = "harness-community/content/harness//modules/triggers"
  version = "~> 0.1"

  name            = "Push"
  organization_id = module.organization.details.id
  project_id      = module.project_{{ slugify .Name }}.details.id
  pipeline_id     = module.pipeline_{{ slugify .Name }}.details.id
  {{- /* TODO: support more than GitHub */}}
  yaml_data       = <<EOT
source:
  type: "Webhook"
  spec:
    type: "Github"
    spec:
      type: "Push"
      spec:
        connectorRef: "{{- $.Connectors.Repo -}}"
        autoAbortPreviousExecutions: false
        payloadConditions:
          - key: targetBranch
            operator: Equals
            value: main
        headerConditions: []
        repoName: "{{ $.Account.Organization }}/{{ .Name }}"
        actions: []
inputYaml: |
  pipeline:
    identifier: "module.pipeline_{{ slugify .Name }}.details.id"
    properties:
      ci:
        codebase:
          build:
            type: branch
            spec:
              branch: <+trigger.branch>"
EOT
}
{{ end }}
