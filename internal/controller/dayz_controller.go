/*
Copyright 2025.

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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	gameserverv1 "github.com/templarfelix/gameserver-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DayzReconciler reconciles a Dayz object
type DayzReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=dayzs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=dayzs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=dayzs/finalizers,verbs=update

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
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *DayzReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx).WithValues("dayz", req.Name)

	instance := &gameserverv1.Dayz{}

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

	if err := r.reconcileDeployment(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.reconcileServices(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// reconcilePVC wraps ReconcilePVC with logging for concurrency conflicts
func (r *DayzReconciler) reconcilePVC(ctx context.Context, instance *gameserverv1.Dayz) error {
	logger := logf.FromContext(ctx)
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
func (r *DayzReconciler) reconcileServices(ctx context.Context, instance *gameserverv1.Dayz) error {
	logger := logf.FromContext(ctx)
	if err := ReconcileServices(ctx, r.Client, instance, instance.Spec.Ports, instance.Spec.LoadBalancerIP); err != nil {
		// Log concurrent modification conflicts
		if errors.IsConflict(err) {
			logger.Info("Services conflict detected, will retry")
		}
		return err
	}
	return nil
}

func (r *DayzReconciler) reconcileDeployment(ctx context.Context, instance *gameserverv1.Dayz) error {
	logger := logf.FromContext(ctx)

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
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup: func(i int64) *int64 { return &i }(1000),
					},

					InitContainers: []corev1.Container{
						{
							Name:    "config-writer",
							Image:   SetupContainerImage,
							Command: []string{"sh", "-c"},
							Args:    []string{r.generateDayzConfigSetupScript(instance)},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "tmp-configs", MountPath: "/tmp/configs"},
							},
						},
						{
							Name:    SetupContainerName,
							Image:   SetupContainerImage,
							Command: []string{"sh", "-c"},
							Args: []string{r.generateDayzSetupScript(instance)},
							SecurityContext: &corev1.SecurityContext{
								RunAsUser:  func(i int64) *int64 { return &i }(1000),
								RunAsGroup: func(i int64) *int64 { return &i }(1000),
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: DataVolumeName, MountPath: "/data"},
								{Name: "tmp-configs", MountPath: "/tmp/configs"},
							},
						},
					},
					Containers: []corev1.Container{
						GetSecureGameServerContainer("server", instance.Spec.Image, instance.Spec.Resources, containerPorts),
						GetSecureCodeServerContainer(instance.Spec.EditorPassword),
					},
					Volumes: []corev1.Volume{
						{
							Name: "tmp-configs", // Direct file path volume
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
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

// generateDayzConfigSetupScript creates a shell script that writes config files to the tmp-configs volume
func (r *DayzReconciler) generateDayzConfigSetupScript(instance *gameserverv1.Dayz) string {
	script := `set -eu
mkdir -p /tmp/configs

# Write config files directly to tmp-configs volume
`

	// Add all configuration files
	if instance.Spec.Config != nil {
		for filepath, content := range instance.Spec.Config {
			script += fmt.Sprintf("mkdir -p $(dirname '/tmp/configs%s')\n", filepath)
			script += fmt.Sprintf("cat > '/tmp/configs%s' << 'EOF'\n%s\nEOF\n", filepath, content)
		}
	}

	script += "echo 'Config files written to tmp-configs volume successfully'\n"
	return script
}

// generateDayzSetupScript creates a shell script that copies config files and runs additional commands
func (r *DayzReconciler) generateDayzSetupScript(instance *gameserverv1.Dayz) string {
	script := `set -eu

# Create DayZ specific directories
mkdir -p /data/config-lgsm/dayzserver /data/serverfiles/cfg

# Copy all config files from tmp-configs to their respective locations
# Game config files go to /data/serverfiles/cfg/
# LinuxGSM config files go to /data/config-lgsm/dayzserver/
# Find all files in the tmp-configs directory (including subdirectories)
find /tmp/configs -type f | while read file; do
  # Extract the target path from filename (remove /tmp/configs prefix)
  # The files are written with full paths like /tmp/configs/data/config-lgsm/dayzserver/dayzserver.cfg
  # So we need to remove the /tmp/configs prefix to get the correct target path
  target_path="${file#/tmp/configs/}"

  # Create parent directory for target path
  target_dir=$(dirname "$target_path")
  mkdir -p "$target_dir"

  # Copy file to target location
  cp "$file" "$target_path"
  echo "Copied $file to $target_path"
done

# Set ownership for linuxgsm user (1000:1000)
chown -R 1000:1000 /data/config-lgsm/dayzserver /data/serverfiles/cfg

# Run additional commands if specified
`

	// Add additional commands
	for _, command := range instance.Spec.PostCopyCommands {
		script += fmt.Sprintf("%s\n", command)
	}

	script += "echo 'DayZ config setup completed successfully'\n"
	return script
}

// SetupWithManager sets up the controller with the Manager.
func (r *DayzReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Temporarily disabled webhooks due to certificate issues
	// if err := (&DayzValidator{}).SetupWebhookWithManager(mgr); err != nil {
	//	return err
	// }

	return ctrl.NewControllerManagedBy(mgr).
		For(&gameserverv1.Dayz{}).
		Complete(r)
}
