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

// ArkSpec defines the desired state of Ark
type ArkSpec struct {
	//+kubebuilder:default="gameservermanagers/gameserver:ark"
	Image string `json:"image"`

	Base `json:",inline"`

	Config ArkConfig `json:"config,omitempty"`
}

// ArkConfig defines configuration for Ark & LinuxGSM
type ArkConfig struct {
	// Game configuration settings (GameUserSettings.ini)
	GameUserSettings ArkGameUserSettings `json:"gameUserSettings,omitempty"`

	// Advanced game configuration (Game.ini)
	Game AdvancedGameSettings `json:"game,omitempty"`

	// LinuxGSM specific configuration
	GSM ArkGSMConfig `json:"gsm,omitempty"`
}

// ArkGameUserSettings defines ARK GameUserSettings.ini configuration
type ArkGameUserSettings struct {
	// Server settings
	ServerSettings ServerSettings `json:"serverSettings,omitempty"`

	// Admin password
	//+kubebuilder:default="adminpassword"
	AdminPassword string `json:"adminPassword,omitempty"`

	// Enable/disable mods
	//+kubebuilder:default=true
	ModInstaller *bool `json:"modInstaller,omitempty"`

	// Mod IDs to install (comma-separated)
	Mods string `json:"mods,omitempty"`

	// Server Map
	//+kubebuilder:validation:Enum=TheIsland;TheCenter;ScorchedEarth;Aberration;Extinction;Genesis;TheIslandII;Fjordur;LostIsland;Valguero;CrystalIsles;Genesis2;Ragnarok;Newtastic
	//+kubebuilder:default="TheIsland"
	ServerMap string `json:"serverMap,omitempty"`

	// Max number of players
	//+kubebuilder:default=20
	MaxPlayers int32 `json:"maxPlayers,omitempty"`

	// Difficulty level (0.2 - 1.0)
	//+kubebuilder:default="1.0"
	Difficulty string `json:"difficulty,omitempty"`

	// Enable/disable Crossplay
	//+kubebuilder:default=true
	Crossplay *bool `json:"crossplay,omitempty"`

	// Server port (default: 7777)
	//+kubebuilder:default=7777
	ServerPort int32 `json:"serverPort,omitempty"`

	// RCON port (default: 27020)
	//+kubebuilder:default=27020
	RconPort int32 `json:"rconPort,omitempty"`

	// Enable/disable RCON
	//+kubebuilder:default=true
	EnableRcon *bool `json:"enableRcon,omitempty"`
}

// ServerSettings defines ARK server configuration settings
type ServerSettings struct {
	// Server name
	ServerName string `json:"serverName,omitempty"`

	// Server description
	ServerDescription string `json:"serverDescription,omitempty"`

	// Night time speed scale
	//+kubebuilder:default="1.0"
	NightTimeSpeed string `json:"nightTimeSpeed,omitempty"`

	// Day cycle speed scale
	//+kubebuilder:default="1.0"
	DayCycleSpeedScale string `json:"dayCycleSpeedScale,omitempty"`

	// Single player settings
	//+kubebuilder:default=false
	Singleplayer bool `json:"singleplayer,omitempty"`

	// Show player map location
	//+kubebuilder:default=true
	ShowPlayerMapLocation *bool `json:"showPlayerMapLocation,omitempty"`

	// Enable PVP
	//+kubebuilder:default=true
	PVP *bool `json:"pvp,omitempty"`

	// Allow friendly fire
	//+kubebuilder:default=true
	AllowFriendlyFire *bool `json:"allowFriendlyFire,omitempty"`

	// Tamed dino damage multiplier
	//+kubebuilder:default="1.0"
	TamedDinoDamageMultiplier string `json:"tamedDinoDamageMultiplier,omitempty"`

	// Wild dino damage multiplier
	//+kubebuilder:default="1.0"
	WildDinoDamageMultiplier string `json:"wildDinoDamageMultiplier,omitempty"`

	// Disable PVP option
	//+kubebuilder:default=false
	DisablePvPOption bool `json:"disablePvPOption,omitempty"`

	// Enable/disable loot crates
	//+kubebuilder:default=true
	DisableLootCrates *bool `json:"disableLootCrates,omitempty"`

	// Disable structure placement collision
	//+kubebuilder:default=false
	DisableStructurePlacementCollision bool `json:"disableStructurePlacementCollision,omitempty"`

	// Max number of tamed dinos
	//+kubebuilder:default=4000
	MaxNumberOfPlayers int32 `json:"maxNumberOfPlayers,omitempty"`

	// Increase harvesting damage
	//+kubebuilder:default=false
	UseSinglePlayerSettings bool `json:"useSinglePlayerSettings,omitempty"`

	// Server PVE (overrides PVP setting)
	//+kubebuilder:default=false
	ServerPVE bool `json:"serverPVE,omitempty"`

	// Show map location
	//+kubebuilder:default=true
	ShowMapLocation *bool `json:"showMapLocation,omitempty"`

	// Prevent tribe alliances
	//+kubebuilder:default=false
	PreventTribeAlliances bool `json:"preventTribeAlliances,omitempty"`

	// Player damage multiplier
	//+kubebuilder:default="1.0"
	PlayerDamageMultiplier string `json:"playerDamageMultiplier,omitempty"`

	// Structure damage multiplier
	//+kubebuilder:default="1.0"
	StructureDamageMultiplier string `json:"structureDamageMultiplier,omitempty"`

	// Player resistance
	//+kubebuilder:default="1.0"
	PlayerResistance string `json:"playerResistance,omitempty"`

	// Auto save period minutes
	//+kubebuilder:default=15
	AutoSavePeriodMinutes int32 `json:"autoSavePeriodMinutes,omitempty"`

	// Disable PVP balance
	//+kubebuilder:default=false
	DisablePvPAutoBalance bool `json:"disablePvPAutoBalance,omitempty"`

	// Dino food drain multiplier
	//+kubebuilder:default="1.0"
	DinoFoodDrainMultiplier string `json:"dinoFoodDrainMultiplier,omitempty"`

	// Change max dinos
	//+kubebuilder:default=4000
	MaximumNumberOfPlayers int32 `json:"maximumNumberOfPlayers,omitempty"`

	// Dino count multiplier
	//+kubebuilder:default="1.0"
	DinoCountMultiplier string `json:"dinoCountMultiplier,omitempty"`
}

// AdvancedGameSettings defines ARK Game.ini configuration
type AdvancedGameSettings struct {
	// Resource harvesting rates
	Harvesting HarvestingRates `json:"harvesting,omitempty"`

	// Breeding rates
	Breeding BreedingRates `json:"breeding,omitempty"`

	// Difficulty settings
	Difficulty DifficultySettings `json:"difficulty,omitempty"`

	// Custom rate settings
	CustomRates CustomRateSettings `json:"customRates,omitempty"`
}

// HarvestingRates defines resource harvesting multipliers
type HarvestingRates struct {
	// Day cycle speed scale
	DayCycleSpeedScale string `json:"dayCycleSpeedScale,omitempty"`

	// Day time speed scale
	DayTimeSpeedScale string `json:"dayTimeSpeedScale,omitempty"`

	// Harvesting damage multiplier
	DamageMultiplier string `json:"damageMultiplier,omitempty"`

	// Use single player harvesting damage
	UseSinglePlayerDamage bool `json:"useSinglePlayerDamage,omitempty"`

	// Metal harvesting quantity
	MetalHarvestQuantity string `json:"metalHarvestQuantity,omitempty"`

	// Wood harvesting quantity
	WoodHarvestQuantity string `json:"woodHarvestQuantity,omitempty"`

	// Stone harvesting quantity
	StoneHarvestQuantity string `json:"stoneHarvestQuantity,omitempty"`

	// Fiber harvesting quantity
	ThatchHarvestQuantity string `json:"thatchHarvestQuantity,omitempty"`

	// Flint harvesting quantity
	FlintHarvestQuantity string `json:"flintHarvestQuantity,omitempty"`

	// Crystal harvesting quantity
	CrystalHarvestQuantity string `json:"crystalHarvestQuantity,omitempty"`

	// Oil harvesting quantity
	OilHarvestQuantity string `json:"oilHarvestQuantity,omitempty"`

	// Obsidian harvesting quantity
	ObsidianHarvestQuantity string `json:"obsidianHarvestQuantity,omitempty"`
}

// BreedingRates defines breeding multipliers
type BreedingRates struct {
	// Mating interval multiplier
	MatingIntervalMultiplier string `json:"matingIntervalMultiplier,omitempty"`

	// Egg hatch speed scale
	EggHatchSpeedScale string `json:"eggHatchSpeedScale,omitempty"`

	// Baby maturation speed scale
	BabyMaturationSpeedScale string `json:"babyMaturationSpeedScale,omitempty"`

	// Imprint period multiplier
	ImprintPeriodMultiplier string `json:"imprintPeriodMultiplier,omitempty"`

	// Enable/disable single baby gestation
	SingleBabyGestation bool `json:"singleBabyGestation,omitempty"`
}

// DifficultySettings defines difficulty-related multipliers
type DifficultySettings struct {
	// Override official difficulty
	OverrideOfficialDifficulty string `json:"overrideOfficialDifficulty,omitempty"`

	// Max difficulty
	MaxDifficulty bool `json:"maxDifficulty,omitempty"`

	// Difficulty level
	DifficultyLevel string `json:"difficultyLevel,omitempty"`

	// Use single player settings
	UseSinglePlayerSettings bool `json:"useSinglePlayerSettings,omitempty"`

	// Prevent spawning dinos without saddle
	PreventSpawningDinosWithoutSaddle bool `json:"preventSpawningDinosWithoutSaddle,omitempty"`

	// Prevent level progression beyond max difficulty
	DontUseDifficulty bool `json:"dontUseDifficulty,omitempty"`
}

// CustomRateSettings defines various customized rates
type CustomRateSettings struct {
	// Tamed dino level multiplier
	TamedDinoLevelMultiplier string `json:"tamedDinoLevelMultiplier,omitempty"`

	// Wild dino level multiplier
	WildDinoLevelMultiplier string `json:"wildDinoLevelMultiplier,omitempty"`

	// XP multiplier
	XPMultiplier string `json:"xpMultiplier,omitempty"`

	// Day cycle speed scale
	DayCycleSpeedScale string `json:"dayCycleSpeedScale,omitempty"`

	// Random supply crate points
	RandomSupplyCratePoints bool `json:"randomSupplyCratePoints,omitempty"`

	// Disable PVP balance
	DisablePvPAutoBalance bool `json:"disablePvPAutoBalance,omitempty"`

	// Disable structure placement collision
	DisableStructurePlacementCollision bool `json:"disableStructurePlacementCollision,omitempty"`
}

// ArkGSMConfig defines LinuxGSM specific configuration
type ArkGSMConfig struct {
	// Custom LinuxGSM configuration file content
	ConfigFile string `json:"configFile,omitempty"`
}

// ArkStatus defines the observed state of Ark
type ArkStatus struct {
	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Ark is the Schema for the arks API
type Ark struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArkSpec   `json:"spec,omitempty"`
	Status ArkStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ArkList contains a list of Ark
type ArkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ark `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ark{}, &ArkList{})
}
