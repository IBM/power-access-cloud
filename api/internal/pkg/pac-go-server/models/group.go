package models

type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Membership is a flag to indicate if the user is a member of the group
	Membership bool `json:"membership"`
}
