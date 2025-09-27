package controller

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// GetCodeServerContainer creates a code-server container
func GetCodeServerContainer(password string) corev1.Container {
	return corev1.Container{
		Name:  "code-server",
		Image: "codercom/code-server:latest",
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

// CompareConfigMaps checks if two ConfigMaps have the same data
func CompareConfigMaps(a, b *corev1.ConfigMap) bool {
	if len(a.Data) != len(b.Data) {
		return false
	}
	for key, valueA := range a.Data {
		valueB, exists := b.Data[key]
		if !exists || valueA != valueB {
			return false
		}
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

// getSecureCodeServerContainer returns a code-server container
func getSecureCodeServerContainer(password string) corev1.Container {
	return corev1.Container{
		Name:  "code-server",
		Image: "codercom/code-server:latest",
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

// getSecureGameServerContainer returns a game server container
func getSecureGameServerContainer(name, image string, resources corev1.ResourceRequirements, ports []corev1.ContainerPort) corev1.Container {
	return corev1.Container{
		Name:      name,
		Image:     image,
		Resources: applyResourceDefaults(resources),
		Ports:     ports,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "data",
				MountPath: "/data",
			},
		},
	}
}

// ReconcileConfigMap handles ConfigMap reconciliation
func ReconcileConfigMap(ctx context.Context, c client.Client, owner metav1.Object, name string, configData map[string]string) error {
	logger := log.FromContext(ctx)

	desired := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: owner.GetNamespace(),
		},
		Data: configData,
	}

	if err := controllerutil.SetControllerReference(owner, desired, c.Scheme()); err != nil {
		return err
	}

	found := &corev1.ConfigMap{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating new ConfigMap", "namespace", desired.Namespace, "name", desired.Name)
		return c.Create(ctx, desired)
	} else if err != nil {
		return err
	}

	if !CompareConfigMaps(found, desired) {
		logger.Info("Updating ConfigMap", "namespace", found.Namespace, "name", found.Name)
		found.Data = desired.Data
		if err := c.Update(ctx, found); err != nil {
			if errors.IsConflict(err) {
				logger.Info("Conflict updating ConfigMap, will retry")
			}
			return err
		}
	}

	logger.V(4).Info("ConfigMap already exists and is up to date", "namespace", found.Namespace, "name", found.Name)
	return nil
}
