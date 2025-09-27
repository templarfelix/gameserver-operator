package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

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
	// Storage size for the persistent volume (deprecated: use storageConfig.size)
	Storage string `json:"storage,omitempty"`

	// Storage configuration
	StorageConfig StorageConfig `json:"storageConfig,omitempty"`

	//+kubebuilder:default=false
	PreserveOnDelete bool `json:"preserveOnDelete,omitempty"`
}

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
