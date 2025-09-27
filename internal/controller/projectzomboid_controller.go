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
	logger := log.FromContext(ctx).WithValues("projectzomboid", req.NamespacedName.Name)

	instance := &gameserverv1alpha1.ProjectZomboid{}

	err := r.Client.Get(ctx, req.NamespacedName, instance)
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
			err := r.Client.Get(ctx, types.NamespacedName{Name: pvcName, Namespace: instance.Namespace}, pvc)
			if err != nil && !errors.IsNotFound(err) {
				logger.Error(err, "Failed to get PVC")
				return reconcile.Result{}, err
			}
			if err == nil { // PVC exists
				if instance.Spec.Persistence.PreserveOnDelete {
					// Remove owner reference to preserve PVC
					pvc.OwnerReferences = nil // Remove all owner refs, or filter
					if err := r.Client.Update(ctx, pvc); err != nil {
						logger.Error(err, "Failed to remove owner reference from PVC")
						return reconcile.Result{}, err
					}
					logger.Info("Preserved PVC by removing owner reference")
				} // else let GC delete it
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(instance, finalizer)
			if err := r.Client.Update(ctx, instance); err != nil {
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
		if err := r.Client.Update(ctx, instance); err != nil {
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

	if err := r.reconcileConfigMap(ctx, instance); err != nil {
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
						{
							Name:  "fix-permissions",
							Image: "busybox",
							Command: []string{
								"sh", "-c", `
								mkdir -p /data/config-lgsm/pzserver/ &&
								mkdir -p /data/serverfiles/ &&
								cp /tmp/config-gsm/pzserver.cfg /data/config-lgsm/pzserver/pzserver.cfg &&
								cp /tmp/config-server/server.ini /data/serverfiles/server.ini &&
								chown 1000:1000 /data/config-lgsm/pzserver/pzserver.cfg &&
								chown 1000:1000 /data/serverfiles/server.ini
								`,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "tmp",
									MountPath: "/tmp",
								},
								{
									Name:      "data",
									MountPath: "/data",
								},
								{
									Name:      "config-server",
									MountPath: "/tmp/config-server",
								},
								{
									Name:      "config-gsm",
									MountPath: "/tmp/config-gsm",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:      "server",
							Image:     instance.Spec.Image,
							Resources: instance.Spec.Resources,
							Ports: []corev1.ContainerPort{
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
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
								},
							},
						},
						GetCodeServerContainer(instance.Spec.EditorPassword),
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
	err := r.Client.Get(ctx, client.ObjectKey{Name: k8sResource.Name, Namespace: k8sResource.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Deployment", "Namespace", k8sResource.Namespace, "Name", k8sResource.Name)
		err = r.Client.Create(ctx, k8sResource)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Check if the Deployment needs update
	if !compareDeployments(found, k8sResource) {
		logger.Info("Updating Deployment", "Namespace", found.Namespace, "Name", found.Name)
		found.Spec = k8sResource.Spec
		return r.Client.Update(ctx, found)
	}

	logger.Info("Skip reconcile: Deployment already exists and is up to date", "Namespace", found.Namespace, "Name", found.Name)

	return nil
}

func (r *ProjectZomboidReconciler) reconcileConfigMap(ctx context.Context, instance *gameserverv1alpha1.ProjectZomboid) error {
	logger := log.FromContext(ctx)

	k8sResource := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-configmap",
			Namespace: instance.Namespace,
		},
		Data: map[string]string{
			"pzserver.cfg": instance.Spec.Config.GSM,
			"server.ini":   instance.Spec.Config.Server,
		},
	}

	if err := controllerutil.SetControllerReference(instance, k8sResource, r.Scheme); err != nil {
		return err
	}

	found := &corev1.ConfigMap{}
	err := r.Client.Get(ctx, client.ObjectKey{Name: k8sResource.Name, Namespace: k8sResource.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Configmap", "Namespace", k8sResource.Namespace, "Name", k8sResource.Name)
		err = r.Client.Create(ctx, k8sResource)
		if err != nil {
			return err
		}
		// ConfigMap created successfully, no need to check for updates
		return nil
	} else if err != nil {
		return err
	}

	// Check if the ConfigMap needs update
	if !compareConfigMaps(found, k8sResource) {
		logger.Info("Updating Configmap", "Namespace", found.Namespace, "Name", found.Name)
		found.Data = k8sResource.Data
		return r.Client.Update(ctx, found)
	}

	logger.Info("Skip reconcile: Configmap already exists and is up to date", "Namespace", found.Namespace, "Name", found.Name)

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectZomboidReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gameserverv1alpha1.ProjectZomboid{}).
		Complete(r)
}
