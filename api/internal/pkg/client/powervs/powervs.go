package powervs

import "github.com/IBM-Cloud/power-go-client/power/models"

type PowerVS interface {
	GetAllInstance() (*models.PVMInstances, error)
	GetAllDHCPServers() (models.DHCPServers, error)
	GetDHCPServer(id string) (*models.DHCPServerDetail, error)
	GetImageByName(name string) (*models.ImageReference, error)
	GetNetworkByName(name string) (*models.NetworkReference, error)
	GetNetworks() (*models.Networks, error)
	GetNetwork(id string) (*models.Network, error)
	CreateNetwork(body *models.NetworkCreate) (*models.Network, error)
	CreateVM(opts *models.PVMInstanceCreate) (*models.PVMInstanceList, error)
	GetVM(id string) (*models.PVMInstance, error)
	DeleteVM(id string) error
	DeleteVMWithVolumes(id string) error
	// Volume operations
	CreateVolume(body *models.CreateDataVolume) (*models.Volume, error)
	DeleteVolume(id string) error
}
