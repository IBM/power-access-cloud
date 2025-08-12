/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"regexp"
	"strconv"

	"github.com/pkg/errors"

	appv1alpha1 "github.com/PDeXchange/pac/apis/app/v1alpha1"
)

var (
	IBMResourceCRNRegexp = regexp.MustCompile(`^crn:v[0-9]:(?P<cloudName>[^:]*):(?P<cloudType>[^:]*):(?P<serviceName>[^:]*):(?P<location>[^:]*):a/(?P<account>[^:]*):(?P<guid>[^:]*)::$`)
	availableSysType     = []string{"s922", "e980"}
	availableProcType    = []string{"dedicated", "shared", "capped"}
)

// ParsePowerVSCRN parses powervs crn to extract guid, zone and account information
func ParsePowerVSCRN(crn string) (string, string, string, error) {
	matches := IBMResourceCRNRegexp.FindStringSubmatch(crn)
	if matches == nil {
		return "", "", "", errors.New("could not parse crn with generic crn regex")
	}

	return matches[IBMResourceCRNRegexp.SubexpIndex("guid")], matches[IBMResourceCRNRegexp.SubexpIndex("location")], matches[IBMResourceCRNRegexp.SubexpIndex("account")], nil
}

func ValidateSysType(sysType string) error {
	for _, st := range availableSysType {
		if st == sysType {
			return nil
		}
	}
	return errors.Errorf("sys type %s is not supported", sysType)
}

func ValidateProcType(procType string) error {
	for _, pt := range availableProcType {
		if pt == procType {
			return nil
		}
	}
	return errors.Errorf("processor type %s is not supported", procType)
}

func ValidateVMCapacity(catalogCapacity *appv1alpha1.Capacity, vmCapacity *appv1alpha1.Capacity) error {
	if vmCapacity.CPU == "" {
		vmCapacity.CPU = catalogCapacity.CPU
	} else {
		catalogCPUCapacity, err := strconv.ParseFloat(catalogCapacity.CPU, 32)
		if err != nil {
			return errors.Wrap(err, "error parsing catalog cpu capacity")
		}

		vmCPUCapacity, err := strconv.ParseFloat(vmCapacity.CPU, 32)
		if err != nil {
			return errors.Wrap(err, "error parsing vm cpu capacity")
		}

		if vmCPUCapacity > catalogCPUCapacity {
			return errors.Errorf("vm cpu capacity should not exceed catalog cpu capacity. catalog cpu capacity: %f, vm cpu capacity: %f", catalogCPUCapacity, vmCPUCapacity)
		}
	}

	if vmCapacity.Memory == 0 {
		vmCapacity.Memory = catalogCapacity.Memory
	} else if vmCapacity.Memory > catalogCapacity.Memory {
		return errors.Errorf("vm memory capacity should not exceed catalog memory capacity. catalog memory capacity: %d, vm memory capacity: %d", catalogCapacity.Memory, vmCapacity.Memory)
	}

	return nil
}
