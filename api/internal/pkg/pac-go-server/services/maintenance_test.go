package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetMaintenanceWindows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, _, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name       string
		mockFunc   func()
		httpStatus int
	}{
		{
			name: "fetched all maintenance windows successfully",
			mockFunc: func() {
				mockDBClient.EXPECT().GetAllMaintenanceWindows().Return(getResource("get-all-maintenance-windows", nil).([]models.MaintenanceWindow), nil).Times(1)
			},
			httpStatus: http.StatusOK,
		},
		{
			name: "empty maintenance windows list",
			mockFunc: func() {
				mockDBClient.EXPECT().GetAllMaintenanceWindows().Return([]models.MaintenanceWindow{}, nil).Times(1)
			},
			httpStatus: http.StatusOK,
		},
		{
			name: "database error",
			mockFunc: func() {
				mockDBClient.EXPECT().GetAllMaintenanceWindows().Return(nil, errors.New("database error")).Times(1)
			},
			httpStatus: http.StatusInternalServerError,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/api/v1/maintenance", nil)
			if err != nil {
				t.Fatal(err)
			}
			c.Request = req
			dbCon = mockDBClient
			GetMaintenanceWindows(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestCreateMaintenanceWindow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		request        *models.MaintenanceCreateRequest
	}{
		{
			name: "maintenance window created successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUserID().Return("admin-user-123").Times(1)
				mockDBClient.EXPECT().GetAllMaintenanceWindows().Return([]models.MaintenanceWindow{}, nil).Times(1) // No overlaps
				mockDBClient.EXPECT().CreateMaintenanceWindow(gomock.Any()).Return(nil).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Return(nil).Times(1)
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus: http.StatusCreated,
			request:    getResource("create-maintenance-window-request", nil).(*models.MaintenanceCreateRequest),
		},
		{
			name: "invalid request body",
			mockFunc: func() {
				// No mock expectations as validation fails before GetUserID call
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus: http.StatusBadRequest,
			request:    &models.MaintenanceCreateRequest{}, // Empty request
		},
		{
			name: "end time before start time",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUserID().Return("admin-user-123").Times(1)
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus: http.StatusBadRequest,
			request:    getResource("create-maintenance-window-invalid-time", nil).(*models.MaintenanceCreateRequest),
		},
		{
			name: "overlapping maintenance window",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUserID().Return("admin-user-123").Times(1)
				mockDBClient.EXPECT().GetAllMaintenanceWindows().Return(getResource("get-all-maintenance-windows", nil).([]models.MaintenanceWindow), nil).Times(1)
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus: http.StatusBadRequest,
			request:    getResource("create-maintenance-window-request", nil).(*models.MaintenanceCreateRequest),
		},
		{
			name: "database error",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUserID().Return("admin-user-123").Times(1)
				mockDBClient.EXPECT().GetAllMaintenanceWindows().Return([]models.MaintenanceWindow{}, nil).Times(1)
				mockDBClient.EXPECT().CreateMaintenanceWindow(gomock.Any()).Return(errors.New("database error")).Times(1)
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus: http.StatusInternalServerError,
			request:    getResource("create-maintenance-window-request", nil).(*models.MaintenanceCreateRequest),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			marshalledRequest, _ := json.Marshal(tc.request)
			req, err := http.NewRequest(http.MethodPost, "/api/v1/maintenance", bytes.NewBuffer(marshalledRequest))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			dbCon = mockDBClient
			CreateMaintenanceWindow(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestUpdateMaintenanceWindow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		request        *models.MaintenanceUpdateRequest
		requestParams  gin.Param
	}{
		{
			name: "maintenance window updated successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUserID().Return("admin-user-123").Times(1)
				mockDBClient.EXPECT().GetMaintenanceWindowByID(gomock.Any()).Return(getResource("get-maintenance-window-by-id", nil).(*models.MaintenanceWindow), nil).Times(1)
				mockDBClient.EXPECT().GetAllMaintenanceWindows().Return([]models.MaintenanceWindow{}, nil).Times(1) // No overlaps
				mockDBClient.EXPECT().UpdateMaintenanceWindow(gomock.Any()).Return(nil).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Return(nil).Times(1)
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus:    http.StatusOK,
			request:       getResource("update-maintenance-window-request", nil).(*models.MaintenanceUpdateRequest),
			requestParams: gin.Param{Key: "id", Value: "507f1f77bcf86cd799439011"},
		},
		{
			name: "maintenance window not found",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUserID().Return("admin-user-123").Times(1)
				mockDBClient.EXPECT().GetMaintenanceWindowByID(gomock.Any()).Return(nil, nil).Times(1)
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus:    http.StatusNotFound,
			request:       getResource("update-maintenance-window-request", nil).(*models.MaintenanceUpdateRequest),
			requestParams: gin.Param{Key: "id", Value: "507f1f77bcf86cd799439011"},
		},
		{
			name: "invalid request body",
			mockFunc: func() {
				// No mock expectations as validation fails before GetUserID call
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus:    http.StatusBadRequest,
			request:       nil, // Will cause JSON parsing error
			requestParams: gin.Param{Key: "id", Value: "507f1f77bcf86cd799439011"},
		},
		{
			name: "database error on get",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUserID().Return("admin-user-123").Times(1)
				mockDBClient.EXPECT().GetMaintenanceWindowByID(gomock.Any()).Return(nil, errors.New("database error")).Times(1)
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus:    http.StatusInternalServerError,
			request:       getResource("update-maintenance-window-request", nil).(*models.MaintenanceUpdateRequest),
			requestParams: gin.Param{Key: "id", Value: "507f1f77bcf86cd799439011"},
		},
		{
			name: "database error on update",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUserID().Return("admin-user-123").Times(1)
				mockDBClient.EXPECT().GetMaintenanceWindowByID(gomock.Any()).Return(getResource("get-maintenance-window-by-id", nil).(*models.MaintenanceWindow), nil).Times(1)
				mockDBClient.EXPECT().GetAllMaintenanceWindows().Return([]models.MaintenanceWindow{}, nil).Times(1)
				mockDBClient.EXPECT().UpdateMaintenanceWindow(gomock.Any()).Return(errors.New("database error")).Times(1)
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus:    http.StatusInternalServerError,
			request:       getResource("update-maintenance-window-request", nil).(*models.MaintenanceUpdateRequest),
			requestParams: gin.Param{Key: "id", Value: "507f1f77bcf86cd799439011"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			var marshalledRequest []byte
			if tc.request != nil {
				marshalledRequest, _ = json.Marshal(tc.request)
			} else {
				marshalledRequest = []byte("invalid json")
			}
			req, err := http.NewRequest(http.MethodPut, "/api/v1/maintenance/"+tc.requestParams.Value, bytes.NewBuffer(marshalledRequest))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			dbCon = mockDBClient
			UpdateMaintenanceWindow(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestDeleteMaintenanceWindow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		requestParams  gin.Param
	}{
		{
			name: "maintenance window deleted successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUserID().Return("admin-user-123").Times(1)
				mockDBClient.EXPECT().GetMaintenanceWindowByID(gomock.Any()).Return(getResource("get-maintenance-window-by-id", nil).(*models.MaintenanceWindow), nil).Times(1)
				mockDBClient.EXPECT().DeleteMaintenanceWindow(gomock.Any(), "admin-user-123", gomock.Any()).Return(nil).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Return(nil).Times(1)
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus:    http.StatusNoContent,
			requestParams: gin.Param{Key: "id", Value: "507f1f77bcf86cd799439011"},
		},
		{
			name: "maintenance window not found",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUserID().Return("admin-user-123").Times(1)
				mockDBClient.EXPECT().GetMaintenanceWindowByID(gomock.Any()).Return(nil, nil).Times(1)
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus:    http.StatusNotFound,
			requestParams: gin.Param{Key: "id", Value: "507f1f77bcf86cd799439011"},
		},
		{
			name: "database error on get",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUserID().Return("admin-user-123").Times(1)
				mockDBClient.EXPECT().GetMaintenanceWindowByID(gomock.Any()).Return(nil, errors.New("database error")).Times(1)
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus:    http.StatusInternalServerError,
			requestParams: gin.Param{Key: "id", Value: "507f1f77bcf86cd799439011"},
		},
		{
			name: "database error on delete",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUserID().Return("admin-user-123").Times(1)
				mockDBClient.EXPECT().GetMaintenanceWindowByID(gomock.Any()).Return(getResource("get-maintenance-window-by-id", nil).(*models.MaintenanceWindow), nil).Times(1)
				mockDBClient.EXPECT().DeleteMaintenanceWindow(gomock.Any(), "admin-user-123", gomock.Any()).Return(errors.New("database error")).Times(1)
			},
			requestContext: formContext(customValues{
				"keycloak_hostname":     "127.0.0.1",
				"keycloak_access_token": "Bearer test-token",
				"keycloak_realm":        "test-pac",
			}),
			httpStatus:    http.StatusInternalServerError,
			requestParams: gin.Param{Key: "id", Value: "507f1f77bcf86cd799439011"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodDelete, "/api/v1/maintenance/"+tc.requestParams.Value, nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			dbCon = mockDBClient
			DeleteMaintenanceWindow(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}
