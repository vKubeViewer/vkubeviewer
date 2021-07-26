/*
Copyright 2021.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HostInfoSpec defines the desired state of HostInfo
type HostInfoSpec struct {
	Hostname string `json:"hostname"`
}

// HostInfoStatus defines the observed state of HostInfo
type HostInfoStatus struct {
	TotalCPU          int64  `json:"total_cpu,omitempty"`
	FreeCPU           int64  `json:"free_cpu,omitempty"`
	TotalMemory       string `json:"total_memory,omitempty"`
	FreeMemory        string `json:"free_memory,omitempty"`
	TotalStorage      string `json:"total_storage,omitempty"`
	FreeStorage       string `json:"free_storage,omitempty"`
	InMaintenanceMode bool   `json:"in_maintenance_mode"`
}

// +kubebuilder:validation:Optional
// +kubebuilder:resource:shortName={"hi"}
// +kubebuilder:printcolumn:name="Hostname",type=string,JSONPath=`.spec.hostname`
// +kubebuilder:printcolumn:name="TotalCPU",type=string,JSONPath=`.status.total_cpu`
// +kubebuilder:printcolumn:name="FreeCPU",type=string,JSONPath=`.status.free_cpu`
// +kubebuilder:printcolumn:name="TotalMemory",type=string,JSONPath=`.status.total_memory`
// +kubebuilder:printcolumn:name="FreeMemory",type=string,JSONPath=`.status.free_memory`
// +kubebuilder:printcolumn:name="InMaintenanceMode",type=string,JSONPath=`.status.in_maintenance_mode`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// HostInfo is the Schema for the hostinfoes API
type HostInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HostInfoSpec   `json:"spec,omitempty"`
	Status HostInfoStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HostInfoList contains a list of HostInfo
type HostInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HostInfo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HostInfo{}, &HostInfoList{})
}
