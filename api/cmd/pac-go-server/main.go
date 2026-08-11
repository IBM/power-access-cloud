package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	flag "github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/client/kubernetes"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/db/mongodb"
	log "github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/logger"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/router"
	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/services"

	_ "github.com/joho/godotenv/autoload"
)

var servicePort string

// shutdownGrace is the time the server gives active WebSocket sessions and
// background workers to finish before the process exits.  Long enough for
// in-flight DB writes and a clean WS close frame; short enough to satisfy
// most orchestrator restart budgets (Kubernetes default terminationGracePeriodSeconds=30s).
const shutdownGrace = 10 * time.Second

func initFlags() {
	flag.StringVar(&servicePort, "port", "8000", "port to run the service on")
	flag.StringSliceVar(&models.ExcludeGroups, "exclude-groups", []string{"admin"}, "comma separated list of groups to exclude")
	flag.DurationVar(&models.ExpiryNotificationDuration, "expiry-notification-duration", 48*time.Hour,
		`set duration for notification for about-to-expire services,
		e.g. 45s, 2m, 1h30m, 20h, default: 48h which means that user will start receiving expiry notifications 48 hrs before service expiry, once a day`)
	flag.Parse()
}

func main() {
	initFlags()
	logger := log.GetLogger()
	logger.Info("Starting PAC server...")

	logger.Info("Attempting to connect to MongoDB...")
	db := mongodb.New()
	if err := db.Connect(); err != nil {
		logger.Fatal("failed to connect to MongoDB", zap.Error(err))
	}

	defer func() {
		if err := db.Disconnect(); err != nil {
			logger.Error("failed to disconnect from MongoDB", zap.Error(err))
		}
	}()
	services.SetDB(db)

	logger.Info("Attempting to connect to Kubernetes cluster...")
	kubeClient := kubernetes.NewClient()
	services.SetKubeClient(kubeClient)

	// workerCtx is cancelled on shutdown to signal background workers to
	// drain their queues and exit.
	workerCtx, workerCancel := context.WithCancel(context.Background())

	logger.Info("Starting service expiry notifier")
	go services.ExpiryNotification(workerCtx)

	go services.StartAdminReplyWorker(workerCtx)

	srv := &http.Server{
		Addr:    ":" + servicePort,
		Handler: router.CreateRouter(),
	}

	// Start serving in a background goroutine so the main goroutine can block
	// on the OS signal channel instead.
	go func() {
		logger.Info("PAC server is up and running", zap.String("port", servicePort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	// Block until SIGINT or SIGTERM is received.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	logger.Info("shutdown signal received, draining active connections",
		zap.String("signal", sig.String()),
		zap.Duration("grace", shutdownGrace),
	)

	// Cancel background workers first so they can drain concurrently with the
	// HTTP server draining its active connections within the grace window.
	workerCancel()

	// Stop accepting new connections; give existing sessions the grace period.
	shutCtx, shutCancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Error("server forced to shut down before grace period expired", zap.Error(err))
	}

	logger.Info("PAC server exited cleanly")
}
