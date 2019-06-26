package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2019-02-01/containerservice"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/suppress"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmOpenShiftCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmOpenShiftClusterCreateUpdate,
		Read:   resourceArmOpenShiftClusterRead,
		Update: resourceArmOpenShiftClusterCreateUpdate,
		Delete: resourceArmOpenShiftClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"openshift_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"auth_profile": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"providers": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"providers": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validate.NoEmptyStrings,
												},

												"provider": {
													Type:         schema.TypeMap,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validate.NoEmptyStrings,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"kind": {
																Type:     schema.TypeString,
																Required: true,
																ForceNew: true,
																ValidateFunc: validation.StringInSlice([]string{
																	string(containerservice.KindAADIdentityProvider),
																	string(containerservice.KindOpenShiftManagedClusterBaseIdentityProvider),
																}, true),
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},

			"master_pool_profile": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validate.OpenShiftMasterPoolName,
						},

						"count": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(1, 100),
						},

						"vm_size": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     validate.NoEmptyStrings,
						},

						"os_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  string(containerservice.Linux),
							ValidateFunc: validation.StringInSlice([]string{
								string(containerservice.Linux),
								string(containerservice.Windows),
							}, true),
							DiffSuppressFunc: suppress.CaseDifference,
						},

						"subnet_cidr": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validate.NoEmptyStrings,
						},
					},
				},
			},

			"agent_pool_profile": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validate.OpenShiftAgentPoolName,
						},

						"count": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(1, 100),
						},

						"vm_size": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     validate.NoEmptyStrings,
						},

						"os_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  string(containerservice.Linux),
							ValidateFunc: validation.StringInSlice([]string{
								string(containerservice.Linux),
								string(containerservice.Windows),
							}, true),
							DiffSuppressFunc: suppress.CaseDifference,
						},

						"subnet_cidr": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validate.NoEmptyStrings,
						},

						"role": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(containerservice.Compute),
								string(containerservice.Infra),
							}, true),
						},
					},
				},
			},

			"network_profile": {
				Type:     schema.TypeList,
				Required: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vnet_cidr": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validate.NoEmptyStrings,
						},

						"peer_vnet_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validate.NoEmptyStrings,
						},
					},
				},
			},

			"router_profile": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validate.NoEmptyStrings,
						},
					},
				},
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceArmOpenShiftClusterCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).containers.OpenShiftClustersClient
	ctx := meta.(*ArmClient).StopContext
	tenantId := meta.(*ArmClient).tenantId

	log.Printf("[INFO] preparing arguments for Azure Red Hat OpenShift Cluster create/update.")

	resourceGroupName := d.Get("resource_group_name").(string)
	name := d.Get("name").(string)

	if requireResourcesToBeImported && d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroupName, name)

		if err != nil && !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("Error checking for presence of existing Azure Red Hat OpenShift Cluster %q (Resource Group %q): %s", name, resourceGroupName, err)
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_openshift_cluster", *existing.ID)
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	openshiftVersion := d.Get("openshift_version").(string)
	authProfile := expandOpenShiftClusterAuthProfile(d, tenantId)
	masterProfile := expandOpenShiftClusterMasterPoolProfile(d)
	agentProfiles := expandOpenShiftClusterAgentPoolProfiles(d)
	networkProfile := expandOpenShiftClusterNetworkProfile(d)
	routerProfiles := expandOpenShiftClusterRouterProfiles(d)
	tags := d.Get("tags").(map[string]interface{})

	parameters := containerservice.OpenShiftManagedCluster{
		//Plan: ,
		Name:     &name,
		Location: &location,
		OpenShiftManagedClusterProperties: &containerservice.OpenShiftManagedClusterProperties{
			OpenShiftVersion:  utils.String(openshiftVersion),
			AuthProfile:       authProfile,
			MasterPoolProfile: masterProfile,
			AgentPoolProfiles: &agentProfiles,
			NetworkProfile:    networkProfile,
			RouterProfiles:    &routerProfiles,
		},
		Tags: expandTags(tags),
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroupName, name, parameters)

	if err != nil {
		return fmt.Errorf("Error creating/updating Azure Red Hat OpenShift Cluster %q (Resource Group %q): %+v", name, resourceGroupName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for completion of Azure Red Hat OpenShift Cluster %q (Resource Group %q): %+v", name, resourceGroupName, err)
	}

	read, err := client.Get(ctx, resourceGroupName, name)

	if err != nil {
		return fmt.Errorf("Error retrieving Azure Red Hat OpenShift Cluster %q (Resource Group %q): %+v", name, resourceGroupName, err)
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read ID for Azure Red Hat OpenShift Cluster %q (Resource Group %q)", name, resourceGroupName)
	}

	d.SetId(*read.ID)

	return resourceArmOpenShiftClusterRead(d, meta)
}

func resourceArmOpenShiftClusterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).containers.OpenShiftClustersClient
	ctx := meta.(*ArmClient).StopContext
	id, err := parseAzureResourceID(d.Id())

	if err != nil {
		return err
	}

	resGroup := id.ResourceGroup
	name := id.Path["managedClusters"]
	resp, err := client.Get(ctx, resGroup, name)

	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Azure Red Hat OpenShift Cluster %q was not found in Resource Group %q - removing from state!", name, resGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving Azure Red Hat OpenShift Cluster %q (Resource Group %q): %+v", name, resGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resGroup)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.OpenShiftManagedClusterProperties; props != nil {
		d.Set("openshift_version", props.OpenShiftVersion)

		authProfile := flattenOpenShiftClusterAuthProfile(props.AuthProfile)

		if err := d.Set("auth_profile", authProfile); err != nil {
			return fmt.Errorf("Error setting `auth_profile`: %+v", err)
		}

		masterPoolProfile := flattenOpenShiftClusterMasterPoolProfile(props.MasterPoolProfile)

		if err := d.Set("master_pool_profile", masterPoolProfile); err != nil {
			return fmt.Errorf("Error setting `master_pool_profile`: %+v", err)
		}

		agentPoolProfiles := flattenOpenShiftClusterAgentPoolProfiles(props.AgentPoolProfiles)

		if err := d.Set("agent_pool_profile", agentPoolProfiles); err != nil {
			return fmt.Errorf("Error setting `agent_pool_profile`: %+v", err)
		}

		networkProfile := flattenOpenShiftClusterNetworkProfile(props.NetworkProfile)

		if err := d.Set("network_profile", networkProfile); err != nil {
			return fmt.Errorf("Error setting `network_profile`: %+v", err)
		}

		routerProfiles := flattenOpenShiftClusterRouterProfiles(props.RouterProfiles)

		if err := d.Set("router_profile", routerProfiles); err != nil {
			return fmt.Errorf("Error setting `router_profile`: %+v", err)
		}
	}

	flattenAndSetTags(d, resp.Tags)

	return nil
}

func resourceArmOpenShiftClusterDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).containers.OpenShiftClustersClient
	ctx := meta.(*ArmClient).StopContext
	id, err := parseAzureResourceID(d.Id())

	if err != nil {
		return err
	}

	resGroup := id.ResourceGroup
	name := id.Path["managedClusters"]
	future, err := client.Delete(ctx, resGroup, name)

	if err != nil {
		return fmt.Errorf("Error deleting Azure Red Hat OpenShift Cluster %q (Resource Group %q): %+v", name, resGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for the deletion of Azure Red Hat OpenShift Cluster %q (Resource Group %q): %+v", name, resGroup, err)
	}

	return nil
}

func expandOpenShiftClusterMasterPoolProfile(d *schema.ResourceData) *containerservice.OpenShiftManagedClusterMasterPoolProfile {
	profiles := d.Get("master_pool_profile").([]interface{})

	if len(profiles) == 0 {
		return nil
	}

	config := profiles[0].(map[string]interface{})

	name := config["name"].(string)
	count := int32(config["count"].(int))
	vmSize := config["vm_size"].(string)
	osType := config["os_type"].(string)
	subnetCidr := config["subnet_cidr"].(string)

	profile := containerservice.OpenShiftManagedClusterMasterPoolProfile{
		Name:       utils.String(name),
		Count:      utils.Int32(count),
		VMSize:     containerservice.OpenShiftContainerServiceVMSize(vmSize),
		SubnetCidr: utils.String(subnetCidr),
		OsType:     containerservice.OSType(osType),
	}

	return &profile
}

func expandOpenShiftClusterAuthProfile(d *schema.ResourceData, tenantId string) *containerservice.OpenShiftManagedClusterAuthProfile {
	configs := d.Get("auth_profile").([]interface{})

	if len(configs) == 0 {
		return nil
	}

	config := configs[0].(map[string]interface{})
	providersConfig := config["providers"].([]interface{})
	providers := make([]containerservice.OpenShiftManagedClusterIdentityProvider, 0)

	for provider_id := range providersConfig {
		providerConfig := providersConfig[provider_id].(map[string]interface{})
		provider := providerConfig["provider"].(map[string]interface{})

		name := providerConfig["name"].(string)
		identityProvider := expandOpenShiftClusterIdentityProvider(provider, tenantId)

		profile := containerservice.OpenShiftManagedClusterIdentityProvider{
			Name:     utils.String(name),
			Provider: identityProvider,
		}

		providers = append(providers, profile)
	}

	profile := containerservice.OpenShiftManagedClusterAuthProfile{
		IdentityProviders: &providers,
	}

	return &profile
}

func expandOpenShiftClusterIdentityProvider(provider map[string]interface{}, tenantId string) containerservice.BasicOpenShiftManagedClusterBaseIdentityProvider {
	kind := provider["kind"].(string)

	switch kind {
	case string(containerservice.KindAADIdentityProvider):
		clientId := provider["client_id"].(string)
		clientSecret := provider["client_secret"].(string)
		groupId := provider["group_id"].(string)

		return containerservice.OpenShiftManagedClusterAADIdentityProvider{
			ClientID:             utils.String(clientId),
			Secret:               utils.String(clientSecret),
			TenantID:             utils.String(tenantId),
			CustomerAdminGroupID: utils.String(groupId),
			Kind:                 containerservice.KindAADIdentityProvider,
		}
	default:
		return containerservice.OpenShiftManagedClusterBaseIdentityProvider{
			Kind: containerservice.KindOpenShiftManagedClusterBaseIdentityProvider,
		}
	}
}

func expandOpenShiftClusterAgentPoolProfiles(d *schema.ResourceData) []containerservice.OpenShiftManagedClusterAgentPoolProfile {
	configs := d.Get("agent_pool_profile").([]interface{})
	profiles := make([]containerservice.OpenShiftManagedClusterAgentPoolProfile, 0)

	for config_id := range configs {
		config := configs[config_id].(map[string]interface{})
		name := config["name"].(string)
		count := int32(config["count"].(int))
		vmSize := config["vm_size"].(string)
		osType := config["os_type"].(string)
		subnetCidr := config["subnet_cidr"].(string)
		role := config["role"].(string)

		profile := containerservice.OpenShiftManagedClusterAgentPoolProfile{
			Name:       utils.String(name),
			Count:      utils.Int32(count),
			VMSize:     containerservice.OpenShiftContainerServiceVMSize(vmSize),
			SubnetCidr: utils.String(subnetCidr),
			OsType:     containerservice.OSType(osType),
			Role:       containerservice.OpenShiftAgentPoolProfileRole(role),
		}

		profiles = append(profiles, profile)
	}

	return profiles
}

func expandOpenShiftClusterNetworkProfile(d *schema.ResourceData) *containerservice.NetworkProfile {
	profiles := d.Get("network_profile").([]interface{})

	if len(profiles) == 0 {
		return nil
	}

	config := profiles[0].(map[string]interface{})

	vnetCidr := config["vnet_cidr"].(string)
	peerVnetId := config["peer_vnet_id"].(string)

	profile := containerservice.NetworkProfile{
		VnetCidr:   utils.String(vnetCidr),
		PeerVnetID: utils.String(peerVnetId),
	}

	return &profile
}

func expandOpenShiftClusterRouterProfiles(d *schema.ResourceData) []containerservice.OpenShiftRouterProfile {
	configs := d.Get("router_profile").([]interface{})
	profiles := make([]containerservice.OpenShiftRouterProfile, 0)

	for config_id := range configs {
		config := configs[config_id].(map[string]interface{})
		name := config["name"].(string)
		subdomain := config["public_subdomain"].(string)

		profile := containerservice.OpenShiftRouterProfile{
			Name:            utils.String(name),
			PublicSubdomain: utils.String(subdomain),
		}

		profiles = append(profiles, profile)
	}

	return profiles
}

func flattenOpenShiftClusterAuthProfile(profile *containerservice.OpenShiftManagedClusterAuthProfile) interface{} {
	if profile == nil {
		return nil
	}

	authProfile := make(map[string]interface{})

	if profile.IdentityProviders == nil {
		return authProfile
	}

	providers := make([]interface{}, 0)

	for _, provider := range *profile.IdentityProviders {
		providerConfig := expandOpenShiftClusterIdentityProviderConfig(provider)
		providers = append(providers, providerConfig)
	}

	authProfile["providers"] = providers

	return authProfile
}

func expandOpenShiftClusterIdentityProviderConfig(identityProvider containerservice.OpenShiftManagedClusterIdentityProvider) map[string]interface{} {
	baseProvider, _ := identityProvider.Provider.AsOpenShiftManagedClusterBaseIdentityProvider()

	provider := make(map[string]interface{})
	provider["kind"] = string(baseProvider.Kind)

	if baseProvider.Kind == containerservice.KindAADIdentityProvider {
		aadProvider, _ := identityProvider.Provider.AsOpenShiftManagedClusterAADIdentityProvider()

		provider["client_id"] = string(*aadProvider.ClientID)
		provider["client_secret"] = string(*aadProvider.Secret)
		provider["group_id"] = string(*aadProvider.CustomerAdminGroupID)
	}

	providerConfig := make(map[string]interface{})
	providerConfig["name"] = *identityProvider.Name
	providerConfig["provider"] = provider

	return providerConfig
}

func flattenOpenShiftClusterMasterPoolProfile(profile *containerservice.OpenShiftManagedClusterMasterPoolProfile) interface{} {
	if profile == nil {
		return nil
	}

	masterPoolProfile := make(map[string]interface{})

	if profile.Name != nil {
		masterPoolProfile["name"] = *profile.Name
	}

	if profile.Count != nil {
		masterPoolProfile["count"] = int(*profile.Count)
	}

	if profile.VMSize != "" {
		masterPoolProfile["vm_size"] = string(profile.VMSize)
	}

	if profile.OsType != "" {
		masterPoolProfile["os_type"] = string(profile.OsType)
	}

	if profile.SubnetCidr != nil {
		masterPoolProfile["subnet_cidr"] = string(*profile.SubnetCidr)
	}

	return masterPoolProfile
}

func flattenOpenShiftClusterAgentPoolProfiles(profiles *[]containerservice.OpenShiftManagedClusterAgentPoolProfile) []interface{} {
	if profiles == nil {
		return []interface{}{}
	}

	agentPoolProfiles := make([]interface{}, 0)

	for _, profile := range *profiles {
		agentPoolProfile := make(map[string]interface{})

		if profile.Name != nil {
			agentPoolProfile["name"] = *profile.Name
		}

		if profile.Count != nil {
			agentPoolProfile["count"] = int(*profile.Count)
		}

		if profile.VMSize != "" {
			agentPoolProfile["vm_size"] = string(profile.VMSize)
		}

		if profile.OsType != "" {
			agentPoolProfile["os_type"] = string(profile.OsType)
		}

		if profile.SubnetCidr != nil {
			agentPoolProfile["subnet_cidr"] = string(*profile.SubnetCidr)
		}

		if profile.Role != "" {
			agentPoolProfile["role"] = string(profile.Role)
		}

		agentPoolProfiles = append(agentPoolProfiles, agentPoolProfile)
	}

	return agentPoolProfiles
}

func flattenOpenShiftClusterNetworkProfile(profile *containerservice.NetworkProfile) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	values := make(map[string]interface{})

	if profile.VnetCidr != nil {
		values["vnet_cidr"] = *profile.VnetCidr
	}

	if profile.PeerVnetID != nil {
		values["peer_vnet_id"] = *profile.PeerVnetID
	}

	return []interface{}{values}
}

func flattenOpenShiftClusterRouterProfiles(profiles *[]containerservice.OpenShiftRouterProfile) []interface{} {
	if profiles == nil {
		return []interface{}{}
	}

	routerProfiles := make([]interface{}, 0)

	for _, profile := range *profiles {
		routerProfile := make(map[string]interface{})

		if profile.Name != nil {
			routerProfile["name"] = *profile.Name
		}

		if profile.PublicSubdomain != nil {
			routerProfile["public_subdomain"] = *profile.PublicSubdomain
		}

		routerProfiles = append(routerProfiles, routerProfile)
	}

	return routerProfiles
}
