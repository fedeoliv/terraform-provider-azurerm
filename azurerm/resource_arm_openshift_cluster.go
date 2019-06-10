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
		// CustomizeDiff: ,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"dns_prefix": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.KubernetesDNSPrefix,
			},

			"kubernetes_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.NoEmptyStrings,
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
							ValidateFunc: validate.KubernetesAgentPoolName,
						},

						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  string(containerservice.AvailabilitySet),
							ValidateFunc: validation.StringInSlice([]string{
								string(containerservice.AvailabilitySet),
								string(containerservice.VirtualMachineScaleSets),
							}, false),
						},

						"count": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(1, 100),
						},

						// TODO: remove this field in the next major version
						"dns_prefix": {
							Type:       schema.TypeString,
							Computed:   true,
							Deprecated: "This field has been removed by Azure",
						},

						"fqdn": {
							Type:       schema.TypeString,
							Computed:   true,
							Deprecated: "This field has been deprecated. Use the parent `fqdn` instead",
						},

						"vm_size": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     validate.NoEmptyStrings,
						},

						"os_disk_size_gb": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Computed:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},

						"vnet_subnet_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: azure.ValidateResourceID,
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

						"max_pods": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},

			// Optional
			"addon_profile": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_application_routing": {
							Type:     schema.TypeList,
							MaxItems: 1,
							ForceNew: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										ForceNew: true,
										Required: true,
									},
									"http_application_routing_zone_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"oms_agent": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"log_analytics_workspace_id": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: azure.ValidateResourceID,
									},
								},
							},
						},

						"aci_connector_linux": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"subnet_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validate.NoEmptyStrings,
									},
								},
							},
						},
					},
				},
			},

			"linux_profile": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"admin_username": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validate.KubernetesAdminUserName,
						},
						"ssh_key": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,

							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key_data": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validate.NoEmptyStrings,
									},
								},
							},
						},
					},
				},
			},

			"network_profile": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_plugin": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(containerservice.Azure),
								string(containerservice.Kubenet),
							}, false),
						},

						"network_policy": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(containerservice.NetworkPolicyCalico),
								string(containerservice.NetworkPolicyAzure),
							}, false),
						},

						"dns_service_ip": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validate.IPv4Address,
						},

						"docker_bridge_cidr": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validate.CIDR,
						},

						"pod_cidr": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validate.CIDR,
						},

						"service_cidr": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validate.CIDR,
						},
					},
				},
			},

			"role_based_access_control": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
							ForceNew: true,
						},
						"azure_active_directory": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_app_id": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validate.UUID,
									},

									"server_app_id": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validate.UUID,
									},

									"server_app_secret": {
										Type:         schema.TypeString,
										ForceNew:     true,
										Required:     true,
										Sensitive:    true,
										ValidateFunc: validate.NoEmptyStrings,
									},

									"tenant_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
										// OrEmpty since this can be sourced from the client config if it's not specified
										ValidateFunc: validate.UUIDOrEmpty,
									},
								},
							},
						},
					},
				},
			},

			"tags": tagsSchema(),

			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// Computed
			"kube_admin_config": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"client_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_key": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"cluster_ca_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"kube_admin_config_raw": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"kube_config": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"client_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_key": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"cluster_ca_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"kube_config_raw": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"node_resource_group": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"api_server_authorized_ip_ranges": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validate.CIDR,
				},
			},
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
	fqdn := d.Get("fqdn").(string)
	authProfile := expandOpenShiftClusterAuthProfile(d, tenantId)
	masterProfile := expandOpenShiftClusterMasterPoolProfile(d)
	agentProfiles := expandOpenShiftClusterAgentPoolProfiles(d)
	networkProfile := expandOpenShiftClusterNetworkProfile(d)
	tags := d.Get("tags").(map[string]interface{})

	parameters := containerservice.OpenShiftManagedCluster{
		//Plan: ,
		Name:     &name,
		Location: &location,
		OpenShiftManagedClusterProperties: &containerservice.OpenShiftManagedClusterProperties{
			OpenShiftVersion: utils.String(openshiftVersion),
			// PublicHostname: ,
			Fqdn: &fqdn,
			// RouterProfiles:    routerProfiles,
			AuthProfile:       authProfile,
			MasterPoolProfile: masterProfile,
			AgentPoolProfiles: &agentProfiles,
			NetworkProfile:    networkProfile,
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
		// d.Set("dns_prefix", props.DNSPrefix)
		d.Set("fqdn", props.Fqdn)
		d.Set("openshift_version", props.OpenShiftVersion)
		// d.Set("node_resource_group", props.NodeResourceGroup)

		agentPoolProfiles := flattenOpenShiftClusterAgentPoolProfiles(props.AgentPoolProfiles, resp.Fqdn)

		if err := d.Set("agent_pool_profile", agentPoolProfiles); err != nil {
			return fmt.Errorf("Error setting `agent_pool_profile`: %+v", err)
		}

		networkProfile := flattenOpenShiftClusterNetworkProfile(props.NetworkProfile)

		if err := d.Set("network_profile", networkProfile); err != nil {
			return fmt.Errorf("Error setting `network_profile`: %+v", err)
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

	profile := containerservice.OpenShiftManagedClusterMasterPoolProfile{
		Name:   utils.String(name),
		Count:  utils.Int32(count),
		VMSize: containerservice.OpenShiftContainerServiceVMSize(vmSize),
		// SubnetCidr:
		OsType: containerservice.OSType(osType),
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
		role := config["role"].(string)

		profile := containerservice.OpenShiftManagedClusterAgentPoolProfile{
			Name:   utils.String(name),
			Count:  utils.Int32(count),
			VMSize: containerservice.OpenShiftContainerServiceVMSize(vmSize),
			// SubnetCidr:
			OsType: containerservice.OSType(osType),
			Role:   containerservice.OpenShiftAgentPoolProfileRole(role),
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

	profile := containerservice.NetworkProfile{
		VnetCidr: utils.String(vnetCidr),
		// PeerVnetID: ,
	}

	return &profile
}

func flattenOpenShiftClusterAgentPoolProfiles(profiles *[]containerservice.OpenShiftManagedClusterAgentPoolProfile, fqdn *string) []interface{} {
	if profiles == nil {
		return []interface{}{}
	}

	agentPoolProfiles := make([]interface{}, 0)

	for _, profile := range *profiles {
		agentPoolProfile := make(map[string]interface{})

		if profile.Count != nil {
			agentPoolProfile["count"] = int(*profile.Count)
		}

		if fqdn != nil {
			// temporarily persist the parent FQDN here until `fqdn` is removed from the `agent_pool_profile`
			agentPoolProfile["fqdn"] = *fqdn
		}

		if profile.Name != nil {
			agentPoolProfile["name"] = *profile.Name
		}

		if profile.VMSize != "" {
			agentPoolProfile["vm_size"] = string(profile.VMSize)
		}

		if profile.OsType != "" {
			agentPoolProfile["os_type"] = string(profile.OsType)
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
