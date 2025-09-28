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
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gameserverv1alpha1 "github.com/templarfelix/gameserver-operator/api/v1alpha1"
)

// GmodReconciler reconciles a Gmod object
type GmodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=gmods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=gmods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=gmods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Gmod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *GmodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get the Gmod instance
	instance := &gameserverv1alpha1.Gmod{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Gmod resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Gmod")
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling Garry's Mod server", "name", instance.Name)

	// Reconcile PVC
	if err := r.reconcilePVC(ctx, instance); err != nil {
		logger.Error(err, "Failed to reconcile PVC")
		return ctrl.Result{}, err
	}

	// Reconcile Services
	if err := r.reconcileServices(ctx, instance); err != nil {
		logger.Error(err, "Failed to reconcile Services")
		return ctrl.Result{}, err
	}

	// Reconcile Deployment
	podSpec, err := r.generateDeploymentSpec(instance)
	if err != nil {
		logger.Error(err, "Failed to generate deployment spec")
		return ctrl.Result{}, err
	}

	if err := r.reconcileDeployment(ctx, instance, podSpec); err != nil {
		logger.Error(err, "Failed to reconcile Deployment")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled Garry's Mod server", "name", instance.Name)
	return ctrl.Result{}, nil
}

// reconcilePVC creates or updates the PVC for game data storage
func (r *GmodReconciler) reconcilePVC(ctx context.Context, instance *gameserverv1alpha1.Gmod) error {
	logger := log.FromContext(ctx)

	size := resource.MustParse(instance.Spec.Persistence.StorageConfig.Size)
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-pvc",
			Namespace: instance.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: size,
				},
			},
		},
	}

	if instance.Spec.Persistence.StorageConfig.StorageClassName != "" {
		pvc.Spec.StorageClassName = &instance.Spec.Persistence.StorageConfig.StorageClassName
	}

	if err := controllerutil.SetControllerReference(instance, pvc, r.Scheme); err != nil {
		return err
	}

	found := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, client.ObjectKey{Name: pvc.Name, Namespace: pvc.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating PVC", "Namespace", pvc.Namespace, "Name", pvc.Name)
		err = r.Create(ctx, pvc)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Check if PVC needs update (only if not preserveOnDelete)
	if !instance.Spec.Persistence.PreserveOnDelete {
		found.Spec = pvc.Spec
		err = r.Update(ctx, found)
		if err != nil {
			return err
		}
	}

	logger.V(4).Info("PVC is up to date", "namespace", found.Namespace, "name", found.Name)
	return nil
}

// reconcileServices creates or updates the Services for the game server
func (r *GmodReconciler) reconcileServices(ctx context.Context, instance *gameserverv1alpha1.Gmod) error {
	logger := log.FromContext(ctx)

	services := []*corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.Name + "-tcp",
				Namespace: instance.Namespace,
				Labels: map[string]string{
					"app":       instance.Name + "-gmod",
					"component": "gameserver",
				},
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{
						Name:       "game",
						Port:       instance.Spec.Ports[0].Port,
						TargetPort: intstr.FromInt(int(instance.Spec.Ports[0].Port)),
						Protocol:   corev1.ProtocolTCP,
					},
				},
				Selector: map[string]string{
					"app": instance.Name + "-gmod",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.Name + "-udp",
				Namespace: instance.Namespace,
				Labels: map[string]string{
					"app":       instance.Name + "-gmod",
					"component": "gameserver",
				},
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{
						Name:       "game",
						Port:       instance.Spec.Ports[0].Port,
						TargetPort: intstr.FromInt(int(instance.Spec.Ports[0].Port)),
						Protocol:   corev1.ProtocolUDP,
					},
				},
				Selector: map[string]string{
					"app": instance.Name + "-gmod",
				},
			},
		},
	}

	if instance.Spec.LoadBalancerIP != "" {
		for _, svc := range services {
			svc.Spec.LoadBalancerIP = instance.Spec.LoadBalancerIP
		}
	}

	for _, svc := range services {
		if err := controllerutil.SetControllerReference(instance, svc, r.Scheme); err != nil {
			return err
		}

		found := &corev1.Service{}
		err := r.Get(ctx, client.ObjectKey{Name: svc.Name, Namespace: svc.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			logger.Info("Creating Service", "Namespace", svc.Namespace, "Name", svc.Name)
			err = r.Create(ctx, svc)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			// Update service if needed
			found.Spec.Ports = svc.Spec.Ports
			if found.Spec.LoadBalancerIP != svc.Spec.LoadBalancerIP {
				found.Spec.LoadBalancerIP = svc.Spec.LoadBalancerIP
			}
			err = r.Update(ctx, found)
			if err != nil {
				return err
			}
		}
	}

	logger.V(4).Info("Services reconciled", "count", len(services))
	return nil
}

// generateDeploymentSpec creates the Pod spec for the Gmod deployment
func (r *GmodReconciler) generateDeploymentSpec(instance *gameserverv1alpha1.Gmod) (*corev1.PodSpec, error) {

	// Default environment variables
	envVars := []corev1.EnvVar{
		{
			Name:  "PUID",
			Value: "1000",
		},
		{
			Name:  "PGID",
			Value: "1000",
		},
		{
			Name:  "TZ",
			Value: "UTC",
		},
	}

	var tolerations []corev1.Toleration
	var affinity *corev1.Affinity
	var nodeSelector map[string]string
	var annotations map[string]string

	if instance.Spec.Resources.Requests == nil {
		instance.Spec.Resources.Requests = corev1.ResourceList{}
	}
	if instance.Spec.Resources.Limits == nil {
		instance.Spec.Resources.Limits = corev1.ResourceList{}
	}

	if instance.Spec.Tolerations != nil {
		tolerations = instance.Spec.Tolerations
	}
	if instance.Spec.Affinity != nil {
		affinity = instance.Spec.Affinity
	}
	if instance.Spec.NodeSelector != nil {
		nodeSelector = instance.Spec.NodeSelector
	}
	if instance.Spec.Annotations != nil {
		annotations = instance.Spec.Annotations
	}

	podSpec := &corev1.PodSpec{
		Tolerations:  tolerations,
		Affinity:     affinity,
		NodeSelector: nodeSelector,
		Containers: []corev1.Container{
			{
				Name:  "gmod-server",
				Image: instance.Spec.Image,
				Ports: []corev1.ContainerPort{
					{
						Name:          "game-tcp",
						ContainerPort: instance.Spec.Ports[0].Port,
						Protocol:      corev1.ProtocolTCP,
					},
					{
						Name:          "game-udp",
						ContainerPort: instance.Spec.Ports[0].Port,
						Protocol:      corev1.ProtocolUDP,
					},
				},
				Env: envVars,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "data",
						MountPath: "/data",
					},
					{
						Name:      "configs",
						MountPath: "/data/serverfiles/garrysmod/cfg",
					},
					{
						Name:      "linuxgsm-configs",
						MountPath: "/data/serverfiles",
					},
				},
				Resources: instance.Spec.Resources,
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: instance.Name + "-pvc",
					},
				},
			},
			{
				Name: "configs",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: instance.Name + "-config",
						},
					},
				},
			},
			{
				Name: "linuxgsm-configs",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: instance.Name + "-gsm-config",
						},
					},
				},
			},
		},
	}

	if annotations != nil {
		podSpec.Containers[0].Env = append(envVars, corev1.EnvVar{
			Name:  "EDITOR_PASSWORD",
			Value: instance.Spec.EditorPassword,
		})
	}

	return podSpec, nil
}

// reconcileDeployment creates or updates the Deployment for the game server
func (r *GmodReconciler) reconcileDeployment(ctx context.Context, instance *gameserverv1alpha1.Gmod, podSpec *corev1.PodSpec) error {
	logger := log.FromContext(ctx)

	// Create ConfigMaps for server configs
	if err := r.reconcileConfigMaps(ctx, instance); err != nil {
		return err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-deployment",
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app":       instance.Name + "-gmod",
				"component": "gameserver",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { i := int32(1); return &i }(),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": instance.Name + "-gmod",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": instance.Name + "-gmod",
					},
				},
				Spec: *podSpec,
			},
		},
	}

	if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
		return err
	}

	found := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Name: deployment.Name, Namespace: deployment.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Deployment", "Namespace", deployment.Namespace, "Name", deployment.Name)
		err = r.Create(ctx, deployment)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Check if the Deployment needs update
	if !CompareDeployments(found, deployment) {
		logger.Info("Updating Deployment", "Namespace", found.Namespace, "Name", found.Name)
		found.Spec = deployment.Spec
		return r.Update(ctx, found)
	}

	logger.V(4).Info("Deployment already exists and is up to date", "namespace", found.Namespace, "name", found.Name)

	return nil
}

// reconcileConfigMaps creates ConfigMaps for server configuration files
func (r *GmodReconciler) reconcileConfigMaps(ctx context.Context, instance *gameserverv1alpha1.Gmod) error {
	logger := log.FromContext(ctx)

	// Server config ConfigMap (server.cfg)
	serverConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-config",
			Namespace: instance.Namespace,
		},
		Data: r.generateGmodConfigData(instance),
	}

	if err := controllerutil.SetControllerReference(instance, serverConfig, r.Scheme); err != nil {
		return err
	}

	found := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{Name: serverConfig.Name, Namespace: serverConfig.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating server config ConfigMap", "Namespace", serverConfig.Namespace, "Name", serverConfig.Name)
		err = r.Create(ctx, serverConfig)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// Update ConfigMap if needed
		found.Data = serverConfig.Data
		err = r.Update(ctx, found)
		if err != nil {
			return err
		}
	}

	// LinuxGSM config ConfigMap (gmodserver.cfg)
	gsmConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-gsm-config",
			Namespace: instance.Namespace,
		},
		Data: map[string]string{
			"gmodserver.cfg": generateGmodGSMConfig(instance),
		},
	}

	if err := controllerutil.SetControllerReference(instance, gsmConfig, r.Scheme); err != nil {
		return err
	}

	found = &corev1.ConfigMap{}
	err = r.Get(ctx, client.ObjectKey{Name: gsmConfig.Name, Namespace: gsmConfig.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating GSM config ConfigMap", "Namespace", gsmConfig.Namespace, "Name", gsmConfig.Name)
		err = r.Create(ctx, gsmConfig)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// Update ConfigMap if needed
		found.Data = gsmConfig.Data
		err = r.Update(ctx, found)
		if err != nil {
			return err
		}
	}

	return nil
}

// generateGmodConfigData creates all necessary configuration files for Garry's Mod
func (r *GmodReconciler) generateGmodConfigData(instance *gameserverv1alpha1.Gmod) map[string]string {
	configData := make(map[string]string)

	// Generate server.cfg (main server configuration)
	configData["server.cfg"] = generateGmodServerConfig(&instance.Spec.Config.Game)

	return configData
}

// generateGmodServerConfig creates server.cfg content from CRD spec
func generateGmodServerConfig(settings *gameserverv1alpha1.GmodServerConfig) string {
	var lines []string

	lines = append(lines, "// Garry's Mod Server Configuration")
	lines = append(lines, "// Generated by GameServer Operator")

	lines = append(lines, fmt.Sprintf("hostname \"%s\"", settings.Hostname))

	if settings.ServerDescription != "" {
		lines = append(lines, fmt.Sprintf("// %s", settings.ServerDescription))
	}

	if settings.Password != "" {
		lines = append(lines, fmt.Sprintf("sv_password \"%s\"", settings.Password))
	}

	lines = append(lines, fmt.Sprintf("sv_region %d", settings.Region))
	lines = append(lines, fmt.Sprintf("gamemode %s", settings.GameMode))
	lines = append(lines, fmt.Sprintf("map %s", settings.DefaultMap))
	lines = append(lines, fmt.Sprintf("sv_lan %d", settings.LANOnly))
	lines = append(lines, fmt.Sprintf("maxplayers %d", settings.MaxPlayers))

	// Lua settings
	if settings.AllowCSLua != nil {
		lines = append(lines, fmt.Sprintf("sv_allowcslua %s", boolToString(*settings.AllowCSLua)))
	}

	// Workshop collection
	if settings.WorkshopCollectionID != "" {
		lines = append(lines, fmt.Sprintf("host_workshop_collection \"%s\"", settings.WorkshopCollectionID))
	}

	// Sandbox settings
	if settings.FallDamage != nil {
		lines = append(lines, fmt.Sprintf("mp_falldamage %s", boolToString(*settings.FallDamage)))
	}

	if settings.AllowVehicles != nil {
		lines = append(lines, fmt.Sprintf("sbox_maxvehicles %d", boolToInt(*settings.AllowVehicles)*6)) // max 6 vehicles
		lines = append(lines, fmt.Sprintf("sbox_noclip %s", boolToString(*settings.AllowVehicles)))
	}

	if settings.AllowNPCs != nil {
		lines = append(lines, fmt.Sprintf("sbox_maxnpcs %d", boolToInt(*settings.AllowNPCs)*10)) // max 10 NPCs
		lines = append(lines, fmt.Sprintf("sbox_godmode %s", boolToString(false)))               // NPCs don't need godmode
	}

	if settings.AllowWeapons != nil {
		lines = append(lines, fmt.Sprintf("sbox_weapons %s", boolToString(*settings.AllowWeapons)))
	}

	if settings.AllowProps != nil {
		lines = append(lines, fmt.Sprintf("sbox_maxprops %d", settings.MaxPropsPerPlayer))
	}

	if settings.AllowEffects != nil {
		lines = append(lines, fmt.Sprintf("sbox_maxeffects 50")) // arbitrary limit
	}

	if settings.AllowRagdolls != nil {
		lines = append(lines, fmt.Sprintf("sbox_maxragdolls 10")) // arbitrary limit
	}

	// Admin settings
	if settings.AdminSteamIDs != "" {
		lines = append(lines, fmt.Sprintf("ulx_superadmins \"%s\"", settings.AdminSteamIDs))
	}

	if settings.ServerOwnerSteamID != "" {
		lines = append(lines, fmt.Sprintf("sv_owner \"%s\"", settings.ServerOwnerSteamID))
	}

	// Mod-specific settings
	if settings.M9KWeapons != nil && *settings.M9KWeapons {
		lines = append(lines, "// M9K Weapons enabled")
		lines = append(lines, "m9k_enable 1")
	}

	if settings.Ads != nil && !*settings.Ads {
		lines = append(lines, "enable_ads 0")
	}

	if settings.DarkRPEconomy != nil && !*settings.DarkRPEconomy {
		lines = append(lines, "darkrp_economy 0")
	}

	// General settings
	if settings.MapCycle != nil {
		lines = append(lines, fmt.Sprintf("mp_mapcycle_enable %s", boolToString(*settings.MapCycle)))
	}

	if settings.MapCycleFile != "" {
		lines = append(lines, fmt.Sprintf("mapcyclefile \"%s\"", settings.MapCycleFile))
	}

	if settings.VoiceIcon != nil {
		lines = append(lines, fmt.Sprintf("mp_show_voice_icon %s", boolToString(*settings.VoiceIcon)))
	}

	if settings.Alltalk != nil {
		lines = append(lines, fmt.Sprintf("sv_alltalk %s", boolToString(*settings.Alltalk)))
	}

	lines = append(lines, fmt.Sprintf("sv_voice_distance %d", settings.VoiceDistance))

	if settings.FastDLURL != "" {
		lines = append(lines, fmt.Sprintf("sv_downloadurl \"%s\"", settings.FastDLURL))
	}

	if settings.MOTD != "" {
		lines = append(lines, fmt.Sprintf("motd \"%s\"", settings.MOTD))
	}

	if settings.PvP != nil {
		lines = append(lines, fmt.Sprintf("sbox_playershurtplayers %s", boolToString(*settings.PvP)))
	}

	lines = append(lines, fmt.Sprintf("sv_logfile %s", boolToString(settings.LogFile != nil && *settings.LogFile)))
	lines = append(lines, fmt.Sprintf("sv_log_onefile %s", boolToString(false)))

	if settings.LoggingLevel > 0 {
		lines = append(lines, fmt.Sprintf("sv_logecho %s", boolToString(settings.LoggingLevel >= 2)))
		lines = append(lines, fmt.Sprintf("sv_logflush %s", boolToString(settings.LoggingLevel >= 3)))
	}

	if settings.ServerTags != "" {
		lines = append(lines, fmt.Sprintf("sv_tags \"%s\"", settings.ServerTags))
	}

	return strings.Join(lines, "\n")
}

// generateGmodGSMConfig creates LinuxGSM config file content
func generateGmodGSMConfig(instance *gameserverv1alpha1.Gmod) string {
	if instance.Spec.Config.GSM.ConfigFile != "" {
		return instance.Spec.Config.GSM.ConfigFile
	}

	var lines []string

	lines = append(lines, "# LinuxGSM configuration for Garry's Mod")
	lines = append(lines, "# Generated by GameServer Operator")
	lines = append(lines, "")

	lines = append(lines, "# Server details")
	lines = append(lines, "servicename=\"gmodserver\"")
	lines = append(lines, "appid=\"4020\"")
	lines = append(lines, "")

	lines = append(lines, "# Server ports")
	lines = append(lines, fmt.Sprintf("port=\"%d\"", instance.Spec.Config.GSM.Port))
	lines = append(lines, fmt.Sprintf("clientport=\"%d\"", instance.Spec.Config.GSM.ClientPort))
	lines = append(lines, fmt.Sprintf("sourcetvport=\"%d\"", instance.Spec.Config.GSM.SourceTVPort))
	lines = append(lines, "")

	lines = append(lines, "# Game settings")
	lines = append(lines, fmt.Sprintf("tickrate=\"%d\"", instance.Spec.Config.GSM.TickRate))
	lines = append(lines, fmt.Sprintf("gamemode=\"%s\"", instance.Spec.Config.Game.GameMode))
	lines = append(lines, fmt.Sprintf("maxplayers=\"%d\"", instance.Spec.Config.Game.MaxPlayers))
	lines = append(lines, fmt.Sprintf("defaultmap=\"%s\"", instance.Spec.Config.Game.DefaultMap))
	lines = append(lines, "")

	// Workshop collection (prefer GSM config over server.cfg)
	collectionID := instance.Spec.Config.GSM.WorkshopCollectionID
	if collectionID == "" {
		collectionID = instance.Spec.Config.Game.WorkshopCollectionID
	}
	lines = append(lines, fmt.Sprintf("wscollectionid=\"%s\"", collectionID))
	lines = append(lines, "")

	// GSLT
	if instance.Spec.Config.GSM.GSLT != "" {
		lines = append(lines, fmt.Sprintf("gslt=\"%s\"", instance.Spec.Config.GSM.GSLT))
	} else {
		lines = append(lines, "gslt=\"\"")
	}

	lines = append(lines, "")

	lines = append(lines, "# LinuxGSM Stats")
	if instance.Spec.Config.GSM.Stats != nil && *instance.Spec.Config.GSM.Stats {
		lines = append(lines, "stats=\"on\"")
	} else {
		lines = append(lines, "stats=\"off\"")
	}

	lines = append(lines, "")

	lines = append(lines, "# Notification Alerts")
	lines = append(lines, "# (on|off)")
	lines = append(lines, "postalert=\"off\"")
	lines = append(lines, "statusalert=\"off\"")

	return strings.Join(lines, "\n")
}

// SetupWithManager sets up the controller with the Manager.
func (r *GmodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Temporarily disabled webhooks due to certificate issues
	// if err := (&GmodValidator{}).SetupWebhookWithManager(mgr); err != nil {
	//	return err
	// }

	return ctrl.NewControllerManagedBy(mgr).
		For(&gameserverv1alpha1.Gmod{}).
		Complete(r)
}
