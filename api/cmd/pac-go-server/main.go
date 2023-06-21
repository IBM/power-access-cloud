package main

import (
	"go.uber.org/zap"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/db/mongodb"
	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/router"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/services"

	_ "github.com/joho/godotenv/autoload"
)

const servicePort = "8000"

func main() {
	logger := log.GetLogger()
	logger.Info("Attempting to connect to MongoDB...")
	db := mongodb.New()
	if err := db.Connect(); err != nil {
		panic(err)
	}

	defer func() {
		if err := db.Disconnect(); err != nil {
			panic(err)
		}
	}()
	services.SetDB(db)

	var appRouter = router.CreateRouter()
	logger.Info("PAC server is up and running", zap.String("port", servicePort))
	logger.Fatal("Error encountered while routing", zap.Error(appRouter.Run(":"+servicePort)))
}
