package services

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, _, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
	}{
		{
			name: "fetched all users",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUsers().Return(getResource("get-all-users", nil), nil).Times(1)
			},
			httpStatus: http.StatusOK,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/users", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			GetUsers(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestGetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, _, mockKCClient, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		requestParams  gin.Param
	}{
		{
			name: "fetched user successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUsers().Return(getResource("get-all-users", nil), nil).Times(1)
			},
			httpStatus:    http.StatusOK,
			requestParams: gin.Param{Key: "id", Value: "12345"},
		},
		{
			name: "user not found",
			mockFunc: func() {
				mockKCClient.EXPECT().GetUsers().Return(getResource("get-all-users", nil), nil).Times(1)
			},
			httpStatus:    http.StatusNotFound,
			requestParams: gin.Param{Key: "id", Value: "1235"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/users", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			GetUser(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}
