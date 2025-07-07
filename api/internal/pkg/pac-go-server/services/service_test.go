package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pac "github.com/PDeXchange/pac/apis/app/v1alpha1"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetAllServices(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockClient, _, tearDown := setUp(t)
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
			},
			requestContext: formContext(customValues{
				"userid":                "12345",
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

func TestCreateService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockClient, mockDBClient, tearDown := setUp(t)
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
				mockDBClient.EXPECT().GetKeyByUserID(gomock.Any()).Return(getResource("get-key-by-userid", nil).([]models.Key), nil).Times(1)
				mockDBClient.EXPECT().GetGroupsQuota(gomock.Any()).Return(getResource("get-groups-quota", nil).([]models.Quota), nil).Times(1)
			},
			service: getResource("create-service", nil).(models.Service),
			requestContext: formContext(customValues{
				"userid":                "122343",
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
	mockClient, mockDBClient, tearDown := setUp(t)
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
				mockClient.EXPECT().DeleteService(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBClient.EXPECT().GetRequestByServiceName(gomock.Any()).Return(getResource("get-request-by-service-name", nil).([]models.Request), nil).Times(1)
			},
			requestContext: formContext(customValues{
				"userid": "12345",
				"roles":  []string{"manager"},
			}),
			requestParams: gin.Param{Key: "name", Value: "test-service"},
			httpStatus:    http.StatusNoContent,
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
