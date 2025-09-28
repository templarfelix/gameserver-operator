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

// ProjectZomboidSpec defines the desired state of ProjectZomboid
type ProjectZomboidSpec struct {
	//+kubebuilder:default="gameservermanagers/gameserver:project-zomboid"
	Image string `json:"image"`

	Base `json:",inline"`

	Config ProjectZomboidConfig `json:"config,omitempty"`
}

// ProjectZomboidConfig defines configuration for Project Zomboid & LinuxGSM
type ProjectZomboidConfig struct {
	// Game server configuration (server.ini)
	Game ProjectZomboidServerConfig `json:"game,omitempty"`

	// LinuxGSM specific configuration
	GSM ArkGSMConfig `json:"gsm,omitempty"`
}

// ProjectZomboidServerConfig defines server settings for Project Zomboid
type ProjectZomboidServerConfig struct {
	// Server name
	//+kubebuilder:default="Project Zomboid Server"
	ServerName string `json:"serverName,omitempty"`

	// Server description
	//+kubebuilder:default="A Project Zomboid Server"
	ServerDescription string `json:"serverDescription,omitempty"`

	// Password for joining the server
	Password string `json:"password,omitempty"`

	// Admin password
	//+kubebuilder:default="adminpassword"
	AdminPassword string `json:"adminPassword,omitempty"`

	// Maximum players
	//+kubebuilder:default=32
	MaxPlayers int32 `json:"maxPlayers,omitempty"`

	// Port for the server
	//+kubebuilder:default=16261
	DefaultPort int32 `json:"defaultPort,omitempty"`

	// UDP port for the server
	//+kubebuilder:default=16262
	UDPPort int32 `json:"udpPort,omitempty"`

	// Reset ID - increments by 1 every time server configuration is reset
	ResetID int32 `json:"resetId,omitempty"`

	// Save interval (minutes)
	//+kubebuilder:default=15
	SaveWorldEveryMinutes int32 `json:"saveWorldEveryMinutes,omitempty"`

	// Player respawn time in days
	//+kubebuilder:default=0
	PlayerRespawnWithSelf int32 `json:"playerRespawnWithSelf,omitempty"`

	// Player respawn time in hours
	//+kubebuilder:default=0
	PlayerRespawnWithOther int32 `json:"playerRespawnWithOther,omitempty"`

	// Drop off white-listed objects when player dies
	DropOffWhiteListedObjects bool `json:"dropOffWhiteListedObjects,omitempty"`

	// Fast forward multiplier (advanced)
	//+kubebuilder:default=1
	FastForwardMultiplier int32 `json:"fastForwardMultiplier,omitempty"`

	// Pause when no players online
	//+kubebuilder:default=false
	PauseOnEmptyServer bool `json:"pauseOnEmptyServer,omitempty"`

	// Maximum accounts per user
	//+kubebuilder:default=0
	MaxAccountsPerUser int32 `json:"maxAccountsPerUser,omitempty"`

	// Enable PVP
	//+kubebuilder:default=true
	PVP bool `json:"pvp,omitempty"`

	// Safehouse allow respawn
	//+kubebuilder:default=true
	SafehouseAllowRespawn bool `json:"safehouseAllowRespawn,omitempty"`

	// Safehouse allow non-members claim
	//+kubebuilder:default=false
	SafehouseAllowTrepass bool `json:"safehouseAllowTrepass,omitempty"`

	// Sleep allowed
	//+kubebuilder:default=true
	SleepAllowed bool `json:"sleepAllowed,omitempty"`

	// Damage multiplier
	//+kubebuilder:default="3"
	DamageMultiplier string `json:"damageMultiplier,omitempty"`

	// Bleeding chance (0-100)
	//+kubebuilder:default="100"
	BleedingChance string `json:"bleedingChance,omitempty"`

	// Minutes per page (for reading books etc)
	//+kubebuilder:default="1.0"
	MinutesPerPage string `json:"minutesPerPage,omitempty"`

	// Hours for loot respawn
	//+kubebuilder:default="0.0"
	HoursForLootRespawn string `json:"hoursForLootRespawn,omitempty"`

	// Max items for loot respawn
	//+kubebuilder:default="4"
	MaxItemsForLootRespawn string `json:"maxItemsForLootRespawn,omitempty"`

	// Construction pre-requisites
	//+kubebuilder:default=true
	ConstructionPreRequisites bool `json:"constructionPreRequisites,omitempty"`

	// Nutrition enabled
	//+kubebuilder:default=true
	Nutrition bool `json:"nutrition,omitempty"`

	// Food rot speed multiplier
	//+kubebuilder:default="1.0"
	FoodRotSpeed string `json:"foodRotSpeed,omitempty"`

	// World erase speed
	//+kubebuilder:default=90
	WorldEraseSpeed int32 `json:"worldEraseSpeed,omitempty"`

	// Player safehouse cooldown
	//+kubebuilder:default="2"
	PlayerSafehouseCooldown string `json:"playerSafehouseCooldown,omitempty"`

	// Admin safehouse cooldown
	//+kubebuilder:default="2"
	AdminSafehouseCooldown string `json:"adminSafehouseCooldown,omitempty"`

	// Safehouse day survived to claim
	//+kubebuilder:default="0"
	SafehouseDaySurvivor string `json:"safehouseDaySurvivor,omitempty"`

	// Remove expire zombie
	//+kubebuilder:default=false
	RemoveExpiredZombies bool `json:"removeExpiredZombies,omitempty"`

	// Allow non-members to destroy safehouses
	//+kubebuilder:default=false
	SafehouseAllowDestroy bool `json:"safehouseAllowDestroy,omitempty"`

	// Enable punching
	//+kubebuilder:default=true
	AllowDestructionBySledgehammer bool `json:"allowDestructionBySledgehammer,omitempty"`

	// Minutes surviving bonus multiplier
	//+kubebuilder:default="1.0"
	MinutesPerDay string `json:"minutesPerDay,omitempty"`

	// Zombie lure distance
	//+kubebuilder:default="0.0"
	ZombieLureDistance string `json:"zombieLureDistance,omitempty"`

	// Zombie lure interval
	//+kubebuilder:default=0
	ZombieLureInterval int32 `json:"zombieLureInterval,omitempty"`

	// Enable global chat
	//+kubebuilder:default=false
	GlobalChat bool `json:"globalChat,omitempty"`

	// Chat streams
	//+kubebuilder:default=0
	ChatStreams int32 `json:"chatStreams,omitempty"`

	// Server welcome message
	ServerWelcomeMessage string `json:"serverWelcomeMessage,omitempty"`

	// Open whitelist mod
	//+kubebuilder:default="server"
	OpenWhitelistMod string `json:"openWhitelistMod,omitempty"`

	// Ban kick time
	//+kubebuilder:default=1
	BannedPlayerKickedTime int32 `json:"bannedPlayerKickedTime,omitempty"`

	// Server player ID
	ServerPlayerID int32 `json:"serverPlayerID,omitempty"`

	// Ping limit
	//+kubebuilder:default=400
	PingLimit int32 `json:"pingLimit,omitempty"`

	// Workshop items (comma-separated Steam workshop IDs)
	WorkshopItems string `json:"workshopItems,omitempty"`

	// Mods (comma-separated mod names)
	Mods string `json:"mods,omitempty"`

	// Map (what world generator to use)
	//+kubebuilder:validation:Enum="Muldraugh, KY"
	//+kubebuilder:default="Muldraugh, KY"
	Map string `json:"map,omitempty"`

	// Zombie population
	//+kubebuilder:default=3
	ZombiePopulation int32 `json:"zombiePopulation,omitempty"`

	// Zombie migrate distance
	//+kubebuilder:default=1000
	ZombieMigrateDistance int32 `json:"zombieMigrateDistance,omitempty"`

	// Zombies respawn rate
	//+kubebuilder:default=-1
	ZombieRespawnRate int32 `json:"zombieRespawnRate,omitempty"`

	// Zombies respawn period
	//+kubebuilder:default=1000
	ZombieRespawnPeriod int32 `json:"zombieRespawnPeriod,omitempty"`
}

// ProjectZomboidStatus defines the observed state of ProjectZomboid
type ProjectZomboidStatus struct {
	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ProjectZomboid is the Schema for the projectzomboids API
type ProjectZomboid struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectZomboidSpec   `json:"spec,omitempty"`
	Status ProjectZomboidStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProjectZomboidList contains a list of ProjectZomboid
type ProjectZomboidList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectZomboid `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProjectZomboid{}, &ProjectZomboidList{})
}
