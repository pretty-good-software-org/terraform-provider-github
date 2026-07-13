resource "github_organization_security_configuration_default" "example" {
  configuration_id             = github_organization_security_configuration.example.configuration_id
  default_for_new_repositories = "public"
}
