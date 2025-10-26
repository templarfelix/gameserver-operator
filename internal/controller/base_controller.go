package controller

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gcpaddress "github.com/upbound/provider-gcp/apis/compute/v1beta1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	"github.com/go-logr/logr"
	gameserverv1 "github.com/templarfelix/gameserver-operator/api/v1"
)

// initializeDefaultPersistence ensures that all persistence fields have safe defaults
func initializeDefaultPersistence(persistence *gameserverv1.Persistence, logger logr.Logger, ownerName string) {
	// Ensure storageConfig is initialized with defaults
	if persistence.StorageConfig.Size == "" {
		logger.V(4).Info("Setting default storage size", "owner", ownerName, "size", "10G")
		persistence.StorageConfig.Size = "10G"
	}

	// Validate that the size can be parsed
	if _, err := resource.ParseQuantity(persistence.StorageConfig.Size); err != nil {
		logger.Error(err, "Invalid storage size, using default", "owner", ownerName, "size", persistence.StorageConfig.Size)
		persistence.StorageConfig.Size = "10G"
	}
}

// ReconcilePVC creates or updates a PersistentVolumeClaim for game data storage
func ReconcilePVC(ctx context.Context, k8sClient client.Client, owner metav1.Object, persistence *gameserverv1.Persistence) error {
	logger := log.FromContext(ctx)
	pvcName := owner.GetName() + "-pvc"

	// Safety check for nil persistence
	if persistence == nil {
		logger.Error(nil, "Persistence configuration is nil", "owner", owner.GetName())
		return fmt.Errorf("persistence configuration cannot be nil")
	}

	// Initialize defaults and validate configuration
	initializeDefaultPersistence(persistence, logger, owner.GetName())

	// Create desired PVC spec
	storageSize := persistence.StorageConfig.Size
	parsedSize, err := resource.ParseQuantity(storageSize)
	if err != nil {
		logger.Error(err, "Invalid storage size, using default", "size", storageSize)
		parsedSize, _ = resource.ParseQuantity("10G") // This should not fail
	}

	var storageClassName *string
	if persistence.StorageConfig.StorageClassName != "" {
		storageClassName = &persistence.StorageConfig.StorageClassName
	}

	desired := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: owner.GetNamespace(),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			StorageClassName: storageClassName,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: parsedSize,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(owner, desired, k8sClient.Scheme()); err != nil {
		return err
	}

	// Check if PVC already exists
	found := &corev1.PersistentVolumeClaim{}
	err = k8sClient.Get(ctx, types.NamespacedName{Name: pvcName, Namespace: owner.GetNamespace()}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating new PVC", "namespace", owner.GetNamespace(), "name", pvcName)
		return k8sClient.Create(ctx, desired)
	} else if err != nil {
		return err
	}

	logger.V(4).Info("PVC already exists", "namespace", found.Namespace, "name", found.Name)
	return nil
}

// ReconcileServices creates or updates Services for exposing the game server
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

	serviceSpec := corev1.ServiceSpec{
		Selector: map[string]string{
			"app": owner.GetName(),
		},
		Type:                  corev1.ServiceTypeLoadBalancer,
		Ports:                 ports,
		ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyLocal,
	}

	// Only set LoadBalancerIP if it's not empty
	if loadBalancerIP != "" {
		serviceSpec.LoadBalancerIP = loadBalancerIP
	}

	desired := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: owner.GetNamespace(),
			Annotations: map[string]string{
				"cloud.google.com/load-balancer-type": "External",
			},
		},
		Spec: serviceSpec,
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
		switch port.Protocol {
		case corev1.ProtocolTCP:
			tcpPorts = append(tcpPorts, port)
		case corev1.ProtocolUDP:
			udpPorts = append(udpPorts, port)
		}
	}
	return tcpPorts, udpPorts
}

// validatePrerequisites checks if all prerequisites for creating ComputeAddress are met
func validatePrerequisites(ctx context.Context, k8sClient client.Client, providerConfigName string) error {
	logger := log.FromContext(ctx)
	
	// For now, just log that we're validating - the actual validation will happen during Address creation
	// TODO: Add proper ProviderConfig validation in the future
	logger.V(1).Info("Validating prerequisites for ComputeAddress creation", "providerConfig", providerConfigName)
	
	return nil
}

// CreateGCPComputeAddress creates a GCP ComputeAddress resource using the new provider
// This function creates a ComputeAddress resource for static IP allocation in GCP
func CreateGCPComputeAddress(ctx context.Context, k8sClient client.Client, owner metav1.Object, name string) error {
	logger := log.FromContext(ctx)
	
	providerConfigName := "default" // TODO: Make this configurable in the future
	
	// Validate prerequisites first
	if err := validatePrerequisites(ctx, k8sClient, providerConfigName); err != nil {
		logger.Error(err, "Prerequisites validation failed", "name", name)
		return err
	}

	// Check if ComputeAddress already exists
	existing := &gcpaddress.Address{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: owner.GetNamespace()}, existing)
	if err == nil {
		logger.V(4).Info("ComputeAddress already exists", "name", name)
		return nil
	}
	if !errors.IsNotFound(err) {
		// Check if it's a CRD not found error (common when Upbound Provider is not installed)
		errMsg := err.Error()
		// Log with Info level to ensure it's visible
		logger.Info("🔍 DEBUGGING ComputeAddress Check Error", 
			"name", name, 
			"error", errMsg,
			"errorType", fmt.Sprintf("%T", err),
			"namespace", owner.GetNamespace(),
			"fullError", err.Error())
		
		// Also log as Error for consistency
		logger.Error(err, "Failed to check existing ComputeAddress - DETAILED ERROR", 
			"name", name, 
			"error", errMsg,
			"errorType", fmt.Sprintf("%T", err),
			"namespace", owner.GetNamespace())
		
		if strings.Contains(errMsg, "no matches for kind") || 
		   strings.Contains(errMsg, "could not find the requested resource") ||
		   strings.Contains(errMsg, "Address") {
			return fmt.Errorf("Upbound Provider GCP CRDs not installed. Install with: kubectl apply -f https://raw.githubusercontent.com/upbound/provider-gcp/main/package/crds/compute.gcp.upbound.io_addresses.yaml. Original error: %w", err)
		}
		if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "timeout") {
			return fmt.Errorf("Kubernetes API connection issue. Check if cluster is accessible. Original error: %w", err)
		}
		if strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "unauthorized") {
			return fmt.Errorf("RBAC permissions issue. Check if operator has permissions to access Address resources. Original error: %w", err)
		}
		
		return fmt.Errorf("failed to check existing ComputeAddress %s: %w", name, err)
	}

	// Create the ComputeAddress object using the new provider type
	addressType := "EXTERNAL"
	region := "southamerica-east1" // TODO: Make this configurable in the future

	computeAddress := &gcpaddress.Address{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: owner.GetNamespace(),
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":      "gameserver-operator",
				"gameserver.templarfelix.com/owner": owner.GetName(),
			},
		},
		Spec: gcpaddress.AddressSpec{
			ForProvider: gcpaddress.AddressParameters{
				AddressType: &addressType,
				Region:      &region,
			},
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{
					Name: providerConfigName,
				},
			},
		},
	}
	
	logger.Info("Creating ComputeAddress", "name", name, "region", region, "providerConfig", providerConfigName)

	// Set the owner reference so the ComputeAddress is cleaned up when the owner is deleted
	if err := controllerutil.SetControllerReference(owner, computeAddress, k8sClient.Scheme()); err != nil {
		logger.Error(err, "Failed to set controller reference", "name", name)
		return err
	}

	// Try to create the ComputeAddress
	logger.Info("Creating new ComputeAddress", "name", name, "region", region, "providerConfig", providerConfigName)
	
	// Debug: Log the complete object being created
	logger.V(1).Info("ComputeAddress object details", 
		"name", computeAddress.Name,
		"namespace", computeAddress.Namespace,
		"addressType", *computeAddress.Spec.ForProvider.AddressType,
		"region", *computeAddress.Spec.ForProvider.Region,
		"providerConfigRef", computeAddress.Spec.ResourceSpec.ProviderConfigReference.Name)
	
	err = k8sClient.Create(ctx, computeAddress)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			logger.V(4).Info("ComputeAddress already exists (race condition)", "name", name)
			return nil
		}
		
		// Enhanced error logging
		errMsg := err.Error()
		logger.Error(err, "Failed to create ComputeAddress - detailed error", 
			"name", name, 
			"error", errMsg,
			"region", region,
			"providerConfig", providerConfigName)
		
		// Check for common error patterns and provide specific solutions
		if strings.Contains(errMsg, "ProviderConfig") || strings.Contains(errMsg, "providerconfig") {
			return fmt.Errorf("failed to create ComputeAddress %s: ProviderConfig '%s' not found. Please ensure the ProviderConfig exists: kubectl get providerconfig. Error: %w", name, providerConfigName, err)
		}
		if strings.Contains(errMsg, "no matches for kind") || strings.Contains(errMsg, "Address") {
			return fmt.Errorf("failed to create ComputeAddress %s: Upbound Provider GCP CRDs not installed. Install with: kubectl apply -f https://raw.githubusercontent.com/upbound/provider-gcp/main/package/crds/compute.gcp.upbound.io_addresses.yaml. Error: %w", name, err)
		}
		if strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "credentials") || strings.Contains(errMsg, "unauthorized") {
			return fmt.Errorf("failed to create ComputeAddress %s: GCP authentication failed. Check service account credentials in secret 'gcp-secret'. Error: %w", name, err)
		}
		if strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "permission") {
			return fmt.Errorf("failed to create ComputeAddress %s: Insufficient GCP permissions. Service account needs 'compute.addresses.create' permission. Error: %w", name, err)
		}
		if strings.Contains(errMsg, "project") {
			return fmt.Errorf("failed to create ComputeAddress %s: GCP project issue. Check if project ID is correct in ProviderConfig. Error: %w", name, err)
		}
		if strings.Contains(errMsg, "region") || strings.Contains(errMsg, "location") {
			return fmt.Errorf("failed to create ComputeAddress %s: Invalid region '%s'. Check if region exists in your GCP project. Error: %w", name, region, err)
		}
		
		return fmt.Errorf("failed to create ComputeAddress %s: %w", name, err)
	}

	logger.Info("Successfully created ComputeAddress", "name", name)
	return nil
}

// IsGCPComputeAddressReady checks if a GCP ComputeAddress resource is ready and has an IP
func IsGCPComputeAddressReady(ctx context.Context, k8sClient client.Client, name, namespace string) (bool, string, error) {
	logger := log.FromContext(ctx)

	// Get the ComputeAddress object
	computeAddress := &gcpaddress.Address{}

	err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, computeAddress)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(4).Info("ComputeAddress not found", "name", name, "namespace", namespace)
			return false, "", nil
		}
		logger.Error(err, "Failed to get ComputeAddress", "name", name, "namespace", namespace)
		return false, "", fmt.Errorf("failed to get ComputeAddress %s/%s: %w", namespace, name, err)
	}

	// Log current status for debugging
	logger.V(4).Info("ComputeAddress status", "name", name, "conditions", len(computeAddress.Status.Conditions))

	// Check if the resource is ready by examining the status conditions
	for _, condition := range computeAddress.Status.Conditions {
		logger.V(4).Info("Checking condition", "type", string(condition.Type), "status", string(condition.Status), "reason", condition.Reason)
		if condition.Type == "Ready" {
			if condition.Status == "True" {
				// Ready, return the address if available
				if computeAddress.Status.AtProvider.Address != nil {
					ip := *computeAddress.Status.AtProvider.Address
					logger.Info("ComputeAddress is ready with IP", "name", name, "ip", ip)
					return true, ip, nil
				}
				logger.Info("ComputeAddress is ready but no IP assigned yet", "name", name)
				return true, "", nil
			} else {
				logger.Info("ComputeAddress not ready", "name", name, "status", string(condition.Status), "reason", condition.Reason, "message", condition.Message)
				return false, "", nil
			}
		}
	}

	logger.V(4).Info("No Ready condition found", "name", name)
	return false, "", nil
}
