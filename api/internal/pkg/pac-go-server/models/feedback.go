package models

import (
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Rating string

const (
	Positive Rating = "positive"
	Negative Rating = "negative"
	Neutral  Rating = "neutral"
)

type Feedback struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID  string             `json:"user_id" bson:"user_id,omitempty"`
	Rating  Rating             `json:"rating" bson:"rating,omitempty"`
	Comment string             `json:"comment" bson:"comment,omitempty" binding:"max=250"`
	// CreatedAt is the time the event was created
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type FeedbackResponse struct {
	// TotalPages is the total number of pages
	TotalPages int64 `json:"total_pages"`
	// TotalItems is the total number of items
	TotalItems int64 `json:"total_items"`
	// Feedbacks is the list of feedback
	Feedbacks []Feedback `json:"feedbacks"`
	// Links contains the links for the current page, next page and last page
	Links Links `json:"links"`
}
type FeedbacksFilter struct {
	UserID string
}

func (f Feedback) ValidateFeedback() error {
	var err error
	if !f.Rating.IsValidRating() {
		return fmt.Errorf("invalid rating %s, allowed values: Negative, Neutral, Positive", f.Rating)
	}
	return err
}

func (r Rating) IsValidRating() bool {
	switch strings.ToLower(string(r)) {
	case string(Positive), string(Negative), string(Neutral):
		return true
	default:
		return false
	}
}
