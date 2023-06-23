package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Quota struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	GroupID  string             `json:"group_id" bson:"group_id"`
	Capacity Capacity           `json:"capacity" bson:"capacity"`
}
