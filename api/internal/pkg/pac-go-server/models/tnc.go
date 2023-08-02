package models

import (
	"time"
)

type TermsAndConditions struct {
	// UserID is the user who accepted the terms and conditions
	UserID string `json:"user_id" bson:"user_id"`
	// Accepted is the flag that indicates if the user accepted the terms and conditions
	Accepted bool `json:"accepted" bson:"accepted"`
	// AcceptedAt is the timestamp when the user accepted the terms and conditions
	AcceptedAt *time.Time `json:"accepted_at,omitempty" bson:"accepted_at"`
	// TODO: Add checksum of the terms and conditions to ensure that the user accepted the latest version
}
