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

type KeyClockClient struct {
	ctx         context.Context
	client      *gocloak.GoCloak
	accessToken string
	realm       string
}

func NewKeyClockClient(ctx context.Context) *KeyClockClient {
	hostname := ctx.Value("keycloak_hostname").(string)

	return &KeyClockClient{
		ctx:         ctx,
		client:      gocloak.NewClient(hostname),
		accessToken: ctx.Value("keycloak_access_token").(string),
		realm:       ctx.Value("keycloak_realm").(string),
	}
}

func (k *KeyClockClient) GetClient() *gocloak.GoCloak {
	return k.client
}

// GetUsers for listing all the users from keycloak
func (k *KeyClockClient) GetUsers() ([]*gocloak.User, error) {
	return k.client.GetUsers(k.ctx, k.accessToken, k.realm, gocloak.GetUsersParams{})
}

// GetUsers for listing all the users from keycloak
func (k *KeyClockClient) GetUser(id string) (*gocloak.User, error) {
	return k.client.GetUser(k.ctx, k.accessToken, k.realm, id)
}

func (k *KeyClockClient) GetGroups() ([]*gocloak.Group, error) {
	return k.client.GetGroups(k.ctx, k.accessToken, k.realm, gocloak.GetGroupsParams{})
}

func (k *KeyClockClient) GetUserInfo() (*gocloak.UserInfo, error) {
	return k.client.GetUserInfo(k.ctx, k.accessToken, k.realm)
}

func (k *KeyClockClient) AddUserToGroup(userID, groupID string) error {
	return k.client.AddUserToGroup(k.ctx, k.accessToken, k.realm, userID, groupID)
}

func (k *KeyClockClient) DeleteUserFromGroup(userID, groupID string) error {
	return k.client.DeleteUserFromGroup(k.ctx, k.accessToken, k.realm, userID, groupID)
}

func (k *KeyClockClient) GetUserGroups(userID string) ([]*gocloak.Group, error) {
	return k.client.GetUserGroups(k.ctx, k.accessToken, k.realm, userID, gocloak.GetGroupsParams{})
}

func (k *KeyClockClient) DeleteUser(userID string) error {
	return k.client.DeleteUser(k.ctx, k.accessToken, k.realm, userID)
}

func (k *KeyClockClient) IsRole(name string) bool {
	roles := k.ctx.Value("roles").([]string)

	for _, role := range roles {
		if role == name {
			return true
		}
	}
	return false
}

func (k *KeyClockClient) GetUserID() string {
	return k.ctx.Value("userid").(string)
}
