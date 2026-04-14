package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MaintenanceWindow represents a single maintenance notification window
type MaintenanceWindow struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	Enabled   bool               `bson:"enabled" json:"enabled"`
	StartTime time.Time          `bson:"start_time" json:"start_time"`
	EndTime   time.Time          `bson:"end_time" json:"end_time"`
	Message   string             `bson:"message" json:"message"`
	CreatedBy string             `bson:"created_by" json:"created_by"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedBy string             `bson:"updated_by" json:"updated_by"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedBy string             `bson:"deleted_by,omitempty" json:"deleted_by,omitempty"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// MaintenanceResponse is the public response for a single maintenance window
type MaintenanceResponse struct {
	ID        string    `json:"id"`
	Enabled   bool      `json:"enabled"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Message   string    `json:"message"`
	IsActive  bool      `json:"is_active"`
}

// MaintenanceAdminResponse is the admin response with audit fields
type MaintenanceAdminResponse struct {
	ID        string     `json:"id"`
	Enabled   bool       `json:"enabled"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	Message   string     `json:"message"`
	IsActive  bool       `json:"is_active"`
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedBy string     `json:"updated_by,omitempty"`
	UpdatedAt time.Time  `json:"updated_at,omitempty"`
	DeletedBy string     `json:"deleted_by,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// MaintenanceListResponse is the public response for listing all maintenance windows
type MaintenanceListResponse struct {
	Maintenances []MaintenanceResponse `json:"maintenances"`
}

// MaintenanceAdminListResponse is the admin response for listing all maintenance windows with pagination
type MaintenanceAdminListResponse struct {
	TotalPages   int64                      `json:"total_pages"`
	TotalItems   int64                      `json:"total_items"`
	Maintenances []MaintenanceAdminResponse `json:"maintenances"`
	Links        Links                      `json:"links"`
}

// MaintenanceCreateRequest is the request body for creating a new maintenance window
type MaintenanceCreateRequest struct {
	Enabled   bool      `json:"enabled"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	Message   string    `json:"message" binding:"required"`
}

// MaintenanceUpdateRequest is the request body for updating an existing maintenance window
type MaintenanceUpdateRequest struct {
	Enabled   bool      `json:"enabled"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Message   string    `json:"message,omitempty"`
}
