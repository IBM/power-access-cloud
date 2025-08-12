package utils

import "errors"

var (
	errInvalidCapacity       = errors.New("minimum supported values for CPU and memory capacity on PowerVS is 0.25C and 2GB respectively")
	errInvalidCPUMultiple    = errors.New("the CPU cores that can be provisoned on PowerVC is multiples of 0.25")
	ErrResourceNotFound      = errors.New("requested resource not found")
	ErrResourceAlreadyExists = errors.New("requested resource already exists")
	ErrNotAuthorized         = errors.New("user does not have permission to delete this key")
)
