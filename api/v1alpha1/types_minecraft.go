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

// MinecraftSpec defines the desired state of Minecraft
type MinecraftSpec struct {
	//+kubebuilder:default="gameservermanagers/gameserver:minecraft"
	Image string `json:"image"`

	Base `json:",inline"`

	Config MinecraftConfig `json:"config,omitempty"`
}

// MinecraftConfig defines configuration for Minecraft & LinuxGSM
type MinecraftConfig struct {
	// Server type (vanilla, spigot, paper, bukkit, forge, fabric)
	//+kubebuilder:validation:Enum=vanilla;spigot;paper;bukkit;forge;fabric
	//+kubebuilder:default=vanilla
	ServerType string `json:"serverType,omitempty"`

	// Server properties configuration
	ServerProperties MinecraftServerProperties `json:"serverProperties,omitempty"`

	// Java virtual machine configuration
	JVM JVMConfig `json:"jvm,omitempty"`

	// Plugins configuration (for Spigot/Paper/Pukkit servers)
	Plugins MinecraftPlugins `json:"plugins,omitempty"`

	// LinuxGSM specific configuration
	GSM MinecraftGSMConfig `json:"gsm,omitempty"`
}

// MinecraftServerProperties defines Minecraft server.properties configuration
type MinecraftServerProperties struct {
	// Server port (default: 25565)
	//+kubebuilder:default=25565
	ServerPort int32 `json:"serverPort,omitempty"`

	// Maximum number of players
	//+kubebuilder:default=20
	MaxPlayers int32 `json:"maxPlayers,omitempty"`

	// Motd (message of the day)
	Motd string `json:"motd,omitempty"`

	// Difficulty level (peaceful, easy, normal, hard)
	//+kubebuilder:validation:Enum=peaceful;easy;normal;hard
	//+kubebuilder:default=normal
	Difficulty string `json:"difficulty,omitempty"`

	// Game mode (survival, creative, adventure, spectator)
	//+kubebuilder:validation:Enum=survival;creative;adventure;spectator
	//+kubebuilder:default=survival
	GameMode string `json:"gameMode,omitempty"`

	// Whether to allow PvP
	//+kubebuilder:default=true
	Pvp *bool `json:"pvp,omitempty"`

	// Whether to require online-mode authentication
	//+kubebuilder:default=true
	OnlineMode *bool `json:"onlineMode,omitempty"`

	// View distance
	//+kubebuilder:default=10
	ViewDistance int32 `json:"viewDistance,omitempty"`

	// Simulation distance (Minecraft 1.18+)
	//+kubebuilder:default=10
	SimulationDistance int32 `json:"simulationDistance,omitempty"`

	// Whether to enable command blocks
	//+kubebuilder:default=false
	EnableCommandBlocks *bool `json:"enableCommandBlocks,omitempty"`

	// Whether to enforce whitelist
	//+kubebuilder:default=false
	EnforceWhitelist *bool `json:"enforceWhitelist,omitempty"`

	// Server icon (base64 encoded image)
	ServerIcon string `json:"serverIcon,omitempty"`

	// Resource pack URL
	ResourcePack string `json:"resourcePack,omitempty"`

	// Resource pack SHA-1 hash
	ResourcePackSha1 string `json:"resourcePackSha1,omitempty"`

	// Whether to allow nether
	//+kubebuilder:default=true
	AllowNether *bool `json:"allowNether,omitempty"`

	// Whether to allow end
	//+kubebuilder:default=true
	AllowEnd *bool `json:"allowEnd,omitempty"`

	// Level seed
	LevelSeed string `json:"levelSeed,omitempty"`

	// Level type (default, flat, largebiomes, amplified, buffet)
	//+kubebuilder:validation:Enum=default;flat;largebiomes;amplified;buffet
	LevelType string `json:"levelType,omitempty"`
}

// JVMConfig defines Java Virtual Machine configuration
type JVMConfig struct {
	// Maximum heap size (default: "2G")
	//+kubebuilder:default="2G"
	MaxHeapSize string `json:"maxHeapSize,omitempty"`

	// Minimum heap size (default: "1G")
	//+kubebuilder:default="1G"
	MinHeapSize string `json:"minHeapSize,omitempty"`

	// Additional JVM arguments
	ExtraArgs string `json:"extraArgs,omitempty"`
}

// MinecraftPlugins defines plugins configuration
type MinecraftPlugins struct {
	// List of plugins to install from Spiget.org or direct URLs
	Install []MinecraftPlugin `json:"install,omitempty"`

	// Custom plugin configuration files as ConfigMap/Secret references
	ConfigMaps []string `json:"configMaps,omitempty"`
}

// MinecraftPlugin defines a plugin to install
type MinecraftPlugin struct {
	// Plugin name (for known plugins)
	Name string `json:"name,omitempty"`

	// Plugin ID from Spiget.org (if using Spiget)
	SpigetID int32 `json:"spigetId,omitempty"`

	// Direct download URL (alternative to Spiget)
	URL string `json:"url,omitempty"`

	// Plugin version (if specific version required)
	Version string `json:"version,omitempty"`

	// Whether to enable the plugin
	//+kubebuilder:default=true
	Enable *bool `json:"enable,omitempty"`
}

// MinecraftGSMConfig defines LinuxGSM specific configuration
type MinecraftGSMConfig struct {
	// Custom LinuxGSM configuration file content
	ConfigFile string `json:"configFile,omitempty"`
}

// MinecraftStatus defines the observed state of Minecraft
type MinecraftStatus struct {
	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Minecraft is the Schema for the minecrafts API
type Minecraft struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinecraftSpec   `json:"spec,omitempty"`
	Status MinecraftStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MinecraftList contains a list of Minecraft
type MinecraftList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Minecraft `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Minecraft{}, &MinecraftList{})
}
