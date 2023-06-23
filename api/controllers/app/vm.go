/*
Copyright 2022.

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

package controllers

import (
	"fmt"
	"github.com/IBM-Cloud/power-go-client/power/models"
	appv1alpha1 "github.com/PDeXchange/pac/apis/app/v1alpha1"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

func extractPVMInstance(scope *ControllerScope, pvmInstance *models.PVMInstance) {
	scope.Service.Status.VM.InstanceID = *pvmInstance.PvmInstanceID
	for _, nw := range pvmInstance.Networks {
		scope.Service.Status.VM.IPAddress = nw.IPAddress
	}
	scope.Service.Status.VM.State = *pvmInstance.Status
}

func createVM(scope *ControllerScope) error {
	vmSpec := scope.Catalog.Spec.VM
	memory := float64(vmSpec.Capacity.Memory)
	processors, _ := strconv.ParseFloat(vmSpec.Capacity.CPU, 64)
	createOpts := &models.PVMInstanceCreate{
		ServerName: &scope.Service.Name,
		ImageID:    &vmSpec.Image,
		NetworkIDs: []string{vmSpec.Network},
		Memory:     &memory,
		Processors: &processors,
		SysType:    vmSpec.SystemType,
		ProcType:   &vmSpec.ProcessorType,
		UserData:   strings.Join(scope.Service.Spec.SSHKeys, "\n"),
	}

	pvmInstanceList, err := scope.PowerVSClient.CreateVM(createOpts)
	if err != nil {
		return err
	}

	for _, i := range *pvmInstanceList {
		extractPVMInstance(scope, i)
	}

	scope.Service.Status.State = appv1alpha1.ServiceStateInProgress
	return nil
}

func updateStatus(scope *ControllerScope, pvmInstance *models.PVMInstance) {
	extractPVMInstance(scope, pvmInstance)

	switch *pvmInstance.Status {
	case "ACTIVE":
		scope.Service.Status.State = appv1alpha1.ServiceStateCreated
		scope.Service.Status.AccessInfo = appv1alpha1.VMAccessInfoTemplate(scope.Service.Status.VM.IPAddress)
	case "ERROR":
		scope.Service.Status.State = appv1alpha1.ServiceStateFailed
		scope.Service.Status.AccessInfo = ""
	default:
		scope.Service.Status.State = appv1alpha1.ServiceStateInProgress
		scope.Service.Status.AccessInfo = ""
	}

	scope.Service.Status.Message = fmt.Sprintf("vm health info from powervs service from ibm cloud. status: %s, reason: %s", pvmInstance.Health.Status, pvmInstance.Health.Reason)
}

func isVMExpired(scope *ControllerScope, pvmInstance *models.PVMInstance) bool {
	currentTime := time.Now()
	if currentTime.After(scope.Service.Spec.Expiry.Time) {
		return true
	}

	return false
}

func cleanupVM(scope *ControllerScope) error {
	return scope.PowerVSClient.DeleteVM(scope.Service.Status.VM.InstanceID)
}

func ReconcileVM(scope *ControllerScope) error {
	if scope.Service.Status.VM.InstanceID == "" {
		if err := createVM(scope); err != nil {
			return errors.Wrap(err, "error creating vm")
		}
	}

	pvmInstance, err := scope.PowerVSClient.GetVM(scope.Service.Status.VM.InstanceID)
	if err != nil {
		return errors.Wrap(err, "error get vm")
	}

	updateStatus(scope, pvmInstance)

	if isVMExpired(scope, pvmInstance) {
		scope.Logger.Info("VM expired", "name", scope.Service.ObjectMeta.Name)
		scope.Service.Status.State = appv1alpha1.ServiceStateExpired
		scope.Service.Status.Expired = true

		if err = cleanupVM(scope); err != nil {
			return errors.Wrap(err, "error cleaning up vm")
		}

		scope.Service.Status.AccessInfo = ""
	}

	return nil
}

func ReconcileDeleteVM(scope *ControllerScope) (bool, error) {
	if err := cleanupVM(scope); err != nil {
		return false, errors.Wrap(err, "error cleaning up vm")
	}
	scope.Service.Status.AccessInfo = ""

	return true, nil
}
