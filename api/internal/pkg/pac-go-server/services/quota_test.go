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

func TestGetQuota(t *testing.T) {
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
			name: "quota fetched successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).AnyTimes()
				mockDBClient.EXPECT().GetQuotaForGroupID(gomock.Any()).Return(getResource("get-quota-by-groupid", nil).(*models.Quota), nil).Times(1)
			},
			httpStatus:    http.StatusOK,
			requestParams: gin.Param{Key: "id", Value: "test-group"},
		},
		{
			name:          "group not found",
			mockFunc:      func() {},
			httpStatus:    http.StatusNotFound,
			requestParams: gin.Param{Key: "id", Value: "test-group-2"},
		},
		{
			name: "quota policy does not exist for the group id",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).AnyTimes()
				mockDBClient.EXPECT().GetQuotaForGroupID(gomock.Any()).Return(nil, nil).Times(1)
			},
			httpStatus:    http.StatusNotFound,
			requestParams: gin.Param{Key: "id", Value: "test-group"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/quota", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			dbCon = mockDBClient
			GetQuota(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestCreateQuota(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		requestParams  gin.Param
		quota          *models.Quota
	}{
		{
			name: "created quota successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).AnyTimes()
				mockDBClient.EXPECT().GetQuotaForGroupID(gomock.Any()).Return(nil, nil).Times(1)
				mockDBClient.EXPECT().NewQuota(gomock.Any()).Return(nil).AnyTimes()
			},
			httpStatus:    http.StatusCreated,
			requestParams: gin.Param{Key: "id", Value: "test-group"},
			quota: getResource("create-quota", customValues{"Capacity": models.Capacity{
				CPU:    10,
				Memory: 10,
			}}).(*models.Quota),
		},
		{
			name: "user not part of group",
			mockFunc: func() {
				mockDBClient.EXPECT().GetQuotaForGroupID(gomock.Any()).Return(getResource("create-quota", nil).(*models.Quota), nil).Times(1)
			},
			httpStatus:    http.StatusConflict,
			requestParams: gin.Param{Key: "id", Value: "test-group"},
			quota: getResource("create-quota", customValues{"Capacity": models.Capacity{
				CPU:    10,
				Memory: 10,
			}}).(*models.Quota),
		},
		{
			name: "group not found",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).AnyTimes()
			},
			httpStatus:    http.StatusNotFound,
			requestParams: gin.Param{Key: "id", Value: "test-group-2"},
		},
		{
			name: "quota validation has failed",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).AnyTimes()
			},
			httpStatus:    http.StatusBadRequest,
			requestParams: gin.Param{Key: "id", Value: "test-group"},
			quota: getResource("create-quota", customValues{"Capacity": models.Capacity{
				CPU:    1,
				Memory: 1,
			}}).(*models.Quota),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			marshalledQuota, _ := json.Marshal(tc.quota)
			req, err := http.NewRequest(http.MethodGet, "/quota", bytes.NewBuffer(marshalledQuota))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			dbCon = mockDBClient
			CreateQuota(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestUpdateQuota(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		requestParams  gin.Param
		quota          *models.Quota
	}{
		{
			name: "updated quota successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).AnyTimes()
				mockDBClient.EXPECT().GetQuotaForGroupID(gomock.Any()).Return(getResource("create-quota", nil).(*models.Quota), nil).Times(1)
				mockDBClient.EXPECT().UpdateQuota(gomock.Any()).Return(nil).AnyTimes()
			},
			httpStatus:    http.StatusCreated,
			requestParams: gin.Param{Key: "id", Value: "test-group"},
			quota: getResource("create-quota", customValues{"Capacity": models.Capacity{
				CPU:    10,
				Memory: 10,
			}}).(*models.Quota),
		},
		{
			name: "no quota policy exists for the groupid",
			mockFunc: func() {
				mockDBClient.EXPECT().GetQuotaForGroupID(gomock.Any()).Return(nil, nil).Times(1)
			},
			httpStatus:    http.StatusConflict,
			requestParams: gin.Param{Key: "id", Value: "test-group"},
			quota: getResource("create-quota", customValues{"Capacity": models.Capacity{
				CPU:    10,
				Memory: 10,
			}}).(*models.Quota),
		},
		{
			name: "group not found",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).AnyTimes()
			},
			httpStatus:    http.StatusNotFound,
			requestParams: gin.Param{Key: "id", Value: "test-group-2"},
		},
		{
			name: "quota validation has failed",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).AnyTimes()
			},
			httpStatus:    http.StatusBadRequest,
			requestParams: gin.Param{Key: "id", Value: "test-group"},
			quota: getResource("create-quota", customValues{"Capacity": models.Capacity{
				CPU:    1,
				Memory: 1,
			}}).(*models.Quota),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			marshalledQuota, _ := json.Marshal(tc.quota)
			req, err := http.NewRequest(http.MethodGet, "/quota", bytes.NewBuffer(marshalledQuota))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			dbCon = mockDBClient
			UpdateQuota(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestDeleteQuota(t *testing.T) {
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
			name: "quota deleted successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).AnyTimes()
				mockDBClient.EXPECT().DeleteQuota(gomock.Any()).Return(nil).AnyTimes()
			},
			httpStatus:    http.StatusNoContent,
			requestParams: gin.Param{Key: "id", Value: "test-group"},
		},
		{
			name: "group id does not exists",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).AnyTimes()
			},
			httpStatus:    http.StatusNotFound,
			requestParams: gin.Param{Key: "id", Value: "test-group-1"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/quota", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			dbCon = mockDBClient
			// keyCloakClient = mockKCClient
			DeleteQuota(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestGetUserQuota(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockClient, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		requestParams  gin.Param
	}{
		{
			name: "user quota fetched successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).AnyTimes()
				mockClient.EXPECT().GetServices(gomock.Any()).Return(getResource("get-all-services", nil).(pac.ServiceList), nil).AnyTimes()
				mockClient.EXPECT().GetCatalog(gomock.Any()).Return(getResource("get-catalog", nil).(pac.Catalog), nil).Times(1)
				mockKCClient.EXPECT().GetUserID().Return("test-user").AnyTimes()
			},
			httpStatus:    http.StatusOK,
			requestParams: gin.Param{Key: "id", Value: "test-group"},
			requestContext: formContext(customValues{
				"userid": "test-user",
				"roles":  []string{"manager"},
			}),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/quota", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			kubeClient = mockClient
			dbCon = mockDBClient
			GetUserQuota(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}
