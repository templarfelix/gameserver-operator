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

	gameserverv1alpha1 "github.com/templarfelix/gameserver-operator/api/v1alpha1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DayzSpec defines the desired state of Dayz
type DayzSpec struct {
	//+kubebuilder:default="gameservermanagers/gameserver:dayz"
	Image string `json:"image"`

	gameserverv1alpha1.Base `json:",inline"`

	// Game server configuration
	Config DayzConfig `json:"config,omitempty"`
}

// +kubebuilder:object:generate=true

// DayzConfig defines configuration as a map of file paths to content
type DayzConfig map[string]string

// DayzStatus defines the observed state of Dayz
type DayzStatus struct {
	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:generate=true

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Dayz is the Schema for the dayzs API
type Dayz struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DayzSpec   `json:"spec,omitempty"`
	Status DayzStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DayzList contains a list of Dayz
type DayzList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dayz `json:"items"`
}

func init() {
	gameserverv1alpha1.SchemeBuilder.Register(&Dayz{}, &DayzList{})
}
