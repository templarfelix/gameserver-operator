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

// SdtdSpec defines the desired state of Sdtd
type SdtdSpec struct {
	//+kubebuilder:default="gameservermanagers/gameserver:7dtd"
	Image string `json:"image"`

	Base `json:",inline"`

	Config SdtdConfig `json:"config,omitempty"`
}

// SdtdConfig defines configuration for 7 Days to Die & LinuxGSM
type SdtdConfig struct {
	// Game server configuration (serverconfig.xml)
	Game SdtdServerConfig `json:"game,omitempty"`

	// LinuxGSM specific configuration
	GSM ArkGSMConfig `json:"gsm,omitempty"`
}

// SdtdServerConfig defines server settings for 7 Days to Die
type SdtdServerConfig struct {
	// GENERAL SERVER SETTINGS

	// Server representation
	//+kubebuilder:default="7 Days to Die Server"
	ServerName string `json:"serverName,omitempty"`

	// Server description shown in server browser
	//+kubebuilder:default="A 7 Days to Die server"
	ServerDescription string `json:"serverDescription,omitempty"`

	// Server website URL
	ServerWebsiteURL string `json:"serverWebsiteURL,omitempty"`

	// Password to gain entry to the server
	ServerPassword string `json:"serverPassword,omitempty"`

	// Server login confirmation text
	ServerLoginConfirmationText string `json:"serverLoginConfirmationText,omitempty"`

	// The region this server is in
	//+kubebuilder:validation:Enum=NorthAmericaEast;NorthAmericaWest;CentralAmerica;SouthAmerica;Europe;Russia;Asia;MiddleEast;Africa;Oceania
	//+kubebuilder:default="NorthAmericaEast"
	Region string `json:"region,omitempty"`

	// Primary language for players on this server
	//+kubebuilder:default="English"
	Language string `json:"language,omitempty"`

	// Networking
	//+kubebuilder:default=26900
	ServerPort int32 `json:"serverPort,omitempty"`

	// Server visibility (2=public, 1=friends only, 0=unlisted)
	//+kubebuilder:default=2
	ServerVisibility int32 `json:"serverVisibility,omitempty"`

	// Networking protocols to disable
	//+kubebuilder:default="SteamNetworking"
	ServerDisabledNetworkProtocols string `json:"serverDisabledNetworkProtocols,omitempty"`

	// Maximum world transfer speed in KiB/s
	//+kubebuilder:default=512
	ServerMaxWorldTransferSpeedKiBs int32 `json:"serverMaxWorldTransferSpeedKiBs,omitempty"`

	// Slots
	//+kubebuilder:default=8
	ServerMaxPlayerCount int32 `json:"serverMaxPlayerCount,omitempty"`

	// Reserved slots count (out of MaxPlayerCount)
	//+kubebuilder:default=0
	ServerReservedSlots int32 `json:"serverReservedSlots,omitempty"`

	// Required permission level to use reserved slots
	//+kubebuilder:default=100
	ServerReservedSlotsPermission int32 `json:"serverReservedSlotsPermission,omitempty"`

	// Admin slots (can join when server full)
	//+kubebuilder:default=0
	ServerAdminSlots int32 `json:"serverAdminSlots,omitempty"`

	// Required permission level for admin slots
	//+kubebuilder:default=0
	ServerAdminSlotsPermission int32 `json:"serverAdminSlotsPermission,omitempty"`

	// Admin interfaces - Web Dashboard
	//+kubebuilder:default=false
	WebDashboardEnabled bool `json:"webDashboardEnabled,omitempty"`

	// Port of the web dashboard
	//+kubebuilder:default=8080
	WebDashboardPort int32 `json:"webDashboardPort,omitempty"`

	// External URL for web dashboard
	WebDashboardUrl string `json:"webDashboardUrl,omitempty"`

	// Enable map rendering for dashboard
	//+kubebuilder:default=false
	EnableMapRendering bool `json:"enableMapRendering,omitempty"`

	// Telnet
	//+kubebuilder:default=true
	TelnetEnabled bool `json:"telnetEnabled,omitempty"`

	// Telnet port
	//+kubebuilder:default=8081
	TelnetPort int32 `json:"telnetPort,omitempty"`

	// Telnet password
	TelnetPassword string `json:"telnetPassword,omitempty"`

	// Telnet failed login limit before blocking
	//+kubebuilder:default=10
	TelnetFailedLoginLimit int32 `json:"telnetFailedLoginLimit,omitempty"`

	// Telnet block duration in seconds
	//+kubebuilder:default=10
	TelnetFailedLoginsBlocktime int32 `json:"telnetFailedLoginsBlocktime,omitempty"`

	// Terminal window (Windows only)
	//+kubebuilder:default=true
	TerminalWindowEnabled bool `json:"terminalWindowEnabled,omitempty"`

	// Admin file name
	//+kubebuilder:default="serveradmin.xml"
	AdminFileName string `json:"adminFileName,omitempty"`

	// Technical settings
	//+kubebuilder:default=true
	EACEnabled bool `json:"eacEnabled,omitempty"`

	// Hide command execution log (0=all, 1=hide from telnet, 2=hide from clients, 3=hide all)
	//+kubebuilder:default=0
	HideCommandExecutionLog int32 `json:"hideCommandExecutionLog,omitempty"`

	// Max uncovered map chunks per player
	//+kubebuilder:default=131072
	MaxUncoveredMapChunksPerPlayer int32 `json:"maxUncoveredMapChunksPerPlayer,omitempty"`

	// Persistent player profiles
	//+kubebuilder:default=false
	PersistentPlayerProfiles bool `json:"persistentPlayerProfiles,omitempty"`

	// GAMEPLAY - World
	//+kubebuilder:default="Navezgane"
	GameWorld string `json:"gameWorld,omitempty"`

	// World generation seed (for RWG worlds)
	GameWorldSeed string `json:"gameWorldSeed,omitempty"`

	// World generation size
	//+kubebuilder:default=6144
	GameWorldSize int32 `json:"gameWorldSize,omitempty"`

	// Game name affects save game and seed
	GameName string `json:"gameName,omitempty"`

	// GameModeSurvival
	//+kubebuilder:default="GameModeSurvival"
	GameMode string `json:"gameMode,omitempty"`

	// Difficulty
	//+kubebuilder:default=1
	GameDifficulty int32 `json:"gameDifficulty,omitempty"`

	// Block damage by players (%)
	//+kubebuilder:default=100
	BlockDamagePlayer int32 `json:"blockDamagePlayer,omitempty"`

	// Block damage by AI (%)
	//+kubebuilder:default=100
	BlockDamageAI int32 `json:"blockDamageAI,omitempty"`

	// Block damage by AI during blood moon (%)
	//+kubebuilder:default=100
	BlockDamageAIBM int32 `json:"blockDamageAIBM,omitempty"`

	// XP multiplier (%)
	//+kubebuilder:default=100
	XPMultiplier int32 `json:"xpMultiplier,omitempty"`

	// Player safe zone level
	//+kubebuilder:default=5
	PlayerSafeZoneLevel int32 `json:"playerSafeZoneLevel,omitempty"`

	// Player safe zone hours
	//+kubebuilder:default=5
	PlayerSafeZoneHours int32 `json:"playerSafeZoneHours,omitempty"`

	// Cheat mode
	//+kubebuilder:default=false
	BuildCreate bool `json:"buildCreate,omitempty"`

	// Real time minutes per game day
	//+kubebuilder:default=60
	DayNightLength int32 `json:"dayNightLength,omitempty"`

	// Game hours of sunlight per day
	//+kubebuilder:default=18
	DayLightLength int32 `json:"dayLightLength,omitempty"`

	// Drop on death (0=nothing, 1=everything, 2=toolbelt, 3=backpack, 4=delete all)
	//+kubebuilder:default=1
	DropOnDeath int32 `json:"dropOnDeath,omitempty"`

	// Drop on quit (0=nothing, 1=everything, 2=toolbelt, 3=backpack)
	//+kubebuilder:default=0
	DropOnQuit int32 `json:"dropOnQuit,omitempty"`

	// Bedroll deadzone size
	//+kubebuilder:default=15
	BedrollDeadZoneSize int32 `json:"bedrollDeadZoneSize,omitempty"`

	// Bedroll expiry time in days
	//+kubebuilder:default=45
	BedrollExpiryTime int32 `json:"bedrollExpiryTime,omitempty"`

	// Performance - Zombie limits
	//+kubebuilder:default=64
	MaxSpawnedZombies int32 `json:"maxSpawnedZombies,omitempty"`

	// Animal limit
	//+kubebuilder:default=50
	MaxSpawnedAnimals int32 `json:"maxSpawnedAnimals,omitempty"`

	// Max allowed view distance
	//+kubebuilder:default=12
	ServerMaxAllowedViewDistance int32 `json:"serverMaxAllowedViewDistance,omitempty"`

	// Max queued mesh layers
	//+kubebuilder:default=1000
	MaxQueuedMeshLayers int32 `json:"maxQueuedMeshLayers,omitempty"`

	// Zombie settings
	//+kubebuilder:default=true
	EnemySpawnMode bool `json:"enemySpawnMode,omitempty"`

	// Enemy difficulty (0=Normal, 1=Feral)
	//+kubebuilder:default=0
	EnemyDifficulty int32 `json:"enemyDifficulty,omitempty"`

	// Zombie feral sense (0=Off, 1=Day, 2=Night, 3=All)
	//+kubebuilder:default=0
	ZombieFeralSense int32 `json:"zombieFeralSense,omitempty"`

	// Zombie movement (0=walk, 1=jog, 2=run, 3=sprint, 4=nightmare)
	//+kubebuilder:default=0
	ZombieMove int32 `json:"zombieMove,omitempty"`

	// Zombie movement at night
	//+kubebuilder:default=3
	ZombieMoveNight int32 `json:"zombieMoveNight,omitempty"`

	// Zombie feral movement
	//+kubebuilder:default=3
	ZombieFeralMove int32 `json:"zombieFeralMove,omitempty"`

	// Zombie blood moon movement
	//+kubebuilder:default=3
	ZombieBMMove int32 `json:"zombieBMMove,omitempty"`

	// Blood moon frequency (days, 0=none)
	//+kubebuilder:default=7
	BloodMoonFrequency int32 `json:"bloodMoonFrequency,omitempty"`

	// Blood moon range (days deviation)
	//+kubebuilder:default=0
	BloodMoonRange int32 `json:"bloodMoonRange,omitempty"`

	// Blood moon warning hour
	//+kubebuilder:default=8
	BloodMoonWarning int32 `json:"bloodMoonWarning,omitempty"`

	// Blood moon enemy count per player
	//+kubebuilder:default=8
	BloodMoonEnemyCount int32 `json:"bloodMoonEnemyCount,omitempty"`

	// Loot settings
	//+kubebuilder:default=100
	LootAbundance int32 `json:"lootAbundance,omitempty"`

	// Loot respawn days
	//+kubebuilder:default=7
	LootRespawnDays int32 `json:"lootRespawnDays,omitempty"`

	// Air drop frequency (hours)
	//+kubebuilder:default=72
	AirDropFrequency int32 `json:"airDropFrequency,omitempty"`

	// Air drop marker
	//+kubebuilder:default=true
	AirDropMarker bool `json:"airDropMarker,omitempty"`

	// Multiplayer
	//+kubebuilder:default=100
	PartySharedKillRange int32 `json:"partySharedKillRange,omitempty"`

	// Player killing mode (0=No Killing, 1=Allies Only, 2=Strangers Only, 3=Everyone)
	//+kubebuilder:default=3
	PlayerKillingMode int32 `json:"playerKillingMode,omitempty"`

	// Land claim options
	//+kubebuilder:default=3
	LandClaimCount int32 `json:"landClaimCount,omitempty"`

	// Land claim size in blocks
	//+kubebuilder:default=41
	LandClaimSize int32 `json:"landClaimSize,omitempty"`

	// Land claim deadzone distance
	//+kubebuilder:default=30
	LandClaimDeadZone int32 `json:"landClaimDeadZone,omitempty"`

	// Land claim expiry time (days)
	//+kubebuilder:default=7
	LandClaimExpiryTime int32 `json:"landClaimExpiryTime,omitempty"`

	// Land claim decay mode (0=Slow, 1=Fast, 2=None)
	//+kubebuilder:default=0
	LandClaimDecayMode int32 `json:"landClaimDecayMode,omitempty"`

	// Land claim online durability modifier
	//+kubebuilder:default=4
	LandClaimOnlineDurabilityModifier int32 `json:"landClaimOnlineDurabilityModifier,omitempty"`

	// Land claim offline durability modifier
	//+kubebuilder:default=4
	LandClaimOfflineDurabilityModifier int32 `json:"landClaimOfflineDurabilityModifier,omitempty"`

	// Land claim offline delay (minutes)
	//+kubebuilder:default=0
	LandClaimOfflineDelay int32 `json:"landClaimOfflineDelay,omitempty"`

	// Dynamic mesh system
	//+kubebuilder:default=true
	DynamicMeshEnabled bool `json:"dynamicMeshEnabled,omitempty"`

	// Dynamic mesh only in land claims
	//+kubebuilder:default=true
	DynamicMeshLandClaimOnly bool `json:"dynamicMeshLandClaimOnly,omitempty"`

	// Dynamic mesh land claim buffer
	//+kubebuilder:default=3
	DynamicMeshLandClaimBuffer int32 `json:"dynamicMeshLandClaimBuffer,omitempty"`

	// Dynamic mesh max item cache
	//+kubebuilder:default=3
	DynamicMeshMaxItemCache int32 `json:"dynamicMeshMaxItemCache,omitempty"`

	// Twitch integration permission level
	//+kubebuilder:default=90
	TwitchServerPermission int32 `json:"twitchServerPermission,omitempty"`

	// Allow twitch actions during blood moon
	//+kubebuilder:default=false
	TwitchBloodMoonAllowed bool `json:"twitchBloodMoonAllowed,omitempty"`

	// Chunk age limit in days (-1=disabled)
	//+kubebuilder:default=-1
	MaxChunkAge int32 `json:"maxChunkAge,omitempty"`

	// Save data limit in MB (-1=disabled)
	//+kubebuilder:default=-1
	SaveDataLimit int32 `json:"saveDataLimit,omitempty"`
}

// SdtdStatus defines the observed state of Sdtd
type SdtdStatus struct {
	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Sdtd is the Schema for the sdtds API
type Sdtd struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SdtdSpec   `json:"spec,omitempty"`
	Status SdtdStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SdtdList contains a list of Sdtd
type SdtdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sdtd `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sdtd{}, &SdtdList{})
}
