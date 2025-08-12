package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	pac "github.com/PDeXchange/pac/apis/app/v1alpha1"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetAllRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
	}{
		{
			name: "get all requests successfully",
			mockFunc: func() {
				mockDBClient.EXPECT().GetRequestsByUserID(gomock.Any(), gomock.Any()).Return(getResource("get-requests-by-user-id", nil).([]models.Request), nil).Times(1)
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(true).AnyTimes()
			},
			requestContext: formContext(customValues{
				"userid": "12345",
			}),
			httpStatus: http.StatusOK,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/requests", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			dbCon = mockDBClient
			GetAllRequests(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestGetRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, _, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
	}{
		{
			name: "get request successfully",
			mockFunc: func() {
				mockDBClient.EXPECT().GetRequestByID(gomock.Any()).Return(getResource("get-request-by-id", nil).(*models.Request), nil).Times(1)
			},
			httpStatus: http.StatusOK,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/requests", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			dbCon = mockDBClient
			GetRequest(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestUpdateServiceExpiryRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockClient, mockDBClient, _, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		request        *models.Request
		requestParams  gin.Param
	}{
		{
			name: "service expiry request successfull",
			mockFunc: func() {
				mockClient.EXPECT().GetService(gomock.Any()).Return(getResource("get-service", nil).(pac.Service), nil).Times(1)
				mockClient.EXPECT().GetCatalog(gomock.Any()).Return(getResource("get-catalog", nil).(pac.Catalog), nil).Times(1)
				mockDBClient.EXPECT().GetRequestByServiceName(gomock.Any()).Return(getResource("get-request-by-service-name", nil).([]models.Request), nil).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Times(1)
				mockDBClient.EXPECT().NewRequest(gomock.Any()).Return("123", nil).Times(1)
			},
			requestContext: formContext(customValues{
				"userid": "12345",
			}),
			httpStatus: http.StatusCreated,
			request:    getResource("get-request-by-id", nil).(*models.Request),
		},
		{
			name:          "justification not set",
			mockFunc:      func() {},
			requestParams: gin.Param{Key: "justification", Value: ""},
			requestContext: formContext(customValues{
				"userid": "12345",
			}),
			httpStatus: http.StatusBadRequest,
		},
		{
			name:          "expiry not set",
			mockFunc:      func() {},
			requestParams: gin.Param{Key: "expiry", Value: ""},
			requestContext: formContext(customValues{
				"userid": "12345",
			}),
			httpStatus: http.StatusBadRequest,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			marshalledRequest, _ := json.Marshal(tc.request)
			req, err := http.NewRequest(http.MethodPost, "/requests", bytes.NewBuffer(marshalledRequest))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			kubeClient = mockClient
			dbCon = mockDBClient
			UpdateServiceExpiryRequest(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}

}

func TestNewGroupRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		request        *models.Request
		requestParams  gin.Param
	}{
		{
			name: "group request created successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).AnyTimes()
				mockDBClient.EXPECT().GetRequestByGroupIDAndUserID(gomock.Any(), gomock.Any()).Return(getResource("get-requests-by-user-id", nil).([]models.Request), nil).AnyTimes()
				mockDBClient.EXPECT().NewRequest(gomock.Any()).Return("123", nil).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Times(1)
			},
			request:       getResource("get-request-by-id", nil).(*models.Request),
			httpStatus:    http.StatusCreated,
			requestParams: gin.Param{Key: "id", Value: "test-group"},
		},
		{
			name:          "group not found",
			mockFunc:      func() {},
			request:       getResource("get-request-by-id", nil).(*models.Request),
			httpStatus:    http.StatusNotFound,
			requestParams: gin.Param{Key: "id", Value: "test-group-2"},
		},
		{
			name:          "justification not set",
			mockFunc:      func() {},
			requestParams: gin.Param{Key: "justification", Value: ""},
			httpStatus:    http.StatusBadRequest,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			marshalledRequest, _ := json.Marshal(tc.request)
			req, err := http.NewRequest(http.MethodPost, "/requests", bytes.NewBuffer(marshalledRequest))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			dbCon = mockDBClient
			NewGroupRequest(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

// func TestExitGroup(t *testing.T) {
// 	gin.SetMode(gin.TestMode)
// 	_, mockDBClient, mockKCClient, tearDown := setUp(t)
// 	defer tearDown()

// 	testcases := []struct {
// 		name           string
// 		mockFunc       func()
// 		requestContext testContext
// 		httpStatus     int
// 		request        *models.Request
// 		requestParams  gin.Param
// 	}{
// 		{
// 			name: "successfully created request",
// 			mockFunc: func() {
// 				mockKCClient.EXPECT().AddUserToGroup("test-user", "test-group").Return(nil).AnyTimes()
// 				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).AnyTimes()
// 				mockDBClient.EXPECT().GetRequestByGroupIDAndUserID(gomock.Any(), gomock.Any()).Return(getResource("get-requests-by-user-id", nil).([]models.Request), nil).AnyTimes()
// 				mockDBClient.EXPECT().NewRequest(gomock.Any()).Return("123", nil).Times(1)
// 				mockDBClient.EXPECT().NewEvent(gomock.Any()).Times(1)
// 			},
// 			httpStatus:    http.StatusCreated,
// 			request:       getResource("add-to-group-request", nil).(*models.Request),
// 			requestParams: gin.Param{Key: "id", Value: "test-group"},
// 			requestContext: formContext(customValues{
// 				"username": "test-user",
// 			}),
// 		},
// 	}
// 	for _, tc := range testcases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			tc.mockFunc()
// 			c, _ := gin.CreateTestContext(httptest.NewRecorder())
// 			marshalledRequest, _ := json.Marshal(tc.request)
// 			req, err := http.NewRequest(http.MethodPost, "/requests", bytes.NewBuffer(marshalledRequest))
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			ctx := getContext(tc.requestContext)
// 			c.Request = req.WithContext(ctx)
// 			c.Params = gin.Params{tc.requestParams}
// 			dbCon = mockDBClient
// 			ExitGroup(c)
// 			assert.Equal(t, tc.httpStatus, c.Writer.Status())
// 		})
// 	}
// }

func TestDeleteUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockClient, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		request        *models.Request
	}{
		{
			name: "request to delete user created successfully",
			mockFunc: func() {
				mockClient.EXPECT().GetServices(gomock.Any()).Return(getResource("get-all-services", nil).(pac.ServiceList), nil).Times(1)
				mockClient.EXPECT().DeleteService(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBClient.EXPECT().GetRequestByServiceName(gomock.Any()).Return(getResource("get-request-by-service-name", nil).([]models.Request), nil).Times(1)
				mockDBClient.EXPECT().GetKeyByUserID(gomock.Any()).Return(getResource("get-key-by-userid", nil).([]models.Key), nil).Times(1)
				mockDBClient.EXPECT().GetKeyByID(gomock.Any()).Return(getResource("get-key-by-id", nil).(*models.Key), nil).Times(1)
				mockDBClient.EXPECT().DeleteKey(gomock.Any()).Return(nil).AnyTimes()
				mockDBClient.EXPECT().NewRequest(gomock.Any()).Return("123", nil).AnyTimes()
				mockDBClient.EXPECT().NewEvent(gomock.Any()).AnyTimes()
				mockKCClient.EXPECT().GetUserID().Return("12345").Times(2)
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(true).Times(2)
			},
			httpStatus: http.StatusCreated,
			request:    getResource("get-request-by-id", nil).(*models.Request),
			requestContext: formContext(customValues{
				"userid": "12345",
				"roles":  []string{"manager"},
			}),
		},
		{
			name: "not authorized to delete key",
			mockFunc: func() {
				mockClient.EXPECT().GetServices(gomock.Any()).Return(getResource("get-all-services", nil).(pac.ServiceList), nil).Times(1)
				mockClient.EXPECT().DeleteService(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBClient.EXPECT().GetRequestByServiceName(gomock.Any()).Return(getResource("get-request-by-service-name", nil).([]models.Request), nil).Times(1)
				mockDBClient.EXPECT().GetKeyByUserID(gomock.Any()).Return(getResource("get-key-by-userid", nil).([]models.Key), nil).Times(1)
				mockDBClient.EXPECT().GetKeyByID(gomock.Any()).Return(getResource("get-key-by-id", nil).(*models.Key), nil).Times(1)
				mockDBClient.EXPECT().DeleteKey(gomock.Any()).Return(nil).AnyTimes()
				mockKCClient.EXPECT().GetUserID().Return("12345").Times(3)
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(false).Times(3)
			},
			httpStatus: http.StatusInternalServerError,
			request:    getResource("get-request-by-id", nil).(*models.Request),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			marshalledRequest, _ := json.Marshal(tc.request)
			req, err := http.NewRequest(http.MethodPost, "/requests", bytes.NewBuffer(marshalledRequest))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			kubeClient = mockClient
			dbCon = mockDBClient
			DeleteUser(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestApproveRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockClient, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		request        *models.Request
	}{
		{
			name: "successfully approved request",
			mockFunc: func() {
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(true).Times(1)
				mockDBClient.EXPECT().GetRequestByID(gomock.Any()).Return(getResource("get-request-by-id", nil).(*models.Request), nil).Times(1)
				mockClient.EXPECT().UpdateServiceExpiry(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBClient.EXPECT().UpdateRequestState(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Times(1)
			},
			httpStatus: http.StatusNoContent,
			requestContext: formContext(customValues{
				"userid": "12345",
				"roles":  []string{"manager"},
			}),
		},
		{
			name: "not authorized to approve request",
			mockFunc: func() {
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(false).Times(1)
			},
			httpStatus: http.StatusUnauthorized,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			marshalledRequest, _ := json.Marshal(tc.request)
			req, err := http.NewRequest(http.MethodPost, "/requests", bytes.NewBuffer(marshalledRequest))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			kubeClient = mockClient
			dbCon = mockDBClient
			ApproveRequest(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestRejectRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		request        *models.Request
	}{
		{
			name: "successfully rejected request",
			mockFunc: func() {
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(true).Times(2)
				mockDBClient.EXPECT().GetRequestByID(gomock.Any()).Return(getResource("get-request-by-id", nil).(*models.Request), nil).Times(1)
				mockDBClient.EXPECT().UpdateRequestStateWithComment(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Times(1)
			},
			httpStatus: http.StatusNoContent,
			requestContext: formContext(customValues{
				"userid": "12345",
				"roles":  []string{"manager"},
			}),
			request: getResource("get-request-by-id", nil).(*models.Request),
		},
		{
			name:       "comment required to reject request",
			mockFunc:   func() {},
			httpStatus: http.StatusBadRequest,
			requestContext: formContext(customValues{
				"userid": "12345",
				"roles":  []string{"manager"},
			}),
			request: getResource("get-request-by-id", nil).(*models.Request),
		},
		{
			name: "not authorized to approve request",
			mockFunc: func() {
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(false).Times(1)
			},
			httpStatus: http.StatusUnauthorized,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			if tc.name == "comment required to reject request" {
				tc.request.Comment = ""
			}
			marshalledRequest, _ := json.Marshal(tc.request)
			req, err := http.NewRequest(http.MethodPost, "/requests", bytes.NewBuffer(marshalledRequest))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			dbCon = mockDBClient
			RejectRequest(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestDeleteRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		request        *models.Request
	}{
		{
			name: "successfully deleted request",
			mockFunc: func() {
				mockDBClient.EXPECT().GetRequestByID(gomock.Any()).Return(getResource("get-request-by-id", nil).(*models.Request), nil).Times(1)
				mockKCClient.EXPECT().GetUserID().Return("12345").Times(1)
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(true).Times(1)
				mockDBClient.EXPECT().DeleteRequest(gomock.Any()).Return(nil).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Times(1)
			},
			httpStatus: http.StatusNoContent,
		},
		{
			name: "not authorized to approve request",
			mockFunc: func() {
				mockDBClient.EXPECT().GetRequestByID(gomock.Any()).Return(getResource("get-request-by-id", nil).(*models.Request), nil).Times(1)
				mockKCClient.EXPECT().GetUserID().Return("1234").Times(1)
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(false).Times(1)
			},
			httpStatus: http.StatusUnauthorized,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			marshalledRequest, _ := json.Marshal(tc.request)
			req, err := http.NewRequest(http.MethodPost, "/requests", bytes.NewBuffer(marshalledRequest))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			dbCon = mockDBClient
			DeleteRequest(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}
