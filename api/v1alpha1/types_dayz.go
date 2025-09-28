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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DayzSpec defines the desired state of Dayz
type DayzSpec struct {
	//+kubebuilder:default="gameservermanagers/gameserver:dayz"
	Image string `json:"image"`

	Base `json:",inline"`

	Config DayzConfig `json:"config,omitempty"`
}

// DayzConfig defines configuration for Dayz & LinuxGSM
type DayzConfig struct {
	// Game server configuration (server.cfg)
	Game GameServerConfig `json:"game,omitempty"`

	// LinuxGSM specific configuration
	GSM ArkGSMConfig `json:"gsm,omitempty"`
}

// GameServerConfig defines common game server settings for DayZ
type GameServerConfig struct {
	// Server hostname
	//+kubebuilder:default="DayZ Server"
	Hostname string `json:"hostname,omitempty"`

	// Server password (blank = no password)
	Password string `json:"password,omitempty"`

	// Admin password for server administration
	//+kubebuilder:default="adminpassword"
	AdminPassword string `json:"adminPassword,omitempty"`

	// Maximum number of players
	//+kubebuilder:default=60
	MaxPlayers int32 `json:"maxPlayers,omitempty"`

	// Disable third person view
	//+kubebuilder:default=false
	Disable3rdPerson bool `json:"disable3rdPerson,omitempty"`

	// Disable crosshair
	//+kubebuilder:default=false
	DisableCrosshair bool `json:"disableCrosshair,omitempty"`

	// Enable server VOIP
	//+kubebuilder:default=true
	DisableVoIP bool `json:"disableVoIP,omitempty"`

	// Enable voice over IP on the server
	//+kubebuilder:default=true
	VON int32 `json:"von,omitempty"`

	// Server time acceleration
	//+kubebuilder:default="12"
	ServerTimeAcceleration string `json:"serverTimeAcceleration,omitempty"`

	// Server night time acceleration
	//+kubebuilder:default="4"
	ServerNightTimeAcceleration string `json:"serverNightTimeAcceleration,omitempty"`

	// Respawn time
	//+kubebuilder:default=300
	RespawnTime int32 `json:"respawnTime,omitempty"`

	// Server time
	//+kubebuilder:default=8
	ServerTime string `json:"serverTime,omitempty"`

	// Login timeout
	//+kubebuilder:default=60
	SteamQueryPort int32 `json:"steamQueryPort,omitempty"`

	// Enable BattlEye
	//+kubebuilder:default=true
	BattlEye bool `json:"battlEye,omitempty"`

	// Verify signatures
	//+kubebuilder:default=true
	VerifySignatures bool `json:"verifySignatures,omitempty"`

	// Force same build
	//+kubebuilder:default=true
	ForceSameBuild bool `json:"forceSameBuild,omitempty"`

	// Allowed build mismatch version
	//+kubebuilder:default=""
	AllowedBuild string `json:"allowedBuild,omitempty"`

	// Enable AI
	//+kubebuilder:default=false
	EnableAI bool `json:"enableAI,omitempty"`

	// Max AI count
	//+kubebuilder:default=100
	MaxAI int32 `json:"maxAI,omitempty"`

	// Force voice codec
	//+kubebuilder:default=0
	ForceVoiceCodec int32 `json:"forceVoiceCodec,omitempty"`

	// Limit FPS
	//+kubebuilder:default=30
	LimitFPS int32 `json:"limitFPS,omitempty"`

	// Allow statistics
	//+kubebuilder:default=true
	Stats bool `json:"stats,omitempty"`

	// Statistics port
	//+kubebuilder:default=2302
	StatsPort int32 `json:"statsPort,omitempty"`

	// Statistics IP
	StatsIP string `json:"statsIP,omitempty"`

	// Statistics password
	StatsPassword string `json:"statsPassword,omitempty"`

	// Required build
	RequiredBuild int32 `json:"requiredBuild,omitempty"`

	// Enable protection for persisted data
	PersistencyDisabled bool `json:"persistencyDisabled,omitempty"`

	// Disable AI gathering around players
	DisableAI int32 `json:"disableAI,omitempty"`

	// Show stream statistics
	ShowStreamStatistics bool `json:"showStreamStatistics,omitempty"`

	// Server time persistent
	//+kubebuilder:default=true
	ServerTimePersistent bool `json:"serverTimePersistent,omitempty"`

	// Disable automatic server gathering
	DisableAutoGroup bool `json:"disableAutoGroup,omitempty"`
}

// DayzStatus defines the observed state of Dayz
type DayzStatus struct {
	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Dayz is the Schema for the dayzs API
type Dayz struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DayzSpec   `json:"spec,omitempty"`
	Status DayzStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DayzList contains a list of Dayz
type DayzList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dayz `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dayz{}, &DayzList{})
}
