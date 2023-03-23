package powervs

import (
	"context"
	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/PDeXchange/pac/internal/pkg/client/iam"
	"github.com/PDeXchange/pac/internal/pkg/client/utils"
)

var _ PowerVS = &Client{}

type Client struct {
	instanceClient *instance.IBMPIInstanceClient
	networkClient  *instance.IBMPINetworkClient
	dhcpClient     *instance.IBMPIDhcpClient
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

type Options struct {
	CloudInstanceID string
	Zone            string
	Debug           bool
}

func NewClient(ctx context.Context, options Options) (*Client, error) {
	auth, err := iam.GetIAMAuth()
	if err != nil {
		return nil, err
	}

	account, err := utils.GetAccountID(ctx, auth)
	if err != nil {
		return nil, err
	}

	opt := &ibmpisession.IBMPIOptions{
		Authenticator: auth,
		UserAccount:   account,
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
	}, nil
}
