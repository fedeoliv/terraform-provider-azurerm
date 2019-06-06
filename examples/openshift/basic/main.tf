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

  service_principal {
    client_id     = "${var.openshift_client_id}"
    client_secret = "${var.openshift_client_secret}"
  }

  tags = {
    environment = "Development"
  }
}
