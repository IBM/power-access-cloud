package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/pkg/errors"
	appv1alpha1 "github.com/IBM/power-access-cloud/api/apis/app/v1alpha1"
	"github.com/IBM/power-access-cloud/api/controllers/app/scope"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
)

const (
	publicNetworkPrefix = "pac-public-network"
)

var (
	ErroNoPublicNetwork = errors.New("no public network available to use for vm creation")
	dnsServers          = []string{"9.9.9.9", "1.1.1.1"}
)

func getAvailablePubNetwork(scope *scope.ServiceScope) (string, error) {
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

	return "", ErroNoPublicNetwork
}

func generateNetworkName() string {
	return fmt.Sprintf("%s-%s", publicNetworkPrefix, utilrand.String(5))
}

var _ Interface = &VM{}

type VM struct {
	scope *scope.ServiceScope
}

func NewVM(scope *scope.ServiceScope) Interface {
	return &VM{
		scope: scope,
	}
}

func (s *VM) Reconcile(ctx context.Context) error {
	if s.scope.Service.Status.VM.InstanceID == "" {
		if err := createVM(s.scope); err != nil {
			return errors.Wrap(err, "error creating vm")
		}
	}

	pvmInstance, err := s.scope.PowerVSClient.GetVM(s.scope.Service.Status.VM.InstanceID)
	if err != nil {
		return errors.Wrap(err, "error get vm")
	}

	updateStatus(s.scope, pvmInstance)

	return nil
}

func (s *VM) Delete(ctx context.Context) (bool, error) {
	if s.scope.Service.Status.VM.InstanceID == "" {
		s.scope.Logger.Info("vm instanceID is empty, nothing to clean up")
		return true, nil
	}

	if err := cleanupVM(s.scope); err != nil {
		return false, errors.Wrap(err, "error cleaning up vm")
	}
	s.scope.Service.Status.ClearVMStatus()

	return true, nil
}

func cleanupVM(scope *scope.ServiceScope) error {
	return scope.PowerVSClient.DeleteVM(scope.Service.Status.VM.InstanceID)
}

func updateStatus(scope *scope.ServiceScope, pvmInstance *models.PVMInstance) {
	extractPVMInstance(scope, pvmInstance)

	switch *pvmInstance.Status {
	case "ACTIVE":
		scope.Service.Status.SetSuccessful()
		scope.Service.Status.State = appv1alpha1.ServiceStateCreated
		scope.Service.Status.AccessInfo = appv1alpha1.VMAccessInfoTemplate(scope.Service.Status.VM.ExternalIPAddress, scope.Service.Status.VM.IPAddress)
		scope.Service.Status.Message = ""
	case "ERROR":
		scope.Service.Status.State = appv1alpha1.ServiceStateFailed
		if pvmInstance.Fault != nil {
			scope.Service.Status.Message = fmt.Sprintf("vm creation failed with reason: %s", pvmInstance.Fault.Message)
		}
		scope.Service.Status.AccessInfo = ""
	default:
		scope.Service.Status.State = appv1alpha1.ServiceStateInProgress
		scope.Service.Status.Message = "vm creation started, will update the access info once vm is ready"
	}
}

func extractPVMInstance(scope *scope.ServiceScope, pvmInstance *models.PVMInstance) {
	scope.Service.Status.VM.InstanceID = *pvmInstance.PvmInstanceID
	for _, nw := range pvmInstance.Networks {
		scope.Service.Status.VM.ExternalIPAddress = nw.ExternalIP
		scope.Service.Status.VM.IPAddress = nw.IPAddress
	}
	scope.Service.Status.VM.State = *pvmInstance.Status
}

func createVM(scope *scope.ServiceScope) error {
	// check if vm already exists and return if it does
	instances, err := scope.PowerVSClient.GetAllInstance()
	if err != nil {
		return err
	}

	for _, instance := range instances.PvmInstances {
		if *instance.ServerName == scope.Service.ObjectMeta.Name {
			scope.Logger.Info("vm already exists, hence skipping the vm creation", "name", scope.Service.ObjectMeta.Name)
			scope.Service.Status.VM.InstanceID = *instance.PvmInstanceID
			return nil
		}
	}

	vmSpec := scope.Catalog.Spec.VM
	var networkID string
	if vmSpec.Network == "" {
		var err error
		networkID, err = getAvailablePubNetwork(scope)
		if err != nil && err != ErroNoPublicNetwork {
			return errors.Wrap(err, "error retrieving available public network in powervs instance")
		} else if err == ErroNoPublicNetwork {
			// create a public network and use it
			network, err := scope.PowerVSClient.CreateNetwork(&models.NetworkCreate{
				Name:       generateNetworkName(),
				Type:       core.StringPtr("pub-vlan"),
				DNSServers: dnsServers,
			})
			if err != nil {
				return errors.Wrap(err, "error creating public network")
			}
			networkID = *network.NetworkID
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
	scope.Service.Status.Message = "vm creation started, will update the access info once vm is ready"

	if len(*pvmInstanceList) != 1 {
		return errors.New("error creating vm, expected 1 vm to be created")
	}
	scope.Service.Status.VM.InstanceID = *(*pvmInstanceList)[0].PvmInstanceID
	return nil
}
