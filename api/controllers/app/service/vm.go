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
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/utils"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	ctrl "sigs.k8s.io/controller-runtime"
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

func (s *VM) Reconcile(ctx context.Context) (ctrl.Result, error) {
	if s.scope.Service.Status.VM.InstanceID == "" {
		if err := createVM(s.scope); err != nil {
			err := s.scope.NotifyServiceCreationFailure(err.Error())
			s.scope.Logger.Error(err, "Error notifying VM creation failure")
			return ctrl.Result{}, errors.Wrap(err, "error creating vm")
		}
	}

	pvmInstance, err := s.scope.PowerVSClient.GetVM(s.scope.Service.Status.VM.InstanceID)
	if err != nil {
		// Check if this is a transient provisioning error (e.g., volume attachment in progress)
		if utils.IsVolumeAttachementInProcessError(err) {
			s.scope.Logger.Info("Error during volume attachment, will retry in 10s", "instanceID", s.scope.Service.Status.VM.InstanceID)
			// Set state to IN_PROGRESS and return custom requeue interval
			s.scope.Service.Status.State = appv1alpha1.ServiceStateInProgress
			s.scope.Service.Status.Message = "VM provisioning in progress with volume attachment"
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
		return ctrl.Result{}, errors.Wrap(err, "error getting vm")
	}

	updateStatus(s.scope, pvmInstance)

	return ctrl.Result{}, nil
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

func updateStatus(scope *scope.ServiceScope, pvmInstance *models.PVMInstance) {
	extractPVMInstance(scope, pvmInstance)

	switch *pvmInstance.Status {
	case "ACTIVE":
		handleActiveStatus(scope)
	case "ERROR":
		scope.Service.Status.State = appv1alpha1.ServiceStateFailed
		serverName := "unknown"
		if pvmInstance.ServerName != nil {
			serverName = *pvmInstance.ServerName
		}
		instanceID := "unknown"
		if pvmInstance.PvmInstanceID != nil {
			instanceID = *pvmInstance.PvmInstanceID
		}

		errorMsg := fmt.Sprintf(`VM Creation Failed

Service Details:
- Server Name: %s
- Instance ID: %s`, serverName, instanceID)
		if pvmInstance.Fault != nil {
			errorMsg += fmt.Sprintf(`

Error Details:
%s`, pvmInstance.Fault.Message)
			scope.Service.Status.Message = fmt.Sprintf("VM creation failed with reason: %s", pvmInstance.Fault.Message)
		}
		scope.Service.Status.AccessInfo = ""
		if err := scope.NotifyServiceCreationFailure(errorMsg); err != nil {
			scope.Logger.Error(err, "failed to create failure notification event")
		}
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
	scope.ClearNotificationCache()
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

	// Create volumes from catalog spec if specified
	volumeIDs, err := createVolumesFromCatalog(scope, vmSpec, imageRef)
	if err != nil {
		return errors.Wrap(err, "failed to create volumes")
	}

	createOpts := &models.PVMInstanceCreate{
		ServerName: &scope.Service.Name,
		ImageID:    imageRef.ImageID,
		NetworkIDs: []string{networkID},
		Memory:     &memory,
		Processors: &processors,
		SysType:    vmSpec.SystemType,
		ProcType:   &vmSpec.ProcessorType,
		UserData:   userData,
		VolumeIDs:  volumeIDs,
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

// createVolumes creates volumes based on the provided volume specifications
// Returns a list of created volume IDs
// imageStoragePool is the storage pool where the image (and boot volume) resides
func createVolumes(scope *scope.ServiceScope, volumeSpecs []appv1alpha1.VolumeSpec, imageStoragePool string) ([]string, error) {
	if len(volumeSpecs) == 0 {
		return nil, nil
	}

	var createdVolumeIDs []string

	// Setup cleanup on error
	defer func() {
		if r := recover(); r != nil {
			cleanupVolumes(scope, createdVolumeIDs)
			panic(r)
		}
	}()

	for _, volSpec := range volumeSpecs {
		volumeName := fmt.Sprintf("%s-%s", scope.Service.Name, volSpec.VolumeNameSuffix)

		volumeID, err := createVolume(scope, volSpec, volumeName, imageStoragePool)
		if err != nil {
			cleanupVolumes(scope, createdVolumeIDs)
			return nil, errors.Wrapf(err, "failed to create volume %s", volumeName)
		}

		createdVolumeIDs = append(createdVolumeIDs, volumeID)
	}

	scope.Logger.Info("all volumes created successfully", "count", len(createdVolumeIDs), "volumeIDs", createdVolumeIDs)
	return createdVolumeIDs, nil
}

// createSingleVolume creates a single volume
func createVolume(scope *scope.ServiceScope, volSpec appv1alpha1.VolumeSpec, volumeName, imageStoragePool string) (string, error) {
	// Build volume parameters
	sizeFloat := float64(volSpec.Size)
	diskType := volSpec.DiskType
	createParams := &models.CreateDataVolume{
		Name:     &volumeName,
		Size:     &sizeFloat,
		DiskType: diskType,
	}

	// Use image storage pool for volume affinity
	if imageStoragePool != "" {
		createParams.VolumePool = imageStoragePool
		scope.Logger.Info("creating volume in image storage pool", "name", volumeName, "size", volSpec.Size, "diskType", diskType, "storagePool", imageStoragePool)
	} else {
		scope.Logger.Info("creating volume without storage pool", "name", volumeName, "size", volSpec.Size, "diskType", diskType)
	}

	// Create volume - fail if it already exists
	volume, err := scope.PowerVSClient.CreateVolume(createParams)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create volume %s", volumeName)
	}

	scope.Logger.Info("volume created successfully", "volumeID", *volume.VolumeID, "name", volumeName)
	return *volume.VolumeID, nil
}

func cleanupVM(scope *scope.ServiceScope) error {
	scope.Logger.Info("deleting VM with all attached volumes", "instanceID", scope.Service.Status.VM.InstanceID)

	// Delete VM and all its attached data volumes in one API call
	err := scope.PowerVSClient.DeleteVMWithVolumes(scope.Service.Status.VM.InstanceID)

	if err != nil && !utils.IsNotFoundError(err) {
		return errors.Wrap(err, "error deleting VM with volumes")
	}

	if err == nil {
		scope.Logger.Info("VM and all attached volumes deleted successfully", "instanceID", scope.Service.Status.VM.InstanceID)
	} else {
		scope.Logger.Info("VM already deleted or not found", "instanceID", scope.Service.Status.VM.InstanceID)
	}

	return nil
}

// cleanupVolumes deletes the specified volumes
func cleanupVolumes(scope *scope.ServiceScope, volumeIDs []string) {
	if len(volumeIDs) == 0 {
		return
	}
	scope.Logger.Info("cleaning up volumes due to error", "count", len(volumeIDs))
	for _, volID := range volumeIDs {
		if err := scope.PowerVSClient.DeleteVolume(volID); err != nil {
			scope.Logger.Error(err, "failed to cleanup volume", "volumeID", volID)
		}
	}
}

// createVolumesFromCatalog creates volumes from catalog spec using image's storage pool
func createVolumesFromCatalog(scope *scope.ServiceScope, vmCatalog appv1alpha1.VMCatalog, imageRef *models.ImageReference) ([]string, error) {
	if len(vmCatalog.Volumes) == 0 {
		return nil, nil
	}

	// Get image storage pool for volume affinity
	imageStoragePool := ""
	if imageRef.StoragePool != nil {
		imageStoragePool = *imageRef.StoragePool
	}

	scope.Logger.Info("creating volumes from catalog", "count", len(vmCatalog.Volumes), "storagePool", imageStoragePool)
	return createVolumes(scope, vmCatalog.Volumes, imageStoragePool)
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
