package controller

import (
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
)

func ReconcilePVC(ctx context.Context, k8sClient client.Client, owner metav1.Object, storage string) error {
	logger := log.FromContext(ctx)
	pvcName := owner.GetName() + "-pvc"

	desired := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: owner.GetNamespace(),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(storage),
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(owner, desired, k8sClient.Scheme()); err != nil {
		return err
	}

	// Verifica se o PVC já existe
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

func ReconcileServices(ctx context.Context, k8sClient client.Client, owner metav1.Object, ports []corev1.ServicePort) error {
	// Add code-server port to the service
	allPorts := append(ports, corev1.ServicePort{
		Name:       "code-server",
		Port:       8080,
		TargetPort: intstr.FromInt32(8080),
		Protocol:   corev1.ProtocolTCP,
	})

	return reconcileService(ctx, owner.GetName()+"-lb", k8sClient, owner, allPorts)
}

func reconcileService(ctx context.Context, serviceName string, k8sClient client.Client, owner metav1.Object, ports []corev1.ServicePort) error {
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
			Type:  corev1.ServiceTypeLoadBalancer,
			Ports: ports,
		},
	}

	if err := controllerutil.SetControllerReference(owner, desired, k8sClient.Scheme()); err != nil {
		return err
	}

	found := &corev1.Service{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: owner.GetNamespace()}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Service", "Namespace", owner.GetNamespace(), "Name", desired)
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
