package mongodb

import (
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
)

func (db *MongoDB) GetKeyByUserID(id string) ([]models.Key, error) {
	return []models.Key{}, nil
}
