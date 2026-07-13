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

func TestAccGithubOrganizationSecurityConfigurationDefault(t *testing.T) {
	t.Parallel()
	skipUnlessHasOrgs(t)

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	configurationName := fmt.Sprintf("%ssecurity-config-default-%s", testResourcePrefix, randomID)
	config := fmt.Sprintf(`
resource "github_organization_security_configuration" "test" {
  name              = %[1]q
  secret_scanning   = "enabled"
  advanced_security = "enabled"
}

resource "github_organization_security_configuration_default" "test" {
  configuration_id             = github_organization_security_configuration.test.configuration_id
  default_for_new_repositories = %%q
}
`, configurationName)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(config, "private_and_internal"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("github_organization_security_configuration_default.test", tfjsonpath.New("default_for_new_repositories"), knownvalue.StringExact("private_and_internal")),
				},
			},
			{
				Config: fmt.Sprintf(config, "all"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("github_organization_security_configuration_default.test", tfjsonpath.New("default_for_new_repositories"), knownvalue.StringExact("all")),
				},
			},
			{
				ResourceName:      "github_organization_security_configuration_default.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
