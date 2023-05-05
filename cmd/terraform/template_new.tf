locals {
  projects = {
{{- range .Org.Projects }}
{{ $yaml := fromYaml .Yaml }}
    "{{ printf "%s" .Name }}" = {
      b64_yaml_properties = "{{ base64Encode (toYaml $yaml.pipeline.properties) }}"
      b64_yaml_stages = "{{ base64Encode (toYaml $yaml.pipeline.stages) }}"
      branch = "{{ printf "%s" .Branch }}"
      repo = "{{ printf "%s" .Repo }}"
      slug = "{{ slugify .Name }}"
{{- if .Secrets }}
      secrets = {
{{- range .Secrets }}
        "{{ printf "%s" .Name }}" = {
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
    "{{ printf "%s" .Name }}" = {
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

resource "harness_platform_secret_text" "organization_secrets" {
  for_each = local.secrets

  identifier  = each.value.slug
  name        = each.key
  org_id      = module.organization.details.id
  value_type  = "Inline"
  value       = each.value.value

  secret_manager_identifier = "harnessSecretManager"
}

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
properties:
  ${base64decode(local.projects[each.key].b64_yaml_properties)}
stages:
  ${base64decode(local.projects[each.key].b64_yaml_stages)}
EOT
}

{{ range .Org.Projects }}
resource "harness_platform_secret_text" "project_{{ slugify .Name }}_secrets" {
   for_each = local.projects["{{ .Name }}"].secrets

   identifier = each.value.slug
   name       = each.key
   org_id     = module.organization.details.id
   project_id = module.projects["{{ .Name }}"].details.id
   value      = each.value.value

   secret_manager_identifier = "harnessSecretManager"
   value_type                = "Inline"
}
{{- end }}
