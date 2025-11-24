package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	pac "github.com/IBM/power-access-cloud/api/apis/app/v1alpha1"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
)

func TestGetAllServices(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockClient, _, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
	}{
		{
			name: "get all services succesfully",
			mockFunc: func() {
				mockClient.EXPECT().GetServices(gomock.Any()).Return(getResource("get-all-services", nil).(pac.ServiceList), nil).Times(1)
				mockKCClient.EXPECT().GetUserID().Return("12345").Times(1)
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus: http.StatusOK,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/services", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			kubeClient = mockClient
			GetAllServicesHandler(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestGetService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockClient, _, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestParams  gin.Param
		requestContext testContext
		httpStatus     int
	}{
		{
			name: "get service succesfully",
			mockFunc: func() {
				mockClient.EXPECT().GetService(gomock.Any()).Return(getResource("get-service", nil).(pac.Service), nil).Times(1)
				mockKCClient.EXPECT().GetUserID().Return("12345").Times(1)
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(true).Times(1)
			},
			requestParams: gin.Param{Key: "name", Value: "test-service"},
			httpStatus:    http.StatusOK,
		},
		{
			name:          "service name not set",
			mockFunc:      func() {},
			requestParams: gin.Param{Key: "name", Value: ""},
			httpStatus:    http.StatusBadRequest,
		},
		{
			name: "failed to get service",
			mockFunc: func() {
				mockClient.EXPECT().GetService(gomock.Any()).Return(getResource("get-service", nil).(pac.Service), errors.New("failed to get service")).Times(1)
			},
			requestParams: gin.Param{Key: "name", Value: "test-service"},
			httpStatus:    http.StatusBadRequest,
		},
		{
			name: "user is not admin or owner of the service",
			mockFunc: func() {
				mockClient.EXPECT().GetService(gomock.Any()).Return(getResource("get-service", nil).(pac.Service), nil).Times(1)
				mockKCClient.EXPECT().GetUserID().Return("1231245").Times(1)
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(false).Times(1)
			},
			requestParams: gin.Param{Key: "name", Value: "test-service"},
			httpStatus:    http.StatusUnauthorized,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodPost, "/services", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			kubeClient = mockClient
			GetService(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestCreateService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockClient, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()
	testcases := []struct {
		name           string
		mockFunc       func()
		service        models.Service
		requestContext testContext
		httpStatus     int
	}{
		{
			name: "create service succesfully",
			mockFunc: func() {
				mockClient.EXPECT().GetCatalog(gomock.Any()).Return(getResource("get-catalog", nil).(pac.Catalog), nil).Times(2)
				mockClient.EXPECT().CreateService(gomock.Any()).Return(nil).Times(1)
				mockClient.EXPECT().GetServices(gomock.Any()).Return(getResource("get-all-services", nil).(pac.ServiceList), nil).Times(1)
				mockKCClient.EXPECT().GetUserID().Return("122344").Times(2)
				mockDBClient.EXPECT().GetKeyByUserID(gomock.Any()).Return(getResource("get-key-by-userid", nil).([]models.Key), nil).Times(1)
				mockDBClient.EXPECT().GetGroupsQuota(gomock.Any()).Return(getResource("get-groups-quota", nil).([]models.Quota), nil).Times(1)
			},
			service: getResource("create-service", nil).(models.Service),
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
				"groups": formGroup(customValues{
					"id":         "122343",
					"name":       "silver",
					"membership": true,
					"quota": formQuota(customValues{
						"cpu":    6.0,
						"memory": 4,
					}),
				}),
			}),
			httpStatus: http.StatusCreated,
		},
		{
			name: "catalog is retired",
			mockFunc: func() {
				mockClient.EXPECT().GetCatalog(gomock.Any()).Return(getResource("get-catalog", nil).(pac.Catalog), nil).Times(2)
				mockKCClient.EXPECT().GetUserID().Return("122344").Times(2)
				mockDBClient.EXPECT().GetKeyByUserID(gomock.Any()).Return(getResource("get-key-by-userid", nil).([]models.Key), nil).Times(1)
				mockClient.EXPECT().GetServices(gomock.Any()).Return(getResource("get-all-services", nil).(pac.ServiceList), nil).Times(1)
			},
			service:    getResource("create-service", customValues{"retired": "true"}).(models.Service),
			httpStatus: http.StatusBadRequest,
		},
		{
			name: "catalog not ready",
			mockFunc: func() {
				mockClient.EXPECT().GetCatalog(gomock.Any()).Return(getResource("get-catalog", nil).(pac.Catalog), nil).Times(2)
				mockKCClient.EXPECT().GetUserID().Return("122344").Times(2)
				mockDBClient.EXPECT().GetKeyByUserID(gomock.Any()).Return(getResource("get-key-by-userid", nil).([]models.Key), nil).Times(1)
				mockClient.EXPECT().GetServices(gomock.Any()).Return(getResource("get-all-services", nil).(pac.ServiceList), nil).Times(1)
			},
			service:    getResource("create-service", customValues{"ready": "false"}).(models.Service),
			httpStatus: http.StatusBadRequest,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			marshalledService, _ := json.Marshal(tc.service)
			req, err := http.NewRequest(http.MethodPost, "/services", bytes.NewBuffer(marshalledService))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			kubeClient = mockClient
			dbCon = mockDBClient
			CreateService(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestDeleteService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockClient, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()
	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		requestParams  gin.Param
		httpStatus     int
	}{
		{
			name: "delete service succesfully",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUserID().Return("12345").Times(1)
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(true).Times(1)
				mockClient.EXPECT().DeleteService(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBClient.EXPECT().GetRequestByServiceName(gomock.Any()).Return(getResource("get-request-by-service-name", nil).([]models.Request), nil).Times(1)
			},
			requestParams: gin.Param{Key: "name", Value: "test-service"},
			httpStatus:    http.StatusNoContent,
		},
		{
			name:          "service name not set",
			mockFunc:      func() {},
			requestParams: gin.Param{Key: "name", Value: ""},
			httpStatus:    http.StatusBadRequest,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodPost, "/services", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			kubeClient = mockClient
			dbCon = mockDBClient
			DeleteServiceHandler(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}
