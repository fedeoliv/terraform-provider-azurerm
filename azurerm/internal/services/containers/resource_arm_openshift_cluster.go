package containers

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-02-01/containerservice"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmOpenShiftCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmOpenShiftClusterCreate,
	}
}

func resourceArmOpenShiftClusterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Containers.OpenShiftClustersClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	tenantID := meta.(*clients.Client).Account.TenantId

	log.Printf("[INFO] preparing arguments for Managed OpenShift Cluster creation")

	resourceGroupName := d.Get("resource_group_name").(string)
	resourceName := d.Get("name").(string)

	if features.ShouldResourcesBeImported() {
		existing, err := client.Get(ctx, resGroup, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing OpenShift Cluster %q (Resource Group %q): %s", name, resGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_openshift_cluster", *existing.ID)
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	openShiftVersion := d.Get("openshift_version").(string)

	masterPoolProfileRaw := d.Get("master_pool_profile ").([]interface{})
	masterPoolProfile, err := expandOpenShiftClusterMasterPoolProfile(masterPoolProfileRaw)
	if err != nil {
		return fmt.Errorf("Error expanding `master_pool_profile`: %+v", err)
	}

	agentPoolProfilesRaw := d.Get("agent_pool_profile").([]interface{})
	agentPoolProfiles, err := expandOpenShiftAgentPoolProfiles(agentPoolProfilesRaw)
	if err != nil {
		return fmt.Errorf("Error expanding `agent_pool_profile`: %+v", err)
	}

	networkProfileRaw := d.Get("network_profile").([]interface{})
	networkProfile, err := expandOpenShiftClusterNetworkProfile(networkProfileRaw)
	if err != nil {
		return fmt.Errorf("Error expanding `network_profile`: %+v", err)
	}

	routerProfilesRaw := d.Get("router_profile").([]interface{})
	routerProfiles, err := expandDefaultRouterProfile(routerProfilesRaw)
	if err != nil {
		return fmt.Errorf("Error expanding `router_profile`: %+v", err)
	}

	authProfileRaw := d.Get("auth_profile").([]interface{})
	authProfile, err := expandAuthProfile(authProfileRaw, tenantID)
	if err != nil {
		return fmt.Errorf("Error expanding `auth_profile`: %+v", err)
	}

	t := d.Get("tags").(map[string]interface{})

	parameters := containerservice.OpenShiftManagedCluster{
		Name:     &resourceName,
		Location: &location,
		OpenShiftManagedClusterProperties: &containerservice.OpenShiftManagedClusterProperties{
			OpenShiftVersion:  utils.String(openShiftVersion),
			NetworkProfile:    networkProfile,
			RouterProfiles:    routerProfiles,
			MasterPoolProfile: masterPoolProfile,
			AgentPoolProfiles: agentPoolProfiles,
			AuthProfile:       authProfile,
		},
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroupName, resourceName, parameters)
	if err != nil {
		return fmt.Errorf("Error creating Managed OpenShift Cluster %q (Resource Group %q): %+v", name, resGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for creation of Managed Kubernetes Cluster %q (Resource Group %q): %+v", name, resGroup, err)
	}

	read, err := client.Get(ctx, resGroup, name)
	if err != nil {
		return fmt.Errorf("Error retrieving Managed Kubernetes Cluster %q (Resource Group %q): %+v", name, resGroup, err)
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read ID for Managed Kubernetes Cluster %q (Resource Group %q)", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmKubernetesClusterRead(d, meta)
}

func expandOpenShiftClusterNetworkProfile(input []interface{}) (*containerservice.NetworkProfile, error) {
	if len(input) == 0 {
		return nil, nil
	}

	config := input[0].(map[string]interface{})
	networkProfile := containerservice.NetworkProfile{}

	if v, ok := config["vnet_cidr"]; ok && v.(string) != "" {
		vnetCidr := v.(string)
		networkProfile.VnetCidr = utils.String(vnetCidr)
	}

	if v, ok := config["peer_vnet_id"]; ok && v.(string) != "" {
		peerVnetID := v.(string)
		networkProfile.PeerVnetID = utils.String(peerVnetID)
	}

	if v, ok := config["vnet_id"]; ok && v.(string) != "" {
		vnetID := v.(string)
		networkProfile.VnetID = utils.String(vnetID)
	}

	return &networkProfile, nil
}

func expandDefaultRouterProfile(input []interface{}) (*[]containerservice.OpenShiftRouterProfile, error) {
	if len(input) == 0 {
		return nil, nil
	}

	raw := input[0].(map[string]interface{})
	name := raw["name"].(string)

	profile := containerservice.OpenShiftRouterProfile{
		Name: utils.String(name),
	}

	return &[]containerservice.OpenShiftRouterProfile{
		profile,
	}, nil
}

func expandAuthProfile(input []interface{}, tenantID string) (*containerservice.OpenShiftManagedClusterAuthProfile, error) {
	if len(input) == 0 {
		return nil, nil
	}

	config := input[0].(map[string]interface{})
	identityProviders := expandAuthProfileIdentityProviders(config, tenantID)

	profile := containerservice.OpenShiftManagedClusterAuthProfile{
		IdentityProviders: identityProviders,
	}

	return &profile, nil
}

func expandAuthProfileIdentityProviders(identityProviderConfig map[string]interface{}, tenantID string) *[]containerservice.OpenShiftManagedClusterIdentityProvider {
	clientID := identityProviderConfig["aad_client_id"].(string)
	clientSecret := identityProviderConfig["aad_client_secret"].(string)
	customerAdminGroupID := identityProviderConfig["aad_customer_admin_group_id"].(string)

	aadIdentityProvider := containerservice.OpenShiftManagedClusterAADIdentityProvider{
		ClientID:             utils.String(clientID),
		Secret:               utils.String(clientSecret),
		TenantID:             utils.String(tenantID),
		CustomerAdminGroupID: utils.String(customerAdminGroupID),
		Kind:                 "KindAADIdentityProvider",
	}

	identityProvider := containerservice.OpenShiftManagedClusterIdentityProvider{
		Name:     utils.String("AAD"),
		Provider: aadIdentityProvider,
	}

	return &[]containerservice.OpenShiftManagedClusterIdentityProvider{
		identityProvider,
	}
}
