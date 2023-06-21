package models

import "k8s.io/apimachinery/pkg/util/intstr"

type Capacity struct {
	CPU    intstr.IntOrString `json:"cpu" bson:"cpu,omitempty"`
	Memory int                `json:"memory" bson:"memory,omitempty"`
}
