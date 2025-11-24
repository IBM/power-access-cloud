package utils_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/utils"
)

// TestRequest is the test model to validate BindAndValidate with other tags.
type TestRequest struct {
	Name   string `json:"name" binding:"required"`
	Length string `json:"length" binding:"gte=5,lte=10"`
	Age    int    `json:"age" binding:"gte=18,lte=60"`
}

func performRequest(body string, model any) (*httptest.ResponseRecorder, error) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	err := utils.BindAndValidate(c, model)
	return w, err
}

func TestBindAndValidate_AllCases(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)

	tests := []struct {
		name           string
		body           string
		model          any
		expectOK       bool
		expectHTTPCode int
		expectField    string
		expectError    error
	}{
		{
			name:           "success",
			body:           `{"id":"507f1f77bcf86cd799439011","user_id":"user123","rating":"positive","comment":"great","created_at":"` + now + `"}`,
			model:          &models.Feedback{},
			expectOK:       true,
			expectHTTPCode: 200,
		},
		{
			name:           "comment too long (max=250)",
			body:           `{"id":"507f1f77bcf86cd799439011","user_id":"user123","rating":"neutral","comment":"` + strings.Repeat("a", 300) + `","created_at":"` + now + `"}`,
			model:          &models.Feedback{},
			expectOK:       false,
			expectHTTPCode: 400,
			expectField:    "Comment",
			expectError:    errors.New("Comment must not exceed 250 characters"),
		},
		{
			name:           "rating type mismatch (string expected)",
			body:           `{"id":"507f1f77bcf86cd799439011","user_id":"user123","rating":123,"comment":"ok","created_at":"` + now + `"}`,
			model:          &models.Feedback{},
			expectOK:       false,
			expectHTTPCode: 400,
			expectField:    "rating",
			expectError:    errors.New("Field 'rating' must be string"),
		},
		{
			name:           "created_at type mismatch",
			body:           `{"id":"507f1f77bcf86cd799439011","user_id":"user123","rating":"positive","comment":"ok","created_at":12345}`,
			model:          &models.Feedback{},
			expectOK:       false,
			expectHTTPCode: 400,
			expectError:    errors.New("invalid request body"),
		},
		{
			name:           "malformed JSON",
			body:           `{"id":"5507f1f77bcf86cd799439011","user_id":"user123","rating":"positive","comment":"ok","created_at":"` + now, // broken JSON
			model:          &models.Feedback{},
			expectOK:       false,
			expectHTTPCode: 400,
			expectError:    errors.New("invalid request body"),
		},
		{
			name:           "missing required field",
			body:           `{"length":"abcdef","age":25}`,
			model:          &TestRequest{},
			expectOK:       false,
			expectHTTPCode: 400,
			expectField:    "Name",
			expectError:    errors.New("Name is required"),
		},
		{
			name:           "length too short (gte violation)",
			body:           `{"name":"foo","length":"abc","age":25}`,
			model:          &TestRequest{},
			expectOK:       false,
			expectHTTPCode: 400,
			expectField:    "Length",
			expectError:    errors.New("Length must be greater than or equal to 5"),
		},
		{
			name:           "length too long (lte violation)",
			body:           `{"name":"foo","length":"abcdefghijk","age":25}`,
			model:          &TestRequest{},
			expectOK:       false,
			expectHTTPCode: 400,
			expectField:    "Length",
			expectError:    errors.New("Length must be less than or equal to 10"),
		},
		{
			name:           "age below minimum",
			body:           `{"name":"foo","length":"abcdef","age":15}`,
			model:          &TestRequest{},
			expectOK:       false,
			expectHTTPCode: 400,
			expectField:    "Age",
			expectError:    errors.New("Age must be greater than or equal to 18"),
		},
		{
			name:           "age above maximum",
			body:           `{"name":"foo","length":"abcdef","age":70}`,
			model:          &TestRequest{},
			expectOK:       false,
			expectHTTPCode: 400,
			expectField:    "Age",
			expectError:    errors.New("Age must be less than or equal to 60"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := performRequest(tt.body, tt.model)

			if tt.expectOK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectHTTPCode, w.Code)
				return
			}

			assert.Error(t, err)
			assert.Equal(t, tt.expectHTTPCode, w.Code)

			var resp map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &resp)

			if tt.expectField != "" {
				assert.Equal(t, tt.expectField, resp["field"])
			}
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError.Error(), resp["error"])
			}
		})
	}
}
