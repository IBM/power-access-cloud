package services

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
)

func TestGetAllGroups(t *testing.T) {
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
			name: "get all groups successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).Times(1)
				mockDBClient.EXPECT().GetGroupsQuota(gomock.Any()).Return(getResource("get-groups-quota", nil), nil).Times(1)
			},
			httpStatus: http.StatusOK,
		},
		{
			name: "failed to get groups",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(nil, errors.New("failed to get groups")).Times(1)
			},
			httpStatus: http.StatusInternalServerError,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/group", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			dbCon = mockDBClient
			GetAllGroups(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestGetGroup(t *testing.T) {
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
			name: "group fetched successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).Times(1)
				mockDBClient.EXPECT().GetQuotaForGroupID(gomock.Any()).Return(getResource("get-quota-by-groupid", nil).(*models.Quota), nil).Times(1)
			},
			requestParams: gin.Param{Key: "id", Value: "test-group"},
			httpStatus:    http.StatusOK,
		},
		{
			name: "group not found",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(getResource("get-group-info", nil).([]*gocloak.Group), nil).Times(1)
			},
			requestParams: gin.Param{Key: "id", Value: "test-group-1"},
			httpStatus:    http.StatusNotFound,
		},
		{
			name: "failed to get groups",
			mockFunc: func() {
				mockKCClient.EXPECT().GetGroups().Return(nil, errors.New("failed to get groups")).Times(1)
			},
			httpStatus: http.StatusBadRequest,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/group", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			dbCon = mockDBClient
			GetGroup(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}
