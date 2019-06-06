variable "prefix" {
  description = "A prefix used for all resources in this example"
}

variable "location" {
  description = "The Azure Region in which all resources in this example should be provisioned"
}

variable "network_security_group_name" {
  description = "The Azure Network Security Group name to use for this Managed OpenShift Cluster"
}

variable "openshift_client_id" {
  description = "The Client ID for the Service Principal to use for this Managed OpenShift Cluster"
}

variable "openshift_client_secret" {
  description = "The Client Secret for the Service Principal to use for this Managed OpenShift Cluster"
}

variable "tenant_id" {
  description = "The Azure AD tenant ID where the Managed OpenShift cluster should be created"
}
