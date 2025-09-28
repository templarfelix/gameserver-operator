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

// Game Server Configuration Constants
const (
	// User and group IDs for game servers (linuxgsm user)
	GameServerUserID  int64 = 1000
	GameServerGroupID int64 = 1000

	// InitContainer configuration
	SetupContainerImage = "alpine:latest"
	SetupContainerName  = "config-setup"

	// Volume names
	ConfigsVolumeName      = "configs"
	TmpVolumeName          = "tmp"
	ConfigServerVolumeName = "config-server"
	ConfigGsmVolumeName    = "config-gsm"
	DataVolumeName         = "data"
)

// getGameServerSecurityContext returns the security context for game server containers
func getGameServerSecurityContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		RunAsUser:  func(i int64) *int64 { return &i }(GameServerUserID),
		RunAsGroup: func(i int64) *int64 { return &i }(GameServerGroupID),
	}
}

// getDayZSetupInitContainer returns an init container specifically for DayZ config setup
// DayZ configuration paths:
// - GSM config: /data/config-lgsm/dayzserver/dayzserver.cfg (LinuxGSM config)
// - Server config: /data/serverfiles/cfg/dayzserver.server.cfg (DayZ server config)
// More info: https://linuxgsm.com/lgsm/dayzserver/
func getDayZSetupInitContainer() corev1.Container {
	return corev1.Container{
		Name:    SetupContainerName,
		Image:   SetupContainerImage,
		Command: []string{"sh", "-c"},
		Args: []string{`
			set -eu

			# Create DayZ specific directories
			mkdir -p /data/config-lgsm/dayzserver /data/serverfiles/cfg

			# Always copy latest DayZ config files (ensures ConfigMap updates are applied)
			cp /configs/dayzserver.cfg /data/config-lgsm/dayzserver/dayzserver.cfg
			cp /configs/dayzserver.server.cfg /data/serverfiles/cfg/dayzserver.server.cfg

			# Set ownership for linuxgsm user (1000:1000)
			chown -R 1000:1000 /data/config-lgsm/dayzserver /data/serverfiles/cfg

			echo "DayZ config setup completed successfully"
		`},
		VolumeMounts: []corev1.VolumeMount{
			{Name: DataVolumeName, MountPath: "/data"},
			{Name: ConfigsVolumeName, MountPath: "/configs"},
		},
	}
}

// getProjectZomboidSetupInitContainer returns an init container specifically for Project Zomboid config setup
// Project Zomboid configuration paths (DIFFERENT from DayZ):
// - GSM config: /data/config-lgsm/pzserver/pzserver.cfg (LinuxGSM config)
// - Server config: /data/serverfiles/server.ini (Project Zomboid server config - note: .ini not .cfg!)
// More info: https://linuxgsm.com/lgsm/pzserver/
// Note: Project Zomboid uses different file structure than DayZ
func getProjectZomboidSetupInitContainer() corev1.Container {
	return corev1.Container{
		Name:    SetupContainerName,
		Image:   "busybox", // Keep busybox for Project Zomboid (lighter image)
		Command: []string{"sh", "-c"},
		Args: []string{`
			set -eu

			# Create Project Zomboid specific directories in persistent storage
			mkdir -p /data/config-lgsm/pzserver /data/serverfiles

			# Copy Project Zomboid config files to persistent locations
			cp /tmp/config-gsm/pzserver.cfg /data/config-lgsm/pzserver/pzserver.cfg
			cp /tmp/config-server/server.ini /data/serverfiles/server.ini

			# Set ownership for linuxgsm user (1000:1000)
			chown -R 1000:1000 /data/config-lgsm/pzserver /data/serverfiles

			echo "Project Zomboid config setup completed successfully"
		`},
		VolumeMounts: []corev1.VolumeMount{
			{Name: DataVolumeName, MountPath: "/data"},
			{Name: ConfigGsmVolumeName, MountPath: "/tmp/config-gsm"},
			{Name: ConfigServerVolumeName, MountPath: "/tmp/config-server"},
		},
	}
}

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
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:  func(i int64) *int64 { return &i }(1000),
			RunAsGroup: func(i int64) *int64 { return &i }(1000),
		},
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

// getSecureGameServerContainer returns a game server container - let LinuxGSM handle user setup
func getSecureGameServerContainer(name, image string, resources corev1.ResourceRequirements, ports []corev1.ContainerPort) corev1.Container {
	return corev1.Container{
		Name:            name,
		Image:           image,
		Resources:       applyResourceDefaults(resources),
		Ports:           ports,
		SecurityContext: getGameServerSecurityContext(),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      DataVolumeName,
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

// getMinecraftSetupInitContainer returns an init container specifically for Minecraft config setup
// Minecraft configuration paths:
// - GSM config: /data/config-lgsm/mcserver/mcserver.cfg (LinuxGSM config)
// - Server config: /data/serverfiles/server.properties (Minecraft server properties)
// - JVM args: /data/serverfiles/jvm.args (JVM arguments)
// More info: https://linuxgsm.com/lgsm/mcserver/
func getMinecraftSetupInitContainer() corev1.Container {
	return corev1.Container{
		Name:    SetupContainerName,
		Image:   SetupContainerImage,
		Command: []string{"sh", "-c"},
		Args: []string{`
			set -eu

			# Create Minecraft specific directories
			mkdir -p /data/config-lgsm/mcserver /data/serverfiles
			mkdir -p /data/serverfiles/plugins /data/serverfiles/mods /data/serverfiles/worlds

			# Copy LinuxGSM config if available
			if [ -f "/configs/mcserver.cfg" ]; then
				cp /configs/mcserver.cfg /data/config-lgsm/mcserver/mcserver.cfg
			fi

			# Generate server.properties if config provided
			if [ -f "/configs/server.properties" ]; then
				cp /configs/server.properties /data/serverfiles/server.properties
			fi

			# Copy JVM args if provided
			if [ -f "/configs/jvm.args" ]; then
				cp /configs/jvm.args /data/serverfiles/jvm.args
			fi

			# Copy plugins if configured
			if [ -d "/configs/plugins" ]; then
				cp -r /configs/plugins/* /data/serverfiles/plugins/ 2>/dev/null || true
			fi

			# Copy mods if configured (for Forge servers)
			if [ -d "/configs/mods" ]; then
				cp -r /configs/mods/* /data/serverfiles/mods/ 2>/dev/null || true
			fi

			# Set ownership for linuxgsm user (1000:1000)
			chown -R 1000:1000 /data/config-lgsm/mcserver /data/serverfiles

			echo "Minecraft config setup completed successfully"
		`},
		VolumeMounts: []corev1.VolumeMount{
			{Name: DataVolumeName, MountPath: "/data"},
			{Name: ConfigsVolumeName, MountPath: "/configs"},
		},
	}
}

// getKf2SetupInitContainer returns an init container specifically for Killing Floor 2 config setup
// KF2 configuration paths:
// - GSM config: /data/config-lgsm/kf2server/kf2server.cfg (LinuxGSM config)
// - Server config: /data/serverfiles/KFGame/Config/KFGame.ini (KF2 server config)
// More info: https://linuxgsm.com/lgsm/kf2server/
func getKf2SetupInitContainer() corev1.Container {
	return corev1.Container{
		Name:    SetupContainerName,
		Image:   SetupContainerImage,
		Command: []string{"sh", "-c"},
		Args: []string{`
			set -eu

			# Create Killing Floor 2 specific directories
			mkdir -p /data/config-lgsm/kf2server /data/serverfiles/KFGame/Config

			# Always copy latest KF2 config files (ensures ConfigMap updates are applied)
			cp /configs/kf2server.cfg /data/config-lgsm/kf2server/kf2server.cfg
			cp /configs/KFGame.ini /data/serverfiles/KFGame/Config/KFGame.ini

			# Set ownership for linuxgsm user (1000:1000)
			chown -R 1000:1000 /data/config-lgsm/kf2server /data/serverfiles/KFGame/Config

			echo "KF2 config setup completed successfully"
		`},
		VolumeMounts: []corev1.VolumeMount{
			{Name: DataVolumeName, MountPath: "/data"},
			{Name: ConfigsVolumeName, MountPath: "/configs"},
		},
	}
}

// getArkSetupInitContainer returns an init container specifically for ARK config setup
// ARK configuration paths:
// - GSM config: /data/config-lgsm/arkserver/arkserver.cfg (LinuxGSM config)
// - GameUserSettings.ini: /data/serverfiles/ShooterGame/Saved/Config/WindowsServer/GameUserSettings.ini
// - Game.ini: /data/serverfiles/ShooterGame/Saved/Config/WindowsServer/Game.ini
// More info: https://linuxgsm.com/lgsm/arkserver/
func getArkSetupInitContainer() corev1.Container {
	return corev1.Container{
		Name:    SetupContainerName,
		Image:   SetupContainerImage,
		Command: []string{"sh", "-c"},
		Args: []string{`
			set -eu

			# Create ARK specific directories
			mkdir -p /data/config-lgsm/arkserver /data/serverfiles/ShooterGame/Saved/Config

			# Copy LinuxGSM config if available
			if [ -f "/configs/arkserver.cfg" ]; then
				cp /configs/arkserver.cfg /data/config-lgsm/arkserver/arkserver.cfg
			fi

			# Copy GameUserSettings.ini to ARK directory structure
			if [ -f "/configs/GameUserSettings.ini" ]; then
				mkdir -p /data/serverfiles/ShooterGame/Saved/Config/WindowsServer
				cp /configs/GameUserSettings.ini /data/serverfiles/ShooterGame/Saved/Config/WindowsServer/GameUserSettings.ini
			fi

			# Copy Game.ini to ARK directory structure
			if [ -f "/configs/Game.ini" ]; then
				mkdir -p /data/serverfiles/ShooterGame/Saved/Config/WindowsServer
				cp /configs/Game.ini /data/serverfiles/ShooterGame/Saved/Config/WindowsServer/Game.ini
			fi

			# Set ownership for linuxgsm user (1000:1000)
			chown -R 1000:1000 /data/config-lgsm/arkserver /data/serverfiles

			echo "ARK config setup completed successfully"
		`},
		VolumeMounts: []corev1.VolumeMount{
			{Name: DataVolumeName, MountPath: "/data"},
			{Name: ConfigsVolumeName, MountPath: "/configs"},
		},
	}
}

// getSdtdSetupInitContainer returns an init container specifically for 7 Days to Die config setup
// 7DTD configuration paths:
// - GSM config: /data/config-lgsm/sdtdserver/sdtdserver.cfg (LinuxGSM config)
// - Server config: /data/serverfiles/SDTD/Config/serverconfig.xml (7DTD server config)
// More info: https://linuxgsm.com/lgsm/sdtdserver/
func getSdtdSetupInitContainer() corev1.Container {
	return corev1.Container{
		Name:            SetupContainerName,
		Image:           SetupContainerImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:  func(i int64) *int64 { return &i }(GameServerUserID),
			RunAsGroup: func(i int64) *int64 { return &i }(GameServerGroupID),
		},
		Command: []string{"sh", "-c"},
		Args: []string{`
			set -e

			echo "Starting 7 Days to Die config setup..."

			# Create required directories
			mkdir -p /data/config-lgsm/sdtdserver
			mkdir -p /data/serverfiles/SDTD/Config

			# Copy GSM config if available
			if [ -f "/configs/sdtdserver.cfg" ]; then
				cp /configs/sdtdserver.cfg /data/config-lgsm/sdtdserver/sdtdserver.cfg
			fi

			# Copy server config (serverconfig.xml) to 7DTD directory structure
			if [ -f "/configs/serverconfig.xml" ]; then
				cp /configs/serverconfig.xml /data/serverfiles/SDTD/Config/serverconfig.xml
			fi

			# Set ownership for linuxgsm user (1000:1000)
			chown -R 1000:1000 /data/config-lgsm/sdtdserver /data/serverfiles

			echo "7 Days to Die config setup completed successfully"
		`},
		VolumeMounts: []corev1.VolumeMount{
			{Name: DataVolumeName, MountPath: "/data"},
			{Name: ConfigsVolumeName, MountPath: "/configs"},
		},
	}
}

// boolToInt converts boolean to integer for various configs
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// boolToString converts boolean to string for various configs
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
