package kubernetes

import (
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	pac "github.com/IBM/power-access-cloud/api/apis/app/v1alpha1"
	log "github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/logger"
)

const DefaultNamespace = "default"

type KubeClient struct {
	kubeClient client.Client
}

func NewClient() Client {
	logger := log.GetLogger()
	if err := pac.AddToScheme(scheme.Scheme); err != nil {
		logger.Fatal("Error adding kuberentes schema")
	}

	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatal("Error getting kuberentes configuration", zap.Error(err))
	}

	kubeClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		logger.Fatal("Error getting k8sClient", zap.Error(err))
	}
	return &KubeClient{
		kubeClient: kubeClient,
	}
}
