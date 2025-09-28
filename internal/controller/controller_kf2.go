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

// Kf2Reconciler reconciles a Kf2 object
type Kf2Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=kf2s,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=kf2s/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gameserver.templarfelix.com,resources=kf2s/finalizers,verbs=update

//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Add RBAC for networking resources to fix permission warnings
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Kf2Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("kf2", req.Name)

	instance := &gameserverv1alpha1.Kf2{}

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
	configData := r.generateKf2ConfigData(instance)
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
func (r *Kf2Reconciler) reconcilePVC(ctx context.Context, instance *gameserverv1alpha1.Kf2) error {
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
func (r *Kf2Reconciler) reconcileServices(ctx context.Context, instance *gameserverv1alpha1.Kf2) error {
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

func (r *Kf2Reconciler) reconcileDeployment(ctx context.Context, instance *gameserverv1alpha1.Kf2) error {
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
						getKf2SetupInitContainer(),
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

// generateKf2ConfigData creates all necessary configuration files for KF2
func (r *Kf2Reconciler) generateKf2ConfigData(instance *gameserverv1alpha1.Kf2) map[string]string {
	configData := make(map[string]string)

	// Generate KFGame.ini (main game configuration)
	configData["KFGame.ini"] = generateKf2GameConfig(&instance.Spec.Config.Game)

	// Generate LinuxGSM config (if custom)
	if instance.Spec.Config.GSM.ConfigFile != "" {
		configData["kf2server.cfg"] = instance.Spec.Config.GSM.ConfigFile
	} else {
		configData["kf2server.cfg"] = generateKf2GSMConfig(&instance.Spec.Config.GSM)
	}

	return configData
}

// generateKf2GameConfig creates KFGame.ini content from CRD spec
func generateKf2GameConfig(settings *gameserverv1alpha1.Kf2GameConfig) string {
	var lines []string

	// Engine.GameInfo section
	lines = append(lines, "[Engine.GameInfo]")
	lines = append(lines, fmt.Sprintf("DefaultGame=KFGameContent.KFGameInfo_Survival"))
	lines = append(lines, fmt.Sprintf("DefaultServerGame=KFGameContent.KFGameInfo_Survival"))
	lines = append(lines, fmt.Sprintf("bAdminCanPause=%s", boolToString(!settings.AllowAdminPause))) // inverted logic
	lines = append(lines, fmt.Sprintf("MaxPlayers=%d", settings.MaxPlayers))
	lines = append(lines, fmt.Sprintf("GameDifficulty=%d.000000", settings.Difficulty))
	lines = append(lines, fmt.Sprintf("bChangeLevels=True"))
	lines = append(lines, fmt.Sprintf("MaxSpectators=%d", settings.MaxSpectators))
	lines = append(lines, fmt.Sprintf("MaxIdleTime=0.000000"))
	lines = append(lines, fmt.Sprintf("MaxTimeMargin=0.000000"))
	lines = append(lines, fmt.Sprintf("TimeMarginSlack=1.350000"))
	lines = append(lines, fmt.Sprintf("MinTimeMargin=-1.000000"))
	lines = append(lines, fmt.Sprintf("TotalNetBandwidth=32000"))
	lines = append(lines, fmt.Sprintf("MaxDynamicBandwidth=7000"))
	lines = append(lines, fmt.Sprintf("MinDynamicBandwidth=4000"))
	lines = append(lines, fmt.Sprintf("DefaultGameType=KFGameContent.KFGameInfo_Survival"))
	lines = append(lines, fmt.Sprintf("GoreLevel=%d", settings.GoreLevel))
	lines = append(lines, fmt.Sprintf("TimeBetweenFailedVotes=%d.0", settings.TimeBetweenFailedVotes))
	lines = append(lines, fmt.Sprintf("VoteTime=%d.0", settings.VoteTime))
	lines = append(lines, fmt.Sprintf("bIsStandbyCheckingEnabled=false"))
	lines = append(lines, fmt.Sprintf("KickVotePercentage=%s", settings.KickVotePercentage))
	lines = append(lines, fmt.Sprintf("bKickLiveIdlers=False"))
	lines = append(lines, fmt.Sprintf("ArbitrationHandshakeTimeout=0.000000"))
	lines = append(lines, "")

	// Engine.AccessControl section
	lines = append(lines, "[Engine.AccessControl]")
	lines = append(lines, fmt.Sprintf("IPPolicies=ACCEPT;*"))
	lines = append(lines, fmt.Sprintf("bAuthenticateClients=True"))
	lines = append(lines, fmt.Sprintf("bAuthenticateServer=True"))
	lines = append(lines, fmt.Sprintf("bAuthenticateListenHost=True"))
	lines = append(lines, fmt.Sprintf("MaxAuthRetryCount=3"))
	lines = append(lines, fmt.Sprintf("AuthRetryDelay=5"))
	if settings.AdminPassword != "" {
		lines = append(lines, fmt.Sprintf("AdminPassword=%s", settings.AdminPassword))
	} else {
		lines = append(lines, fmt.Sprintf("AdminPassword="))
	}
	if settings.Password != "" {
		lines = append(lines, fmt.Sprintf("GamePassword=%s", settings.Password))
	} else {
		lines = append(lines, fmt.Sprintf("GamePassword="))
	}
	lines = append(lines, "")

	// DefaultPlayer section
	lines = append(lines, "[DefaultPlayer]")
	lines = append(lines, fmt.Sprintf("Name=Player"))
	lines = append(lines, fmt.Sprintf("Team=255"))
	lines = append(lines, "")

	// Set the game mode based on selection
	var gameModeClass string
	switch settings.GameMode {
	case "WeeklySurvival":
		gameModeClass = "KFGameContent.KFGameInfo_WeeklySurvival"
	case "VersusSurvival":
		gameModeClass = "KFGameContent.KFGameInfo_VersusSurvival"
	default:
		gameModeClass = "KFGameContent.KFGameInfo_Survival"
	}

	// KFGame.KFGameInfo section
	lines = append(lines, "[KFGame.KFGameInfo]")
	lines = append(lines, fmt.Sprintf("GameModes=(FriendlyName=\"Survival\",ClassNameAndPath=\"KFGameContent.KFGameInfo_Survival\",bSoloPlaySupported=true,DifficultyLevels=4,Lengths=4,LocalizeID=0)"))
	lines = append(lines, fmt.Sprintf("GameModes=(FriendlyName=\"Weekly\",ClassNameAndPath=\"KFGameContent.KFGameInfo_WeeklySurvival\",bSoloPlaySupported=true,DifficultyLevels=0,Lengths=0,LocalizeID=1)"))
	lines = append(lines, fmt.Sprintf("GameModes=(FriendlyName=\"Versus\",ClassNameAndPath=\"KFGameContent.KFGameInfo_VersusSurvival\",bSoloPlaySupported=false,DifficultyLevels=0,Lengths=0,LocalizeID=2)"))
	lines = append(lines, fmt.Sprintf("GameLength=%d", settings.GameLength))
	lines = append(lines, fmt.Sprintf("MinNetPlayers=1"))
	lines = append(lines, fmt.Sprintf("GameStartDelay=4"))
	lines = append(lines, fmt.Sprintf("ReadyUpDelay=90"))
	lines = append(lines, fmt.Sprintf("bWaitForNetPlayers=true"))
	lines = append(lines, fmt.Sprintf("bEnableMapObjectives=%s", boolToString(settings.EnableMapObjectives)))
	lines = append(lines, fmt.Sprintf("bUseMapList=True"))
	lines = append(lines, fmt.Sprintf("ActiveMapCycle=0"))
	lines = append(lines, fmt.Sprintf("GameMapCycles=(Maps=(\"%s\"))", strings.Join(settings.MapCycle, "\",\"")))
	lines = append(lines, fmt.Sprintf("EndOfGameDelay=15"))
	lines = append(lines, fmt.Sprintf("FriendlyFireScale=%s", settings.FriendlyFireScale))
	lines = append(lines, fmt.Sprintf("KickVotePercentage=%s", settings.KickVotePercentage))
	lines = append(lines, fmt.Sprintf("TimeBetweenFailedVotes=%d.000000", settings.TimeBetweenFailedVotes))
	lines = append(lines, fmt.Sprintf("VoteTime=%d.0", settings.VoteTime))
	lines = append(lines, fmt.Sprintf("MapVoteDuration=%0.3f", float64(settings.MapVoteDuration)))
	lines = append(lines, fmt.Sprintf("bLogGameBalance=true"))
	lines = append(lines, fmt.Sprintf("BannerLink=http://art.tripwirecdn.com/TestItemIcons/MOTDServer.png"))
	lines = append(lines, fmt.Sprintf("ServerMOTD=\"%s\"", settings.ServerMOTD))
	if strings.TrimSpace(settings.WelcomeMessage) != "" {
		lines = append(lines, fmt.Sprintf("ServerWelcomeMessage=\"%s\"", settings.WelcomeMessage))
	}
	lines = append(lines, fmt.Sprintf("WebsiteLink=http://killingfloor2.com/"))
	lines = append(lines, fmt.Sprintf("ClanMotto=\"%s\"", settings.ClanMotto))
	lines = append(lines, fmt.Sprintf("ServerMOTDColor=(B=254,G=254,R=254,A=192)"))
	lines = append(lines, fmt.Sprintf("WebLinkColor=(B=254,G=254,R=254,A=192)"))
	lines = append(lines, fmt.Sprintf("bEnableDeadToVOIP=%s", boolToString(settings.EnableDeadToVOIP)))
	lines = append(lines, fmt.Sprintf("ServerExpirationForKillWhenEmpty=%d.0", settings.EmptyServerDelay))
	lines = append(lines, fmt.Sprintf("bDisableKickVote=%s", boolToString(settings.DisableKickVoting)))
	lines = append(lines, fmt.Sprintf("bDisablePublicTextChat=False"))
	lines = append(lines, fmt.Sprintf("bDisableVOIP=%s", boolToString(!settings.EnableVOIP)))
	lines = append(lines, fmt.Sprintf("bDisableMapVote=%s", boolToString(!settings.EnableMapVoting)))
	lines = append(lines, fmt.Sprintf("bDisableTeamCollision=False"))
	lines = append(lines, fmt.Sprintf("bDisablePublicVOIPChannel=%s", boolToString(!settings.EnablePublicVOIPChannel)))
	lines = append(lines, fmt.Sprintf("bPartitionSpectators=%s", boolToString(settings.PartitionSpectators)))
	lines = append(lines, fmt.Sprintf("MapVotePercentage=%s", settings.MapVotePercentage))
	lines = append(lines, fmt.Sprintf("MapCycleIndex=-1"))
	lines = append(lines, "")

	// Game mode specific sections
	switch gameModeClass {
	case "KFGameContent.KFGameInfo_Survival":
		lines = append(lines, "[KFGameContent.KFGameInfo_Survival]")
		lines = append(lines, fmt.Sprintf("bEnableGameAnalytics=%s", boolToString(settings.EnableGameAnalytics)))
		lines = append(lines, fmt.Sprintf("bRecordGameStatsFile=%s", boolToString(settings.RecordGameStats)))
		lines = append(lines, fmt.Sprintf("MaxPlayers=%d", settings.MaxPlayers))
		lines = append(lines, fmt.Sprintf("GameModes=(FriendlyName=\"Survival\",ClassNameAndPath=\"KFGameContent.KFGameInfo_Survival\",bSoloPlaySupported=True,DifficultyLevels=4,Lengths=4,LocalizeID=0)"))
		lines = append(lines, fmt.Sprintf("GameModes=(FriendlyName=\"Weekly\",ClassNameAndPath=\"KFGameContent.KFGameInfo_WeeklySurvival\",bSoloPlaySupported=True,DifficultyLevels=0,Lengths=0,LocalizeID=1)"))
		lines = append(lines, fmt.Sprintf("GameModes=(FriendlyName=\"Versus\",ClassNameAndPath=\"KFGameContent.KFGameInfo_VersusSurvival\",bSoloPlaySupported=False,DifficultyLevels=0,Lengths=0,LocalizeID=2)"))
		lines = append(lines, fmt.Sprintf("bWaitForNetPlayers=True"))
		lines = append(lines, fmt.Sprintf("bEnableMapObjectives=True"))
		lines = append(lines, fmt.Sprintf("bLogScoring=False"))
		lines = append(lines, fmt.Sprintf("bLogAIDefaults=False"))
		lines = append(lines, fmt.Sprintf("bLogAICount=False"))
		lines = append(lines, fmt.Sprintf("MinNetPlayers=1"))
		lines = append(lines, fmt.Sprintf("EndOfGameDelay=15"))
		lines = append(lines, fmt.Sprintf("ServerExpirationForKillWhenEmpty=%d.000000", settings.EmptyServerDelay))
		lines = append(lines, fmt.Sprintf("RequiredMobileInputConfigs=(GroupName=\"DebugGroup\",RequireZoneNames=(\"DebugStickMoveZone\",\"DebugStickLookZone\",\"DebugLookZone\"),bIsAttractModeGroup=False)"))
		lines = append(lines, fmt.Sprintf("bIsStandbyCheckingEnabled=False"))
		lines = append(lines, fmt.Sprintf("GoalScore=0"))
		lines = append(lines, fmt.Sprintf("MaxLives=0"))
		lines = append(lines, fmt.Sprintf("TimeLimit=0"))
		lines = append(lines, fmt.Sprintf("StandbyRxCheatTime=0.000000"))
		lines = append(lines, fmt.Sprintf("StandbyTxCheatTime=0.000000"))
		lines = append(lines, fmt.Sprintf("BadPingThreshold=0"))
		lines = append(lines, fmt.Sprintf("PercentMissingForRxStandby=0.000000"))
		lines = append(lines, fmt.Sprintf("PercentMissingForTxStandby=0.000000"))
		lines = append(lines, fmt.Sprintf("PercentForBadPing=0.000000"))
		lines = append(lines, fmt.Sprintf("JoinInProgressStandbyWaitTime=0.000000"))
		lines = append(lines, fmt.Sprintf("DefaultGameType=KFGameContent.KFGameInfo_Survival"))
		lines = append(lines, fmt.Sprintf("AnimTreePoolSize=0"))
	case "KFGameContent.KFGameInfo_WeeklySurvival":
		lines = append(lines, "[KFGameContent.KFGameInfo_WeeklySurvival]")
		lines = append(lines, fmt.Sprintf("bEnableGameAnalytics=%s", boolToString(settings.EnableGameAnalytics)))
		lines = append(lines, fmt.Sprintf("bRecordGameStatsFile=%s", boolToString(settings.RecordGameStats)))
		lines = append(lines, fmt.Sprintf("MaxPlayers=%d", settings.MaxPlayers))
		lines = append(lines, fmt.Sprintf("bEnableDevAnalytics=true"))
		lines = append(lines, fmt.Sprintf("GameModes=(FriendlyName=\"Survival\",ClassNameAndPath=\"KFGameContent.KFGameInfo_Survival\",bSoloPlaySupported=true,DifficultyLevels=4,Lengths=4,LocalizeID=0)"))
		lines = append(lines, fmt.Sprintf("GameModes=(FriendlyName=\"Weekly\",ClassNameAndPath=\"KFGameContent.KFGameInfo_WeeklySurvival\",bSoloPlaySupported=true,DifficultyLevels=0,Lengths=0,LocalizeID=1)"))
		lines = append(lines, fmt.Sprintf("GameModes=(FriendlyName=\"Versus\",ClassNameAndPath=\"KFGameContent.KFGameInfo_VersusSurvival\",bSoloPlaySupported=false,DifficultyLevels=0,Lengths=0,LocalizeID=2)"))
	case "KFGameContent.KFGameInfo_VersusSurvival":
		lines = append(lines, "[KFGameContent.KFGameInfo_VersusSurvival]")
		lines = append(lines, fmt.Sprintf("bEnableGameAnalytics=%s", boolToString(settings.EnableGameAnalytics)))
		lines = append(lines, fmt.Sprintf("bRecordGameStatsFile=%s", boolToString(settings.RecordGameStats)))
		lines = append(lines, fmt.Sprintf("MinNetPlayers=2"))
		lines = append(lines, fmt.Sprintf("bTeamBalanceEnabled=true"))
		lines = append(lines, fmt.Sprintf("MaxPlayers=%d", settings.MaxPlayers))
		lines = append(lines, fmt.Sprintf("TimeUntilNextRound=30"))
		lines = append(lines, fmt.Sprintf("bEnableDevAnalytics=true"))
		lines = append(lines, fmt.Sprintf("ScoreRadius=1000.000000"))
		lines = append(lines, fmt.Sprintf("GameModes=(FriendlyName=\"Survival\",ClassNameAndPath=\"KFGameContent.KFGameInfo_Survival\",bSoloPlaySupported=True,DifficultyLevels=4,Lengths=4,LocalizeID=0)"))
		lines = append(lines, fmt.Sprintf("GameModes=(FriendlyName=\"Weekly\",ClassNameAndPath=\"KFGameContent.KFGameInfo_WeeklySurvival\",bSoloPlaySupported=True,DifficultyLevels=0,Lengths=0,LocalizeID=1)"))
		lines = append(lines, fmt.Sprintf("GameModes=(FriendlyName=\"Versus\",ClassNameAndPath=\"KFGameContent.KFGameInfo_VersusSurvival\",bSoloPlaySupported=False,DifficultyLevels=0,Lengths=0,LocalizeID=2)"))
		lines = append(lines, fmt.Sprintf("bWaitForNetPlayers=True"))
		lines = append(lines, fmt.Sprintf("bEnableMapObjectives=True"))
		lines = append(lines, fmt.Sprintf("bLogScoring=False"))
		lines = append(lines, fmt.Sprintf("bLogAIDefaults=False"))
		lines = append(lines, fmt.Sprintf("bLogAICount=False"))
		lines = append(lines, fmt.Sprintf("EndOfGameDelay=15"))
		lines = append(lines, fmt.Sprintf("ServerExpirationForKillWhenEmpty=%d.000000", settings.EmptyServerDelay))
		lines = append(lines, fmt.Sprintf("RequiredMobileInputConfigs=(GroupName=\"DebugGroup\",RequireZoneNames=(\"DebugStickMoveZone\",\"DebugStickLookZone\",\"DebugLookZone\"),bIsAttractModeGroup=False)"))
		lines = append(lines, fmt.Sprintf("bIsStandbyCheckingEnabled=False"))
		lines = append(lines, fmt.Sprintf("GoalScore=0"))
		lines = append(lines, fmt.Sprintf("MaxLives=0"))
		lines = append(lines, fmt.Sprintf("TimeLimit=0"))
		lines = append(lines, fmt.Sprintf("StandbyRxCheatTime=0.000000"))
		lines = append(lines, fmt.Sprintf("StandbyTxCheatTime=0.000000"))
		lines = append(lines, fmt.Sprintf("BadPingThreshold=0"))
		lines = append(lines, fmt.Sprintf("PercentMissingForRxStandby=0.000000"))
		lines = append(lines, fmt.Sprintf("PercentMissingForTxStandby=0.000000"))
		lines = append(lines, fmt.Sprintf("PercentForBadPing=0.000000"))
		lines = append(lines, fmt.Sprintf("JoinInProgressStandbyWaitTime=0.000000"))
		lines = append(lines, fmt.Sprintf("DefaultGameType=KFGameContent.KFGameInfo_Survival"))
		lines = append(lines, fmt.Sprintf("AnimTreePoolSize=0"))
	}

	// Engine.GameReplicationInfo section
	lines = append(lines, "")
	lines = append(lines, "[Engine.GameReplicationInfo]")
	lines = append(lines, fmt.Sprintf("ServerName=%s", settings.ServerName))
	lines = append(lines, fmt.Sprintf("ShortName=KF2Server"))
	lines = append(lines, "")

	// Engine.GameEngine section
	lines = append(lines, "[Engine.GameEngine]")
	lines = append(lines, fmt.Sprintf("MasterVolumeMultiplier=100"))
	lines = append(lines, fmt.Sprintf("DialogVolumeMultiplier=100"))
	lines = append(lines, fmt.Sprintf("MusicVolumeMultiplier=50"))
	lines = append(lines, fmt.Sprintf("SFxVolumeMultiplier=100"))
	lines = append(lines, fmt.Sprintf("BattleChatterIndex=0"))
	lines = append(lines, fmt.Sprintf("GammaMultiplier=0.68"))
	lines = append(lines, fmt.Sprintf("bMusicVocalsEnabled=false"))
	lines = append(lines, fmt.Sprintf("bMinimalChatter=false"))
	lines = append(lines, fmt.Sprintf("bEnableAdvDebugLines=false"))
	lines = append(lines, fmt.Sprintf("bMuteOnLossOfFocus=True"))
	lines = append(lines, fmt.Sprintf("bShowCrossHair=false"))
	lines = append(lines, fmt.Sprintf("FOVOptionsPercentageValue=1"))
	lines = append(lines, fmt.Sprintf("GoreLevel=%d", settings.GoreLevel))
	lines = append(lines, "")

	// Weapon and gameplay modifiers
	lines = append(lines, "[KFGame.KFGameReplicationInfo]")
	lines = append(lines, fmt.Sprintf("WeaponPickupModifier=%s", settings.WeaponSpawnModifier))
	lines = append(lines, fmt.Sprintf("ZedHealthModifier=%s", settings.ZedHealthModifier))
	lines = append(lines, fmt.Sprintf("ZedHeadHealthModifier=%s", settings.ZedHeadHealthModifier))
	lines = append(lines, fmt.Sprintf("ZedMovementSpeedModifier=%s", settings.ZedMovementSpeedModifier))
	lines = append(lines, fmt.Sprintf("InitialSpawnRateModifier=%s", settings.InitialSpawnRateModifier))
	lines = append(lines, fmt.Sprintf("MaxSpawnRateModifier=%s", settings.MaxSpawnRateModifier))
	lines = append(lines, fmt.Sprintf("bPickupLifespanOverride=%s", boolToString(settings.DisablePickupsWhenFull)))

	return strings.Join(lines, "\n")
}

// generateKf2GSMConfig creates LinuxGSM config file content
func generateKf2GSMConfig(gsmConfig *gameserverv1alpha1.Kf2GSMConfig) string {
	var lines []string

	lines = append(lines, "# LinuxGSM configuration for Killing Floor 2")
	lines = append(lines, "# Generated by GameServer Operator")
	lines = append(lines, "")
	lines = append(lines, "# Server details")
	lines = append(lines, `servicename="kf2server"`)
	if gsmConfig.SteamUser != "" {
		lines = append(lines, fmt.Sprintf("steamuser=\"%s\"", gsmConfig.SteamUser))
	}
	if gsmConfig.SteamPass != "" {
		lines = append(lines, fmt.Sprintf("steampass='%s'", gsmConfig.SteamPass))
	}

	return strings.Join(lines, "\n")
}

// SetupWithManager sets up the controller with the Manager.
func (r *Kf2Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Temporarily disabled webhooks due to certificate issues
	// if err := (&Kf2Validator{}).SetupWebhookWithManager(mgr); err != nil {
	//	return err
	// }

	return ctrl.NewControllerManagedBy(mgr).
		For(&gameserverv1alpha1.Kf2{}).
		Complete(r)
}
