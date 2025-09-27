package controller

import (
	"fmt"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gameserverv1alpha1 "github.com/templarfelix/gameserver-operator/api/v1alpha1"
)

func pointerString(s string) *string {
	return &s
}

func ReconcilePVC(ctx context.Context, k8sClient client.Client, owner metav1.Object, persistence *gameserverv1alpha1.Persistence) error {
	logger := log.FromContext(ctx)
	pvcName := owner.GetName() + "-pvc"

	// Safety check for nil persistence
	if persistence == nil {
		logger.Error(nil, "Persistence configuration is nil", "owner", owner.GetName())
		return fmt.Errorf("persistence configuration cannot be nil")
	}

	desired := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: owner.GetNamespace(),
		},
	}

	{
		// Get storage size with fallback logic (new structure first, then deprecated)
		storageSize := persistence.StorageConfig.Size
		logger.V(4).Info("Persistence values", "storageConfig.Size", storageSize, "storage", persistence.Storage)

		if storageSize == "" {
			logger.Info("Using deprecated storage field", "storage", persistence.Storage)
			storageSize = persistence.Storage
		}
		if storageSize == "" {
			logger.Info("Using default storage size", "default", "10G")
			storageSize = "10G" // default
		}

		// Final safety check for storageSize - this should never happen but prevents panic
		if storageSize == "" {
			logger.Error(nil, "StorageSize is still empty after fallbacks, this should never happen", "owner", owner.GetName())
			storageSize = "10G"
		}

		// Validate the storage size can be parsed before calling MustParse
		parsedSize, err := resource.ParseQuantity(storageSize)
		if err != nil {
			logger.Error(err, "Invalid storage size format, using default", "storageSize", storageSize, "owner", owner.GetName())
			parsedSize = resource.MustParse("10G")
		}

		// Get storage class name with fallback logic
		var storageClassName *string
		if persistence.StorageConfig.StorageClassName != "" {
			storageClassName = &persistence.StorageConfig.StorageClassName
		}

		desired.Spec = corev1.PersistentVolumeClaimSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			StorageClassName: storageClassName,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: parsedSize,
				},
			},
		}
	}

	if err := controllerutil.SetControllerReference(owner, desired, k8sClient.Scheme()); err != nil {
		return err
	}

	// Check if PVC already exists
	found := &corev1.PersistentVolumeClaim{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: pvcName, Namespace: owner.GetNamespace()}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new PVC", "Namespace", owner.GetNamespace(), "Name", pvcName)
		return k8sClient.Create(ctx, desired)
	} else if err != nil {
		return err
	}

	logger.Info("Skip reconcile: PVC already exists", "Namespace", found.Namespace, "Name", found.Name)
	return nil
}

func ReconcileServices(ctx context.Context, k8sClient client.Client, owner metav1.Object, ports []corev1.ServicePort, loadBalancerIP string) error {
	// Add code-server port to TCP service
	tcpPorts, udpPorts := separatePortsByProtocol(ports)

	// Create TCP service with code-server port
	tcpPorts = append(tcpPorts, corev1.ServicePort{
		Name:       "code-server",
		Port:       8080,
		TargetPort: intstr.FromInt32(8080),
		Protocol:   corev1.ProtocolTCP,
	})

	// Create separate services for TCP and UDP
	if err := reconcileService(ctx, owner.GetName()+"-tcp", k8sClient, owner, tcpPorts, loadBalancerIP); err != nil {
		return err
	}

	if len(udpPorts) > 0 {
		if err := reconcileService(ctx, owner.GetName()+"-udp", k8sClient, owner, udpPorts, loadBalancerIP); err != nil {
			return err
		}
	}

	return nil
}

func reconcileService(ctx context.Context, serviceName string, k8sClient client.Client, owner metav1.Object, ports []corev1.ServicePort, loadBalancerIP string) error {
	logger := log.FromContext(ctx)

	desired := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: owner.GetNamespace(),
			Labels: map[string]string{
				"cloud.google.com/load-balancer-type": "External",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": owner.GetName(),
			},
			Type:           corev1.ServiceTypeLoadBalancer,
			LoadBalancerIP: loadBalancerIP,
			Ports:          ports,
		},
	}

	if err := controllerutil.SetControllerReference(owner, desired, k8sClient.Scheme()); err != nil {
		return err
	}

	found := &corev1.Service{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: owner.GetNamespace()}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Service", "Namespace", owner.GetNamespace(), "Name", serviceName)
		return k8sClient.Create(ctx, desired)
	} else if err != nil {
		return err
	}

	logger.Info("Skip reconcile: Service already exists", "Namespace", found.Namespace, "Name", found.Name)
	return nil
}

func separatePortsByProtocol(ports []corev1.ServicePort) (tcpPorts []corev1.ServicePort, udpPorts []corev1.ServicePort) {
	for _, port := range ports {
		if port.Protocol == corev1.ProtocolTCP {
			tcpPorts = append(tcpPorts, port)
		} else if port.Protocol == corev1.ProtocolUDP {
			udpPorts = append(udpPorts, port)
		}
	}
	return tcpPorts, udpPorts
}

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
