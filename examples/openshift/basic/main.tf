resource "azurerm_resource_group" "example" {
  name     = "${var.prefix}-openshift-resources"
  location = "${var.location}"
}

resource "azurerm_network_security_group" "test" {
  name                = "${var.network_security_group_name}"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
}

resource "azurerm_openshift_cluster" "example" {
  name                = "${var.prefix}-openshift"
  location            = "${azurerm_resource_group.example.location}"
  resource_group_name = "${azurerm_resource_group.example.name}"
  tenant_id           = "${var.tenant_id}"
  network_security_group_id = "${azurerm_network_security_group.test.id}"
  openshift_version   = "${var.openshift_version}"

  agent_pool_profile {
    name              = "default"
    count             = 3
    vm_size           = "Standard_D2s_v3"
    os_type           = "Linux"
    os_disk_size_gb   = 30
  }

  service_principal {
    client_id     = "${var.openshift_client_id}"
    client_secret = "${var.openshift_client_secret}"
  }

  tags = {
    environment = "Development"
  }
}
