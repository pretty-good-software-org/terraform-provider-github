# Terraform Provider GitHub

|<img src="https://cloud.githubusercontent.com/assets/98681/24211275/c4ebd04e-0ee8-11e7-8606-061d656a42df.png" width="72" height="" alt="GitHub Logo"> | <img src="https://raw.githubusercontent.com/hashicorp/terraform-website/d841a1e5fca574416b5ca24306f85a0f4f41b36d/content/source/assets/images/logo-terraform-main.svg" width="300px" alt="Terraform Logo">|

This provider manages GitHub resources — repositories, teams, branch protections, actions secrets/variables, organization settings, rulesets, deploy keys, webhooks, and more — using Terraform. It supports both GitHub.com and GitHub Enterprise Server via the REST and GraphQL APIs.

See the [GitHub Provider page on the Terraform Registry](https://registry.terraform.io/providers/integrations/github/) for installation and documentation.

## Quick Start

```hcl
provider "github" {
  owner = "my-org"
}

resource "github_repository" "example" {
  name        = "example-repo"
  description = "Managed by Terraform"
  visibility  = "private"
}
```

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 1.x
- [Go](https://golang.org/doc/install) 1.26.x (to build the provider plugin)

## Usage

Detailed documentation for the GitHub provider can be found in the [Terraform Registry](https://registry.terraform.io/providers/integrations/github).

## Contributing

For instructions on how to contribute to the GitHub Terraform provider, see the [Contributing Guide](CONTRIBUTING.md).

## Roadmap

This project uses [Milestones](https://github.com/integrations/terraform-provider-github/milestones) to scope upcoming features and bug fixes. Issues that receive the most recent discussion or the most reactions will be more likely to be included in an upcoming release.

## Development Toolchain

Development tools are pinned by Mise and locked for six POSIX targets: Linux ARM64, Linux ARM64 musl, Linux x64, Linux x64 musl, macOS ARM64, and macOS x64. Windows tool entries are intentionally absent. The Python lock environment sets `compile = false` and records GitHub attestation provenance for its platform artifacts.

Use `mise run setup:default` to install the locked tools and configure hooks. Safe local checks are available as `mise run lint:default`, `mise run test:default`, and `mise run lint:vulncheck`. Provider documentation remains owned by the existing `tfplugindocs` generation and validation tasks.

## Support

GitHub Support does not provide support for this integration. This is a community-supported project. GitHub's SDK team triages issues and PRs periodically.
