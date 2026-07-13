package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccGithubOrganizationSecurityConfigurationRepositories(t *testing.T) {
	t.Parallel()
	skipUnlessAcceptanceTest(t)
	skipUnlessHasOrgs(t)

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	firstRepositoryName := fmt.Sprintf("%ssecurity-config-first-%s", testResourcePrefix, randomID)
	secondRepositoryName := fmt.Sprintf("%ssecurity-config-second-%s", testResourcePrefix, randomID)
	configurationName := fmt.Sprintf("%ssecurity-config-repositories-%s", testResourcePrefix, randomID)

	config := fmt.Sprintf(`
resource "github_repository" "first" {
  name       = %[1]q
  visibility = "public"
  auto_init  = true
}

resource "github_repository" "second" {
  name       = %[2]q
  visibility = "public"
  auto_init  = true
}

resource "github_organization_security_configuration" "test" {
  name               = %[3]q
  secret_scanning    = "enabled"
  advanced_security  = "enabled"
}

resource "github_organization_security_configuration_repositories" "test" {
  configuration_id = github_organization_security_configuration.test.configuration_id
  repository_ids = %%v
}
`, firstRepositoryName, secondRepositoryName, configurationName)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(config, "[github_repository.first.repo_id]"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("github_organization_security_configuration_repositories.test", tfjsonpath.New("repository_ids"), knownvalue.SetSizeExact(1)),
				},
			},
			{
				Config: fmt.Sprintf(config, "[github_repository.first.repo_id, github_repository.second.repo_id]"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("github_organization_security_configuration_repositories.test", tfjsonpath.New("repository_ids"), knownvalue.SetSizeExact(2)),
				},
			},
			{
				ResourceName:      "github_organization_security_configuration_repositories.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestOrganizationSecurityConfigurationAttachmentIsActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status string
		want   bool
	}{
		{status: "attached", want: true},
		{status: "attaching", want: true},
		{status: "enforced", want: true},
		{status: "updating", want: true},
		{status: "detached", want: false},
		{status: "failed", want: false},
		{status: "removed", want: false},
		{status: "removed_by_enterprise", want: false},
	}

	for _, test := range tests {
		t.Run(test.status, func(t *testing.T) {
			t.Parallel()
			got := organizationSecurityConfigurationAttachmentIsActive(test.status)
			if got != test.want {
				t.Errorf("organizationSecurityConfigurationAttachmentIsActive(%q) = %t; want %t", test.status, got, test.want)
			}
		})
	}
}
