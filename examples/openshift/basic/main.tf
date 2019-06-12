resource "azurerm_resource_group" "example" {
  name     = "${var.prefix}-openshift-resources"
  location = "${var.location}"
}

resource "azurerm_openshift_cluster" "example" {
  name                      = "${var.prefix}-openshift"
  location                  = "${azurerm_resource_group.example.location}"
  resource_group_name       = "${azurerm_resource_group.example.name}"
  openshift_version         = "${var.openshift_version}"
  fqdn                      = ""

  auth_profile {
    providers = [
      {
        name = "Azure AD"
        provider = {
          kind            = "${var.provider_kind}"
          client_id       = "${var.provider_client_id}"
          client_secret   = "${var.provider_client_secret}"
          group_id        = "${var.provider_group_id}"
        }
      }
    ]
  }

  master_pool_profile {
    name              = "default"
    count             = 3
    vm_size           = "Standard_D2s_v3"
    os_type           = "Linux"
    subnet_cidr       = "10.0.0.0/24"
  }

  agent_pool_profile {
    name              = "default"
    count             = 3
    vm_size           = "Standard_D2s_v3"
    os_type           = "Linux"
    subnet_cidr       = "10.0.0.0/24"
    role              = "Compute"
  }

  network_profile {
    vnet_cidr         = "10.0.0.0/8"
    peer_vnet_id      = "10.0.0.0/8"
  }

  router_profile {
    name              = "default"
    public_subdomain  = "b788fade68d345da9b77.location1.int.aksapp.io"
  }

  tags = {
    environment = "Development"
  }
}
