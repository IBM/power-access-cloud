package models

import "time"

// Service will have the details of the user provisioned service from catalog
type Service struct {
	ID          string        `json:"id"`
	UserID      string        `json:"user_id"`
	Name        string        `json:"name"`
	DisplayName string        `json:"display_name"`
	CatalogName string        `json:"catalog_name"`
	Expiry      time.Time     `json:"expiry"`
	Status      ServiceStatus `json:"status"`
}

type ServiceStatus struct {
	State      string `json:"state"`
	Message    string `json:"message"`
	AccessInfo string `json:"access_info"`
}
