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
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/IBM-Cloud/power-go-client/power/models"
	appv1alpha1 "github.com/PDeXchange/pac/apis/app/v1alpha1"
	"github.com/pkg/errors"
)

func extractPVMInstance(scope *ControllerScope, pvmInstance *models.PVMInstance) {
	scope.Service.Status.VM.InstanceID = *pvmInstance.PvmInstanceID
	for _, nw := range pvmInstance.Networks {
		scope.Service.Status.VM.ExternalIPAddress = nw.ExternalIP
		scope.Service.Status.VM.IPAddress = nw.IPAddress
	}
	scope.Service.Status.VM.State = *pvmInstance.Status
}

func getAvailablePubNetwork(scope *ControllerScope) (string, error) {
	networks, err := scope.PowerVSClient.GetNetworks()
	if err != nil {
		return "", errors.Wrap(err, "error get all networks")
	}

	for _, nw := range networks.Networks {
		if *nw.Type == "pub-vlan" {
			network, err := scope.PowerVSClient.GetNetwork(*nw.NetworkID)
			if err != nil {
				return "", errors.Wrapf(err, "error get network with id %s", *nw.NetworkID)
			}

			if *network.IPAddressMetrics.Available > 0 {
				return *network.NetworkID, nil
			}
		}
	}

	return "", errors.New("no public network available to use for vm creation")
}

func createVM(scope *ControllerScope) error {
	vmSpec := scope.Catalog.Spec.VM

	var networkID string
	if vmSpec.Network == "" {
		var err error
		networkID, err = getAvailablePubNetwork(scope)
		if err != nil {
			return errors.Wrap(err, "error retrieving available public network in powervs instance")
		}
	} else {
		nwRef, err := scope.PowerVSClient.GetNetworkByName(vmSpec.Network)
		if err != nil {
			return errors.Wrapf(err, "error retrieving network by name %s", vmSpec.Network)
		}
		networkID = *nwRef.NetworkID
	}

	imageRef, err := scope.PowerVSClient.GetImageByName(vmSpec.Image)
	if err != nil {
		return errors.Wrapf(err, "error retrieving image by name %s", vmSpec.Image)
	}

	memory := float64(vmSpec.Capacity.Memory)
	processors, _ := strconv.ParseFloat(vmSpec.Capacity.CPU, 64)
	createOpts := &models.PVMInstanceCreate{
		ServerName: &scope.Service.Name,
		ImageID:    imageRef.ImageID,
		NetworkIDs: []string{networkID},
		Memory:     &memory,
		Processors: &processors,
		SysType:    vmSpec.SystemType,
		ProcType:   &vmSpec.ProcessorType,
		UserData:   base64.StdEncoding.EncodeToString([]byte(strings.Join(scope.Service.Spec.SSHKeys, "\n"))),
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
		scope.Service.Status.AccessInfo = appv1alpha1.VMAccessInfoTemplate(scope.Service.Status.VM.ExternalIPAddress, scope.Service.Status.VM.IPAddress)
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
	if scope.Service.Status.VM.InstanceID == "" {
		scope.Logger.Info("vm instanceID is empty, nothing to clean up")
		return true, nil
	}

	if err := cleanupVM(scope); err != nil {
		return false, errors.Wrap(err, "error cleaning up vm")
	}
	scope.Service.Status.AccessInfo = ""

	return true, nil
}
