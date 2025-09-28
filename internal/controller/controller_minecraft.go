/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gameserverv1alpha1 "github.com/templarfelix/gameserver-operator/api/v1alpha1"
)

// MinecraftReconciler reconciles a Minecraft object
type MinecraftReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=minecrafts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=minecrafts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=minecrafts/finalizers,verbs=update

//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Add RBAC for networking resources to fix permission warnings
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *MinecraftReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("minecraft", req.Name)

	instance := &gameserverv1alpha1.Minecraft{}

	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	const finalizer = "gameserver.templarfelix.com/finalizer"

	if instance.DeletionTimestamp != nil {
		if controllerutil.ContainsFinalizer(instance, finalizer) {
			// Perform cleanup
			pvcName := instance.Name + "-pvc"
			pvc := &corev1.PersistentVolumeClaim{}
			err := r.Get(ctx, types.NamespacedName{Name: pvcName, Namespace: instance.Namespace}, pvc)
			if err != nil && !errors.IsNotFound(err) {
				logger.Error(err, "Failed to get PVC")
				return reconcile.Result{}, err
			}
			if err == nil { // PVC exists
				if instance.Spec.Persistence.PreserveOnDelete {
					// Remove owner reference to preserve PVC
					pvc.OwnerReferences = nil // Remove all owner refs
					if err := r.Update(ctx, pvc); err != nil {
						logger.Error(err, "Failed to remove owner reference from PVC")
						return reconcile.Result{}, err
					}
					logger.Info("Preserved PVC by removing owner reference")
				} // else let GC delete it
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(instance, finalizer)
			if err := r.Update(ctx, instance); err != nil {
				logger.Error(err, "Failed to remove finalizer")
				return reconcile.Result{}, err
			}
			logger.Info("Finalizer removed, resources will be cleaned up")
			return reconcile.Result{}, nil
		}

		// No finalizer present during deletion, proceed to delete
		return reconcile.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(instance, finalizer) {
		controllerutil.AddFinalizer(instance, finalizer)
		if err := r.Update(ctx, instance); err != nil {
			// Handle concurrent modification conflicts by requeueing
			if errors.IsConflict(err) {
				logger.Info("Conflict adding finalizer, requeueing")
				return reconcile.Result{Requeue: true}, nil
			}
			logger.Error(err, "Failed to add finalizer")
			return reconcile.Result{}, err
		}
		logger.Info("Added finalizer")
		return reconcile.Result{Requeue: true}, nil
	}

	// Normal reconciliation
	if err := r.reconcilePVC(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}

	configMapName := instance.Name + "-configmap"
	configData := r.generateMinecraftConfigData(instance)
	if err := ReconcileConfigMap(ctx, r.Client, instance, configMapName, configData); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.reconcileDeployment(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.reconcileServices(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// generateMinecraftConfigData creates all necessary configuration files for Minecraft
func (r *MinecraftReconciler) generateMinecraftConfigData(instance *gameserverv1alpha1.Minecraft) map[string]string {
	configData := make(map[string]string)

	// Generate server.properties
	configData["server.properties"] = generateServerProperties(&instance.Spec.Config.ServerProperties)

	// Generate JVM args
	configData["jvm.args"] = generateJVMArgs(&instance.Spec.Config.JVM)

	// Generate LinuxGSM config
	configData["mcserver.cfg"] = generateMinecraftGSMConfig(&instance.Spec.Config.GSM)

	return configData
}

// generateServerProperties creates server.properties content from CRD spec
func generateServerProperties(props *gameserverv1alpha1.MinecraftServerProperties) string {
	var lines []string

	lines = append(lines, "#Minecraft server properties")
	lines = append(lines, "#Generated by GameServer Operator")
	lines = append(lines, "")

	// Server configuration
	if props.ServerPort != 0 {
		lines = append(lines, fmt.Sprintf("server-port=%d", props.ServerPort))
	}
	if props.Motd != "" {
		lines = append(lines, fmt.Sprintf("motd=%s", props.Motd))
	}

	// Game settings
	if props.MaxPlayers != 0 {
		lines = append(lines, fmt.Sprintf("max-players=%d", props.MaxPlayers))
	}
	if props.LevelSeed != "" {
		lines = append(lines, fmt.Sprintf("level-seed=%s", props.LevelSeed))
	}
	if props.LevelType != "" {
		lines = append(lines, fmt.Sprintf("level-type=%s", props.LevelType))
	}

	// Difficulty and game mode
	if props.Difficulty != "" {
		lines = append(lines, fmt.Sprintf("difficulty=%s", props.Difficulty))
	}
	if props.GameMode != "" {
		lines = append(lines, fmt.Sprintf("gamemode=%s", props.GameMode))
	}

	// Booleans
	if props.Pvp != nil {
		lines = append(lines, fmt.Sprintf("pvp=%t", *props.Pvp))
	}
	if props.OnlineMode != nil {
		lines = append(lines, fmt.Sprintf("online-mode=%t", *props.OnlineMode))
	}
	if props.AllowNether != nil {
		lines = append(lines, fmt.Sprintf("allow-nether=%t", *props.AllowNether))
	}
	if props.AllowEnd != nil {
		lines = append(lines, fmt.Sprintf("allow-end=%t", *props.AllowEnd))
	}
	if props.EnforceWhitelist != nil {
		lines = append(lines, fmt.Sprintf("enforce-whitelist=%t", *props.EnforceWhitelist))
	}
	if props.EnableCommandBlocks != nil {
		lines = append(lines, fmt.Sprintf("enable-command-blocks=%t", *props.EnableCommandBlocks))
	}

	// Performance settings
	if props.ViewDistance != 0 {
		lines = append(lines, fmt.Sprintf("view-distance=%d", props.ViewDistance))
	}
	if props.SimulationDistance != 0 {
		lines = append(lines, fmt.Sprintf("simulation-distance=%d", props.SimulationDistance))
	}

	// Resource pack settings
	if props.ResourcePack != "" {
		lines = append(lines, fmt.Sprintf("resource-pack=%s", props.ResourcePack))
	}
	if props.ResourcePackSha1 != "" {
		lines = append(lines, fmt.Sprintf("resource-pack-sha1=%s", props.ResourcePackSha1))
	}

	return strings.Join(lines, "\n")
}

// generateJVMArgs creates JVM arguments file content
func generateJVMArgs(jvm *gameserverv1alpha1.JVMConfig) string {
	var args []string

	// Default JVM settings for Minecraft
	args = append(args, "#JVM arguments for Minecraft server")
	args = append(args, "#Generated by GameServer Operator")

	// Memory settings
	if jvm.MinHeapSize != "" {
		args = append(args, fmt.Sprintf("-Xms%s", jvm.MinHeapSize))
	}
	if jvm.MaxHeapSize != "" {
		args = append(args, fmt.Sprintf("-Xmx%s", jvm.MaxHeapSize))
	}

	// Additional arguments
	if jvm.ExtraArgs != "" {
		extraArgs := strings.Fields(jvm.ExtraArgs)
		args = append(args, extraArgs...)
	}

	return strings.Join(args, " ")
}

// generateMinecraftGSMConfig creates LinuxGSM config file content
func generateMinecraftGSMConfig(gsm *gameserverv1alpha1.MinecraftGSMConfig) string {
	if gsm.ConfigFile != "" {
		return gsm.ConfigFile
	}

	// Default LinuxGSM config for Minecraft
	return `# LinuxGSM configuration for Minecraft
# Generated by GameServer Operator

# Minecraft Java settings
startparameters="-Dlog4j2.formatMsgNoLookups=true"

# Server details
servicename="mcserver"
appid="0"

# Notification alerts
# (email and other alerts can be configured here)`

}

// reconcilePVC wraps ReconcilePVC with logging for concurrency conflicts
func (r *MinecraftReconciler) reconcilePVC(ctx context.Context, instance *gameserverv1alpha1.Minecraft) error {
	logger := log.FromContext(ctx)
	if err := ReconcilePVC(ctx, r.Client, instance, &instance.Spec.Persistence); err != nil {
		// Log concurrent modification conflicts
		if errors.IsConflict(err) {
			logger.Info("PVC conflict detected, will retry")
		}
		return err
	}
	return nil
}

// reconcileServices wraps ReconcileServices with logging for concurrency conflicts
func (r *MinecraftReconciler) reconcileServices(ctx context.Context, instance *gameserverv1alpha1.Minecraft) error {
	logger := log.FromContext(ctx)
	if err := ReconcileServices(ctx, r.Client, instance, instance.Spec.Ports, instance.Spec.LoadBalancerIP); err != nil {
		// Log concurrent modification conflicts
		if errors.IsConflict(err) {
			logger.Info("Services conflict detected, will retry")
		}
		return err
	}
	return nil
}

func (r *MinecraftReconciler) reconcileDeployment(ctx context.Context, instance *gameserverv1alpha1.Minecraft) error {
	logger := log.FromContext(ctx)

	// Generate container ports dynamically from CRD ports
	var containerPorts []corev1.ContainerPort
	for _, port := range instance.Spec.Ports {
		containerPort := int32(port.TargetPort.IntValue())
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: containerPort,
			Name:          port.Name,
			Protocol:      port.Protocol,
		})
	}

	// Calculate JVM memory requirements based on requested heap
	jvmMemoryReq := "2Gi"   // Default
	jvmMemoryLimit := "4Gi" // Default

	if instance.Spec.Config.JVM.MaxHeapSize != "" {
		// Add some overhead to JVM heap for JVM itself (about 1.5x JVM heap)
		switch instance.Spec.Config.JVM.MaxHeapSize {
		case "1G":
			jvmMemoryReq = "1536Mi"
			jvmMemoryLimit = "2560Mi"
		case "2G":
			jvmMemoryReq = "3Gi"
			jvmMemoryLimit = "4Gi"
		case "4G":
			jvmMemoryReq = "6Gi"
			jvmMemoryLimit = "8Gi"
		case "8G":
			jvmMemoryReq = "12Gi"
			jvmMemoryLimit = "16Gi"
		}
	}

	// Override resources if JVM config is specified
	resources := instance.Spec.Resources
	if jvmMemoryLimit != "4Gi" {
		resources.Limits = corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(jvmMemoryLimit),
		}
		resources.Requests = corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(jvmMemoryReq),
		}
	}

	k8sResource := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-deployment",
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { r := int32(1); return &r }(),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": instance.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": instance.Name},
				},
				Spec: corev1.PodSpec{
					NodeSelector: instance.Spec.NodeSelector,
					Tolerations:  instance.Spec.Tolerations,
					Affinity:     instance.Spec.Affinity,

					InitContainers: []corev1.Container{
						getMinecraftSetupInitContainer(),
					},
					Containers: []corev1.Container{
						getSecureGameServerContainer("server", instance.Spec.Image, resources, containerPorts),
						getSecureCodeServerContainer(instance.Spec.EditorPassword),
					},
					Volumes: []corev1.Volume{
						{
							Name: ConfigsVolumeName, // Unified config volume
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: instance.Name + "-configmap",
									},
									DefaultMode: func(i int32) *int32 { return &i }(0777),
								},
							},
						},
						{
							Name: DataVolumeName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: instance.Name + "-pvc",
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(instance, k8sResource, r.Scheme); err != nil {
		return err
	}

	found := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Name: k8sResource.Name, Namespace: k8sResource.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Deployment", "Namespace", k8sResource.Namespace, "Name", k8sResource.Name)
		err = r.Create(ctx, k8sResource)
		if err != nil {
			return err
		}
		return nil // Don't update immediately after creation
	} else if err != nil {
		return err
	}

	// Check if the Deployment needs update
	if !CompareDeployments(found, k8sResource) {
		logger.Info("Updating Deployment", "Namespace", found.Namespace, "Name", found.Name)
		found.Spec = k8sResource.Spec
		if err := r.Update(ctx, found); err != nil {
			if errors.IsConflict(err) {
				logger.Info("Conflict updating deployment, will retry")
			}
			return err
		}
	}

	logger.V(4).Info("Deployment already exists and is up to date", "namespace", found.Namespace, "name", found.Name)

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MinecraftReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gameserverv1alpha1.Minecraft{}).
		Complete(r)
}
