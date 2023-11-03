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

var serviceExpiryMsg = "Service %s is expiring on %s. It will be deleted post-expiry, if not extended"

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
			logger.Debug("expiry notification not required for service yet", zap.Any("service", service))
			continue
		}

		if expiryNotificationSentRecently(service.Name, service.Expiry.String()) {
			logger.Debug("notification already sent", zap.Any("service", service))
			continue
		}
		logger.Debug("raising service-expiry-notification for service", zap.Any("service", service.Name))
		event, err := models.NewEvent(service.UserID, service.UserID, models.EventServiceExpiryNotification)
		if err != nil {
			logger.Error("failed to create event", zap.Error(err))
			return
		}
		event.SetLog(models.EventLogLevelINFO, fmt.Sprintf(serviceExpiryMsg, service.Name, service.Expiry.String()))
		if err := dbCon.NewEvent(event); err != nil {
			log.GetLogger().Error("failed to create event", zap.Error(err))
			return
		}
	}
}

func expiryNotificationSentRecently(serviceName string, serviceExpiry string) bool {
	logger := log.GetLogger()
	events, _, err := dbCon.GetEventsByType(models.EventServiceExpiryNotification, pastDay)
	// Return false if unable to fetch events from db which will result in expiry notification getting sent
	if err != nil {
		logger.Error("failed to get service-about-expire events", zap.Error(err))
		return false
	}
	for _, event := range events {
		if event.Log.Message != fmt.Sprintf(serviceExpiryMsg, serviceName, serviceExpiry) {
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

		for {
			select {
			case <-ticker.C:
				raiseNotification()
			}
		}
	}()
}
