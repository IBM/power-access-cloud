package client

import (
	"context"
	"errors"

	"github.com/Nerzal/gocloak/v13"
)

var (
	ErrorGroupNotFound = errors.New("group not found")
	ErrorUserNotFound  = errors.New("user not found")
)

// KeyCloakConfig holds the configuration for Keycloak operations
type KeyCloakConfig struct {
	Hostname    string
	AccessToken string
	Realm       string
	UserID      string
	Roles       []string
}

// KeycloakClient implements KeyCloakInterface
type KeyCloakClient struct {
	ctx    context.Context
	config KeyCloakConfig
	client *gocloak.GoCloak
}

var NewKeyCloakClient = func(config KeyCloakConfig, ctx context.Context) Keycloak {
	return &KeyCloakClient{
		ctx:    ctx,
		config: config,
		client: gocloak.NewClient(config.Hostname),
	}
}

func (k *KeyCloakClient) GetClient() *gocloak.GoCloak {
	return k.client
}

// GetUsers for listing all the users from keycloak
func (k *KeyCloakClient) GetUsers() ([]*gocloak.User, error) {
	return k.client.GetUsers(k.ctx, k.config.AccessToken, k.config.Realm, gocloak.GetUsersParams{})
}

// GetUsers for listing all the users from keycloak
func (k *KeyCloakClient) GetUser(id string) (*gocloak.User, error) {
	return k.client.GetUser(k.ctx, k.config.AccessToken, k.config.Realm, id)
}

func (k *KeyCloakClient) GetGroups() ([]*gocloak.Group, error) {
	return k.client.GetGroups(k.ctx, k.config.AccessToken, k.config.Realm, gocloak.GetGroupsParams{})
}

func (k *KeyCloakClient) GetUserInfo() (*gocloak.UserInfo, error) {
	return k.client.GetUserInfo(k.ctx, k.config.AccessToken, k.config.Realm)
}

func (k *KeyCloakClient) AddUserToGroup(userID, groupID string) error {
	return k.client.AddUserToGroup(k.ctx, k.config.AccessToken, k.config.Realm, userID, groupID)
}

func (k *KeyCloakClient) DeleteUserFromGroup(userID, groupID string) error {
	return k.client.DeleteUserFromGroup(k.ctx, k.config.AccessToken, k.config.Realm, userID, groupID)
}

func (k *KeyCloakClient) GetUserGroups(userID string) ([]*gocloak.Group, error) {
	return k.client.GetUserGroups(k.ctx, k.config.AccessToken, k.config.Realm, userID, gocloak.GetGroupsParams{})
}

func (k *KeyCloakClient) DeleteUser(userID string) error {
	return k.client.DeleteUser(k.ctx, k.config.AccessToken, k.config.Realm, userID)
}

func (k *KeyCloakClient) IsRole(name string) bool {

	for _, role := range k.config.Roles {
		if role == name {
			return true
		}
	}
	return false
}

func (k *KeyCloakClient) GetUserID() string {
	return k.config.UserID
}

// GetConfigFromContext gets config from context
func GetConfigFromContext(ctx context.Context) KeyCloakConfig {

	config := KeyCloakConfig{
		Hostname:    ctx.Value("keycloak_hostname").(string),
		AccessToken: ctx.Value("keycloak_access_token").(string),
		Realm:       ctx.Value("keycloak_realm").(string),
	}

	if userID := ctx.Value("userid"); userID != nil {
		config.UserID = userID.(string)
	}

	if roles := ctx.Value("roles"); roles != nil {
		config.Roles = roles.([]string)
	}

	return config
}
