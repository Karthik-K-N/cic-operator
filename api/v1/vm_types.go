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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VMSpec defines the desired state of VM
type VMSpec struct {
	// ImageID indicates the image ID
	ImageID string `json:"imageID,omitempty"`
	// NetworkID indicates network ID
	NetworkID string `json:"networkID,omitempty"`
	// FlavorID indicates the flavor ID
	FlavorID string `json:"flavorID,omitempty"`
}

// VMStatus defines the observed state of VM
type VMStatus struct {
	// Status indicates the current status of the VM
	Status string `json:"status,omitempty"`
	// ID indicates the VM ID
	ID string `json:"id,omitempty"`
	// IP indicates the VM IP addresses
	IP string `json:"ip,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="Status of the VM"
//+kubebuilder:printcolumn:name="IP",type="string",JSONPath=".status.ip",description="IP address of the VM"

// VM is the Schema for the vms API
type VM struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VMSpec   `json:"spec,omitempty"`
	Status VMStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VMList contains a list of VM
type VMList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VM `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VM{}, &VMList{})
}
