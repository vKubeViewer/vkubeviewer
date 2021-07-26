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

// DatastoreInfoSpec defines the desired state of DatastoreInfo
type DatastoreInfoSpec struct {
	Datastore string `json:"datastore"`
}

// DatastoreInfoStatus defines the observed state of DatastoreInfo
type DatastoreInfoStatus struct {
	Type         string   `json:"type,omitempty"`
	Status       string   `json:"status,omitempty"`
	Capacity     string   `json:"capacity,omitempty"`
	FreeSpace    string   `json:"free_space,omitempty"`
	Accessible   bool     `json:"accessible,omitempty"`
	HostsMounted []string `json:"hosts_mounted,omitempty"`
}

// +kubebuilder:validation:Optional
// +kubebuilder:resource:shortName={"ds"}
// +kubebuilder:printcolumn:name="Datastore",type=string,JSONPath=`.spec.datastore`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.status.type`
// +kubebuilder:printcolumn:name="Capacity",type=string,JSONPath=`.status.capacity`
// +kubebuilder:printcolumn:name="FreeSpace",type=string,JSONPath=`.status.free_space`
// +kubebuilder:printcolumn:name="HostsMounted",type=string,JSONPath=`.status.hosts_mounted`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DatastoreInfo is the Schema for the datastoreinfoes API
type DatastoreInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatastoreInfoSpec   `json:"spec,omitempty"`
	Status DatastoreInfoStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DatastoreInfoList contains a list of DatastoreInfo
type DatastoreInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatastoreInfo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatastoreInfo{}, &DatastoreInfoList{})
}
