package powervs

import (
	"context"

	"github.com/pkg/errors"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/power-access-cloud/api/internal/pkg/client/iam"
	"github.com/IBM/power-access-cloud/api/internal/pkg/client/utils"
)

var _ PowerVS = &Client{}

type Client struct {
	instanceClient *instance.IBMPIInstanceClient
	networkClient  *instance.IBMPINetworkClient
	dhcpClient     *instance.IBMPIDhcpClient
	imageClient    *instance.IBMPIImageClient
}

// GetAllInstance returns all the virtual machine in the Power VS service instance.
func (s *Client) GetAllInstance() (*models.PVMInstances, error) {
	return s.instanceClient.GetAll()
}

// GetAllDHCPServers returns all the DHCP servers in the Power VS service instance.
func (s *Client) GetAllDHCPServers() (models.DHCPServers, error) {
	return s.dhcpClient.GetAll()
}

// GetDHCPServer returns the details for DHCP server associated with id.
func (s *Client) GetDHCPServer(id string) (*models.DHCPServerDetail, error) {
	return s.dhcpClient.Get(id)
}

func (s *Client) GetInstance(id string) (*models.PVMInstance, error) {
	return s.instanceClient.Get(id)
}

// GetImageByName returns *models.ImageReference for given image name if exists, if not will return appropriate error
func (s *Client) GetImageByName(name string) (*models.ImageReference, error) {
	images, err := s.imageClient.GetAll()
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving images")
	}

	for _, image := range images.Images {
		if *image.Name == name {
			return image, nil
		}
	}

	return nil, errors.Errorf("error retrieving image by name %s", name)
}

// GetNetworkByName returns *models.NetworkReference for given network name if exists, if not will return appropriate error
func (s *Client) GetNetworkByName(name string) (*models.NetworkReference, error) {
	networks, err := s.networkClient.GetAll()
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving networks")
	}

	for _, network := range networks.Networks {
		if *network.Name == name {
			return network, nil
		}
	}

	return nil, errors.Errorf("error retrieving network by name %s", name)
}

func (s *Client) GetNetworks() (*models.Networks, error) {
	return s.networkClient.GetAll()
}

func (s *Client) GetNetwork(id string) (*models.Network, error) {
	return s.networkClient.Get(id)
}

func (s *Client) CreateNetwork(body *models.NetworkCreate) (*models.Network, error) {
	return s.networkClient.Create(body)
}

func (s *Client) CreateVM(opts *models.PVMInstanceCreate) (*models.PVMInstanceList, error) {
	return s.instanceClient.Create(opts)
}

func (s *Client) GetVM(id string) (*models.PVMInstance, error) {
	return s.instanceClient.Get(id)
}

func (s *Client) DeleteVM(id string) error {
	return s.instanceClient.Delete(id)
}

type Options struct {
	AccountID       string
	CloudInstanceID string
	Zone            string
	Debug           bool
}

func NewClient(ctx context.Context, options Options) (*Client, error) {
	auth, err := iam.GetIAMAuth()
	if err != nil {
		return nil, err
	}

	var accountID string
	if options.AccountID == "" {
		accountID, err = utils.GetAccountID(ctx, auth)
		if err != nil {
			return nil, err
		}
	} else {
		accountID = options.AccountID
	}

	opt := &ibmpisession.IBMPIOptions{
		Authenticator: auth,
		UserAccount:   accountID,
		Zone:          options.Zone,
		Debug:         options.Debug,
	}
	session, err := ibmpisession.NewIBMPISession(opt)
	if err != nil {
		return nil, err
	}

	return &Client{
		instanceClient: instance.NewIBMPIInstanceClient(ctx, session, options.CloudInstanceID),
		networkClient:  instance.NewIBMPINetworkClient(ctx, session, options.CloudInstanceID),
		dhcpClient:     instance.NewIBMPIDhcpClient(ctx, session, options.CloudInstanceID),
		imageClient:    instance.NewIBMPIImageClient(ctx, session, options.CloudInstanceID),
	}, nil
}
