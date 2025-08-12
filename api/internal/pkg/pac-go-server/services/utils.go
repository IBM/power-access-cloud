package services

import (
	"context"
	"net/http"
	"strings"

	"github.com/Nerzal/gocloak/v13"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/client"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
)

func getGroups(context context.Context) ([]*gocloak.Group, error) {
	config := client.GetConfigFromContext(context)
	groups, err := client.NewKeyCloakClient(config, context).GetGroups()
	if err != nil {
		return nil, err
	}
	return filterExcludedGroups(groups), nil
}

func getGroup(context context.Context, id string) (*gocloak.Group, error) {
	groups, err := getGroups(context)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		if *group.ID == id {
			return group, nil
		}
	}
	return nil, client.ErrorGroupNotFound
}

// isExcludedGroup checks if the group is in the list of excluded groups
func isExcludedGroup(group string) bool {
	for _, excludedGroup := range models.ExcludeGroups {
		if group == excludedGroup {
			return true
		}
	}
	return false
}

// filterExcludedGroups filters out groups that are excluded
func filterExcludedGroups(groups []*gocloak.Group) []*gocloak.Group {
	var filteredGroups []*gocloak.Group
	for _, group := range groups {
		if !isExcludedGroup(*group.Name) {
			filteredGroups = append(filteredGroups, group)
		}
	}
	return filteredGroups
}

// getKeycloakHttpStatus returns keycloak http status code on the basis of error message
func getKeycloakHttpStatus(err error) int {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "unauthorized"):
		return http.StatusUnauthorized

	case strings.Contains(errMsg, "forbidden"):
		return http.StatusForbidden

	case strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "bad request"):
		return http.StatusBadRequest

	case strings.Contains(errMsg, "not found"):
		return http.StatusNotFound

	case strings.Contains(errMsg, "timeout"):
		return http.StatusGatewayTimeout

	case strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "connection reset"):
		return http.StatusServiceUnavailable

	default:
		return http.StatusInternalServerError
	}
}
