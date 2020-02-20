package containers

import (
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-02-01/containerservice"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func expandOpenShiftAgentPoolProfiles(agentPoolProfiles []interface{}) (*[]containerservice.OpenShiftManagedClusterAgentPoolProfile, error) {
	if len(agentPoolProfiles) == 0 {
		return nil, nil
	}

	var profiles []containerservice.OpenShiftManagedClusterAgentPoolProfile

	for _, input := range agentPoolProfiles {
		config := input.(map[string]interface{})

		name := config["name"].(string)
		count := config["node_count"].(int)
		vmSize := config["vm_size"].(string)
		role := config["role"].(string)

		profile := containerservice.OpenShiftManagedClusterAgentPoolProfile{
			Name:   utils.String(name),
			Count:  count,
			VMSize: containerservice.VMSizeTypes(vmSize),
			Role:   containerservice.OpenShiftAgentPoolProfileRole(role),
		}

		if v, ok := config["os_type"]; ok && v.(string) != "" {
			osType := v.(string)
			profile.OsType = containerservice.OSType(osType)
		}

		if v, ok := config["subnet_cidr"]; ok && v.(string) != "" {
			subnetCidr := v.(string)
			profile.SubnetCidr = utils.String(subnetCidr)
		}

		profiles = append(profiles, profile)
	}

	return &profiles, nil
}

func expandOpenShiftClusterMasterPoolProfile(input []interface{}) (*containerservice.OpenShiftManagedClusterMasterPoolProfile, error) {
	if len(input) == 0 {
		return nil, nil
	}

	config := input[0].(map[string]interface{})

	name := config["name"].(string)
	nodeCount := config["node_count"].(int)
	vmSize := config["vm_size"].(string)

	masterPoolProfile := containerservice.OpenShiftManagedClusterMasterPoolProfile{
		Name:   utils.String(name),
		Count:  nodeCount,
		VMSize: utils.String(vmSize),
	}

	if v, ok := config["subnet_cidr"]; ok && v.(string) != "" {
		subnetCidr := v.(string)
		masterPoolProfile.SubnetCidr = utils.String(subnetCidr)
	}

	if v, ok := config["os_type"]; ok && v.(string) != "" {
		osType := v.(string)
		masterPoolProfile.OsType = utils.String(osType)
	}

	return &masterPoolProfile, nil
}
