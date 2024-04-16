package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	pac "github.com/PDeXchange/pac/apis/app/v1alpha1"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateCatalog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockClient, mockDBClient, tearDown := setUp(t)
	defer tearDown()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	testcases := []struct {
		name           string
		mockFunc       func()
		requestContext testContext
		catalog        models.Catalog
		httpStatus     int
	}{
		{
			name: "valid catalog",
			mockFunc: func() {
				mockClient.EXPECT().CreateCatalog(gomock.Any()).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Times(1)
			},
			requestContext: formContext(customValues{"userid": "12345"}),
			catalog:        getResource("create-catalog", nil).(models.Catalog),
			httpStatus:     http.StatusCreated,
		},
		{
			name:           "invalid image thumbnail in catalog",
			mockFunc:       func() {},
			requestContext: formContext(customValues{"userid": "12345"}),
			catalog:        getResource("create-catalog", customValues{"ImageThumbnailReference": "thumbnail"}).(models.Catalog),
			httpStatus:     http.StatusBadRequest,
		},
		{
			name: "catalog already exists",
			mockFunc: func() {
				mockClient.EXPECT().CreateCatalog(gomock.Any()).Return(errors.New("catalog already exists")).Times(1)
			},
			requestContext: formContext(customValues{"userid": "12345"}),
			catalog:        getResource("create-catalog", nil).(models.Catalog),
			httpStatus:     http.StatusBadRequest,
		},
		{
			name: "failed to create event",
			mockFunc: func() {
				mockClient.EXPECT().CreateCatalog(gomock.Any()).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Return(errors.New("failed to create event")).Times(1)
			},
			requestContext: formContext(customValues{"userid": "12345"}),
			catalog:        getResource("create-catalog", nil).(models.Catalog),
			httpStatus:     http.StatusBadRequest,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			marshalledCatalog, _ := json.Marshal(tc.catalog)
			req, err := http.NewRequest(http.MethodPost, "/catalog", bytes.NewBuffer(marshalledCatalog))
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			kubeClient = mockClient
			dbCon = mockDBClient
			CreateCatalog(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestGetAllCatalog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockClient, _, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name       string
		mockFunc   func()
		httpStatus int
	}{
		{
			name: "get all catalogs succesfully",
			mockFunc: func() {
				mockClient.EXPECT().GetCatalogs().Return(getResource("get-all-catalogs", nil).(pac.CatalogList), nil).Times(1)
			},
			httpStatus: http.StatusOK,
		},
		{
			name: "failed to get catalogs",
			mockFunc: func() {
				mockClient.EXPECT().GetCatalogs().Return(getResource("get-all-catalogs", nil).(pac.CatalogList), errors.New("failed to get catalog")).Times(1)
			},
			httpStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			kubeClient = mockClient
			GetAllCatalogs(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestGetCatalog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockClient, _, tearDown := setUp(t)
	defer tearDown()

	testcases := []struct {
		name          string
		mockFunc      func()
		requestParams gin.Param
		httpStatus    int
	}{
		{
			name: "get catalog successfully",
			mockFunc: func() {
				mockClient.EXPECT().GetCatalog(gomock.Any()).Return(getResource("get-catalog", nil).(pac.Catalog), nil).Times(1)
			},
			requestParams: gin.Param{Key: "name", Value: "test-catalog"},
			httpStatus:    http.StatusOK,
		},
		{
			name:          "catalog name not set",
			mockFunc:      func() {},
			requestParams: gin.Param{Key: "name", Value: ""},
			httpStatus:    http.StatusBadRequest,
		},
		{
			name: "failed to get catalog",
			mockFunc: func() {
				mockClient.EXPECT().GetCatalog(gomock.Any()).Return(getResource("get-catalog", nil).(pac.Catalog), errors.New("failed to get catalog")).Times(1)
			},
			requestParams: gin.Param{Key: "name", Value: "test-catalog"},
			httpStatus:    http.StatusBadRequest,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			kubeClient = mockClient
			c.Params = gin.Params{tc.requestParams}
			GetCatalog(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestDeleteCatalog(t *testing.T) {
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
			name: "delete catalog successfully",
			mockFunc: func() {
				mockClient.EXPECT().DeleteCatalog(gomock.Any()).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Times(1)
			},
			requestContext: formContext(customValues{"userid": "12345"}),
			requestParams:  gin.Param{Key: "name", Value: "test-catalog"},
			httpStatus:     http.StatusNoContent,
		},
		{
			name:           "catalog name not set",
			mockFunc:       func() {},
			requestContext: formContext(customValues{"userid": "12345"}),
			requestParams:  gin.Param{Key: "name", Value: ""},
			httpStatus:     http.StatusBadRequest,
		},
		{
			name: "failed to delete catalog",
			mockFunc: func() {
				mockClient.EXPECT().DeleteCatalog(gomock.Any()).Return(errors.New("failed to delete catalog")).Times(1)
			},
			requestContext: formContext(customValues{"userid": "12345"}),
			requestParams:  gin.Param{Key: "name", Value: "test-catalog"},
			httpStatus:     http.StatusBadRequest,
		},
		{
			name: "failed to create event",
			mockFunc: func() {
				mockClient.EXPECT().DeleteCatalog(gomock.Any()).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Return(errors.New("failed to create event")).Times(1)
			},
			requestContext: formContext(customValues{"userid": "12345"}),
			requestParams:  gin.Param{Key: "name", Value: "test-catalog"},
			httpStatus:     http.StatusNoContent,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodPost, "/catalog", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			kubeClient = mockClient
			dbCon = mockDBClient
			DeleteCatalog(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}

func TestRetireCatalog(t *testing.T) {
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
			name: "retire catalog successfully",
			mockFunc: func() {
				mockClient.EXPECT().RetireCatalog(gomock.Any()).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Times(1)
			},
			requestContext: formContext(customValues{"userid": "12345"}),
			requestParams:  gin.Param{Key: "name", Value: "test-catalog"},
			httpStatus:     http.StatusNoContent,
		},
		{
			name:           "catalog name not set",
			mockFunc:       func() {},
			requestContext: formContext(customValues{"userid": "12345"}),
			requestParams:  gin.Param{Key: "name", Value: ""},
			httpStatus:     http.StatusBadRequest,
		},
		{
			name: "failed to retire catalog",
			mockFunc: func() {
				mockClient.EXPECT().RetireCatalog(gomock.Any()).Return(errors.New("failed to retire catalog")).Times(1)
			},
			requestContext: formContext(customValues{"userid": "12345"}),
			requestParams:  gin.Param{Key: "name", Value: "test-catalog"},
			httpStatus:     http.StatusBadRequest,
		},
		{
			name: "failed to create event",
			mockFunc: func() {
				mockClient.EXPECT().RetireCatalog(gomock.Any()).Times(1)
				mockDBClient.EXPECT().NewEvent(gomock.Any()).Return(errors.New("failed to create event")).Times(1)
			},
			requestContext: formContext(customValues{"userid": "12345"}),
			requestParams:  gin.Param{Key: "name", Value: "test-catalog"},
			httpStatus:     http.StatusNoContent,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := http.NewRequest(http.MethodPost, "/catalog", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := getContext(tc.requestContext)
			c.Request = req.WithContext(ctx)
			c.Params = gin.Params{tc.requestParams}
			kubeClient = mockClient
			dbCon = mockDBClient
			RetireCatalog(c)
			assert.Equal(t, tc.httpStatus, c.Writer.Status())
		})
	}
}
