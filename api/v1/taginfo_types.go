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

// TagInfoSpec defines the desired state of TagInfo
type TagInfoSpec struct {
	Tagname string `json:"tagname,omitempty"`
}

// TagInfoStatus defines the observed state of TagInfo
type TagInfoStatus struct {
	HostList      []string `json:"host_list,omitempty"`
	VMList        []string `json:"vm_list,omitempty"`
	DatastoreList []string `json:"datastore_list,omitempty"`
	NetworkList   []string `json:"network_list,omitempty"`
	ClusterList   []string `json:"cluster_list,omitempty"`
	// DatacenterList []string
}

// +kubebuilder:validation:Optional
// +kubebuilder:resource:shortName={"tag"}
// +kubebuilder:printcolumn:name="Nodename",type=string,JSONPath=`.spec.nodename`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TagInfo is the Schema for the taginfoes API
type TagInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TagInfoSpec   `json:"spec,omitempty"`
	Status TagInfoStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TagInfoList contains a list of TagInfo
type TagInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TagInfo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TagInfo{}, &TagInfoList{})
}
