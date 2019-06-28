resource "azurerm_resource_group" "example" {
  name     = "${var.prefix}-openshift-resources"
  location = "${var.location}"
}

resource "azurerm_openshift_cluster" "example" {
  name                = "${var.prefix}-openshift"
  location            = "${azurerm_resource_group.example.location}"
  resource_group_name = "${azurerm_resource_group.example.name}"
  openshift_version   = "${var.openshift_version}"

  network_profile {
    vnet_cidr = "10.0.0.0/8"
  }

  master_pool_profile {
    name        = "master"
    count       = 3
    vm_size     = "Standard_D4s_v3"
    os_type     = "Linux"
    subnet_cidr = "10.0.0.0/24"
  }

  agent_pool_profile {
    name        = "infra"
    role        = "infra"
    count       = 2
    vm_size     = "Standard_D4s_v3"
    os_type     = "Linux"
    subnet_cidr = "10.0.0.0/24"
  }

  agent_pool_profile {
    name        = "compute"
    role        = "compute"
    count       = 4
    vm_size     = "Standard_D4s_v3"
    os_type     = "Linux"
    subnet_cidr = "10.0.0.0/24"
  }

  router_profile {
    name = "default"
  }

  auth_profile {
    providers = [
      {
        name = "Azure AD"
        provider = {
          kind          = "${var.provider_kind}"
          client_id     = "${var.provider_client_id}"
          client_secret = "${var.provider_client_secret}"
          group_id      = "${var.provider_group_id}"
        }
      }
    ]
  }

  tags = {
    environment = "Development"
  }
}
