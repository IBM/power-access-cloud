package db

import (
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
)

type DB interface {
	Connect() error
	Disconnect() error
	GetRequestsByUserID(id string) ([]models.Request, error)
	NewRequest(request *models.Request) error
	GetRequestByGroupIDAndUserID(groupID, userID string) (*models.Request, error)
	GetRequestByID(string) (*models.Request, error)
	DeleteRequest(string) error
	UpdateRequestState(id string, state models.RequestStateType) error
}
