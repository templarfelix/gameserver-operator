package controller

import (
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Game Server Configuration Constants
const (
	// InitContainer configuration
	SetupContainerImage = "alpine:latest"
	SetupContainerName  = "config-setup"

	// Volume names
	DataVolumeName = "data"
)

// CompareDeployments checks if two Deployments have equivalent specs
func CompareDeployments(a, b *appsv1.Deployment) bool {
	// Compare replicas
	aReplicas := int32(1)
	bReplicas := int32(1)
	if a.Spec.Replicas != nil {
		aReplicas = *a.Spec.Replicas
	}
	if b.Spec.Replicas != nil {
		bReplicas = *b.Spec.Replicas
	}
	if aReplicas != bReplicas {
		return false
	}

	// Compare selector labels
	if !reflect.DeepEqual(a.Spec.Selector, b.Spec.Selector) {
		return false
	}

	// Compare pod template spec deeply
	if !reflect.DeepEqual(a.Spec.Template.Spec, b.Spec.Template.Spec) {
		return false
	}

	// Compare pod template metadata labels
	if !reflect.DeepEqual(a.Spec.Template.Labels, b.Spec.Template.Labels) {
		return false
	}

	return true
}

// applyResourceDefaults ensures resources have secure defaults
func applyResourceDefaults(resources corev1.ResourceRequirements) corev1.ResourceRequirements {
	if resources.Requests == nil {
		resources.Requests = make(corev1.ResourceList)
	}
	if resources.Limits == nil {
		resources.Limits = make(corev1.ResourceList)
	}

	// Set CPU defaults if not specified
	if _, exists := resources.Requests[corev1.ResourceCPU]; !exists {
		resources.Requests[corev1.ResourceCPU] = resource.MustParse("500m")
	}
	if _, exists := resources.Limits[corev1.ResourceCPU]; !exists {
		resources.Limits[corev1.ResourceCPU] = resource.MustParse("2000m")
	}

	// Set Memory defaults if not specified
	if _, exists := resources.Requests[corev1.ResourceMemory]; !exists {
		resources.Requests[corev1.ResourceMemory] = resource.MustParse("1Gi")
	}
	if _, exists := resources.Limits[corev1.ResourceMemory]; !exists {
		resources.Limits[corev1.ResourceMemory] = resource.MustParse("4Gi")
	}

	return resources
}

// GetSecureCodeServerContainer returns a code-server container
func GetSecureCodeServerContainer(password string) corev1.Container {
	return corev1.Container{
		Name:  "code-server",
		Image: "codercom/code-server:latest",
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:  func(i int64) *int64 { return &i }(1000),
			RunAsGroup: func(i int64) *int64 { return &i }(1000),
		},
		Ports: []corev1.ContainerPort{{
			ContainerPort: 8080,
			Name:          "code-server",
			Protocol:      corev1.ProtocolTCP,
		}},
		Env: []corev1.EnvVar{{
			Name:  "PASSWORD",
			Value: password,
		}},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "data",
			MountPath: "/data",
		}},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
		},
		ReadinessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/",
					Port: intstr.FromInt32(8080),
				},
			},
			InitialDelaySeconds: 5,
			PeriodSeconds:       10,
		},
		LivenessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/",
					Port: intstr.FromInt32(8080),
				},
			},
			InitialDelaySeconds: 15,
			PeriodSeconds:       20,
		},
	}
}

// GetSecureGameServerContainer returns a game server container - let LinuxGSM handle user setup
func GetSecureGameServerContainer(name, image string, resources corev1.ResourceRequirements, ports []corev1.ContainerPort) corev1.Container {
	return corev1.Container{
		Name:      name,
		Image:     image,
		Resources: applyResourceDefaults(resources),
		Ports:     ports,
		// Adding security context with specific settings to address permission issues
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:  func(i int64) *int64 { return &i }(0), // Run as root to avoid permission issues
			RunAsGroup: func(i int64) *int64 { return &i }(0), // Run as root group
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      DataVolumeName,
				MountPath: "/data",
			},
		},
	}
}
