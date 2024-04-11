package v1alpha1

import corev1 "k8s.io/api/core/v1"

// Base values need
type Base struct {

	//+kubebuilder:default="10G"
	Storage string `json:"storage,omitempty"`

	LoadBalancerIP string `json:"loadBalancerIP,omitempty"`

	Ports []corev1.ServicePort `json:"ports"`
}
