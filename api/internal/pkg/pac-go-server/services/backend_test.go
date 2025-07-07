package services

import (
	"context"
	"reflect"
	"testing"
	"time"

	pac "github.com/PDeXchange/pac/apis/app/v1alpha1"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/client/kubernetes"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/db"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type testContext struct {
	userID              string
	keyCloakHostname    string
	keyCloakAccessToken string
	keyCloakRealm       string
	roles               []string
	groups              []models.Group
	username            string
}

type customValues = map[string]interface{}

// return new mockclients and tearDown to release resource
func setUp(t testing.TB) (mockedKubeClient *kubernetes.MockClient, mockedDBClient *db.MockDB, tearDown func()) {
	// mocking kubeclient
	ctrlKube := gomock.NewController(t)
	mockkubeclient := kubernetes.NewMockClient(ctrlKube)

	// mocking db client
	ctrlDB := gomock.NewController(t)
	mockDBClient := db.NewMockDB(ctrlDB)

	return mockkubeclient, mockDBClient, func() {
		ctrlKube.Finish()
		ctrlDB.Finish()
	}
}

func getResource(apiType string, customValues map[string]interface{}) interface{} {
	switch apiType {

	case "get-all-catalogs":
		catalogList := pac.CatalogList{}
		catCapacity := pac.Capacity{
			CPU:    "2",
			Memory: 2,
		}
		vmCat := pac.VMCatalog{
			CRN:           "test-crn",
			ProcessorType: "test-processor",
			SystemType:    "test-system",
			Image:         "test-image",
			Network:       "test-network",
			Capacity:      catCapacity,
		}
		catStatus := pac.CatalogStatus{
			Ready:   true,
			Message: "catalog is ready",
		}
		catSpec := pac.CatalogSpec{
			Type:                    "test",
			Description:             "catalog for testing",
			Capacity:                catCapacity,
			Retired:                 false,
			Expiry:                  10,
			ImageThumbnailReference: "https://test-catalog",
			VM:                      vmCat,
		}
		testCatalog := pac.Catalog{
			Spec:   catSpec,
			Status: catStatus,
		}
		catalogList.Items = []pac.Catalog{testCatalog}
		return catalogList

	case "create-catalog":
		cap := models.Capacity{
			CPU:    2,
			Memory: 2,
		}
		vm := models.VM{
			CRN:           "test-crn",
			ProcessorType: "ppc",
			SystemType:    "test",
			Image:         "image",
			Network:       "internal",
			Capacity:      cap,
		}
		status := models.CatalogStatus{
			Ready:   true,
			Message: "catalog is ready",
		}
		catalog := models.Catalog{
			ID:                      "1",
			Type:                    "VM",
			Name:                    "test-catalog",
			Description:             "catalog for test",
			Capacity:                cap,
			Retired:                 false,
			Expiry:                  2,
			ImageThumbnailReference: "https://thumbnail",
			VM:                      vm,
			Status:                  status,
		}

		// Update catalog with custom values if provided
		for key, value := range customValues {
			if fieldValue := reflect.ValueOf(&catalog).Elem().FieldByName(key); fieldValue.IsValid() {
				if value != nil {
					fieldValue.Set(reflect.ValueOf(value))
				}
			}
		}
		return catalog
	case "get-catalog":
		catCapacity := pac.Capacity{
			CPU:    "2",
			Memory: 2,
		}
		vmCat := pac.VMCatalog{
			CRN:           "test-crn",
			ProcessorType: "test-processor",
			SystemType:    "test-system",
			Image:         "test-image",
			Network:       "test-network",
			Capacity:      catCapacity,
		}
		catStatus := pac.CatalogStatus{
			Ready:   true,
			Message: "catalog is ready",
		}
		catSpec := pac.CatalogSpec{
			Type:                    "test",
			Description:             "catalog for testing",
			Capacity:                catCapacity,
			Retired:                 false,
			Expiry:                  10,
			ImageThumbnailReference: "https://test-catalog",
			VM:                      vmCat,
		}
		catalog := pac.Catalog{
			Spec:   catSpec,
			Status: catStatus,
		}

		// Update catalog with custom values if provided
		for key, value := range customValues {
			if fieldValue := reflect.ValueOf(&catalog).Elem().FieldByName(key); fieldValue.IsValid() {
				if value != nil {
					fieldValue.Set(reflect.ValueOf(value))
				}
			}
		}
		return catalog
	case "get-all-services":
		serviceList := pac.ServiceList{}
		serviceSpec := pac.ServiceSpec{
			UserID:      "test-user",
			DisplayName: "test-service",
			Expiry:      metav1.Time{},
			Catalog:     corev1.LocalObjectReference{Name: "test-catalog"},
			SSHKeys:     []string{"ssh-key"},
		}
		vm := pac.VM{
			InstanceID:        "test",
			IPAddress:         "1.1.1.1",
			ExternalIPAddress: "2.2.2.2",
			State:             "ready",
		}
		status := pac.ServiceStatus{
			VM:         vm,
			AccessInfo: "access-info",
			Expired:    false,
			Message:    "test service",
			State:      pac.ServiceStateCreated,
			Successful: true,
		}
		service := pac.Service{
			Spec:   serviceSpec,
			Status: status,
		}
		// Update services with custom values if provided
		for key, value := range customValues {
			if fieldValue := reflect.ValueOf(&service).Elem().FieldByName(key); fieldValue.IsValid() {
				if value != nil {
					fieldValue.Set(reflect.ValueOf(value))
				}
			}
		}
		serviceList.Items = []pac.Service{service}
		return serviceList
	case "get-key-by-userid":
		key := models.Key{
			ID:      [12]byte{1},
			UserID:  "12345",
			Name:    "test-key",
			Content: "content",
		}
		// Update key with custom values if provided
		for key, value := range customValues {
			if fieldValue := reflect.ValueOf(&key).Elem().FieldByName(key); fieldValue.IsValid() {
				if value != nil {
					fieldValue.Set(reflect.ValueOf(value))
				}
			}
		}
		return []models.Key{key}
	case "get-groups-quota":
		quota := models.Quota{
			ID:      [12]byte{2},
			GroupID: "122343",
			Capacity: models.Capacity{
				CPU:    10,
				Memory: 10,
			},
		}
		// Update quota with custom values if provided
		for key, value := range customValues {
			if fieldValue := reflect.ValueOf(&quota).Elem().FieldByName(key); fieldValue.IsValid() {
				if value != nil {
					fieldValue.Set(reflect.ValueOf(value))
				}
			}
		}
		return []models.Quota{quota}
	case "create-service":
		serviceStatus := models.ServiceStatus{
			State:      "Ready",
			Message:    "ready to use",
			AccessInfo: "127.0.0.1",
		}
		service := models.Service{
			ID:          "12345",
			UserID:      "122343",
			Name:        "test-user",
			DisplayName: "test-service",
			CatalogName: "test-catalog",
			Expiry:      time.Time{},
			Status:      serviceStatus,
		}
		// Update service with custom values if provided
		for key, value := range customValues {
			if fieldValue := reflect.ValueOf(&service).Elem().FieldByName(key); fieldValue.IsValid() {
				if value != nil {
					fieldValue.Set(reflect.ValueOf(value))
				}
			}
		}
		return service
	case "get-request-by-service-name":
		request := models.Request{
			ID:            [12]byte{1},
			UserID:        "12345",
			Justification: "justification",
			Comment:       "comment",
			CreatedAt:     time.Time{},
			State:         "approved",
			RequestType:   "extension",
			GroupAdmission: &models.GroupAdmission{
				GroupID:   "test-group",
				Group:     "manager",
				Requester: "test-user",
			},
			ServiceExpiry: &models.ServiceExpiry{
				Name:   "test-service",
				Expiry: time.Now(),
			},
		}
		// Update request with custom values if provided
		for key, value := range customValues {
			if fieldValue := reflect.ValueOf(&request).Elem().FieldByName(key); fieldValue.IsValid() {
				if value != nil {
					fieldValue.Set(reflect.ValueOf(value))
				}
			}
		}
		return []models.Request{request}
	default:
		return nil
	}
}

func getContext(requestCtx testContext) context.Context {
	//nolint:staticcheck
	ctx := context.WithValue(context.Background(), "userid", requestCtx.userID)
	//nolint:staticcheck
	ctx = context.WithValue(ctx, "keycloak_hostname", requestCtx.keyCloakHostname)
	//nolint:staticcheck
	ctx = context.WithValue(ctx, "keycloak_access_token", requestCtx.keyCloakAccessToken)
	//nolint:staticcheck
	ctx = context.WithValue(ctx, "keycloak_realm", requestCtx.keyCloakRealm)
	//nolint:staticcheck
	ctx = context.WithValue(ctx, "groups", requestCtx.groups)
	//nolint:staticcheck
	ctx = context.WithValue(ctx, "username", requestCtx.username)
	//nolint:staticcheck
	ctx = context.WithValue(ctx, "roles", requestCtx.roles)
	return ctx
}

func formContext(params customValues) testContext {
	ctx := testContext{}
	for key, val := range params {
		switch key {
		case "userid":
			if v, ok := val.(string); ok {
				ctx.userID = v
			} else {
				panic("userid must be string")
			}
		case "keycloak_access_token":
			if v, ok := val.(string); ok {
				ctx.keyCloakAccessToken = v

			} else {
				panic("keycloak_access_token must be string")
			}
		case "keycloak_realm":
			if v, ok := val.(string); ok {
				ctx.keyCloakRealm = v
			} else {
				panic("keycloak_realm must be string")
			}
		case "keycloak_hostname":
			if v, ok := val.(string); ok {
				ctx.keyCloakHostname = v

			} else {
				panic("keycloak_hostname must be string")
			}
		case "roles":
			if v, ok := val.([]string); ok {
				ctx.roles = v

			} else {
				panic("invalid roles information")
			}
		case "groups":
			if v, ok := val.([]models.Group); ok {
				ctx.groups = v

			} else {
				panic("invalid groups information")
			}
		case "username":
			if v, ok := val.(string); ok {
				ctx.username = v

			} else {
				panic("invalid username information")
			}
		}
	}
	return ctx
}

func formGroup(params customValues) []models.Group {
	group := models.Group{}
	for key, val := range params {
		switch key {
		case "id":
			if v, ok := val.(string); ok {
				group.ID = v
			} else {
				panic("id must be string")
			}
		case "name":
			if v, ok := val.(string); ok {
				group.Name = v

			} else {
				panic("name must be string")
			}
		case "membership":
			if v, ok := val.(bool); ok {
				group.Membership = v

			} else {
				panic("membership must be bool")
			}
		case "quota":
			if v, ok := val.(models.Capacity); ok {
				group.Quota = v

			} else {
				panic("invalid quota information")
			}
		}
	}
	return []models.Group{group}
}

func formQuota(params customValues) models.Capacity {
	cap := models.Capacity{}
	for key, val := range params {
		switch key {
		case "cpu":
			if v, ok := val.(float64); ok {
				cap.CPU = v
			} else {
				panic("cpu must be float64")
			}
		case "Memory":
			if v, ok := val.(int); ok {
				cap.Memory = v

			} else {
				panic("memory must be int")
			}
		}
	}
	return cap
}
