package utils

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	errInvalidCapacity    = errors.New("minimum supported values for CPU and memory capacity on PowerVS is 0.25C and 2GB respectively")
	errInvalidCPUMultiple = errors.New("the CPU cores that can be provisoned on PowerVC is multiples of 0.25")
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
