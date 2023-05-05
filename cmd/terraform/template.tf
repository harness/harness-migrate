locals {
  projects = {
{{- range .Org.Projects }}
{{- $yaml := fromYaml .Yaml }}
    {{ printf "%s" .Name }} = {
      yaml_properties = <<-EOT
        properties:
{{ indent (toYaml $yaml.pipeline.properties) 10 -}}
      EOT
      yaml_stages = <<-EOT
        stages:
{{- /* Escape variable references which conflict with terraform */}}
{{ indent (replace (toYaml $yaml.pipeline.stages) "${" "$${") 10 -}}
      EOT
      branch = "{{ printf "%s" .Branch }}"
      repo = "{{ printf "%s" .Repo }}"
      slug = "{{ slugify .Name }}"
{{- if .Secrets }}
      secrets = {
{{- range .Secrets }}
        {{ printf "%s" .Name }} = {
          slug = "{{ slugify .Name }}"
          value = "{{ printf "%s" .Value }}"
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
    {{ printf "%s" .Name }} = {
      slug = "{{ slugify .Name }}"
      value = "{{ printf "%s" .Value }}"
    }
{{- end }}
  }
{{- end }}
}

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

module "projects" {
  for_each = local.projects

  source  = "harness-community/structure/harness//modules/projects"
  version = "~> 0.1"

  name            = each.key
  organization_id = module.organization.details.id
}

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
