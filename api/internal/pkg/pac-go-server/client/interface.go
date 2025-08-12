package client

import (
	"github.com/Nerzal/gocloak/v13"
)

//go:generate mockgen -destination=mock_keycloak.go -package=client . Keycloak
type Keycloak interface {
	GetClient() *gocloak.GoCloak
	GetUsers() ([]*gocloak.User, error)
	GetUser(id string) (*gocloak.User, error)
	GetGroups() ([]*gocloak.Group, error)
	GetUserInfo() (*gocloak.UserInfo, error)
	AddUserToGroup(userID, groupID string) error
	DeleteUserFromGroup(userID, groupID string) error
	GetUserGroups(userID string) ([]*gocloak.Group, error)
	DeleteUser(userID string) error
	IsRole(name string) bool
	GetUserID() string
}
