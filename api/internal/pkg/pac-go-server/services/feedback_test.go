package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateFeedbackForUserId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, _, tearDown := setUp(t)
	defer tearDown()
	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		httpStatus     int
		feedback       *models.Feedback
	}{
		{
			name: "created feedback successfully",
			mockFunc: func() {
				mockDBClient.EXPECT().InsertFeedback(gomock.Any()).Return(nil).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Times(1)
			},
			requestContext: testContext{
				userID: "user1",
			},
			httpStatus: http.StatusCreated,
			feedback:   getResource("create-feedback", nil).(*models.Feedback),
		},
		{
			name:     "create feedback with invalid rating",
			mockFunc: func() {},
			requestContext: testContext{
				userID: "user1",
			},
			httpStatus: http.StatusBadRequest,
			feedback:   getResource("create-feedback", customValues{"Rating": models.Rating("bad")}).(*models.Feedback),
		},
		{
			name:     "create feedback with invalid comment",
			mockFunc: func() {},
			requestContext: testContext{
				userID: "user1",
			},
			httpStatus: http.StatusBadRequest,
			feedback:   getResource("create-feedback", customValues{"Comment": getComment()}).(*models.Feedback),
		},
		{
			name: "create feedback with InternalServerError",
			mockFunc: func() {
				mockDBClient.EXPECT().InsertFeedback(gomock.Any()).Return(errors.New("")).Times(1)
			},
			requestContext: testContext{
				userID: "user1",
			},
			httpStatus: http.StatusInternalServerError,
			feedback:   getResource("create-feedback", nil).(*models.Feedback),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			marshalledFeedback, _ := json.Marshal(tc.feedback)
			req, err := http.NewRequest(http.MethodPost, "/feedbacks", bytes.NewBuffer(marshalledFeedback))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			dbCon = mockDBClient
			CreateFeedback(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())

		})
	}
}
