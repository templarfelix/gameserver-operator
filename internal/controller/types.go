package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	gameserverv1alpha1 "github.com/templarfelix/gameserver-operator/api/v1alpha1"
)

// KubernetesClient extends client.Client with scheme
type KubernetesClient interface {
	runtime.Object
	Scheme() *runtime.Scheme
}

// GameServerSpec provides common interface for game specs
type GameServerSpec interface {
	GetImage() string
	GetBase() gameserverv1alpha1.Base
}

// GameServer provides common interface for game server CRDs
type GameServer interface {
	metav1.Object
	GetSpec() GameServerSpec
}
