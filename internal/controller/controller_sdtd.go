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

// SdtdReconciler reconciles a Sdtd object
type SdtdReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=sdtds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=sdtds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=sdtds/finalizers,verbs=update

//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Add RBAC for networking resources to fix permission warnings
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Dayz object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *SdtdReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("sdtd", req.Name)

	instance := &gameserverv1alpha1.Sdtd{}

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
	configData := r.generateSdtdConfigData(instance)
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

// reconcilePVC wraps ReconcilePVC with logging for concurrency conflicts
func (r *SdtdReconciler) reconcilePVC(ctx context.Context, instance *gameserverv1alpha1.Sdtd) error {
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
func (r *SdtdReconciler) reconcileServices(ctx context.Context, instance *gameserverv1alpha1.Sdtd) error {
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

func (r *SdtdReconciler) reconcileDeployment(ctx context.Context, instance *gameserverv1alpha1.Sdtd) error {
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
						getSdtdSetupInitContainer(),
					},
					Containers: []corev1.Container{
						getSecureGameServerContainer("server", instance.Spec.Image, instance.Spec.Resources, containerPorts),
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

// generateSdtdConfigData creates all necessary configuration files for 7 Days to Die
func (r *SdtdReconciler) generateSdtdConfigData(instance *gameserverv1alpha1.Sdtd) map[string]string {
	configData := make(map[string]string)

	// Generate serverconfig.xml (server settings)
	configData["serverconfig.xml"] = generateSdtdServerConfig(&instance.Spec.Config.Game)

	// Generate LinuxGSM config (if custom)
	if instance.Spec.Config.GSM.ConfigFile != "" {
		configData["sdtdserver.cfg"] = instance.Spec.Config.GSM.ConfigFile
	} else {
		configData["sdtdserver.cfg"] = generateSdtdGSMConfig()
	}

	return configData
}

// generateSdtdServerConfig creates serverconfig.xml content from CRD spec
func generateSdtdServerConfig(settings *gameserverv1alpha1.SdtdServerConfig) string {
	var properties []string

	// Add XML header
	properties = append(properties, "<?xml version=\"1.0\"?>")
	properties = append(properties, "<ServerSettings>")

	// GENERAL SERVER SETTINGS
	properties = append(properties, "  <!-- GENERAL SERVER SETTINGS -->")

	// Server representation
	properties = append(properties, fmt.Sprintf("  <property name=\"ServerName\" value=\"%s\" />", settings.ServerName))
	properties = append(properties, fmt.Sprintf("  <property name=\"ServerDescription\" value=\"%s\" />", settings.ServerDescription))

	if settings.ServerWebsiteURL != "" {
		properties = append(properties, fmt.Sprintf("  <property name=\"ServerWebsiteURL\" value=\"%s\" />", settings.ServerWebsiteURL))
	}

	if settings.ServerPassword != "" {
		properties = append(properties, fmt.Sprintf("  <property name=\"ServerPassword\" value=\"%s\" />", settings.ServerPassword))
	}

	if settings.ServerLoginConfirmationText != "" {
		properties = append(properties, fmt.Sprintf("  <property name=\"ServerLoginConfirmationText\" value=\"%s\" />", settings.ServerLoginConfirmationText))
	}

	properties = append(properties, fmt.Sprintf("  <property name=\"Region\" value=\"%s\" />", settings.Region))
	properties = append(properties, fmt.Sprintf("  <property name=\"Language\" value=\"%s\" />", settings.Language))

	// Networking
	properties = append(properties, fmt.Sprintf("  <property name=\"ServerPort\" value=\"%d\" />", settings.ServerPort))
	properties = append(properties, fmt.Sprintf("  <property name=\"ServerVisibility\" value=\"%d\" />", settings.ServerVisibility))
	properties = append(properties, fmt.Sprintf("  <property name=\"ServerDisabledNetworkProtocols\" value=\"%s\" />", settings.ServerDisabledNetworkProtocols))
	properties = append(properties, fmt.Sprintf("  <property name=\"ServerMaxWorldTransferSpeedKiBs\" value=\"%d\" />", settings.ServerMaxWorldTransferSpeedKiBs))

	// Slots
	properties = append(properties, fmt.Sprintf("  <property name=\"ServerMaxPlayerCount\" value=\"%d\" />", settings.ServerMaxPlayerCount))
	properties = append(properties, fmt.Sprintf("  <property name=\"ServerReservedSlots\" value=\"%d\" />", settings.ServerReservedSlots))
	properties = append(properties, fmt.Sprintf("  <property name=\"ServerReservedSlotsPermission\" value=\"%d\" />", settings.ServerReservedSlotsPermission))
	properties = append(properties, fmt.Sprintf("  <property name=\"ServerAdminSlots\" value=\"%d\" />", settings.ServerAdminSlots))
	properties = append(properties, fmt.Sprintf("  <property name=\"ServerAdminSlotsPermission\" value=\"%d\" />", settings.ServerAdminSlotsPermission))

	// Admin interfaces
	properties = append(properties, fmt.Sprintf("  <property name=\"WebDashboardEnabled\" value=\"%t\" />", settings.WebDashboardEnabled))
	properties = append(properties, fmt.Sprintf("  <property name=\"WebDashboardPort\" value=\"%d\" />", settings.WebDashboardPort))
	if settings.WebDashboardUrl != "" {
		properties = append(properties, fmt.Sprintf("  <property name=\"WebDashboardUrl\" value=\"%s\" />", settings.WebDashboardUrl))
	}
	properties = append(properties, fmt.Sprintf("  <property name=\"EnableMapRendering\" value=\"%t\" />", settings.EnableMapRendering))

	properties = append(properties, fmt.Sprintf("  <property name=\"TelnetEnabled\" value=\"%t\" />", settings.TelnetEnabled))
	properties = append(properties, fmt.Sprintf("  <property name=\"TelnetPort\" value=\"%d\" />", settings.TelnetPort))
	if settings.TelnetPassword != "" {
		properties = append(properties, fmt.Sprintf("  <property name=\"TelnetPassword\" value=\"%s\" />", settings.TelnetPassword))
	}
	properties = append(properties, fmt.Sprintf("  <property name=\"TelnetFailedLoginLimit\" value=\"%d\" />", settings.TelnetFailedLoginLimit))
	properties = append(properties, fmt.Sprintf("  <property name=\"TelnetFailedLoginsBlocktime\" value=\"%d\" />", settings.TelnetFailedLoginsBlocktime))

	properties = append(properties, fmt.Sprintf("  <property name=\"TerminalWindowEnabled\" value=\"%t\" />", settings.TerminalWindowEnabled))

	// Folder and file locations
	properties = append(properties, fmt.Sprintf("  <property name=\"AdminFileName\" value=\"%s\" />", settings.AdminFileName))

	// Technical settings
	properties = append(properties, fmt.Sprintf("  <property name=\"EACEnabled\" value=\"%t\" />", settings.EACEnabled))
	properties = append(properties, fmt.Sprintf("  <property name=\"HideCommandExecutionLog\" value=\"%d\" />", settings.HideCommandExecutionLog))
	properties = append(properties, fmt.Sprintf("  <property name=\"MaxUncoveredMapChunksPerPlayer\" value=\"%d\" />", settings.MaxUncoveredMapChunksPerPlayer))
	properties = append(properties, fmt.Sprintf("  <property name=\"PersistentPlayerProfiles\" value=\"%t\" />", settings.PersistentPlayerProfiles))

	// GAMEPLAY - World
	properties = append(properties, "  <!-- World -->")
	properties = append(properties, fmt.Sprintf("  <property name=\"GameWorld\" value=\"%s\" />", settings.GameWorld))
	if settings.GameWorldSeed != "" {
		properties = append(properties, fmt.Sprintf("  <property name=\"WorldGenSeed\" value=\"%s\" />", settings.GameWorldSeed))
	}
	properties = append(properties, fmt.Sprintf("  <property name=\"WorldGenSize\" value=\"%d\" />", settings.GameWorldSize))
	if settings.GameName != "" {
		properties = append(properties, fmt.Sprintf("  <property name=\"GameName\" value=\"%s\" />", settings.GameName))
	}
	properties = append(properties, fmt.Sprintf("  <property name=\"GameMode\" value=\"%s\" />", settings.GameMode))

	// Difficulty
	properties = append(properties, "  <!-- Difficulty -->")
	properties = append(properties, fmt.Sprintf("  <property name=\"GameDifficulty\" value=\"%d\" />", settings.GameDifficulty))
	properties = append(properties, fmt.Sprintf("  <property name=\"BlockDamagePlayer\" value=\"%d\" />", settings.BlockDamagePlayer))
	properties = append(properties, fmt.Sprintf("  <property name=\"BlockDamageAI\" value=\"%d\" />", settings.BlockDamageAI))
	properties = append(properties, fmt.Sprintf("  <property name=\"BlockDamageAIBM\" value=\"%d\" />", settings.BlockDamageAIBM))
	properties = append(properties, fmt.Sprintf("  <property name=\"XPMultiplier\" value=\"%d\" />", settings.XPMultiplier))
	properties = append(properties, fmt.Sprintf("  <property name=\"PlayerSafeZoneLevel\" value=\"%d\" />", settings.PlayerSafeZoneLevel))
	properties = append(properties, fmt.Sprintf("  <property name=\"PlayerSafeZoneHours\" value=\"%d\" />", settings.PlayerSafeZoneHours))

	properties = append(properties, fmt.Sprintf("  <property name=\"BuildCreate\" value=\"%t\" />", settings.BuildCreate))
	properties = append(properties, fmt.Sprintf("  <property name=\"DayNightLength\" value=\"%d\" />", settings.DayNightLength))
	properties = append(properties, fmt.Sprintf("  <property name=\"DayLightLength\" value=\"%d\" />", settings.DayLightLength))
	properties = append(properties, fmt.Sprintf("  <property name=\"DropOnDeath\" value=\"%d\" />", settings.DropOnDeath))
	properties = append(properties, fmt.Sprintf("  <property name=\"DropOnQuit\" value=\"%d\" />", settings.DropOnQuit))
	properties = append(properties, fmt.Sprintf("  <property name=\"BedrollDeadZoneSize\" value=\"%d\" />", settings.BedrollDeadZoneSize))
	properties = append(properties, fmt.Sprintf("  <property name=\"BedrollExpiryTime\" value=\"%d\" />", settings.BedrollExpiryTime))

	// Performance
	properties = append(properties, fmt.Sprintf("  <property name=\"MaxSpawnedZombies\" value=\"%d\" />", settings.MaxSpawnedZombies))
	properties = append(properties, fmt.Sprintf("  <property name=\"MaxSpawnedAnimals\" value=\"%d\" />", settings.MaxSpawnedAnimals))
	properties = append(properties, fmt.Sprintf("  <property name=\"ServerMaxAllowedViewDistance\" value=\"%d\" />", settings.ServerMaxAllowedViewDistance))
	properties = append(properties, fmt.Sprintf("  <property name=\"MaxQueuedMeshLayers\" value=\"%d\" />", settings.MaxQueuedMeshLayers))

	// Zombie settings
	properties = append(properties, "  <!-- Zombie settings -->")
	properties = append(properties, fmt.Sprintf("  <property name=\"EnemySpawnMode\" value=\"%t\" />", settings.EnemySpawnMode))
	properties = append(properties, fmt.Sprintf("  <property name=\"EnemyDifficulty\" value=\"%d\" />", settings.EnemyDifficulty))
	properties = append(properties, fmt.Sprintf("  <property name=\"ZombieFeralSense\" value=\"%d\" />", settings.ZombieFeralSense))
	properties = append(properties, fmt.Sprintf("  <property name=\"ZombieMove\" value=\"%d\" />", settings.ZombieMove))
	properties = append(properties, fmt.Sprintf("  <property name=\"ZombieMoveNight\" value=\"%d\" />", settings.ZombieMoveNight))
	properties = append(properties, fmt.Sprintf("  <property name=\"ZombieFeralMove\" value=\"%d\" />", settings.ZombieFeralMove))
	properties = append(properties, fmt.Sprintf("  <property name=\"ZombieBMMove\" value=\"%d\" />", settings.ZombieBMMove))
	properties = append(properties, fmt.Sprintf("  <property name=\"BloodMoonFrequency\" value=\"%d\" />", settings.BloodMoonFrequency))
	properties = append(properties, fmt.Sprintf("  <property name=\"BloodMoonRange\" value=\"%d\" />", settings.BloodMoonRange))
	properties = append(properties, fmt.Sprintf("  <property name=\"BloodMoonWarning\" value=\"%d\" />", settings.BloodMoonWarning))
	properties = append(properties, fmt.Sprintf("  <property name=\"BloodMoonEnemyCount\" value=\"%d\" />", settings.BloodMoonEnemyCount))

	// Loot settings
	properties = append(properties, "  <!-- Loot -->")
	properties = append(properties, fmt.Sprintf("  <property name=\"LootAbundance\" value=\"%d\" />", settings.LootAbundance))
	properties = append(properties, fmt.Sprintf("  <property name=\"LootRespawnDays\" value=\"%d\" />", settings.LootRespawnDays))
	properties = append(properties, fmt.Sprintf("  <property name=\"AirDropFrequency\" value=\"%d\" />", settings.AirDropFrequency))
	properties = append(properties, fmt.Sprintf("  <property name=\"AirDropMarker\" value=\"%t\" />", settings.AirDropMarker))

	// Multiplayer
	properties = append(properties, "  <!-- Multiplayer -->")
	properties = append(properties, fmt.Sprintf("  <property name=\"PartySharedKillRange\" value=\"%d\" />", settings.PartySharedKillRange))
	properties = append(properties, fmt.Sprintf("  <property name=\"PlayerKillingMode\" value=\"%d\" />", settings.PlayerKillingMode))

	// Land claim options
	properties = append(properties, "  <!-- Land claim options -->")
	properties = append(properties, fmt.Sprintf("  <property name=\"LandClaimCount\" value=\"%d\" />", settings.LandClaimCount))
	properties = append(properties, fmt.Sprintf("  <property name=\"LandClaimSize\" value=\"%d\" />", settings.LandClaimSize))
	properties = append(properties, fmt.Sprintf("  <property name=\"LandClaimDeadZone\" value=\"%d\" />", settings.LandClaimDeadZone))
	properties = append(properties, fmt.Sprintf("  <property name=\"LandClaimExpiryTime\" value=\"%d\" />", settings.LandClaimExpiryTime))
	properties = append(properties, fmt.Sprintf("  <property name=\"LandClaimDecayMode\" value=\"%d\" />", settings.LandClaimDecayMode))
	properties = append(properties, fmt.Sprintf("  <property name=\"LandClaimOnlineDurabilityModifier\" value=\"%d\" />", settings.LandClaimOnlineDurabilityModifier))
	properties = append(properties, fmt.Sprintf("  <property name=\"LandClaimOfflineDurabilityModifier\" value=\"%d\" />", settings.LandClaimOfflineDurabilityModifier))
	properties = append(properties, fmt.Sprintf("  <property name=\"LandClaimOfflineDelay\" value=\"%d\" />", settings.LandClaimOfflineDelay))

	// Dynamic mesh
	properties = append(properties, fmt.Sprintf("  <property name=\"DynamicMeshEnabled\" value=\"%t\" />", settings.DynamicMeshEnabled))
	properties = append(properties, fmt.Sprintf("  <property name=\"DynamicMeshLandClaimOnly\" value=\"%t\" />", settings.DynamicMeshLandClaimOnly))
	properties = append(properties, fmt.Sprintf("  <property name=\"DynamicMeshLandClaimBuffer\" value=\"%d\" />", settings.DynamicMeshLandClaimBuffer))
	properties = append(properties, fmt.Sprintf("  <property name=\"DynamicMeshMaxItemCache\" value=\"%d\" />", settings.DynamicMeshMaxItemCache))

	properties = append(properties, fmt.Sprintf("  <property name=\"TwitchServerPermission\" value=\"%d\" />", settings.TwitchServerPermission))
	properties = append(properties, fmt.Sprintf("  <property name=\"TwitchBloodMoonAllowed\" value=\"%t\" />", settings.TwitchBloodMoonAllowed))

	properties = append(properties, fmt.Sprintf("  <property name=\"MaxChunkAge\" value=\"%d\" />", settings.MaxChunkAge))
	properties = append(properties, fmt.Sprintf("  <property name=\"SaveDataLimit\" value=\"%d\" />", settings.SaveDataLimit))

	properties = append(properties, "</ServerSettings>")

	return strings.Join(properties, "\n")
}

// generateSdtdGSMConfig creates LinuxGSM config file content
func generateSdtdGSMConfig() string {
	return `# LinuxGSM configuration for 7 Days to Die
# Generated by GameServer Operator

# Server details
servicename="sdtdserver"
appid="294420"
executable="./7DaysToDieServer.x86_64"

# Server parameters
startparameters="-quit -batchmode -nographics -dedicated -configfile=${servercfgfullpath}"

# Required ports
port="26900"

# Server cfg name
servercfg="serverconfig.xml"
servercfgfullpath="${servercfgdir}/${servercfg}"

# Update permissions
steamcmdforcewindows="no"

# Stop command
stopmode="8"

# Query settings
querymode="2"
querytype="protocol-valve"

# Console settings
consoleverbose="yes"
consoleinteract="no"

# Notification alerts
# (email and other alerts can be configured here)`
}

// SetupWithManager sets up the controller with the Manager.
func (r *SdtdReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Temporarily disabled webhooks due to certificate issues
	// if err := (&DayzValidator{}).SetupWebhookWithManager(mgr); err != nil {
	//	return err
	// }

	return ctrl.NewControllerManagedBy(mgr).
		For(&gameserverv1alpha1.Dayz{}).
		Complete(r)
}
