package mongodb

import (
	"github.com/PDeXchange/pac/internal/pkg/pac-go-server/models"
)

func (db *MongoDB) GetUserQuota(s string) (models.Quota, error) {
	return models.Quota{
		GroupID: "123",
		Capacity: models.Capacity{
			CPU:    5,
			Memory: 10,
		},
	}, nil
}
