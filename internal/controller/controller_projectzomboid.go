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

// ProjectZomboidReconciler reconciles a ProjectZomboid object
type ProjectZomboidReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=projectzomboids,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=projectzomboids/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=projectzomboids/finalizers,verbs=update

//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ProjectZomboidReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("projectzomboid", req.Name)

	instance := &gameserverv1alpha1.ProjectZomboid{}

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
					pvc.OwnerReferences = nil // Remove all owner refs, or filter
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
			logger.Error(err, "Failed to add finalizer")
			return reconcile.Result{}, err
		}
		logger.Info("Added finalizer")
		return reconcile.Result{Requeue: true}, nil
	}

	// Normal reconciliation
	if err := ReconcilePVC(ctx, r.Client, instance, &instance.Spec.Persistence); err != nil {
		return reconcile.Result{}, err
	}

	configMapName := instance.Name + "-configmap"
	configData := r.generateProjectZomboidConfigData(instance)
	if err := ReconcileConfigMap(ctx, r.Client, instance, configMapName, configData); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.reconcileDeployment(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}

	if err := ReconcileServices(ctx, r.Client, instance, instance.Spec.Ports, instance.Spec.LoadBalancerIP); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ProjectZomboidReconciler) reconcileDeployment(ctx context.Context, instance *gameserverv1alpha1.ProjectZomboid) error {
	logger := log.FromContext(ctx)

	k8sResource := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-deployment",
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
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
						getProjectZomboidSetupInitContainer(),
					},
					Containers: []corev1.Container{
						getSecureGameServerContainer("server", instance.Spec.Image, instance.Spec.Resources, []corev1.ContainerPort{
							{
								ContainerPort: 16261,
								Name:          "port-16261-tcp",
								Protocol:      corev1.ProtocolTCP,
							},
							{
								ContainerPort: 16262,
								Name:          "port-16262-udp",
								Protocol:      corev1.ProtocolUDP,
							},
						}),
						getSecureCodeServerContainer(instance.Spec.EditorPassword),
					},
					Volumes: []corev1.Volume{
						{
							Name: "tmp",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "config-server",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: instance.Name + "-configmap",
									},
									DefaultMode: func(i int32) *int32 { return &i }(0777),
									Items: []corev1.KeyToPath{
										{
											Key:  "server.ini",
											Path: "server.ini",
											Mode: func(i int32) *int32 { return &i }(0777),
										},
									},
								},
							},
						},
						{
							Name: "config-gsm",
							VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: instance.Name + "-configmap",
								},
								DefaultMode: func(i int32) *int32 { return &i }(0777),
								Items: []corev1.KeyToPath{
									{
										Key:  "pzserver.cfg",
										Path: "pzserver.cfg",
										Mode: func(i int32) *int32 { return &i }(0777),
									},
								},
							},
							},
						},
						{
							Name: "data",
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
	} else if err != nil {
		return err
	}

	// Check if the Deployment needs update
	if !CompareDeployments(found, k8sResource) {
		logger.Info("Updating Deployment", "Namespace", found.Namespace, "Name", found.Name)
		found.Spec = k8sResource.Spec
		return r.Update(ctx, found)
	}

	logger.V(4).Info("Deployment already exists and is up to date", "namespace", found.Namespace, "name", found.Name)

	return nil
}

// generateProjectZomboidConfigData creates all necessary configuration files for Project Zomboid
func (r *ProjectZomboidReconciler) generateProjectZomboidConfigData(instance *gameserverv1alpha1.ProjectZomboid) map[string]string {
	configData := make(map[string]string)

	// Generate server.ini (main server configuration)
	configData["server.ini"] = generateProjectZomboidServerConfig(&instance.Spec.Config.Game)

	// Generate LinuxGSM config (if custom)
	if instance.Spec.Config.GSM.ConfigFile != "" {
		configData["pzserver.cfg"] = instance.Spec.Config.GSM.ConfigFile
	} else {
		configData["pzserver.cfg"] = generateProjectZomboidGSMConfig()
	}

	return configData
}

// generateProjectZomboidServerConfig creates server.ini content from CRD spec
func generateProjectZomboidServerConfig(settings *gameserverv1alpha1.ProjectZomboidServerConfig) string {
	var lines []string

	lines = append(lines, fmt.Sprintf("ServerName=%s", settings.ServerName))
	lines = append(lines, fmt.Sprintf("ServerDescription=%s", settings.ServerDescription))

	if settings.Password != "" {
		lines = append(lines, fmt.Sprintf("Password=%s", settings.Password))
	}
	if settings.AdminPassword != "" {
		lines = append(lines, fmt.Sprintf("AdminPassword=%s", settings.AdminPassword))
	}

	lines = append(lines, fmt.Sprintf("MaxPlayers=%d", settings.MaxPlayers))
	lines = append(lines, fmt.Sprintf("DefaultPort=%d", settings.DefaultPort))
	lines = append(lines, fmt.Sprintf("UDPPort=%d", settings.UDPPort))

	if settings.ResetID != 0 {
		lines = append(lines, fmt.Sprintf("ResetID=%d", settings.ResetID))
	}

	lines = append(lines, fmt.Sprintf("SaveWorldEveryMinutes=%d", settings.SaveWorldEveryMinutes))
	lines = append(lines, fmt.Sprintf("PlayerRespawnWithSelf=%d", settings.PlayerRespawnWithSelf))
	lines = append(lines, fmt.Sprintf("PlayerRespawnWithOther=%d", settings.PlayerRespawnWithOther))

	lines = append(lines, fmt.Sprintf("DropOffWhiteListedObjects=%s", boolToString(settings.DropOffWhiteListedObjects)))
	lines = append(lines, fmt.Sprintf("FastForwardMultiplier=%d", settings.FastForwardMultiplier))
	lines = append(lines, fmt.Sprintf("PauseOnEmptyServer=%s", boolToString(settings.PauseOnEmptyServer)))
	lines = append(lines, fmt.Sprintf("MaxAccountsPerUser=%d", settings.MaxAccountsPerUser))

	lines = append(lines, fmt.Sprintf("PVP=%s", boolToString(settings.PVP)))
	lines = append(lines, fmt.Sprintf("SafehouseAllowRespawn=%s", boolToString(settings.SafehouseAllowRespawn)))
	lines = append(lines, fmt.Sprintf("SafehouseAllowTrepass=%s", boolToString(settings.SafehouseAllowTrepass)))

	lines = append(lines, fmt.Sprintf("SleepAllowed=%s", boolToString(settings.SleepAllowed)))
	lines = append(lines, fmt.Sprintf("DamageMultiplier=%s", settings.DamageMultiplier))
	lines = append(lines, fmt.Sprintf("BleedingChance=%s", settings.BleedingChance))
	lines = append(lines, fmt.Sprintf("MinutesPerPage=%s", settings.MinutesPerPage))

	lines = append(lines, fmt.Sprintf("HoursForLootRespawn=%s", settings.HoursForLootRespawn))
	lines = append(lines, fmt.Sprintf("MaxItemsForLootRespawn=%s", settings.MaxItemsForLootRespawn))
	lines = append(lines, fmt.Sprintf("ConstructionPreRequisites=%s", boolToString(settings.ConstructionPreRequisites)))

	lines = append(lines, fmt.Sprintf("Nutrition=%s", boolToString(settings.Nutrition)))
	lines = append(lines, fmt.Sprintf("FoodRotSpeed=%s", settings.FoodRotSpeed))
	lines = append(lines, fmt.Sprintf("WorldEraseSpeed=%d", settings.WorldEraseSpeed))

	lines = append(lines, fmt.Sprintf("PlayerSafehouseCooldown=%s", settings.PlayerSafehouseCooldown))
	lines = append(lines, fmt.Sprintf("AdminSafehouseCooldown=%s", settings.AdminSafehouseCooldown))
	lines = append(lines, fmt.Sprintf("SafehouseDaySurvivor=%s", settings.SafehouseDaySurvivor))

	lines = append(lines, fmt.Sprintf("RemoveExpiredZombies=%s", boolToString(settings.RemoveExpiredZombies)))
	lines = append(lines, fmt.Sprintf("SafehouseAllowDestroy=%s", boolToString(settings.SafehouseAllowDestroy)))
	lines = append(lines, fmt.Sprintf("AllowDestructionBySledgehammer=%s", boolToString(settings.AllowDestructionBySledgehammer)))

	lines = append(lines, fmt.Sprintf("MinutesPerDay=%s", settings.MinutesPerDay))
	lines = append(lines, fmt.Sprintf("ZombieLureDistance=%s", settings.ZombieLureDistance))
	lines = append(lines, fmt.Sprintf("ZombieLureInterval=%d", settings.ZombieLureInterval))

	lines = append(lines, fmt.Sprintf("GlobalChat=%s", boolToString(settings.GlobalChat)))
	lines = append(lines, fmt.Sprintf("ChatStreams=%d", settings.ChatStreams))

	if settings.ServerWelcomeMessage != "" {
		lines = append(lines, fmt.Sprintf("ServerWelcomeMessage=%s", settings.ServerWelcomeMessage))
	}

	lines = append(lines, fmt.Sprintf("OpenWhitelistMod=%s", settings.OpenWhitelistMod))
	lines = append(lines, fmt.Sprintf("BannedPlayerKickedTime=%d", settings.BannedPlayerKickedTime))

	if settings.ServerPlayerID != 0 {
		lines = append(lines, fmt.Sprintf("ServerPlayerID=%d", settings.ServerPlayerID))
	}

	lines = append(lines, fmt.Sprintf("PingLimit=%d", settings.PingLimit))

	if settings.WorkshopItems != "" {
		lines = append(lines, fmt.Sprintf("WorkshopItems=%s", settings.WorkshopItems))
	}
	if settings.Mods != "" {
		lines = append(lines, fmt.Sprintf("Mods=%s", settings.Mods))
	}

	lines = append(lines, fmt.Sprintf("Map=%s", settings.Map))
	lines = append(lines, fmt.Sprintf("ZombiePopulation=%d", settings.ZombiePopulation))
	lines = append(lines, fmt.Sprintf("ZombieMigrateDistance=%d", settings.ZombieMigrateDistance))
	lines = append(lines, fmt.Sprintf("ZombieRespawnRate=%d", settings.ZombieRespawnRate))
	lines = append(lines, fmt.Sprintf("ZombieRespawnPeriod=%d", settings.ZombieRespawnPeriod))

	return strings.Join(lines, "\n")
}

// generateProjectZomboidGSMConfig creates LinuxGSM config file content
func generateProjectZomboidGSMConfig() string {
	return `# LinuxGSM configuration for Project Zomboid
# Generated by GameServer Operator

# Server details
servicename="pzserver"
appid="108600"

# SteamCMD Branch
branch="-beta"

# Required ports
port="16261"
queryport="16261"

# Notification alerts
# (email and other alerts can be configured here)`
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectZomboidReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Temporarily disabled webhooks due to certificate issues
	// if err := (&ProjectZomboidValidator{}).SetupWebhookWithManager(mgr); err != nil {
	//	return err
	// }

	return ctrl.NewControllerManagedBy(mgr).
		For(&gameserverv1alpha1.ProjectZomboid{}).
		Complete(r)
}
