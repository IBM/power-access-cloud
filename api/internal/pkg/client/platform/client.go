package platform

import (
	"context"

	"github.com/pkg/errors"

	"github.com/IBM/platform-services-go-sdk/iamidentityv1"
	"github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"

	"github.com/IBM/power-access-cloud/api/internal/pkg/client/iam"
)

type Client struct {
	resourceControllerClient *resourcecontrollerv2.ResourceControllerV2
	iamIdentityClient        *iamidentityv1.IamIdentityV1
}

func (c *Client) GetResourceInstance(ctx context.Context, id string) (*resourcecontrollerv2.ResourceInstance, error) {
	instance, _, err := c.resourceControllerClient.GetResourceInstanceWithContext(ctx, &resourcecontrollerv2.GetResourceInstanceOptions{ID: &id})
	return instance, err
}

func NewClient() (*Client, error) {
	auth, err := iam.GetIAMAuth()
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve authenticator")
	}

	rcClient, err := resourcecontrollerv2.NewResourceControllerV2(&resourcecontrollerv2.ResourceControllerV2Options{Authenticator: auth})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create resource controller client")
	}

	iamClient, err := iamidentityv1.NewIamIdentityV1(&iamidentityv1.IamIdentityV1Options{
		Authenticator: auth,
		URL:           iamidentityv1.DefaultServiceURL,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create iam identity client")
	}

	return &Client{
		resourceControllerClient: rcClient,
		iamIdentityClient:        iamClient,
	}, nil
}

// NewIAMIdentityClient creates iam identity client.
func NewIAMIdentityClient() (*iamidentityv1.IamIdentityV1, error) {
	auth, err := iam.GetIAMAuth()
	if err != nil {
		return nil, err
	}
	return iamidentityv1.NewIamIdentityV1(&iamidentityv1.IamIdentityV1Options{
		Authenticator: auth,
		URL:           iamidentityv1.DefaultServiceURL,
	})
}
