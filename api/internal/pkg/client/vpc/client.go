package vpc

import (
	"context"
	"fmt"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/vpc-go-sdk/vpcv1"
	"github.com/PDeXchange/pac/internal/pkg/client/iam"
)

var _ VPC = &Client{}

type Client struct {
	vpcService *vpcv1.VpcV1
}

// ListLoadBalancers returns list of load balancers in a region.
func (s *Client) ListLoadBalancers(options *vpcv1.ListLoadBalancersOptions) (*vpcv1.LoadBalancerCollection, *core.DetailedResponse, error) {
	return s.vpcService.ListLoadBalancers(options)
}

// GetLoadBalancer returns a load balancer.
func (s *Client) GetLoadBalancer(options *vpcv1.GetLoadBalancerOptions) (*vpcv1.LoadBalancer, *core.DetailedResponse, error) {
	return s.vpcService.GetLoadBalancer(options)
}

// CreateLoadBalancerPoolMember creates a new member and adds the member to the pool.
func (s *Client) CreateLoadBalancerPoolMember(options *vpcv1.CreateLoadBalancerPoolMemberOptions) (*vpcv1.LoadBalancerPoolMember, *core.DetailedResponse, error) {
	return s.vpcService.CreateLoadBalancerPoolMember(options)
}

// DeleteLoadBalancerPoolMember deletes a member from the load balancer pool.
func (s *Client) DeleteLoadBalancerPoolMember(options *vpcv1.DeleteLoadBalancerPoolMemberOptions) (*core.DetailedResponse, error) {
	return s.vpcService.DeleteLoadBalancerPoolMember(options)
}

func (s *Client) ListLoadBalancerPools(options *vpcv1.ListLoadBalancerPoolsOptions) (*vpcv1.LoadBalancerPoolCollection, *core.DetailedResponse, error) {
	return s.vpcService.ListLoadBalancerPools(options)
}

// ListLoadBalancerPoolMembers returns members of a load balancer pool.
func (s *Client) ListLoadBalancerPoolMembers(options *vpcv1.ListLoadBalancerPoolMembersOptions) (*vpcv1.LoadBalancerPoolMemberCollection, *core.DetailedResponse, error) {
	return s.vpcService.ListLoadBalancerPoolMembers(options)
}

// ListLoadBalancerListeners returns the LoadBalancer Listeners
func (s *Client) ListLoadBalancerListeners(options *vpcv1.ListLoadBalancerListenersOptions) (*vpcv1.LoadBalancerListenerCollection, *core.DetailedResponse, error) {
	return s.vpcService.ListLoadBalancerListeners(options)
}

// CreateLoadBalancerListener creates the LoadBalancer Listener
func (s *Client) CreateLoadBalancerListener(options *vpcv1.CreateLoadBalancerListenerOptions) (*vpcv1.LoadBalancerListener, *core.DetailedResponse, error) {
	return s.vpcService.CreateLoadBalancerListener(options)
}

// CreateLoadBalancerPool creates the LoadBalancer Pool
func (s *Client) CreateLoadBalancerPool(options *vpcv1.CreateLoadBalancerPoolOptions) (*vpcv1.LoadBalancerPool, *core.DetailedResponse, error) {
	return s.vpcService.CreateLoadBalancerPool(options)
}

type Options struct {
	Region string
}

func NewClient(ctx context.Context, options Options) (*Client, error) {
	client := &Client{}
	auth, err := iam.GetIAMAuth()
	if err != nil {
		return nil, err
	}

	client.vpcService, err = vpcv1.NewVpcV1(&vpcv1.VpcV1Options{
		Authenticator: auth,
		URL:           fmt.Sprintf("https://%s.iaas.cloud.ibm.com/v1", options.Region),
	})

	return client, err
}
