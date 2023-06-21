package models

import "time"

// Service will have the details of the user provisioned service from catalog
type Service struct {
	ID          string        `json:"id"`
	UserID      string        `json:"user_id"`
	DisplayName string        `json:"display_name"`
	CatalogID   string        `json:"catalog_id"`
	Expiry      time.Time     `json:"expiry"`
	Status      ServiceStatus `json:"status"`
}

type ServiceStatus struct {
	State      string `json:"state"`
	Message    string `json:"message"`
	AccessInfo string `json:"access_info"`
}
