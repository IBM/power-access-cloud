package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestStateType string
type RequestType string

const (
	RequestStateNew            RequestStateType = "NEW"
	RequestStateApproved       RequestStateType = "APPROVED"
	RequestStateRejected       RequestStateType = "REJECTED"
	RequestAddToGroup          RequestType      = "GROUP"
	RequestExitFromGroup       RequestType      = "GROUP_EXIT"
	RequestExtendServiceExpiry RequestType      = "SERVICE_EXPIRY"
	RequestStateExpired        RequestStateType = "EXPIRED"
)

type Request struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID         string             `json:"user_id" bson:"user_id,omitempty"`
	Justification  string             `json:"justification" bson:"justification,omitempty"`
	Comment        string             `json:"comment" bson:"comment,omitempty"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at,omitempty"`
	State          RequestStateType   `json:"state" bson:"state,omitempty"`
	RequestType    RequestType        `json:"type" bson:"type,omitempty"`
	GroupAdmission *GroupAdmission    `json:"group,omitempty" bson:"group,omitempty"`
	ServiceExpiry  *ServiceExpiry     `json:"service,omitempty" bson:"service,omitempty"`
}

type ServiceExpiry struct {
	Name   string    `json:"name" bson:"name,omitempty"`
	Expiry time.Time `json:"expiry" bson:"expiry,omitempty"`
}

type GroupAdmission struct {
	GroupID   string `json:"group_id" bson:"group_id,omitempty"`
	Group     string `json:"group" bson:"group,omitempty"`
	Requester string `json:"requester" bson:"requester,omitempty"`
}

type NewRequest struct {
	Justification string `json:"justification"`
}

func GetNewRequest() NewRequest {
	var request NewRequest
	return request
}

func GetRequest() Request {
	var request Request
	return request
}

func GetRequests() []Request {
	var requests []Request
	return requests
}
