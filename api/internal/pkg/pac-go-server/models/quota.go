package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Quota struct {
	ID       bson.ObjectID `json:"id" bson:"_id,omitempty"`
	GroupID  string             `json:"group_id" bson:"group_id"`
	Capacity Capacity           `json:"capacity" bson:"capacity"`
}
