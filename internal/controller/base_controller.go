package controller

import (
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func ReconcilePVC(ctx context.Context, k8sClient client.Client, owner metav1.Object, storage string) error {
	logger := log.FromContext(ctx)
	pvcName := owner.GetName() + "-pvc"

	// Define o PVC desejado
	desired := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: owner.GetNamespace(),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(storage),
				},
			},
		},
	}

	// Define o objeto como proprietário do PVC
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

	// Se o PVC já existir, loga que está pulando a criação e retorna nil
	logger.Info("Skip reconcile: PVC already exists", "Namespace", found.Namespace, "Name", found.Name)
	return nil
}

func ReconcileServices(ctx context.Context, k8sClient client.Client, owner metav1.Object, ports []corev1.ServicePort) error {
	logger := log.FromContext(ctx)
	serviceNameTCP := owner.GetName() + "-tcp"
	serviceNameUDP := owner.GetName() + "-tcp"

	tcpPorts, udpPorts := separatePortsByProtocol(ports)

	// TCP
	desiredTCP := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceNameTCP,
			Namespace: owner.GetNamespace(),
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": owner.GetName(),
			},
			Ports: tcpPorts,
		},
	}

	// TCP
	desiredUDP := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceNameUDP,
			Namespace: owner.GetNamespace(),
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": owner.GetName(),
			},
			Ports: udpPorts,
		},
	}

	// TCP
	if err := controllerutil.SetControllerReference(owner, desiredTCP, k8sClient.Scheme()); err != nil {
		return err
	}

	foundTCP := &corev1.PersistentVolumeClaim{}
	errTCP := k8sClient.Get(ctx, types.NamespacedName{Name: serviceNameTCP, Namespace: owner.GetNamespace()}, foundTCP)
	if errTCP != nil && errors.IsNotFound(errTCP) {
		logger.Info("Creating a new Service", "Namespace", owner.GetNamespace(), "Name", desiredTCP)
		return k8sClient.Create(ctx, desiredTCP)
	} else if errTCP != nil {
		return errTCP
	}
	logger.Info("Skip reconcile: Service already exists", "Namespace", foundTCP.Namespace, "Name", foundTCP.Name)

	// UDP
	if err := controllerutil.SetControllerReference(owner, desiredUDP, k8sClient.Scheme()); err != nil {
		return err
	}

	foundUDP := &corev1.PersistentVolumeClaim{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: serviceNameUDP, Namespace: owner.GetNamespace()}, foundUDP)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Service", "Namespace", owner.GetNamespace(), "Name", desiredTCP)
		return k8sClient.Create(ctx, desiredTCP)
	} else if err != nil {
		return err
	}

	logger.Info("Skip reconcile: Service already exists", "Namespace", foundUDP.Namespace, "Name", foundUDP.Name)
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
