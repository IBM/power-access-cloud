package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Key struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID  string             `json:"user_id" bson:"user_id,omitempty"`
	Name    string             `json:"name" bson:"name,omitempty"`
	Content string             `json:"content" bson:"content,omitempty"`
}

func GetNewKey() Key {
	var key Key
	return key
}
