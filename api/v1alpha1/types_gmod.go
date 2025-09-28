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

// GmodSpec defines the desired state of Gmod
type GmodSpec struct {
	//+kubebuilder:default="gameservermanagers/gameserver:gmod"
	Image string `json:"image"`

	Base `json:",inline"`

	Config GmodConfig `json:"config,omitempty"`
}

// GmodConfig defines configuration for Garry's Mod & LinuxGSM
type GmodConfig struct {
	// Game server configuration (server.cfg)
	Game GmodServerConfig `json:"game,omitempty"`

	// LinuxGSM specific configuration
	GSM GmodGSMConfig `json:"gsm,omitempty"`
}

// GmodServerConfig defines server settings for Garry's Mod
type GmodServerConfig struct {
	// Server hostname
	//+kubebuilder:default="Garry's Mod Server"
	Hostname string `json:"hostname,omitempty"`

	// Server password (blank = no password)
	Password string `json:"password,omitempty"`

	// Server description
	//+kubebuilder:default="A Garry's Mod Server"
	ServerDescription string `json:"serverDescription,omitempty"`

	// Maximum number of players
	//+kubebuilder:default=16
	MaxPlayers int32 `json:"maxPlayers,omitempty"`

	// Server region (0=US East, 1=US West, 2=South America, 3=Europe, 4=Asia, 5=Australia, 6=Middle East, 7=Africa, 255=World)
	//+kubebuilder:default=255
	Region int32 `json:"region,omitempty"`

	// Game mode
	//+kubebuilder:default="sandbox"
	GameMode string `json:"gameMode,omitempty"`

	// Default map
	//+kubebuilder:default="gm_construct"
	DefaultMap string `json:"defaultMap,omitempty"`

	// Enable LAN only (0=Internet, 1=LAN only)
	//+kubebuilder:default=0
	LANOnly int32 `json:"lanOnly,omitempty"`

	// Allow custom CSS file loading
	//+kubebuilder:default=true
	AllowCSLua *bool `json:"allowCSLua,omitempty"`

	// Workshop collection ID for addons
	WorkshopCollectionID string `json:"workshopCollectionId,omitempty"`

	// Enable/disable fall damage
	//+kubebuilder:default=true
	FallDamage *bool `json:"fallDamage,omitempty"`

	// Enable/disable vehicle spawning
	//+kubebuilder:default=true
	AllowVehicles *bool `json:"allowVehicles,omitempty"`

	// Enable/disable NPC spawning
	//+kubebuilder:default=true
	AllowNPCs *bool `json:"allowNPCs,omitempty"`

	// Enable/disable weapons spawning
	//+kubebuilder:default=true
	AllowWeapons *bool `json:"allowWeapons,omitempty"`

	// Enable/disable props spawning
	//+kubebuilder:default=true
	AllowProps *bool `json:"allowProps,omitempty"`

	// Enable/disable effects spawning
	//+kubebuilder:default=true
	AllowEffects *bool `json:"allowEffects,omitempty"`

	// Enable/disable ragdolls spawning
	//+kubebuilder:default=true
	AllowRagdolls *bool `json:"allowRagdolls,omitempty"`

	// Admin SteamID64 (comma-separated for multiple admins)
	AdminSteamIDs string `json:"adminSteamIds,omitempty"`

	// Enable/disable M9K weapons
	//+kubebuilder:default=false
	M9KWeapons *bool `json:"m9kWeapons,omitempty"`

	// Enable/disable custom admin messages
	//+kubebuilder:default=false
	Ads *bool `json:"ads,omitempty"`

	// Enable/disable darkRP economy system
	//+kubebuilder:default=false
	DarkRPEconomy *bool `json:"darkRpEconomy,omitempty"`

	// Enable/disable map cycling
	//+kubebuilder:default=true
	MapCycle *bool `json:"mapCycle,omitempty"`

	// Map cycle file (relative path)
	//+kubebuilder:default="mapcycle.txt"
	MapCycleFile string `json:"mapCycleFile,omitempty"`

	// Voice icon display
	//+kubebuilder:default=true
	VoiceIcon *bool `json:"voiceIcon,omitempty"`

	// Dead players can hear live players
	//+kubebuilder:default=true
	Alltalk *bool `json:"alltalk,omitempty"`

	// Voice distance in units
	//+kubebuilder:default=0
	VoiceDistance int32 `json:"voiceDistance,omitempty"`

	// Fast download URL for workshop items
	FastDLURL string `json:"fastDlUrl,omitempty"`

	// Server owner SteamID64
	ServerOwnerSteamID string `json:"serverOwnerSteamId,omitempty"`

	// MOTD file URL or text
	MOTD string `json:"motd,omitempty"`

	// Enable/disable PvP
	//+kubebuilder:default=true
	PvP *bool `json:"pvp,omitempty"`

	// Logging level (1=minimal, 2=normal, 3=verbose)
	//+kubebuilder:default=1
	LoggingLevel int32 `json:"loggingLevel,omitempty"`

	// Maximum number of props per player
	//+kubebuilder:default=200
	MaxPropsPerPlayer int32 `json:"maxPropsPerPlayer,omitempty"`

	// Enable/disable prop deletion logging
	//+kubebuilder:default=false
	LogFile *bool `json:"logFile,omitempty"`

	// Server tags (comma-separated)
	ServerTags string `json:"serverTags,omitempty"`
}

// GmodGSMConfig defines LinuxGSM specific configuration
type GmodGSMConfig struct {
	// Custom LinuxGSM configuration file content
	ConfigFile string `json:"configFile,omitempty"`

	// Server port
	//+kubebuilder:default=27015
	Port int32 `json:"port,omitempty"`

	// Client port (used for connection)
	//+kubebuilder:default=27005
	ClientPort int32 `json:"clientPort,omitempty"`

	// SourceTV port
	//+kubebuilder:default=27020
	SourceTVPort int32 `json:"sourceTvPort,omitempty"`

	// Server tickrate
	//+kubebuilder:default=66
	TickRate int32 `json:"tickRate,omitempty"`

	// Workshop collection ID (alternative to server.cfg)
	WorkshopCollectionID string `json:"workshopCollectionId,omitempty"`

	// Game Login Token (GSLT) for public servers
	GSLT string `json:"gslt,omitempty"`

	// Enable/disable LinuxGSM stats reporting
	//+kubebuilder:default=false
	Stats *bool `json:"stats,omitempty"`
}

// GmodStatus defines the observed state of Gmod
type GmodStatus struct {
	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Gmod is the Schema for the gmods API
type Gmod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GmodSpec   `json:"spec,omitempty"`
	Status GmodStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GmodList contains a list of Gmod
type GmodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gmod `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Gmod{}, &GmodList{})
}
