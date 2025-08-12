package services

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetEvents(t *testing.T) {
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
			name: "get events successfully",
			mockFunc: func() {
				mockKCClient.EXPECT().IsRole(gomock.Any()).Return(true).Times(1)
				mockDBClient.EXPECT().GetEventsByUserID(gomock.Any(), gomock.Any(), gomock.Any()).Return(getResource("get-events-by-userid", nil).([]models.Event), int64(1), nil).Times(1)
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
			req, err := http.NewRequest(http.MethodGet, "/events", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			dbCon = mockDBClient
			GetEvents(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}
