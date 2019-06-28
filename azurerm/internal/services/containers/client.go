package containers

import (
	"github.com/Azure/azure-sdk-for-go/services/containerinstance/mgmt/2018-10-01/containerinstance"
	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2017-10-01/containerregistry"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2019-06-01/containerservice"
)

type Client struct {
	KubernetesClustersClient   containerservice.ManagedClustersClient
	OpenShiftClustersClient    containerservice.OpenShiftManagedClustersClient
	GroupsClient               containerinstance.ContainerGroupsClient
	RegistryClient             containerregistry.RegistriesClient
	RegistryReplicationsClient containerregistry.ReplicationsClient
	ServicesClient             containerservice.ContainerServicesClient
}
