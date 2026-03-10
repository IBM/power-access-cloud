package service

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"
	appv1alpha1 "github.com/IBM/power-access-cloud/api/apis/app/v1alpha1"
	"github.com/IBM/power-access-cloud/api/controllers/app/scope"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
)

const (
	publicNetworkPrefix = "pac-public-network"
)

//go:embed templates/userdata.yaml
var userDataTemplate string

var (
	ErroNoPublicNetwork = errors.New("no public network available to use for vm creation")
	dnsServers          = []string{"9.9.9.9", "1.1.1.1"}
	IbmiUsername        = "PAC"
	IbmiOS              = "ibmi"
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
		handleActiveStatus(scope)
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

// handleActiveStatus processes VM in ACTIVE state with IBMi initialization delay logic
func handleActiveStatus(scope *scope.ServiceScope) {
	// If service was already successfully created, keep it that way (don't revert to IN_PROGRESS)
	if scope.Service.Status.Successful {
		scope.Service.Status.State = appv1alpha1.ServiceStateCreated
		scope.Service.Status.AccessInfo = appv1alpha1.VMAccessInfoTemplate(
			scope.Service.Status.VM.ExternalIPAddress,
			scope.Service.Status.VM.IPAddress)
		scope.Service.Status.Message = ""
		return
	}

	// Track when VM first became ACTIVE (set once, never overwrite)
	if scope.Service.Status.VM.CreatedAt == nil {
		now := metav1.Now()
		scope.Service.Status.VM.CreatedAt = &now
		scope.Logger.Info("VM became ACTIVE, tracking creation time",
			"instanceID", scope.Service.Status.VM.InstanceID,
			"createdAt", now)
	} else {
		scope.Logger.Info("VM CreatedAt already set, preserving timestamp",
			"instanceID", scope.Service.Status.VM.InstanceID,
			"createdAt", scope.Service.Status.VM.CreatedAt)
	}

	// Check if this is IBMi and if we need to wait 40 minutes (only for new deployments)
	if isIBMiOS(scope) {
		elapsed := time.Since(scope.Service.Status.VM.CreatedAt.Time)
		waitTime := 40 * time.Minute

		if elapsed < waitTime {
			// Keep in IN_PROGRESS state - controller will requeue every 2 minutes
			scope.Service.Status.State = appv1alpha1.ServiceStateInProgress
			remaining := waitTime - elapsed
			scope.Service.Status.Message = fmt.Sprintf(
				"IBMi instance is initializing. Time remaining: %d minutes (elapsed: %d minutes)",
				int(remaining.Minutes()), int(elapsed.Minutes()))
			scope.Logger.Info("IBMi still initializing",
				"elapsed", elapsed, "remaining", remaining)
			return // Don't mark as CREATED yet
		}

		scope.Logger.Info("IBMi initialization complete", "elapsed", elapsed)
	}

	// Mark as created after wait period (or immediately for non-IBMi)
	scope.Service.Status.SetSuccessful()
	scope.Service.Status.State = appv1alpha1.ServiceStateCreated
	scope.Service.Status.AccessInfo = appv1alpha1.VMAccessInfoTemplate(
		scope.Service.Status.VM.ExternalIPAddress,
		scope.Service.Status.VM.IPAddress)
	scope.Service.Status.Message = ""
	scope.Logger.Info("Service marked as CREATED")
}

func isIBMiOS(scope *scope.ServiceScope) bool {
	return strings.ToLower(scope.Catalog.Spec.VM.OS) == IbmiOS
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

	userData := getUserData(vmSpec.OS, scope)
	createOpts := &models.PVMInstanceCreate{
		ServerName: &scope.Service.Name,
		ImageID:    imageRef.ImageID,
		NetworkIDs: []string{networkID},
		Memory:     &memory,
		Processors: &processors,
		SysType:    vmSpec.SystemType,
		ProcType:   &vmSpec.ProcessorType,
		UserData:   userData,
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

func getUserData(OS string, scope *scope.ServiceScope) string {
	var userData string
	switch OS {
	case IbmiOS:
		userData = getIBMiUserData(scope.Service.Spec.SSHKeys, IbmiUsername, scope)
	default:
		userData = base64.StdEncoding.EncodeToString(
			[]byte(strings.Join(scope.Service.Spec.SSHKeys, "\n")),
		)
	}
	return userData
}

func getIBMiUserData(sshKeys []string, username string, scope *scope.ServiceScope) string {
	allKeys := strings.Join(sshKeys, "\\n")
	cloudInitTemplate, err := template.New("userdata").Parse(userDataTemplate)
	if err != nil {
		scope.Logger.Error(err, "error parsing cloud-init template")
		return ""
	}
	var buf bytes.Buffer
	err = cloudInitTemplate.Execute(&buf, struct {
		Username string
		Keys     string
	}{
		Username: username,
		Keys:     allKeys,
	})
	if err != nil {
		scope.Logger.Error(err, "error executing cloud-init template")
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
