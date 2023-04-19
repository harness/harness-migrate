provider "harness" {
  endpoint         = "{{ .Auth.Endpoint }}"
  account_id       = "{{ .Account.ID }}"
  platform_api_key = "{{ .Auth.Token }}"
}

resource "harness_platform_organization" "this" {
  identifier  = "{{ slugify .Org.Name }}"
  name        = "{{ .Org.Name }}"
}

{{ range .Org.Secrets }}
resource "harness_platform_secret_text" "inline" {
  org_id      = "" # TODO
  identifier  = "{{ slugify .Name }}"
  name        = "{{ .Name }}"
  description = "{{ .Desc }}"
  value_type  = "Inline"
  value       = "{{ .Value }}"
  secret_manager_identifier = "harnessSecretManager"
}
{{ end }}

{{ range .Org.Projects }}
# TODO create project
# TODO loop through and create project secrets
{{ end }}
