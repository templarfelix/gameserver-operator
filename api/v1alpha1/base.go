package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

// Base values need
type Base struct {

	//+kubebuilder:default="10G"
	Storage string `json:"storage,omitempty"`

	Ports []corev1.ServicePort `json:"ports,omitempty"`

	LoadBalancerIP string `json:"loadBalancerIP,omitempty"`

	Resources corev1.ResourceRequirements `json:"resources"`

	// NodeSelector is a selector which must be true for the pod to fit on a node
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations are the tolerations for the pod
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity is the affinity for the pod
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// EditorPassword is the password for the code-server editor
	EditorPassword string `json:"editorPassword,omitempty"`
}
