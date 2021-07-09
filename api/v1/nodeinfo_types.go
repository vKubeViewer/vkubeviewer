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
	GuestId    string `json:"guestId"`
	TotalCPU   int64  `json:"totalCPU"`
	ResvdCPU   int64  `json:"resvdCPU"`
	TotalMem   int64  `json:"totalMem"`
	ResvdMem   int64  `json:"resvdMem"`
	PowerState string `json:"powerState"`
	HwVersion  string `json:"hwVersion"`
	IpAddress  string `json:"ipAddress"`
	PathToVM   string `json:"pathToVM"`
}

// +kubebuilder:validation:Optional
// +kubebuilder:resource:shortName={"ch"}
// +kubebuilder:printcolumn:name="Nodename",type=string,JSONPath=`.spec.nodename`
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
