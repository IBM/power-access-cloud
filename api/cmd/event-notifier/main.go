package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/PDeXchange/pac/internal/pkg/notifier/client/mail"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/db/mongodb"
	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/services"
)

var (
	realm                  = os.Getenv("KEYCLOAK_REALM")
	hostname               = os.Getenv("KEYCLOAK_HOSTNAME")
	serviceAccount         = os.Getenv("KEYCLOAK_SERVICE_ACCOUNT")
	serviceAccountPassword = os.Getenv("KEYCLOAK_SERVICE_ACCOUNT_PASSWORD")
	envPrefixForEmail      = os.Getenv("EMAIL_SUBJECT_PREFIX")
)

const cappedEvents = 30000

func validateEnvVars() error {
	globalVars := map[string]string{
		"KEYCLOAK_REALM":                    realm,
		"KEYCLOAK_HOSTNAME":                 hostname,
		"KEYCLOAK_SERVICE_ACCOUNT":          serviceAccount,
		"KEYCLOAK_SERVICE_ACCOUNT_PASSWORD": serviceAccountPassword,
	}
	for k, v := range globalVars {
		if v == "" {
			return fmt.Errorf("%s not provided", k)
		}
	}
	return nil
}

func main() {
	l := log.GetLogger()
	l.Info("Starting event notifier")
	if err := validateEnvVars(); err != nil {
		l.Fatal("Env variable not set or empty", zap.Error(err))
	}
	db := mongodb.New()
	if err := db.Connect(); err != nil {
		l.Fatal("Error connecting to MongoDB", zap.Error(err))
	}
	disconnect := func() {
		if err := db.Disconnect(); err != nil {
			l.Debug("Db disconnect failed", zap.Any("error", err))
		}
	}
	defer disconnect()
	services.SetDB(db)
	if err := db.SetEventCapping(cappedEvents); err != nil {
		l.Fatal("Capping event notifier failed", zap.Error(err))
	}
	mailClient := mail.New(envPrefixForEmail)
	notifier(db, mailClient)
}
