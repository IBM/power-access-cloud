package models

import (
	"errors"
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
	Comment string             `json:"comment" bson:"comment,omitempty"`
	// CreatedAt is the time the event was created
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type FeedbacksFilter struct {
	UserID string
}

func (f Feedback) ValidateFeedback() []error {
	var errs []error
	if !f.Rating.IsValidRating() {
		errs = append(errs, fmt.Errorf("invalid rating %s, allowed values: Negative, Neutral, Positive", f.Rating))
	}
	if len(f.Comment) > 250 {
		errs = append(errs, errors.New("comment must not exceed 250 characters"))
	}
	return errs
}

func (r Rating) IsValidRating() bool {
	switch strings.ToLower(string(r)) {
	case string(Positive), string(Negative), string(Neutral):
		return true
	default:
		return false
	}
}
