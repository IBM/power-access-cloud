package services

import (
	"fmt"

	// "net/http"

	"time"

	"go.uber.org/zap"

	pac "github.com/PDeXchange/pac/apis/app/v1alpha1"
	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
)

const (
	everyDay = 24
	pastDay  = 24
)

var (
	serviceExpiryMsg  = "Service %s is expiring on %s. It will be deleted post-expiry, if not extended. Kindly ensure that all necessary data is downloaded prior to expiry, as all associated data will be permanently deleted upon service-deletion. Please refer this link for service extension: https://github.com/PDeXchange/pac-support/blob/main/docs/FAQ.md"
	requestExpiryMsg  = "Service is expired, hence request is no longer needed"
	serviceExpiredMsg = "Service %s is expired. It is going to be deleted."
)

func raiseNotification() {
	logger := log.GetLogger()
	var services pac.ServiceList
	var err error

	logger.Debug("raising service-expiry-notification if required")
	services, err = kubeClient.GetServices("")
	if err != nil {
		logger.Error("failed to get services", zap.Error(err))
		return
	}

	serviceItems := convertToServices(services)
	logger.Debug("fetched services", zap.Any("services", serviceItems))

	for _, service := range serviceItems {
		fewDaysAgo := service.Expiry.Add(time.Duration(-(models.ExpiryNotificationDuration)) * time.Hour)

		if !fewDaysAgo.Before(time.Now()) {
			logger.Debug("expiry notification not required for service", zap.Any("service", service.Name))
			continue
		}
		if isServiceExpired(service) {
			logger.Debug("service is expired, expiry notification is not required", zap.Any("service", service.Name))
			updateExpiryRequestinDB(service.Name)
			generateEvent(service, models.EventServiceExpiredNotification, fmt.Sprintf(serviceExpiredMsg, service.Name))
			continue
		}
		generateEvent(service, models.EventServiceExpiryNotification, fmt.Sprintf(serviceExpiryMsg, service.Name, service.Expiry.String()))
	}
}

func generateEvent(service models.Service, eventType models.EventType, eventLog string) {
	logger := log.GetLogger()
	logger.Debug("raising notification for service", zap.Any("service", service.Name), zap.Any("notification", eventType), zap.Any("eventLog", eventLog))
	if isNotificationSentRecently(service.Name, eventType, eventLog) {
		logger.Debug("notification already sent", zap.Any("service", service.Name), zap.Any("notification", eventType))
		return
	}
	event, err := models.NewEvent(service.UserID, service.UserID, eventType)
	if err != nil {
		logger.Error("failed to create event", zap.Error(err))
		return
	}
	event.SetLog(models.EventLogLevelINFO, eventLog)
	if err := dbCon.NewEvent(event); err != nil {
		log.GetLogger().Error("failed to create event", zap.Error(err))
		return
	}
}

func isServiceExpired(service models.Service) bool {
	logger := log.GetLogger()
	now := time.Now()
	if now.After(service.Expiry) {
		logger.Error("service expired", zap.String("service name", service.Name))
		return true
	}
	return false
}

func updateExpiryRequestinDB(serviceName string) {
	logger := log.GetLogger()
	req, err := dbCon.GetRequestByServiceName(serviceName)
	if err != nil {
		logger.Error("failed to fetch the request", zap.String("service name", serviceName), zap.Error(err))
		return
	}
	// Updating state of service-expiry-extension request to "EXPIRED" since associated service is expired
	logger.Debug("fetched request", zap.Any("request", req))
	for _, request := range req {
		if request.RequestType == models.RequestExtendServiceExpiry && request.State != models.RequestStateExpired {
			if err := dbCon.UpdateRequestStateWithComment(request.ID.String(), models.RequestStateExpired, requestExpiryMsg); err != nil {
				logger.Error("failed to update request in database", zap.String("id", request.ID.String()), zap.Error(err))
			}
		}

	}
}

func isNotificationSentRecently(serviceName string, notificationType models.EventType, eventLog string) bool {
	logger := log.GetLogger()
	logger.Debug("checking if notification sent recently", zap.Any("service", serviceName), zap.Any("notification", notificationType))
	events, _, err := dbCon.GetEventsByType(notificationType, pastDay)
	// Return false if unable to fetch events from db which will result in expiry notification getting sent
	if err != nil {
		logger.Error("failed to get service-about-expire events", zap.Error(err))
		return false
	}
	for _, event := range events {
		if event.Log.Message != eventLog {
			continue
		}
		return true
	}
	return false
}

// ExpiryNotification raises notification for about-to-expire services
func ExpiryNotification() {
	go func() {
		every := time.Duration(everyDay) * time.Hour
		ticker := time.NewTicker(every)
		defer ticker.Stop()

		for range ticker.C {
			raiseNotification()
		}
	}()
}
