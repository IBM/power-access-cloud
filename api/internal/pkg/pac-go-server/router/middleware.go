package router

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	pacClient "github.com/PDeXchange/pac/internal/pkg/pac-go-server/client"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/utils"
)

func AllowAdminOnly(c *gin.Context) {
	config := pacClient.GetConfigFromContext(c.Request.Context())
	kc := pacClient.NewKeyCloakClient(config, c)
	if !kc.IsRole(utils.ManagerRole) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Not authorized to perform this action"})
		return
	}
	c.Next()
}

func RetrospectKeycloakToken(c *gin.Context) {
	//nolint:staticcheck
	ctx := context.WithValue(c.Request.Context(), "keycloak_client", client)
	//nolint:staticcheck
	ctx = context.WithValue(ctx, "keycloak_realm", realm)
	//nolint:staticcheck
	ctx = context.WithValue(ctx, "keycloak_hostname", hostname)

	authHeader := c.GetHeader("Authorization")
	if len(authHeader) < 1 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No Authorization header provided"})
		return
	}
	accessToken := strings.Split(authHeader, " ")[1]
	//nolint:staticcheck
	ctx = context.WithValue(ctx, "keycloak_access_token", accessToken)
	rptResult, err := client.RetrospectToken(c.Request.Context(), accessToken, clientId, clientSecret, realm)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	isTokenValid := *rptResult.Active
	if !isTokenValid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is not valid"})
		return
	}

	// Get username of a user and append it to the context
	{
		user, err := client.GetUserInfo(c.Request.Context(), accessToken, realm)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//nolint:staticcheck
		ctx = context.WithValue(ctx, "username", *user.PreferredUsername)
		//nolint:staticcheck
		ctx = context.WithValue(ctx, "userid", *user.Sub)
	}

	// Get groups of a user and append them to the context
	{
		groups, err := client.GetAccountGroups(c.Request.Context(), accessToken, realm)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to get groups of a user: %s", err.Error())})
			return
		}

		getGroups := func() []models.Group {
			var grps []models.Group
			for _, group := range groups {
				grps = append(grps, models.Group{ID: *group.ID, Name: *group.Name})
			}
			return grps
		}
		//nolint:staticcheck
		ctx = context.WithValue(ctx, "groups", getGroups())
	}

	_, claim, err := client.DecodeAccessToken(ctx, accessToken, realm)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to parse the access token: %s", err.Error())})
		return
	}

	getRoles := func() ([]string, error) {
		realmAccess, ok := (*claim)["realm_access"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("no realm_access in token")
		}
		roles, ok := realmAccess["roles"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("no roles in realm_access")
		}
		var rolesList []string
		for _, role := range roles {
			rolesList = append(rolesList, role.(string))
		}
		return rolesList, nil
	}

	if roles, err := getRoles(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to get roles of a user: %s", err.Error())})
		return
	} else {
		//nolint:staticcheck
		ctx = context.WithValue(ctx, "roles", roles)
	}

	// Replace the request context with the new context
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
