package db

import (
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
)

//go:generate mockgen -destination=mock_db_client.go -package=db . DB
type DB interface {
	Connect() error
	Disconnect() error

	GetRequestsByUserID(id, requestType string) ([]models.Request, error)
	NewRequest(request *models.Request) (string, error)
	GetRequestByGroupIDAndUserID(groupID, userID string) ([]models.Request, error)
	GetRequestByID(string) (*models.Request, error)
	DeleteRequest(string) error
	UpdateRequestState(id string, state models.RequestStateType) error
	UpdateRequestStateWithComment(id string, state models.RequestStateType, comment string) error
	GetRequestByServiceName(string) ([]models.Request, error)

	GetKeyByID(id string) (*models.Key, error)
	GetKeyByUserID(userid string) ([]models.Key, error)
	CreateKey(key *models.Key) error
	DeleteKey(string) error

	// Implementations for group quota.
	NewQuota(*models.Quota) error
	UpdateQuota(*models.Quota) error
	DeleteQuota(string) error
	GetQuotaForGroupID(string) (*models.Quota, error)
	GetGroupsQuota([]string) ([]models.Quota, error)

	NewEvent(*models.Event) error
	GetEventsByUserID(string, int64, int64) ([]models.Event, int64, error)
	GetEventsByType(models.EventType, uint) ([]models.Event, int64, error)
	WatchEvents(chan<- *models.Event) error
	MarkEventAsNotified(string) error

	AcceptTermsAndConditions(*models.TermsAndConditions) error
	GetTermsAndConditionsByUserID(string) (*models.TermsAndConditions, error)
	DeleteTermsAndConditionsByUserID(string) error
}
