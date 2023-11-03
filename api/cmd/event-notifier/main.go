package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/PDeXchange/pac/internal/pkg/notifier/client/mail"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/db/mongodb"
	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
)

var (
	realm                  = os.Getenv("KEYCLOAK_REALM")
	hostname               = os.Getenv("KEYCLOAK_HOSTNAME")
	serviceAccount         = os.Getenv("KEYCLOAK_SERVICE_ACCOUNT")
	serviceAccountPassword = os.Getenv("KEYCLOAK_SERVICE_ACCOUNT_PASSWORD")
)

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
	defer db.Disconnect()
	mailClient := mail.New()
	notifier(db, mailClient)
}
