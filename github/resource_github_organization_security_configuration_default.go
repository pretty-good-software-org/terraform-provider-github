package github

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/go-github/v88/github"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubOrganizationSecurityConfigurationDefault() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGithubOrganizationSecurityConfigurationDefaultCreateOrUpdate,
		ReadContext:   resourceGithubOrganizationSecurityConfigurationDefaultRead,
		UpdateContext: resourceGithubOrganizationSecurityConfigurationDefaultCreateOrUpdate,
		DeleteContext: resourceGithubOrganizationSecurityConfigurationDefaultDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceGithubOrganizationSecurityConfigurationDefaultImport,
		},

		Description: "Resource to manage a GitHub code security configuration as an organization's default for new repositories.",

		Schema: map[string]*schema.Schema{
			"configuration_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Numeric ID of the code security configuration.",
			},
			"default_for_new_repositories": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Repository visibility for which the configuration is the default. Can be one of 'all', 'public', or 'private_and_internal'.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
					"all", "public", "private_and_internal",
				}, false)),
			},
		},
	}
}

func resourceGithubOrganizationSecurityConfigurationDefaultCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := checkOrganization(m); err != nil {
		return diag.FromErr(err)
	}

	configurationID, err := organizationSecurityConfigurationID(d)
	if err != nil {
		return diag.FromErr(err)
	}
	defaultForNewRepositories, err := organizationSecurityConfigurationString(d, "default_for_new_repositories")
	if err != nil {
		return diag.FromErr(err)
	}
	meta, err := organizationSecurityConfigurationOwner(m)
	if err != nil {
		return diag.FromErr(err)
	}
	_, _, err = meta.v3client.Organizations.SetDefaultCodeSecurityConfiguration(ctx, meta.name, configurationID, defaultForNewRepositories)
	if err != nil {
		return diag.Errorf("setting code security configuration %d as the default for new repositories: %s", configurationID, err)
	}

	d.SetId(strconv.FormatInt(configurationID, 10))
	return nil
}

func resourceGithubOrganizationSecurityConfigurationDefaultRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := checkOrganization(m); err != nil {
		return diag.FromErr(err)
	}

	configurationID, err := organizationSecurityConfigurationID(d)
	if err != nil {
		return diag.FromErr(err)
	}
	meta, err := organizationSecurityConfigurationOwner(m)
	if err != nil {
		return diag.FromErr(err)
	}
	defaults, response, err := meta.v3client.Organizations.ListDefaultCodeSecurityConfigurations(ctx, meta.name)
	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			tflog.Info(ctx, "Removing organization code security configuration default from state because the configuration no longer exists", map[string]any{"configuration_id": configurationID})
			d.SetId("")
			return nil
		}
		return diag.Errorf("listing default code security configurations: %s", err)
	}

	for _, defaultConfiguration := range defaults {
		if defaultConfiguration.GetConfiguration().GetID() != configurationID {
			continue
		}
		if err := d.Set("default_for_new_repositories", defaultConfiguration.GetDefaultForNewRepos()); err != nil {
			return diag.Errorf("setting default repository visibility for code security configuration %d: %s", configurationID, err)
		}
		return nil
	}

	tflog.Info(ctx, "Removing organization code security configuration default from state because it is no longer a default", map[string]any{"configuration_id": configurationID})
	d.SetId("")
	return nil
}

func resourceGithubOrganizationSecurityConfigurationDefaultDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := checkOrganization(m); err != nil {
		return diag.FromErr(err)
	}

	configurationID, err := organizationSecurityConfigurationID(d)
	if err != nil {
		return diag.FromErr(err)
	}
	meta, err := organizationSecurityConfigurationOwner(m)
	if err != nil {
		return diag.FromErr(err)
	}
	_, response, err := meta.v3client.Organizations.SetDefaultCodeSecurityConfiguration(ctx, meta.name, configurationID, "none")
	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			return nil
		}
		if githubError, ok := errors.AsType[*github.ErrorResponse](err); ok && githubError.Response.StatusCode == http.StatusNotFound {
			return nil
		}
		return diag.Errorf("removing code security configuration %d as the default for new repositories: %s", configurationID, err)
	}

	return nil
}

func resourceGithubOrganizationSecurityConfigurationDefaultImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	configurationID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return nil, errors.New("organization code security configuration default import ID must be a numeric configuration ID")
	}
	if err := d.Set("configuration_id", configurationID); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
