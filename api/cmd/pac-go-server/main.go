package main

import (
	"time"

	flag "github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/client/kubernetes"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/db/mongodb"
	log "github.com/PDeXchange/pac/internal/pkg/pac-go-server/logger"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/router"
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/services"

	_ "github.com/joho/godotenv/autoload"
)

var (
	servicePort = "8000"
)

func initFlags() {
	flag.StringVar(&servicePort, "port", "8000", "port to run the service on")
	flag.StringSliceVar(&models.ExcludeGroups, "exclude-groups", []string{"admin"}, "comma separated list of groups to exclude")
	flag.DurationVar(&models.ExpiryNotificationDuration, "expiry-notification-duration", 48*time.Hour,
		`set duration for notification for about-to-expire services,
		e.g. 45s, 2m, 1h30m, 20h, default: 48h which means that user will start receiving expiry notifications 48 hrs before service expiry, once a day`)
	flag.Parse()
}

func main() {
	logger := log.GetLogger()
	logger.Info("Starting PAC server...")
	initFlags()

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

	logger.Info("Attempting to connect to Kubernetes cluster...")
	kubeClient := kubernetes.NewClient()
	services.SetKubeClient(kubeClient)

	logger.Info("Starting service expiry notifier")
	go services.ExpiryNotification()

	var appRouter = router.CreateRouter()
	logger.Info("PAC server is up and running", zap.String("port", servicePort))
	logger.Fatal("Error encountered while routing", zap.Error(appRouter.Run(":"+servicePort)))
}
