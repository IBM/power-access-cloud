package main

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/Nerzal/gocloak/v13"
	mailclient "github.com/PDeXchange/pac/internal/pkg/notifier/client"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/client"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/db"
	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
)

func notifier(db db.DB, mailClient mailclient.Notifier) {
	l := log.GetLogger()
	eventCh := make(chan *models.Event)
	go func() {
		err := db.WatchEvents(eventCh)
		if err != nil {
			l.Fatal("error watching events", zap.Error(err))
		}
	}()

	for event := range eventCh {
		var err error
		l.Info("Received an event:", zap.Any("Document", event))
		// skip if already notified
		if event.Notified {
			l.Info("Already notified")
			continue
		}
		l.Info("Notifying")
		if event.UserEmail, err = getEmailForEvent(event); err != nil {
			l.Error("Error retrieving user information", zap.Error(err))
		}
		if err = mailClient.Notify(*event); err != nil {
			l.Error("Error notifying", zap.Error(err))
		}
		if err = db.MarkEventAsNotified(event.ID.Hex()); err != nil {
			l.Error("Error updating event", zap.Error(err))
		}
	}
}

func getEmailForEvent(event *models.Event) (string, error) {
	kc := gocloak.NewClient(hostname)
	l := log.GetLogger()
	token, err := kc.LoginAdmin(context.Background(), serviceAccount, serviceAccountPassword, realm)
	if err != nil {
		l.Error("failed to get access token", zap.Error(err))
		return "", errors.New("ailed to get access token")
	}

	// Build Context
	//nolint:staticcheck
	ctx := context.WithValue(context.Background(), "keycloak_client", kc)
	//nolint:staticcheck
	ctx = context.WithValue(ctx, "keycloak_realm", realm)
	//nolint:staticcheck
	ctx = context.WithValue(ctx, "keycloak_hostname", hostname)
	//nolint:staticcheck
	ctx = context.WithValue(ctx, "keycloak_access_token", token.AccessToken)

	config := client.GetConfigFromContext(ctx)
	user, err := client.NewKeyCloakClient(config, ctx).GetUser(event.UserID)
	if err != nil {
		return "", err
	}
	return *user.Email, nil

}
