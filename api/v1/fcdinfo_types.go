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

// FCDInfoSpec defines the desired state of FCDInfo
type FCDInfoSpec struct {
	PVId string `json:"pvId"`
}

// FCDInfoStatus defines the observed state of FCDInfo
type FCDInfoStatus struct {
	SizeMB           int64  `json:"sizeMB"`
	FilePath         string `json:"filePath"`
	ProvisioningType string `json:"provisioningType"`
}

// +kubebuilder:validation:Optional
// +kubebuilder:resource:shortName={"fcd"}
// +kubebuilder:printcolumn:name="PVId",type=string,JSONPath=`.spec.pvId`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// FCDInfo is the Schema for the fcdinfoes API
type FCDInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FCDInfoSpec   `json:"spec,omitempty"`
	Status FCDInfoStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FCDInfoList contains a list of FCDInfo
type FCDInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FCDInfo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FCDInfo{}, &FCDInfoList{})
}
