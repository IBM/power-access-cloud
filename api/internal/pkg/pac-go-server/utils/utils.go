package utils

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetFloatValue(val intstr.IntOrString) (float64, error) {
	var result float64
	var err error
	switch val.Type {
	case intstr.Int:
		result = float64(val.IntVal)
	case intstr.String:
		result, err = strconv.ParseFloat(val.StrVal, 64)
		if err != nil {
			return result, fmt.Errorf("failed to convert val %s to float64", val.StrVal)
		}
	}
	return result, nil
}
