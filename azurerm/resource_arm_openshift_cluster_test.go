package azurerm

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
)

func TestAccAzureRMOpenShiftCluster_basic(t *testing.T) {
	resourceName := "azurerm_openshift_cluster.test"
	ri := tf.AccRandTimeInt()

	openshiftVersion := os.Getenv("OPENSHIFT_VERSION")
	clientId := os.Getenv("ARM_CLIENT_ID")
	clientSecret := os.Getenv("ARM_CLIENT_SECRET")
	groupId := os.Getenv("ARM_GROUP_ID")

	config := testAccAzureRMOpenShiftCluster_basic(ri, openshiftVersion, clientId, clientSecret, groupId, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMOpenShiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMOpenShiftClusterExists(resourceName),
					// resource.TestCheckResourceAttr(resourceName, "role_based_access_control.#", "1"),
					// resource.TestCheckResourceAttr(resourceName, "role_based_access_control.0.enabled", "false"),
					// resource.TestCheckResourceAttr(resourceName, "role_based_access_control.0.azure_active_directory.#", "0"),
					// resource.TestCheckResourceAttrSet(resourceName, "kube_config.0.client_key"),
					// resource.TestCheckResourceAttrSet(resourceName, "kube_config.0.client_certificate"),
					// resource.TestCheckResourceAttrSet(resourceName, "kube_config.0.cluster_ca_certificate"),
					// resource.TestCheckResourceAttrSet(resourceName, "kube_config.0.host"),
					// resource.TestCheckResourceAttrSet(resourceName, "kube_config.0.username"),
					// resource.TestCheckResourceAttrSet(resourceName, "kube_config.0.password"),
					// resource.TestCheckResourceAttr(resourceName, "kube_admin_config.#", "0"),
					// resource.TestCheckResourceAttr(resourceName, "kube_admin_config_raw", ""),
					// resource.TestCheckResourceAttrSet(resourceName, "agent_pool_profile.0.max_pods"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckAzureRMOpenShiftClusterExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]

		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for Managed OpenShift Cluster: %s", name)
		}

		client := testAccProvider.Meta().(*ArmClient).containers.OpenShiftClustersClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		aro, err := client.Get(ctx, resourceGroup, name)

		if err != nil {
			return fmt.Errorf("Bad: Get on OpenShiftClustersClient: %+v", err)
		}

		if aro.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: Managed OpenShift Cluster %q (Resource Group: %q) does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testAccAzureRMOpenShiftCluster_basic(rInt int, openshiftVersion string, clientId string, clientSecret string, groupId string, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_openshift_cluster" "test" {
  name                = "acctestopenshift%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  openshift_version   = "%s"

  auth_profile {
    providers = [
      {
        name = "Azure AD"
        provider = {
          kind            = "AADIdentityProvider"
          client_id       = "%s"
          client_secret   = "%s"
          group_id        = "%s"
        }
      }
    ]
  }

  master_pool_profile {
    name 		= "default"
    count		= 1
    vm_size		= "Standard_D2s_v3"
    os_type		= "Linux"
    subnet_cidr	= "10.0.0.0/24"
  }

  agent_pool_profile {
    name		= "default"
    count		= 1
    vm_size		= "Standard_D2s_v3"
    os_type		= "Linux"
    subnet_cidr	= "10.0.0.0/24"
    role		= "Compute"
  }

  network_profile {
    vnet_cidr         = "10.0.0.0/8"
    peer_vnet_id      = "10.0.0.0/8"
  }
}
`, rInt, location, rInt, openshiftVersion, clientId, clientSecret, groupId)
}

func testCheckAzureRMOpenShiftClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).containers.OpenShiftClustersClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_openshift_cluster" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := conn.Get(ctx, resourceGroup, name)

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Managed OpenShift Cluster still exists:\n%#v", resp)
		}
	}

	return nil
}
