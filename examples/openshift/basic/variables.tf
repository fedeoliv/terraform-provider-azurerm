variable "prefix" {
  description = "A prefix used for all resources in this example"
}

variable "location" {
  description = "The Azure Region in which all resources in this example should be provisioned"
}

variable "network_security_group_name" {
  description = "The Azure Network Security Group name to use for this Managed OpenShift Cluster"
}

variable "provider_kind" {
  description = "The ARO cluster identity provider type (AADIdentityProvider or OpenShiftManagedClusterBaseIdentityProvider)"
}

variable "provider_client_id" {
  description = "The client ID associated with the ARO cluster identity provider"
}

variable "provider_client_secret" {
  description = "The client secret associated with the ARO cluster identity provider"
}

variable "provider_group_id" {
  description = "The group id to be granted to the ARO cluster admin role"
}

variable "tenant_id" {
  description = "The Azure AD tenant ID where the Managed OpenShift cluster should be created"
}

variable "openshift_version" {
  description = "The Managed OpenShift version"
}
