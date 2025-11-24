package client

import "github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"

type Notifier interface {
	Notify(event models.Event) error
}
