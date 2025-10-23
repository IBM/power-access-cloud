package utils

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

var (
	errInvalidCapacity       = errors.New("minimum supported values for CPU and memory capacity on PowerVS is 0.25C and 2GB respectively")
	errInvalidCPUMultiple    = errors.New("the CPU cores that can be provisoned on PowerVC is multiples of 0.25")
	ErrResourceNotFound      = errors.New("requested resource not found")
	ErrResourceAlreadyExists = errors.New("requested resource already exists")
	ErrNotAuthorized         = errors.New("user does not have permission to delete this key")
)

func translateValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", fe.Field(), fe.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

func getExpectedType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.String:
		return "string"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "value"
	}
}
