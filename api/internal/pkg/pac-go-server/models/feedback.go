package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Rating string

const (
	Positive Rating = "Positive"
	Negative Rating = "Negative"
	Neutral  Rating = "Neutral"
)

func (r Rating) IsValid() bool {
	switch r {
	case Positive, Negative, Neutral:
		return true
	default:
		return false
	}
}

type Feedback struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID  string             `json:"user_id" bson:"user_id,omitempty"`
	Rating  Rating             `json:"rating" bson:"rating,omitempty"`
	Comment string             `json:"comment" bson:"comment,omitempty"`
}

type FeedbacksFilter struct {
	UserID string
}
