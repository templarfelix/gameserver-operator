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
	"strconv"
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

// ArkReconciler reconciles a Ark object
type ArkReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=arks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=arks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=arks/finalizers,verbs=update

//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Add RBAC for networking resources to fix permission warnings
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ArkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("ark", req.Name)

	instance := &gameserverv1alpha1.Ark{}

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
	configData := r.generateArkConfigData(instance)
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

// generateArkConfigData creates all necessary configuration files for ARK
func (r *ArkReconciler) generateArkConfigData(instance *gameserverv1alpha1.Ark) map[string]string {
	configData := make(map[string]string)

	// Generate GameUserSettings.ini (basic server settings)
	configData["GameUserSettings.ini"] = generateGameUserSettings(&instance.Spec.Config.GameUserSettings)

	// Generate Game.ini (advanced game settings)
	configData["Game.ini"] = generateGameIni(&instance.Spec.Config.Game)

	// Generate LinuxGSM config (if custom)
	if instance.Spec.Config.GSM.ConfigFile != "" {
		configData["arkserver.cfg"] = instance.Spec.Config.GSM.ConfigFile
	} else {
		configData["arkserver.cfg"] = generateArkGSMConfig()
	}

	return configData
}

// generateGameUserSettings creates GameUserSettings.ini content from CRD spec
func generateGameUserSettings(settings *gameserverv1alpha1.ArkGameUserSettings) string {
	var lines []string

	lines = append(lines, "[ServerSettings]")
	lines = append(lines, "ServerAdminPassword="+settings.AdminPassword)

	// Server settings
	serverSettings := settings.ServerSettings
	if serverSettings.ServerName != "" {
		lines = append(lines, "ServerName="+serverSettings.ServerName)
	}
	if serverSettings.ServerDescription != "" {
		lines = append(lines, "ServerDescription="+serverSettings.ServerDescription)
	}

	// Map and game settings
	lines = append(lines, "Map="+settings.ServerMap)
	lines = append(lines, "MaxPlayers="+strconv.Itoa(int(settings.MaxPlayers)))
	lines = append(lines, "Difficulty="+settings.Difficulty)
	lines = append(lines, "ServerPort="+strconv.Itoa(int(settings.ServerPort)))
	lines = append(lines, "RCONPort="+strconv.Itoa(int(settings.RconPort)))

	// RCON settings
	if settings.EnableRcon != nil {
		lines = append(lines, "RCONEnabled="+boolToString(*settings.EnableRcon))
	}

	// Mod settings
	if settings.ModInstaller != nil {
		lines = append(lines, "ModInstaller="+boolToString(*settings.ModInstaller))
	}
	if settings.Mods != "" {
		lines = append(lines, "Mods="+settings.Mods)
	}

	// Gameplay settings
	if settings.Crossplay != nil {
		lines = append(lines, "Crossplay="+boolToString(*settings.Crossplay))
	}
	if serverSettings.NightTimeSpeed != "" {
		lines = append(lines, "NightTimeSpeedScale="+serverSettings.NightTimeSpeed)
	}
	if serverSettings.DayCycleSpeedScale != "" {
		lines = append(lines, "DayCycleSpeedScale="+serverSettings.DayCycleSpeedScale)
	}

	// PVP settings
	if serverSettings.PVP != nil {
		lines = append(lines, "PVP="+boolToString(*serverSettings.PVP))
	}
	if serverSettings.AllowFriendlyFire != nil {
		lines = append(lines, "AllowFriendlyFire="+boolToString(*serverSettings.AllowFriendlyFire))
	}

	// Damage multipliers
	if serverSettings.TamedDinoDamageMultiplier != "" {
		lines = append(lines, "TamedDinoDamageMultiplier="+serverSettings.TamedDinoDamageMultiplier)
	}
	if serverSettings.WildDinoDamageMultiplier != "" {
		lines = append(lines, "WildDinoDamageMultiplier="+serverSettings.WildDinoDamageMultiplier)
	}
	if serverSettings.PlayerDamageMultiplier != "" {
		lines = append(lines, "PlayerDamageMultiplier="+serverSettings.PlayerDamageMultiplier)
	}
	if serverSettings.StructureDamageMultiplier != "" {
		lines = append(lines, "StructureDamageMultiplier="+serverSettings.StructureDamageMultiplier)
	}
	if serverSettings.PlayerResistance != "" {
		lines = append(lines, "PlayerResistance="+serverSettings.PlayerResistance)
	}

	// Dino settings
	if serverSettings.DinoCountMultiplier != "" {
		lines = append(lines, "DinoCountMultiplier="+serverSettings.DinoCountMultiplier)
	}
	if serverSettings.DinoFoodDrainMultiplier != "" {
		lines = append(lines, "DinoFoodDrainMultiplier="+serverSettings.DinoFoodDrainMultiplier)
	}

	// Other settings
	if serverSettings.AutoSavePeriodMinutes != 0 {
		lines = append(lines, "AutoSavePeriodMinutes="+strconv.Itoa(int(serverSettings.AutoSavePeriodMinutes)))
	}
	if serverSettings.DisablePvPOption {
		lines = append(lines, "DisablePvPOption=true")
	}
	if serverSettings.ServerPVE {
		lines = append(lines, "ServerPVE=true")
	}
	if serverSettings.ShowPlayerMapLocation != nil {
		lines = append(lines, "ShowPlayerMapLocation="+boolToString(*serverSettings.ShowPlayerMapLocation))
	}
	if serverSettings.ShowMapLocation != nil {
		lines = append(lines, "ShowMapPlayerLocation="+boolToString(*serverSettings.ShowMapLocation))
	}

	return strings.Join(lines, "\n")
}

// generateGameIni creates Game.ini content from CRD spec
func generateGameIni(game *gameserverv1alpha1.AdvancedGameSettings) string {
	var sections []string

	// Harvesting rates section
	if hasHarvestingRates(game.Harvesting) {
		lines := []string{"[/Script/ShooterGame.ShooterGameMode]"}

		harvesting := game.Harvesting
		if harvesting.DayCycleSpeedScale != "" {
			lines = append(lines, fmt.Sprintf("DayCycleSpeedScale=%s", harvesting.DayCycleSpeedScale))
		}
		if harvesting.DayTimeSpeedScale != "" {
			lines = append(lines, fmt.Sprintf("DayTimeSpeedScale=%s", harvesting.DayTimeSpeedScale))
		}
		if harvesting.DamageMultiplier != "" {
			lines = append(lines, fmt.Sprintf("HarvestingDamageMultiplier=%s", harvesting.DamageMultiplier))
		}
		if harvesting.UseSinglePlayerDamage {
			lines = append(lines, "UseSinglePlayerHarvestDamage=true")
		}

		// Harvesting quantities
		if harvesting.MetalHarvestQuantity != "" {
			lines = append(lines, fmt.Sprintf("MetalHarvestQuantity=%s", harvesting.MetalHarvestQuantity))
		}
		if harvesting.WoodHarvestQuantity != "" {
			lines = append(lines, fmt.Sprintf("WoodHarvestQuantity=%s", harvesting.WoodHarvestQuantity))
		}
		if harvesting.StoneHarvestQuantity != "" {
			lines = append(lines, fmt.Sprintf("StoneHarvestQuantity=%s", harvesting.StoneHarvestQuantity))
		}
		if harvesting.ThatchHarvestQuantity != "" {
			lines = append(lines, fmt.Sprintf("ThatchHarvestQuantity=%s", harvesting.ThatchHarvestQuantity))
		}
		if harvesting.FlintHarvestQuantity != "" {
			lines = append(lines, fmt.Sprintf("FlintHarvestQuantity=%s", harvesting.FlintHarvestQuantity))
		}
		if harvesting.CrystalHarvestQuantity != "" {
			lines = append(lines, fmt.Sprintf("CrystalHarvestQuantity=%s", harvesting.CrystalHarvestQuantity))
		}
		if harvesting.OilHarvestQuantity != "" {
			lines = append(lines, fmt.Sprintf("OilHarvestQuantity=%s", harvesting.OilHarvestQuantity))
		}
		if harvesting.ObsidianHarvestQuantity != "" {
			lines = append(lines, fmt.Sprintf("ObsidianHarvestQuantity=%s", harvesting.ObsidianHarvestQuantity))
		}

		sections = append(sections, strings.Join(lines, "\n"))
	}

	// Breeding rates section
	if hasBreedingRates(game.Breeding) {
		lines := []string{"[/Script/ShooterGame.ShooterGameMode]"}

		breeding := game.Breeding
		if breeding.MatingIntervalMultiplier != "" {
			lines = append(lines, fmt.Sprintf("MatingIntervalMultiplier=%s", breeding.MatingIntervalMultiplier))
		}
		if breeding.EggHatchSpeedScale != "" {
			lines = append(lines, fmt.Sprintf("EggHatchSpeedScale=%s", breeding.EggHatchSpeedScale))
		}
		if breeding.BabyMaturationSpeedScale != "" {
			lines = append(lines, fmt.Sprintf("BabyMaturationSpeedScale=%s", breeding.BabyMaturationSpeedScale))
		}
		if breeding.ImprintPeriodMultiplier != "" {
			lines = append(lines, fmt.Sprintf("ImprintPeriodMultiplier=%s", breeding.ImprintPeriodMultiplier))
		}
		if breeding.SingleBabyGestation {
			lines = append(lines, "SingleBabyGestation=true")
		}

		sections = append(sections, strings.Join(lines, "\n"))
	}

	// Difficulty settings section
	if hasDifficultySettings(game.Difficulty) {
		lines := []string{"[/Script/ShooterGame.ShooterGameMode]"}

		diff := game.Difficulty
		if diff.OverrideOfficialDifficulty != "" {
			lines = append(lines, fmt.Sprintf("OverrideOfficialDifficulty=%s", diff.OverrideOfficialDifficulty))
		}
		if diff.DifficultyLevel != "" {
			lines = append(lines, fmt.Sprintf("DifficultyLevel=%s", diff.DifficultyLevel))
		}
		if diff.MaxDifficulty {
			lines = append(lines, "MaxDifficulty=true")
		}
		if diff.UseSinglePlayerSettings {
			lines = append(lines, "UseSinglePlayerSettings=true")
		}
		if diff.PreventSpawningDinosWithoutSaddle {
			lines = append(lines, "PreventSpawningDinosWithoutSaddle=true")
		}
		if diff.DontUseDifficulty {
			lines = append(lines, "DontUseDifficulty=true")
		}

		sections = append(sections, strings.Join(lines, "\n"))
	}

	// Custom rates section
	if hasCustomRates(game.CustomRates) {
		lines := []string{"[/Script/ShooterGame.ShooterGameMode]"}

		rates := game.CustomRates
		if rates.TamedDinoLevelMultiplier != "" {
			lines = append(lines, fmt.Sprintf("TamedDinoLevelMultiplier=%s", rates.TamedDinoLevelMultiplier))
		}
		if rates.WildDinoLevelMultiplier != "" {
			lines = append(lines, fmt.Sprintf("WildDinoLevelMultiplier=%s", rates.WildDinoLevelMultiplier))
		}
		if rates.XPMultiplier != "" {
			lines = append(lines, fmt.Sprintf("XPMultiplier=%s", rates.XPMultiplier))
		}
		if rates.DayCycleSpeedScale != "" {
			lines = append(lines, fmt.Sprintf("DayCycleSpeedScale=%s", rates.DayCycleSpeedScale))
		}
		if rates.RandomSupplyCratePoints {
			lines = append(lines, "RandomSupplyCratePoints=true")
		}
		if rates.DisablePvPAutoBalance {
			lines = append(lines, "DisablePvPAutoBalance=true")
		}
		if rates.DisableStructurePlacementCollision {
			lines = append(lines, "DisableStructurePlacementCollision=true")
		}

		sections = append(sections, strings.Join(lines, "\n"))
	}

	return strings.Join(sections, "\n")
}

// Helper functions for checking if sections have data
func hasHarvestingRates(h gameserverv1alpha1.HarvestingRates) bool {
	return h.DayCycleSpeedScale != "" || h.DayTimeSpeedScale != "" || h.DamageMultiplier != "" ||
		h.UseSinglePlayerDamage || h.MetalHarvestQuantity != "" || h.WoodHarvestQuantity != "" ||
		h.StoneHarvestQuantity != "" || h.ThatchHarvestQuantity != "" || h.FlintHarvestQuantity != "" ||
		h.CrystalHarvestQuantity != "" || h.OilHarvestQuantity != "" || h.ObsidianHarvestQuantity != ""
}

func hasBreedingRates(b gameserverv1alpha1.BreedingRates) bool {
	return b.MatingIntervalMultiplier != "" || b.EggHatchSpeedScale != "" || b.BabyMaturationSpeedScale != "" ||
		b.ImprintPeriodMultiplier != "" || b.SingleBabyGestation
}

func hasDifficultySettings(d gameserverv1alpha1.DifficultySettings) bool {
	return d.OverrideOfficialDifficulty != "" || d.DifficultyLevel != "" || d.MaxDifficulty ||
		d.UseSinglePlayerSettings || d.PreventSpawningDinosWithoutSaddle || d.DontUseDifficulty
}

func hasCustomRates(c gameserverv1alpha1.CustomRateSettings) bool {
	return c.TamedDinoLevelMultiplier != "" || c.WildDinoLevelMultiplier != "" || c.XPMultiplier != "" ||
		c.DayCycleSpeedScale != "" || c.RandomSupplyCratePoints || c.DisablePvPAutoBalance ||
		c.DisableStructurePlacementCollision
}

// generateArkGSMConfig creates LinuxGSM config file content
func generateArkGSMConfig() string {
	return `# LinuxGSM configuration for ARK
# Generated by GameServer Operator

# Server details
servicename="arkserver"
appid="376030"

# Notification alerts
# (email and other alerts can be configured here)`
}

// reconcilePVC wraps ReconcilePVC with logging for concurrency conflicts
func (r *ArkReconciler) reconcilePVC(ctx context.Context, instance *gameserverv1alpha1.Ark) error {
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
func (r *ArkReconciler) reconcileServices(ctx context.Context, instance *gameserverv1alpha1.Ark) error {
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

func (r *ArkReconciler) reconcileDeployment(ctx context.Context, instance *gameserverv1alpha1.Ark) error {
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
						getArkSetupInitContainer(),
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

// SetupWithManager sets up the controller with the Manager.
func (r *ArkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gameserverv1alpha1.Ark{}).
		Complete(r)
}
