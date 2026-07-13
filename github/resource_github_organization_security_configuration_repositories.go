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
)

func resourceGithubOrganizationSecurityConfigurationRepositories() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGithubOrganizationSecurityConfigurationRepositoriesCreate,
		ReadContext:   resourceGithubOrganizationSecurityConfigurationRepositoriesRead,
		UpdateContext: resourceGithubOrganizationSecurityConfigurationRepositoriesUpdate,
		DeleteContext: resourceGithubOrganizationSecurityConfigurationRepositoriesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceGithubOrganizationSecurityConfigurationRepositoriesImport,
		},

		Description: "Resource to manage the repositories attached to a GitHub code security configuration for an organization.",

		Schema: map[string]*schema.Schema{
			"configuration_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Numeric ID of the code security configuration.",
			},
			"repository_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Set:         schema.HashInt,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Numeric IDs of the repositories that use the code security configuration.",
			},
		},
	}
}

func resourceGithubOrganizationSecurityConfigurationRepositoriesCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := checkOrganization(m); err != nil {
		return diag.FromErr(err)
	}

	configurationID, err := organizationSecurityConfigurationID(d)
	if err != nil {
		return diag.FromErr(err)
	}
	repositoryIDs, err := expandOrganizationSecurityConfigurationRepositoryIDs(d.Get("repository_ids"))
	if err != nil {
		return diag.FromErr(err)
	}
	meta, err := organizationSecurityConfigurationOwner(m)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = meta.v3client.Organizations.AttachCodeSecurityConfigurationToRepositories(ctx, meta.name, configurationID, "selected", repositoryIDs)
	if err != nil {
		return diag.Errorf("attaching code security configuration %d to repositories: %s", configurationID, err)
	}

	d.SetId(strconv.FormatInt(configurationID, 10))
	return nil
}

func resourceGithubOrganizationSecurityConfigurationRepositoriesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
	repositoryIDs := make([]int64, 0)
	options := &github.ListCodeSecurityConfigurationRepositoriesOptions{
		PerPage: maxPerPage,
	}

	for attachment, err := range meta.v3client.Organizations.ListCodeSecurityConfigurationRepositoriesIter(ctx, meta.name, configurationID, options) {
		if err != nil {
			if response, ok := errors.AsType[*github.ErrorResponse](err); ok && response.Response.StatusCode == http.StatusNotFound {
				tflog.Info(ctx, "Removing organization code security configuration repositories from state because the configuration no longer exists", map[string]any{"configuration_id": configurationID})
				d.SetId("")
				return nil
			}
			return diag.Errorf("listing repositories for code security configuration %d: %s", configurationID, err)
		}

		if organizationSecurityConfigurationAttachmentIsActive(attachment.GetStatus()) {
			repositoryIDs = append(repositoryIDs, attachment.GetRepository().GetID())
		}
	}

	if err := d.Set("repository_ids", repositoryIDs); err != nil {
		return diag.Errorf("setting repository IDs for code security configuration %d: %s", configurationID, err)
	}

	return nil
}

func resourceGithubOrganizationSecurityConfigurationRepositoriesUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := checkOrganization(m); err != nil {
		return diag.FromErr(err)
	}

	configurationID, err := organizationSecurityConfigurationID(d)
	if err != nil {
		return diag.FromErr(err)
	}
	oldValue, newValue := d.GetChange("repository_ids")
	oldRepositoryIDs, ok := oldValue.(*schema.Set)
	if !ok {
		return diag.Errorf("reading previous code security configuration repository IDs: expected a set")
	}
	newRepositoryIDs, ok := newValue.(*schema.Set)
	if !ok {
		return diag.Errorf("reading new code security configuration repository IDs: expected a set")
	}
	removedRepositoryIDs, err := expandOrganizationSecurityConfigurationRepositoryIDs(oldRepositoryIDs.Difference(newRepositoryIDs))
	if err != nil {
		return diag.FromErr(err)
	}
	addedRepositoryIDs, err := expandOrganizationSecurityConfigurationRepositoryIDs(newRepositoryIDs.Difference(oldRepositoryIDs))
	if err != nil {
		return diag.FromErr(err)
	}
	meta, err := organizationSecurityConfigurationOwner(m)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(removedRepositoryIDs) > 0 {
		_, err := meta.v3client.Organizations.DetachCodeSecurityConfigurationsFromRepositories(ctx, meta.name, removedRepositoryIDs)
		if err != nil {
			return diag.Errorf("detaching repositories from code security configuration %d: %s", configurationID, err)
		}
	}

	if len(addedRepositoryIDs) > 0 {
		_, err := meta.v3client.Organizations.AttachCodeSecurityConfigurationToRepositories(ctx, meta.name, configurationID, "selected", addedRepositoryIDs)
		if err != nil {
			return diag.Errorf("attaching code security configuration %d to repositories: %s", configurationID, err)
		}
	}

	return nil
}

func resourceGithubOrganizationSecurityConfigurationRepositoriesDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := checkOrganization(m); err != nil {
		return diag.FromErr(err)
	}

	repositoryIDs, err := expandOrganizationSecurityConfigurationRepositoryIDs(d.Get("repository_ids"))
	if err != nil {
		return diag.FromErr(err)
	}
	if len(repositoryIDs) == 0 {
		return nil
	}
	meta, err := organizationSecurityConfigurationOwner(m)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = meta.v3client.Organizations.DetachCodeSecurityConfigurationsFromRepositories(ctx, meta.name, repositoryIDs)
	if err != nil {
		if response, ok := errors.AsType[*github.ErrorResponse](err); ok && response.Response.StatusCode == http.StatusNotFound {
			return nil
		}
		return diag.Errorf("detaching repositories from code security configuration %s: %s", d.Id(), err)
	}

	return nil
}

func resourceGithubOrganizationSecurityConfigurationRepositoriesImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	configurationID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return nil, errors.New("organization code security configuration repositories import ID must be a numeric configuration ID")
	}
	if err := d.Set("configuration_id", configurationID); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func organizationSecurityConfigurationAttachmentIsActive(status string) bool {
	switch status {
	case "attached", "attaching", "enforced", "updating":
		return true
	default:
		return false
	}
}
