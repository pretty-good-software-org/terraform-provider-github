resource "github_organization_security_configuration_repositories" "example" {
  configuration_id = github_organization_security_configuration.example.configuration_id
  repository_ids = [
    github_repository.first.repo_id,
    github_repository.second.repo_id,
  ]
}
