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
	"github.com/IBM-Cloud/power-go-client/power/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ServiceType string

var (
	ServiceTypeVM ServiceType = "vm"
)

type PortType string

var (
	PortTypeTCP   PortType = "tcp"
	PortTypeUDP   PortType = "udp"
	PortTypeHTTP  PortType = "http"
	PortTypeHTTPS PortType = "https"
)

type VirtualMachine struct {

	// Name of the virtual machine
	Name string `json:"name"`

	// ID of the vm
	ID string `json:"ID"`

	// Ports
	Ports []Port `json:"ports,omitempty"`

	// CloudInstanceID is a PowerVS cloud instance ID
	CloudInstanceID string `json:"cloudInstanceID,omitempty"`

	// Zone in which PowerVS cloud instance exist
	Zone string `json:"zone,omitempty"`

	// VPC information
	VPC VPC `json:"vpc,omitempty"`
}

type VPC struct {
	ID           string `json:"ID"`
	Region       string `json:"region"`
	Loadbalancer string `json:"loadbalancer"`
}

// ServiceSpec defines the desired state of Service
type ServiceSpec struct {
	// ManageIQ ID
	ID string `json:"ID"`

	CreatedAt string `json:"createdAt"`

	// Type of service
	Type ServiceType `json:"type,omitempty"`

	// VirtualMachine spec
	// +optional
	VirtualMachine *VirtualMachine `json:"virtualMachine,omitempty"`
}

type Port struct {
	// Port number
	Number uint `json:"number"`

	Type        string `json:"type"`
	Target      uint   `json:"target,omitempty"`
	BackendPool string `json:"backendPool,omitempty"`
}

//type NetworkStatus struct {
//	Port uint `json:"port"`
//	// TODO: Add check
//	Type                 PortType `json:"type"`
//	TargetPort           uint     `json:"targetPort,omitempty"`
//	BackendPool          string   `json:"backendPool,omitempty"`
//	LoadbalancerHostname string   `json:"loadbalancerHostname,omitempty"`
//}

type VirtualMachineStatus struct {
	// InstanceID is the virtual machine instance id
	InstanceID string `json:"instanceID,omitempty"`

	MACAddress   string `json:"MACAddress,omitempty"`
	Network      string `json:"network,omitempty"`
	IPAddress    string `json:"IPAddress,omitempty"`
	Loadbalancer string `json:"loadbalancer,omitempty"`
	Ports        []Port `json:"ports,omitempty"`
}

// ServiceStatus defines the observed state of Service
type ServiceStatus struct {
	// Ready is true when the service is ready.
	// +optional
	Ready bool `json:"ready"`

	// Retired will be true when service is retired
	Retired bool `json:"retired"`

	// Deleted will be true if service not found
	Deleted bool `json:"deleted"`

	// Conditions defines current service state of the service
	// +optional
	Conditions capiv1beta1.Conditions `json:"conditions,omitempty"`

	// VirtualMachine status spec
	VirtualMachine VirtualMachineStatus `json:"virtualMachine,omitempty"`
}

func (s *ServiceStatus) SetReady() {
	s.Ready = true
}

func (s *ServiceStatus) SetVirtualMachineStatusInstanceID(in *models.PVMInstanceReference) {
	s.VirtualMachine.InstanceID = *in.PvmInstanceID
}

func (s *ServiceStatus) SetVirtualMachineStatusMACAddress(in *models.PVMInstanceReference) {
	for _, nw := range in.Networks {
		// skipping fixed ip addresses which are public networks in nature
		if nw.Type == "fixed" {
			continue
		}
		// skip the networks which doesn't have mac address set
		if nw.MacAddress == "" {
			continue
		}

		s.VirtualMachine.MACAddress = nw.MacAddress
		s.VirtualMachine.Network = nw.NetworkName

	}
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of Service"
//+kubebuilder:printcolumn:name="Created At",type="date",JSONPath=".spec.createdAt",description="When the service is created at"
//+kubebuilder:printcolumn:name="Retired",type="string",JSONPath=".status.retired",description="Service retired"
//+kubebuilder:printcolumn:name="Deleted",type="string",JSONPath=".status.deleted",description="Service deleted"
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Service Ready to consume"

// Service is the Schema for the services API
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
}

func (in *Service) SetRetired() {
	in.Status.Retired = true
}

func (in *Service) IsRetired() bool {
	return in.Status.Retired
}

func (in *Service) SetNotReady() {
	in.Status.Ready = false
}

func (in *Service) IsReady() bool {
	return in.Status.Ready
}

func (in *Service) SetDeleted() {
	in.Status.Deleted = true
}

func (in *Service) IsDeleted() bool {
	return in.Status.Deleted
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
