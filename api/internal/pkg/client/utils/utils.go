package utils

import (
	"context"
	"fmt"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/platform-services-go-sdk/iamidentityv1"
	"github.com/IBM/power-access-cloud/api/internal/pkg/client/platform"
)

// GetAccountID returns IBM cloud account ID of API key used.
func GetAccountID(ctx context.Context, auth *core.IamAuthenticator) (string, error) {
	iamv1, err := platform.NewIAMIdentityClient()
	if err != nil {
		return "", err
	}

	apiKeyDetailsOpt := iamidentityv1.GetAPIKeysDetailsOptions{IamAPIKey: &auth.ApiKey}
	apiKey, _, err := iamv1.GetAPIKeysDetailsWithContext(ctx, &apiKeyDetailsOpt)
	if err != nil {
		return "", err
	}
	if apiKey == nil {
		return "", fmt.Errorf("could not retrieve account id")
	}

	return *apiKey.AccountID, nil
}
