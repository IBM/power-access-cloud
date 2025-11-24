package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
)

func TestGetAllKeys(t *testing.T) {
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
			name: "fetched all keys successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(true).Times(1)
				mockDBClient.EXPECT().GetKeyByUserID(gomock.Any()).Return(getResource("get-key-by-userid", nil), nil).Times(1)
			},
			httpStatus: http.StatusOK,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/keys", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			dbCon = mockDBClient
			GetAllKeysHandler(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestGetKey(t *testing.T) {
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
			name: "fetched all keys successfully",
			mockFunc: func() {
				mockDBClient.EXPECT().GetKeyByID(gomock.Any()).Return(getResource("get-key-by-id", nil), nil).Times(1)
			},
			httpStatus: http.StatusOK,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/keys", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			dbCon = mockDBClient
			GetKey(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestCreateKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, _, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		key            *models.Key
		requestParams  gin.Param
	}{
		{
			name: "key created successfully",
			mockFunc: func() {
				mockDBClient.EXPECT().GetKeyByUserID(gomock.Any()).Return(getResource("get-key-by-userid", nil).([]models.Key), nil).Times(1)
				mockDBClient.EXPECT().CreateKey(gomock.Any()).Return(nil).Times(1)
			},
			httpStatus:    http.StatusCreated,
			key:           getResource("get-key-by-id", nil).(*models.Key),
			requestParams: gin.Param{Key: "userid", Value: "12345"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			marshalledKey, _ := json.Marshal(tc.key)
			req, err := http.NewRequest(http.MethodPost, "/keys", bytes.NewBuffer(marshalledKey))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			dbCon = mockDBClient
			CreateKey(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestDeleteKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		key            *models.Key
		requestParams  gin.Param
	}{
		{
			name: "key deleted",
			mockFunc: func() {
				mockDBClient.EXPECT().GetKeyByID(gomock.Any()).Return(getResource("get-key-by-id", nil), nil).Times(1)
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(true).Times(1)
				mockDBClient.EXPECT().DeleteKey(gomock.Any()).Return(nil).Times(1)
			},
			httpStatus: http.StatusNoContent,
		},
		{
			name: "not authorized to delete key",
			mockFunc: func() {
				mockDBClient.EXPECT().GetKeyByID(gomock.Any()).Return(getResource("get-key-by-id", nil), nil).Times(1)
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(false).Times(1)
			},
			httpStatus: http.StatusUnauthorized,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/keys", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			dbCon = mockDBClient
			DeleteKeyHandler(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}
