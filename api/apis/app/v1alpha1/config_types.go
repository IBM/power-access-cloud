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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ConfigType string

var (
	ConfigTypePowerVS ConfigType = "powervs"
)

type PowerVSConfig struct {
	Zone            string `json:"zone"`
	CloudInstanceID string `json:"cloudInstanceID"`
}

type VPCConfig struct {
	Region string `json:"region"`
	//TODO: remove if not required
	Zone           string `json:"zone"`
	ID             string `json:"ID"`
	LoadBalancerID string `json:"loadBalancerID"`
}

// ConfigSpec defines the desired state of Config
// TODO: Add appropriate kubebuilder markers for the field
type ConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// MIQURL used for talking to Manage IQ
	MIQURL string `json:"MIQURL"`

	// MIQUserName is the user name used for talking ManageIQ
	MIQUserName string `json:"MIQUserName"`

	// MIQClientID is the client ID created in the keycloak server for talking to ManageIQ
	MIQClientID string `json:"MIQClientID"`

	// KeycloakURL used for talking to keycloak server
	KeycloakURL string `json:"keycloakURL"`

	// KeycloakRealm  is the realm used for the manageiq
	KeycloakRealm string `json:"keycloakRealm"`

	// CredentialSecret is the secret contains the credential like MIQ password, ClientSecret
	// Secret contains the following data:
	// miq-password: <ManageIQ Password>
	// miq-client-password: <ManageIQ Client Password>
	CredentialSecret corev1.LocalObjectReference `json:"credentialSecret"`

	Type ConfigType `json:"type"`

	PowerVS PowerVSConfig `json:"powerVS"`

	VPC VPCConfig `json:"vpc"`
}

// ConfigStatus defines the observed state of Config
type ConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Config is the Schema for the configs API
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigSpec   `json:"spec,omitempty"`
	Status ConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConfigList contains a list of Config
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Config{}, &ConfigList{})
}
