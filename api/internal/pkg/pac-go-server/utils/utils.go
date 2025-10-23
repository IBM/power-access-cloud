package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	ManagerRole = "manager"

	// DefaultExpirationDays - default expiration days set for Catalog
	DefaultExpirationDays = 5
)

func CastStrToFloat(val string) (float64, error) {
	return strconv.ParseFloat(val, 64)
}

func CastFloatToStr(val float64) string {
	return fmt.Sprintf("%.2f", val)
}

// ValidateQuotaFields : Check if the data provided by admin are appropriate.
// The minimum possible values for CPU and memory for PowerVS instance is 0.25C and 2GB respectively.
func ValidateQuotaFields(c *gin.Context, cpuCap float64, memCap int) error {
	if cpuCap < 0.25 || memCap < 2 {
		return errInvalidCapacity
	}
	if int(cpuCap*100)%25 != 0 {
		return errInvalidCPUMultiple
	}
	return nil
}

func Ptr[T any](v T) *T {
	return &v
}

// BindAndValidate is used to bind unmarshal request and return custom errors if there are any.
func BindAndValidate(c *gin.Context, obj any) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			// validation errors (missing fields, range, etc.)
			fe := ve[0]
			c.JSON(http.StatusBadRequest, gin.H{
				"error": translateValidationError(fe),
				"field": fe.Field(),
			})
			return err
		}

		// handle JSON type mismatch
		if ute, ok := err.(*json.UnmarshalTypeError); ok {
			expectedType := getExpectedType(ute.Type)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Field '%s' must be %s", ute.Field, expectedType),
				"field": ute.Field,
			})
			return err
		}

		// fallback (malformed JSON, syntax error, etc.)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return err
	}
	return nil
}
