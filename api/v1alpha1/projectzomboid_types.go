/*
Copyright 2024.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProjectZomboidSpec defines the desired state of ProjectZomboid
type ProjectZomboidSpec struct {
	//+kubebuilder:default="gameservermanagers/gameserver:project-zomboid"
	Image string `json:"image"`

	Base `json:",inline"`

	Config ProjectZomboidConfig `json:"config,omitempty"`
}

// ProjectZomboidConfig defines configuration for Project Zomboid & LinuxGSM
type ProjectZomboidConfig struct {
	Server string `json:"server,omitempty"`
	GSM    string `json:"gsm,omitempty"`
}

// ProjectZomboidStatus defines the observed state of ProjectZomboid
type ProjectZomboidStatus struct {
	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ProjectZomboid is the Schema for the projectzomboids API
type ProjectZomboid struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectZomboidSpec   `json:"spec,omitempty"`
	Status ProjectZomboidStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProjectZomboidList contains a list of ProjectZomboid
type ProjectZomboidList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectZomboid `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProjectZomboid{}, &ProjectZomboidList{})
}
