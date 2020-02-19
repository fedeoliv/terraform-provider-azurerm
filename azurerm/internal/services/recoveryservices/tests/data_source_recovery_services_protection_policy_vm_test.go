package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
)

func TestAccDataSourceAzureRMRecoveryServicesProtectionPolicyVm_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_recovery_services_protection_policy_vm", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.SupportedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRecoveryServicesProtectionPolicyVm_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMRecoveryServicesProtectionPolicyVmExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "name"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "recovery_vault_name"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "resource_group_name"),
					resource.TestCheckResourceAttr(data.ResourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccDataSourceRecoveryServicesProtectionPolicyVm_basic(data acceptance.TestData) string {
	template := testAccAzureRMRecoveryServicesProtectionPolicyVm_basicDaily(data)
	return fmt.Sprintf(` 
%s

data "azurerm_recovery_services_protection_policy_vm" "test" {
  name                = "${azurerm_recovery_services_protection_policy_vm.test.name}"
  recovery_vault_name = "${azurerm_recovery_services_vault.test.name}"
  resource_group_name = "${azurerm_resource_group.test.name}"
}
`, template)
}
