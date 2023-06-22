package models

type Capacity struct {
	CPU    float64 `json:"cpu" bson:"cpu,omitempty"`
	Memory int     `json:"memory" bson:"memory,omitempty"`
}
