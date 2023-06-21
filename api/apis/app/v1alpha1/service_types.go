/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceState is state of catalog
// +kubebuilder:validation:Enum=NEW;IN_PROGRESS;CREATED;FAILED;EXPIRED
type ServiceState string

const (
	ServiceStateNew        ServiceState = "NEW"
	ServiceStateInProgress ServiceState = "IN_PROGRESS"
	ServiceStateCreated    ServiceState = "CREATED"
	ServiceStateFailed     ServiceState = "FAILED"
	ServiceStateExpired    ServiceState = "EXPIRED"
)

// VM has the detail of provisioned vm service
type VM struct {
	InstanceID string `json:"instance_id,omitempty"`
	IPAddress  string `json:"ip_address,omitempty"`
	State      string `json:"state,omitempty"`
}

// ServiceSpec defines the desired state of Service
type ServiceSpec struct {
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="user_id is immutable"
	UserID      string      `json:"user_id"`
	DisplayName string      `json:"display_name"`
	Expiry      metav1.Time `json:"expiry"`
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="catalog is immutable"
	Catalog corev1.LocalObjectReference `json:"catalog"`
	SSHKeys []string                    `json:"ssh_keys"`
}

// ServiceStatus defines the observed state of Service
type ServiceStatus struct {
	// +kubebuilder:validation:Optional
	VM VM `json:"vm,omitempty"`
	// +optional
	AccessInfo string `json:"accessInfo"`
	// +kubebuilder:validation:Required
	Expired bool `json:"expired,omitempty"`
	// +kubebuilder:validation:Required
	Message string `json:"message,omitempty"`
	// +kubebuilder:validation:Required
	State ServiceState `json:"state,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Service is the Schema for the services API
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec ServiceSpec `json:"spec,omitempty"`
	// +optional
	Status ServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ServiceList contains a list of Service
type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Service `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Service{}, &ServiceList{})
}
