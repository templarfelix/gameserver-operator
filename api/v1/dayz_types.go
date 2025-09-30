/*
Copyright 2025.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DayzSpec defines the desired state of Dayz.
type DayzSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:default="gameservermanagers/gameserver:dayz"
	Image string `json:"image"`

	// Base contains common configuration fields for game server CRDs
	Base `json:",inline"`

	// Game server configuration
	Config DayzConfig `json:"config,omitempty"`
}

// +kubebuilder:object:generate=true

// DayzConfig defines configuration as a map of file paths to content
type DayzConfig map[string]string

// Base contains common configuration fields for game server CRDs
type Base struct {
	Persistence Persistence `json:"persistence,omitempty"`

	Ports []corev1.ServicePort `json:"ports,omitempty"`

	LoadBalancerIP string `json:"loadBalancerIP,omitempty"`

	Resources corev1.ResourceRequirements `json:"resources"`

	// NodeSelector is a selector which must be true for the pod to fit on a node
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations are the tolerations for the pod
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity is the affinity for the pod
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Annotations for the pod template
	Annotations map[string]string `json:"annotations,omitempty"`

	// EditorPassword is the password for the code-server editor
	EditorPassword string `json:"editorPassword,omitempty"`
}

// StorageConfig defines the storage configuration for persistent volumes
type StorageConfig struct {
	// Size of the persistent volume (default: "10G")
	//+kubebuilder:default="10G"
	Size string `json:"size,omitempty"`

	// Storage class name for the volume
	StorageClassName string `json:"storageClassName,omitempty"`
}

// Persistence configures the persistent volume for game data
type Persistence struct {
	// Storage configuration
	StorageConfig StorageConfig `json:"storageConfig,omitempty"`

	//+kubebuilder:default=false
	PreserveOnDelete bool `json:"preserveOnDelete,omitempty"`
}

// DayzStatus defines the observed state of Dayz.
type DayzStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Dayz is the Schema for the dayzs API.
type Dayz struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DayzSpec   `json:"spec,omitempty"`
	Status DayzStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DayzList contains a list of Dayz.
type DayzList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dayz `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dayz{}, &DayzList{})
}
