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

// NodeNetInfoSpec defines the desired state of NodeNetInfo
type NodeNetInfoSpec struct {
	Nodename string `json:"nodename"`
}

// NodeNetInfoStatus defines the observed state of NodeNetInfo
type NodeNetInfoStatus struct {
	NetName string `json:"net_name,omitempty"`
	// NetRef           string `json:"net_ref,omitempty"`
	// this ref can't show as string, just leave it
	NetOverallStatus string `json:"net_overall_status,omitempty"`
	VlanId           int64  `json:"vlan_id,omitempty"`
	SwitchType       string `json:"switch_type,omitempty"`
}

// +kubebuilder:validation:Optional
// +kubebuilder:resource:shortName={"net"}
// +kubebuilder:printcolumn:name="Nodename",type=string,JSONPath=`.spec.nodename`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NodeNetInfo is the Schema for the nodenetinfoes API
type NodeNetInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeNetInfoSpec   `json:"spec,omitempty"`
	Status NodeNetInfoStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeNetInfoList contains a list of NodeNetInfo
type NodeNetInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeNetInfo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeNetInfo{}, &NodeNetInfoList{})
}
