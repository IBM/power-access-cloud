package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetTermsAndConditionsStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, _, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name       string
		mockFunc   func()
		tncContext testContext
		httpStatus int
	}{
		{
			name: "status fetched successfully",
			mockFunc: func() {
				mockDBClient.EXPECT().GetTermsAndConditionsByUserID(gomock.Any()).Return(getResource("get-tnc-by-userid", nil).(*models.TermsAndConditions), nil).Times(1)
			},
			httpStatus: http.StatusOK,
			tncContext: formContext(customValues{
				"userid": "12345",
			}),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodGet, "/tnc", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.tncContext)
			c.Request = req.WithContext(ctx)
			dbCon = mockDBClient
			GetTermsAndConditionsStatus(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestAcceptTermsAndCondition(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDBClient, _, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name       string
		mockFunc   func()
		tncContext testContext
		tnc        *models.TermsAndConditions
		httpStatus int
	}{
		{
			name: "terms and condition accepted",
			mockFunc: func() {
				mockDBClient.EXPECT().GetTermsAndConditionsByUserID(gomock.Any()).Return(getResource("get-tnc-by-userid", nil).(*models.TermsAndConditions), nil).Times(1)
				mockDBClient.EXPECT().AcceptTermsAndConditions(gomock.Any()).Return(nil).Times(1)
			},
			httpStatus: http.StatusCreated,
			tncContext: formContext(customValues{
				"userid": "12345",
			}),
			tnc: getResource("get-tnc-by-userid", nil).(*models.TermsAndConditions),
		},
		{
			name: "terms and condition already accepted",
			mockFunc: func() {
				mockDBClient.EXPECT().GetTermsAndConditionsByUserID(gomock.Any()).Return(getResource("get-tnc-by-userid", customValues{"Accepted": true}).(*models.TermsAndConditions), nil).Times(1)
			},
			httpStatus: http.StatusBadRequest,
			tncContext: formContext(customValues{
				"userid": "12345",
			}),
			tnc: getResource("get-tnc-by-userid", customValues{"Accepted": true}).(*models.TermsAndConditions),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			marshalledRequest, _ := json.Marshal(tc.tnc)
			req, err := http.NewRequest(http.MethodPost, "/tnc", bytes.NewBuffer(marshalledRequest))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.tncContext)
			c.Request = req.WithContext(ctx)
			dbCon = mockDBClient
			AcceptTermsAndConditions(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}
