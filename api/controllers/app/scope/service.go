package scope

import (
	"context"
	"os"
	"strconv"
	"time"

	"fmt"
	"sync"

	"github.com/IBM/power-access-cloud/api/apis/app/v1alpha1"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/db"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"github.com/pkg/errors"
	"sigs.k8s.io/cluster-api/util/patch"
)

var (
	notificationCache  = make(map[string]time.Time)
	cacheMutex         sync.RWMutex
	minIntervalMinutes = 30
	dbCon              db.DB
)

func SetDB(db db.DB) {
	dbCon = db
}

type ServiceScopeParams struct {
	ControllerScopeParams
	Service *v1alpha1.Service
}

type ServiceScope struct {
	ControllerScope
	servicePatchHelper *patch.Helper
	Service            *v1alpha1.Service
}

func (s *ServiceScope) IsExpired() bool {
	currentTime := time.Now()
	return currentTime.After(s.Service.Spec.Expiry.Time)
}

// PatchObject persists the catalog/service configuration and status.
func (m *ServiceScope) PatchServiceObject() error {
	return m.servicePatchHelper.Patch(context.TODO(), m.Service)
}

// NotifyServiceCreationFailure creates an event to notify about service creation failure
func (s *ServiceScope) NotifyServiceCreationFailure(errorMessage string) error {

	var notificationKey string
	if s.Service.Status.VM.InstanceID == "" {
		// No instance ID - use event type as key
		notificationKey = string(models.EventServiceCreateFailed)
		if !shouldNotifyforErrorCode(models.EventServiceCreateFailed) {
			return nil
		}
	} else {
		// Instance ID exists - use it as key
		notificationKey = s.Service.Status.VM.InstanceID
		if !shouldNotifyforInstance(s.Service.Status.VM.InstanceID) {
			return nil
		}
	}
	event, err := models.NewEvent(s.Service.Spec.UserID, s.Service.Spec.UserID, models.EventServiceCreateFailed)
	if err != nil {
		return err
	}

	// Notify only admin
	event.SetNotifyAdmin()

	// Set the error message
	logMessage := fmt.Sprintf("Service '%s' creation failed. Reason: %s", s.Service.Name, errorMessage)
	event.SetLog(models.EventLogLevelERROR, logMessage)

	err = dbCon.ConnectionExists(true)
	if err != nil {
		return err
	}

	err = dbCon.NewEvent(event)
	if err != nil {
		return err
	}
	recordNotification(s.Service.Status.VM.InstanceID)
	s.Logger.Info("Created failure notification event", "service", s.Service.Name, "notificationKey", notificationKey)
	return nil
}



func (s *ServiceScope) ClearNotificationCache() {
	if s.Service.Status.VM.InstanceID != "" {
		clearNotification(s.Service.Status.VM.InstanceID)
		s.Logger.Info("Cleared notification cache for VM instance",
			"instanceID", s.Service.Status.VM.InstanceID,
			"service", s.Service.Name)
	}
}

func NewServiceScope(ctx context.Context, params ServiceScopeParams) (*ServiceScope, error) {
	scope := &ServiceScope{}

	ctrlScope, err := NewControllerScope(ctx, params.ControllerScopeParams)
	if err != nil {
		err = errors.Wrap(err, "failed to init controller scope")
		return scope, err
	}
	scope.ControllerScope = *ctrlScope

	if params.Service == nil {
		err = errors.New("service is required when creating a ServiceScope")
		return scope, err
	}
	scope.Service = params.Service

	serviceHelper, err := patch.NewHelper(params.Service, params.Client)
	if err != nil {
		err = errors.Wrap(err, "failed to init patch helper")
		return scope, err
	}
	scope.servicePatchHelper = serviceHelper

	return scope, nil
}

// shouldNotify returns true if a notification can be sent for the service.
// Implements rate limiting to prevent duplicate notifications within minIntervalMinutes.
func shouldNotifyforInstance(instanceID string) bool {
	if !ShouldNotify() {
		return false
	}
	cacheMutex.RLock()
	lastNotified, exists := notificationCache[instanceID]
	cacheMutex.RUnlock()

	if !exists {
		return true
	}

	return time.Since(lastNotified) > time.Duration(minIntervalMinutes)*time.Minute
}

func shouldNotifyforErrorCode(event models.EventType) bool{
	if !ShouldNotify() {
		return false
	}

	cacheMutex.RLock()
	lastNotified, exists := notificationCache[string(event)]
	cacheMutex.RUnlock()

	if !exists {
		return true
	}
	return time.Since(lastNotified) > time.Duration(minIntervalMinutes)*time.Minute
}

func ShouldNotify() bool {
	notify, err := strconv.ParseBool(os.Getenv("NOTIFY_VM_CREATION_FAILURE"))
	if err != nil {
		return false
	}
	if !notify {
		return false
	}
	return true
}

// recordNotification records the current time as the last notification timestamp for the service.
// Used by rate limiting to track when notifications were sent.
func recordNotification(key string) {
	cacheMutex.Lock()
	notificationCache[key] = time.Now()
	cacheMutex.Unlock()
}

// clearNotification removes the service from the notification cache.
func clearNotification(instanceID string) {
	cacheMutex.Lock()
	delete(notificationCache, instanceID)
	cacheMutex.Unlock()
}
