package github

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func organizationSecurityConfigurationOwner(metadata any) (*Owner, error) {
	owner, ok := metadata.(*Owner)
	if !ok {
		return nil, errors.New("reading GitHub provider metadata: expected owner metadata")
	}
	return owner, nil
}

func organizationSecurityConfigurationID(data *schema.ResourceData) (int64, error) {
	configurationID, ok := data.Get("configuration_id").(int)
	if !ok {
		return 0, errors.New("reading code security configuration ID: expected an integer")
	}
	return int64(configurationID), nil
}

func organizationSecurityConfigurationString(data *schema.ResourceData, attribute string) (string, error) {
	value, ok := data.Get(attribute).(string)
	if !ok {
		return "", fmt.Errorf("reading %s: expected a string", attribute)
	}
	return value, nil
}

func expandOrganizationSecurityConfigurationRepositoryIDs(value any) ([]int64, error) {
	repositoryIDs, ok := value.(*schema.Set)
	if !ok {
		return nil, errors.New("reading code security configuration repository IDs: expected a set")
	}

	ids := make([]int64, 0, repositoryIDs.Len())
	for _, repositoryIDValue := range repositoryIDs.List() {
		repositoryID, ok := repositoryIDValue.(int)
		if !ok {
			return nil, errors.New("reading code security configuration repository ID: expected an integer")
		}
		ids = append(ids, int64(repositoryID))
	}
	return ids, nil
}
