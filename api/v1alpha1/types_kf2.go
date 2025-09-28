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

// Kf2Spec defines the desired state of Kf2
type Kf2Spec struct {
	//+kubebuilder:default="gameservermanagers/gameserver:kf2"
	Image string `json:"image"`

	Base `json:",inline"`

	Config Kf2Config `json:"config,omitempty"`
}

// Kf2Config defines configuration for KF2 & LinuxGSM
type Kf2Config struct {
	// Game server configuration (KFGame.ini)
	Game Kf2GameConfig `json:"game,omitempty"`

	// LinuxGSM specific configuration
	GSM Kf2GSMConfig `json:"gsm,omitempty"`
}

// Kf2GameConfig defines server configuration for Killing Floor 2
type Kf2GameConfig struct {
	// Server name displayed in server browser
	//+kubebuilder:default="KF2 Server"
	ServerName string `json:"serverName,omitempty"`

	// Server password (blank = no password)
	Password string `json:"password,omitempty"`

	// Admin password for server administration
	//+kubebuilder:default="adminpassword"
	AdminPassword string `json:"adminPassword,omitempty"`

	// Maximum number of players
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=12
	//+kubebuilder:default=6
	MaxPlayers int32 `json:"maxPlayers,omitempty"`

	// Server difficulty level (0=Normal, 1=Hard, 2=Suicidal, 3=Hell on Earth)
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=3
	//+kubebuilder:default=1
	Difficulty int32 `json:"difficulty,omitempty"`

	// Game mode type
	//+kubebuilder:validation:Enum="Survival";"WeeklySurvival";"VersusSurvival"
	//+kubebuilder:default="Survival"
	GameMode string `json:"gameMode,omitempty"`

	// Game length (0=Short, 1=Medium Normal, 2=Long)
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=2
	//+kubebuilder:default=1
	GameLength int32 `json:"gameLength,omitempty"`

	// Allow admins to pause the game
	//+kubebuilder:default=false
	AllowAdminPause bool `json:"allowAdminPause,omitempty"`

	// Maximum number of spectators
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=12
	//+kubebuilder:default=2
	MaxSpectators int32 `json:"maxSpectators,omitempty"`

	// Enable kick voting
	//+kubebuilder:default=true
	DisableKickVoting bool `json:"disableKickVoting,omitempty"`

	// Minimum percentage required for successful kick vote (0.5 = 50%)
	//+kubebuilder:default="0.5"
	KickVotePercentage string `json:"kickVotePercentage,omitempty"`

	// Time to wait for failed votes (seconds)
	//+kubebuilder:validation:Minimum=5
	//+kubebuilder:validation:Maximum=60
	//+kubebuilder:default=10
	TimeBetweenFailedVotes int32 `json:"timeBetweenFailedVotes,omitempty"`

	// Time to wait for vote result (seconds)
	//+kubebuilder:validation:Minimum=5
	//+kubebuilder:validation:Maximum=60
	//+kubebuilder:default=30
	VoteTime int32 `json:"voteTime,omitempty"`

	// Map cycle - array of map names
	//+kubebuilder:default={"KF-BurningParis","KF-BioticsLab","KF-Outpost","KF-VolterManor","KF-Catacombs"}
	MapCycle []string `json:"mapCycle,omitempty"`

	// Friendly Fire damage multiplier (0.0 = no friendly fire, 1.0 = full damage)
	//+kubebuilder:default="0.0"
	FriendlyFireScale string `json:"friendlyFireScale,omitempty"`

	// Enable server welcome message
	//+kubebuilder:default=true
	EnableWelcomeMessage bool `json:"enableWelcomeMessage,omitempty"`

	// Server welcome message
	//+kubebuilder:default="Welcome to our Killing Floor 2 server!"
	WelcomeMessage string `json:"welcomeMessage,omitempty"`

	// Server MOTD
	//+kubebuilder:default="Good luck and have fun!"
	ServerMOTD string `json:"serverMOTD,omitempty"`

	// Clan/Server name for display
	//+kubebuilder:default="KF2 Server"
	ClanMotto string `json:"clanMotto,omitempty"`

	// Enable VOIP in game
	//+kubebuilder:default=true
	EnableVOIP bool `json:"enableVOIP,omitempty"`

	// Enable Public VOIP channel
	//+kubebuilder:default=true
	EnablePublicVOIPChannel bool `json:"enablePublicVOIPChannel,omitempty"`

	// Enable spectator VOIP
	//+kubebuilder:default=false
	EnableSpectatorVOIP bool `json:"enableSpectatorVOIP,omitempty"`

	// Enable death to VOIP (dead players can talk but not hear)
	//+kubebuilder:default=true
	EnableDeadToVOIP bool `json:"enableDeadToVOIP,omitempty"`

	// Partition spectators (spectators only talk to other spectators)
	//+kubebuilder:default=false
	PartitionSpectators bool `json:"partitionSpectators,omitempty"`

	// Enable map objectives
	//+kubebuilder:default=true
	EnableMapObjectives bool `json:"enableMapObjectives,omitempty"`

	// Enable map voting
	//+kubebuilder:default=true
	EnableMapVoting bool `json:"enableMapVoting,omitempty"`

	// Map vote duration in seconds
	//+kubebuilder:validation:Minimum=30
	//+kubebuilder:validation:Maximum=120
	//+kubebuilder:default=60
	MapVoteDuration int32 `json:"mapVoteDuration,omitempty"`

	// Percentage required for map vote (0.0 = disable, 1.0 = 100% needed)
	//+kubebuilder:default="0.0"
	MapVotePercentage string `json:"mapVotePercentage,omitempty"`

	// Delay before game ends when empty (seconds)
	//+kubebuilder:validation:Minimum=60
	//+kubebuilder:validation:Maximum=600
	//+kubebuilder:default=120
	EmptyServerDelay int32 `json:"emptyServerDelay,omitempty"`

	// Gore level (0=None, 1=Reduced, 2=Full)
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=2
	//+kubebuilder:default=2
	GoreLevel int32 `json:"goreLevel,omitempty"`

	// Weapons can pick up while full
	//+kubebuilder:default=false
	DisablePickupsWhenFull bool `json:"disablePickupsWhenFull,omitempty"`

	// Weapon spawn rate modifier (0.5 = half, 1.0 = normal, 2.0 = double)
	//+kubebuilder:default="1.0"
	WeaponSpawnModifier string `json:"weaponSpawnModifier,omitempty"`

	// ZED health modifier (0.5 = half health, 1.0 = normal, 2.0 = double health)
	//+kubebuilder:default="1.0"
	ZedHealthModifier string `json:"zedHealthModifier,omitempty"`

	// ZED head health modifier (0.5 = half health, 1.0 = normal, 2.0 = double health)
	//+kubebuilder:default="1.0"
	ZedHeadHealthModifier string `json:"zedHeadHealthModifier,omitempty"`

	// ZED movement speed modifier (0.5 = half speed, 1.0 = normal, 2.0 = double speed)
	//+kubebuilder:default="1.0"
	ZedMovementSpeedModifier string `json:"zedMovementSpeedModifier,omitempty"`

	// Initial spawn rate modifier (controls initial ZED spawn rate)
	//+kubebuilder:default="1.0"
	InitialSpawnRateModifier string `json:"initialSpawnRateModifier,omitempty"`

	// Max spawn rate modifier (controls maximum ZED spawn rate)
	//+kubebuilder:default="1.0"
	MaxSpawnRateModifier string `json:"maxSpawnRateModifier,omitempty"`

	// Enable game analytics reporting
	//+kubebuilder:default=true
	EnableGameAnalytics bool `json:"enableGameAnalytics,omitempty"`

	// Enable recording game statistics
	//+kubebuilder:default=false
	RecordGameStats bool `json:"recordGameStats,omitempty"`
}

// Kf2GSMConfig defines LinuxGSM specific configuration for KF2
type Kf2GSMConfig struct {
	// Custom LinuxGSM config file content
	ConfigFile string `json:"configFile,omitempty"`

	// Steam username for server authentication (optional)
	SteamUser string `json:"steamUser,omitempty"`

	// Steam password for server authentication (optional)
	SteamPass string `json:"steamPass,omitempty"`
}

// Kf2Status defines the observed state of Kf2
type Kf2Status struct {
	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Kf2 is the Schema for the kf2s API
type Kf2 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Kf2Spec   `json:"spec,omitempty"`
	Status Kf2Status `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// Kf2List contains a list of Kf2
type Kf2List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kf2 `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kf2{}, &Kf2List{})
}
