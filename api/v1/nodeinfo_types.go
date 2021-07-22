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

// NodeInfoSpec defines the desired state of NodeInfo
type NodeInfoSpec struct {
	Nodename string `json:"nodename"`
}

// NodeInfoStatus defines the observed state of NodeInfo
type NodeInfoStatus struct {
	// cpu, memory, vmipaddress, powerstate
	ActtachedTag []string `json:"acttached_tag,omitempty"`
	VMGuestId    string   `json:"vm_guest_id,omitempty"`
	VMTotalCPU   int64    `json:"vm_total_cpu,omitempty"`
	VMResvdCPU   int64    `json:"vm_resvd_cpu,omitempty"`
	VMTotalMem   int64    `json:"vm_total_mem,omitempty"`
	VMResvdMem   int64    `json:"vm_resvd_mem,omitempty"`
	VMPowerState string   `json:"vm_power_state,omitempty"`
	VMHwVersion  string   `json:"vm_hw_version,omitempty"`
	VMIpAddress  string   `json:"vm_ip_address,omitempty"`
	PathToVM     string   `json:"path_to_vm,omitempty"`

	NetName          string `json:"net_name,omitempty"`
	NetOverallStatus string `json:"net_overall_status,omitempty"`
	NetSwitchType    string `json:"net_switch_type,omitempty"`
	NetVlanId        int32  `json:"net_vlan_id,omitempty"`
}

// +kubebuilder:validation:Optional
// +kubebuilder:resource:shortName={"nd"}
// +kubebuilder:printcolumn:name="Nodename",type=string,JSONPath=`.spec.nodename`
// +kubebuilder:printcolumn:name="VMTotalCPU",type=string,JSONPath=`.status.vm_total_cpu`
// +kubebuilder:printcolumn:name="VMTotalMem",type=string,JSONPath=`.status.vm_total_mem`
// +kubebuilder:printcolumn:name="VMPowerState",type=string,JSONPath=`.status.vm_power_state`
// +kubebuilder:printcolumn:name="VMIpAddress",type=string,JSONPath=`.status.vm_ip_address`
// +kubebuilder:printcolumn:name="VMHwVersion",type=string,JSONPath=`.status.vm_hw_version`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NodeInfo is the Schema for the nodeinfoes API
type NodeInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeInfoSpec   `json:"spec,omitempty"`
	Status NodeInfoStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeInfoList contains a list of NodeInfo
type NodeInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeInfo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeInfo{}, &NodeInfoList{})
}
