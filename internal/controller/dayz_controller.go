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
	"k8s.io/apimachinery/pkg/api/resource"
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

// DayzReconciler reconciles a Dayz object
type DayzReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=dayzs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=dayzs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=dayzs/finalizers,verbs=update

//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Dayz object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *DayzReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	instance := &gameserverv1alpha1.Dayz{}

	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Objeto não encontrado, pode ter sido deletado após o request de reconciliação. Sair do processamento.
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	if err := r.reconcilePVC(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.reconcileConfigMap(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.reconcileDeployment(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *DayzReconciler) reconcileDeployment(ctx context.Context, instance *gameserverv1alpha1.Dayz) error {
	logger := log.FromContext(ctx)

	resource := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-deployment",
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			//Replicas: ,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": instance.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": instance.Name},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:  "fix-permissions",
							Image: "busybox",
							Command: []string{
								"sh", "-c", `
								mkdir -p /data/config-lgsm/dayzserver/ &&
								mkdir -p /data/serverfiles/cfg/ &&
								ls -la /tmp &&
								ls -la /tmp/config-gsm &&
								ls -la /tmp/config-server 
								#&& 
								#cp /tmp/config-gsm/dayzserver.cfg /data/config-lgsm/dayzserver/dayzserver.cfg &&
								#cp /tmp/config-server/dayzserver.server.cfg /data/serverfiles/cfg/dayzserver.server.cfg &&
								#chown 1000:1000 /data/config-lgsm/dayzserver/dayzserver.cfg &&
								#chown 1000:1000 /data/serverfiles/cfg/dayzserver.server.cfg
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
							Name:  "server",
							Image: instance.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 2302,
									Name:          "port-2302-tcp",
									Protocol:      corev1.ProtocolTCP,
								},
								// Add other ports here
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
								},
							},
						},
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
											Key:  "dayzserver.server.cfg",
											Path: "dayzserver.server.cfg",
											Mode: func(i int32) *int32 { return &i }(0777),
										},
									},
								},
							},
						},
						{
							Name: "config-gsm",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: instance.Name + "-configmap",
									},
									DefaultMode: func(i int32) *int32 { return &i }(0777),
									Items: []corev1.KeyToPath{
										{
											Key:  "dayzserver.cfg",
											Path: "dayzserver.cfg",
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

	if err := controllerutil.SetControllerReference(instance, resource, r.Scheme); err != nil {
		return err
	}

	found := &appsv1.Deployment{}
	err := r.Client.Get(ctx, client.ObjectKey{Name: resource.Name, Namespace: resource.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Deployment %s/%s\n", resource.Namespace, resource.Name)
		err = r.Client.Create(ctx, resource)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	logger.Info("Skip reconcile: Deployment %s/%s already exists", found.Namespace, found.Name)

	return nil
}

func (r *DayzReconciler) reconcilePVC(ctx context.Context, instance *gameserverv1alpha1.Dayz) error {
	logger := log.FromContext(ctx)

	resource := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-pvc",
			Namespace: instance.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(instance.Spec.Storage),
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(instance, resource, r.Scheme); err != nil {
		return err
	}

	found := &corev1.PersistentVolumeClaim{}
	err := r.Client.Get(ctx, client.ObjectKey{Name: resource.Name, Namespace: resource.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new PVC %s/%s\n", resource.Namespace, resource.Name)
		err = r.Client.Create(ctx, resource)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	logger.Info("Skip reconcile: PVC %s/%s already exists", found.Namespace, found.Name)

	return nil
}

func (r *DayzReconciler) reconcileConfigMap(ctx context.Context, instance *gameserverv1alpha1.Dayz) error {
	logger := log.FromContext(ctx)

	resource := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-configmap",
			Namespace: instance.Namespace,
		},
		Data: map[string]string{
			"dayzserver.cfg":        instance.Spec.Config.GSM,
			"dayzserver.server.cfg": instance.Spec.Config.Server,
		},
	}

	if err := controllerutil.SetControllerReference(instance, resource, r.Scheme); err != nil {
		return err
	}

	found := &corev1.ConfigMap{}
	err := r.Client.Get(ctx, client.ObjectKey{Name: resource.Name, Namespace: resource.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Configmap %s/%s\n", resource.Namespace, resource.Name)
		err = r.Client.Create(ctx, resource)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	logger.Info("Skip reconcile: Configmap %s/%s already exists", found.Namespace, found.Name)

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DayzReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gameserverv1alpha1.Dayz{}).
		Complete(r)
}
