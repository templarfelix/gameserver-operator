# GameServer Operator - Games Implementation Roadmap

## 🎯 **Overview**

This document outlines the implementation status of all supported games in the GameServer Operator based on [LinuxGSM supported games](https://linuxgsm.com/servers/) and their configuration repositories.

## ✅ **Currently Implemented Games**

### DayZ
- **Status**: ✅ **FULLY IMPLEMENTED**
- **GSM Config**: `dayzserver.cfg` → `/data/config-lgsm/dayzserver/dayzserver.cfg`
- **Server Config**: `dayzserver.server.cfg` → `/data/serverfiles/cfg/dayzserver.server.cfg`
- **Controller**: `internal/controller/dayz_controller.go`
- **CRD**: `api/v1alpha1/dayz_types.go`
- **Docs**: `_docs/dayz.md`

### Project Zomboid
- **Status**: ✅ **FULLY IMPLEMENTED**
- **GSM Config**: `pzserver.cfg` → `/data/config-lgsm/pzserver/pzserver.cfg`
- **Server Config**: `server.ini` → `/data/serverfiles/server.ini`
- **Controller**: `internal/controller/projectzomboid_controller.go`
- **CRD**: `api/v1alpha1/projectzomboid_types.go`
- **Docs**: `_docs/projectzomboid.md`

### Minecraft
- **Status**: ✅ **FULLY IMPLEMENTED**
- **GSM Config**: `mcserver.cfg` → `/data/config-lgsm/mcserver/mcserver.cfg`
- **Server Config**: `server.properties` → `/data/serverfiles/server.properties` (Generated)
- **JVM Config**: `jvm.args` → `/data/serverfiles/jvm.args` (Generated)
- **Controller**: `internal/controller/minecraft_controller.go`
- **CRD**: `api/v1alpha1/minecraft_types.go`
- **Docs**: `_docs/minecraft.md`

### ARK: Survival Evolved
- **Status**: ✅ **FULLY IMPLEMENTED**
- **GSM Config**: `arkserver.cfg` → `/data/config-lgsm/arkserver/arkserver.cfg`
- **GameUserSettings.ini**: Auto-generated → `/data/serverfiles/ShooterGame/Saved/Config/WindowsServer/`
- **Game.ini**: Auto-generated → `/data/serverfiles/ShooterGame/Saved/Config/WindowsServer/`
- **Controller**: `internal/controller/ark_controller.go`
- **CRD**: `api/v1alpha1/ark_types.go`
- **Docs**: `_docs/ark.md`

## 🔄 **Implementation Architecture**

Each game needs these components:

### 1. **API Types** (`api/v1alpha1/`)
- Custom Resource Definition (CRD)
- Go types for the game spec

### 2. **Controller** (`internal/controller/`)
- Game-specific reconciler
- Deployment creation logic
- InitContainer configuration

### 3. **Utils Functions** (`internal/controller/utils.go`)
- Game-specific init container function
- Configuration paths setup

### 4. **Documentation** (`_docs/`)
- Game configuration guide
- LinuxGSM setup documentation

### 5. **Sample Configs** (`config/samples/`)
- YAML deployment examples

## 🎮 **Games To Implement**

### Action Games

#### ARK: Survival Evolved
- **GSM**: `arkserver` → `/data/config-lgsm/arkserver/`
- **Server**: `GameUserSettings.ini`, `Game.ini` → `/data/serverfiles/ShooterGame/Saved/Config/LinuxServer/`
- **Priority**: 🔴 **High** (Popular game)
- **Difficulty**: 🔴 **High** (Complex multi-config setup)

#### Avorion
- **GSM**: `avserver` → `/data/config-lgsm/avserver/`
- **Server**: `server.ini` → `/data/serverfiles/data/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟢 **Low**

#### Ballistic Overkill
- **GSM**: `boserver` → `/data/config-lgsm/boserver/`
- **Server**: `PCServer-*.ini` → `/data/serverfiles/DedicatedServer/Config/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟢 **Low**

#### Chivalry 2
- **GSM**: `chivalry2server` → `/data/config-lgsm/chivalry2server/`
- **Server**: Custom configs → `/data/serverfiles/Chivalry2/Saved/Config/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟢 **Low**

#### Codename CURE
- **GSM**: `cureserver` → `/data/config-lgsm/cureserver/`
- **Server**: `ServerSettings.ini` → `/data/serverfiles/Config/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### Colony Survival
- **GSM**: `csserver` → `/data/config-lgsm/csserver/`
- **Server**: `config.json` → `/data/serverfiles/server/config/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

### Adventure/RPG Games

#### 7 Days to Die
- **GSM**: `sdtdserver` → `/data/config-lgsm/sdtdserver/`
- **Server**: `serverconfig.xml` → `/data/serverfiles/SDTD/Config/`
- **Priority**: 🔴 **High** (Popular zombie survival)
- **Difficulty**: 🟡 **Medium**

#### SteamCMD Only

#### Unreal Tournament (1999)
- **GSM**: `utserver` → `/data/config-lgsm/utserver/`
- **Server**: `UnrealTournament.ini` → `/data/serverfiles/System/`
- **Priority**: 🔵 **Low** (Legacy game)
- **Difficulty**: 🟢 **Low**

#### Action: Source Dedicated Server
- **GSM**: `acsds` → `/data/config-lgsm/acsds/`
- **Server**: Various .cfg files → `/data/serverfiles/cfg/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

### Battle Royale/Shooter Games

#### BattleBit Remastered
- **GSM**: `bbrserver` → `/data/config-lgsm/bbrserver/`
- **Server**: `DefaultGame.ini`, `DefaultEngine.ini` → `/data/serverfiles/BattleBitRemastered/Saved/Game/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟢 **Low**

#### Dangerous Driving
- **GSM**: `ddserver` → `/data/config-lgsm/ddserver/`
- **Server**: Server configs → `/data/serverfiles/Config/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### Fistful of Frags
- **GSM**: `fofserver` → `/data/config-lgsm/fofserver/`
- **Server**: `server.cfg` → `/data/serverfiles/cfg/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### Garry's Mod
- **GSM**: `gmodserver` → `/data/config-lgsm/gmodserver/`
- **Server**: `server.cfg` → `/data/serverfiles/garrysmod/cfg/`
- **Priority**: 🔴 **High** (Very popular)
- **Difficulty**: 🟡 **Medium**

#### Insurgency: Sandstorm
- **GSM**: `inssserver` → `/data/config-lgsm/inssserver/`
- **Server**: `Game.ini`, `Engine.ini` → `/data/serverfiles/Insurgency/Saved/Config/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟡 **Medium**

#### Killing Floor 2
- **GSM**: `kf2server` → `/data/config-lgsm/kf2server/`
- **Server**: `KFGame.ini` → `/data/serverfiles/KFGame/Config/`
- **Priority**: 🔴 **High** (Co-op shooter)
- **Difficulty**: 🟡 **Medium**

#### Left 4 Dead 2
- **GSM**: `l4d2server` → `/data/config-lgsm/l4d2server/`
- **Server**: `server.cfg`, `host.txt` → `/data/serverfiles/left4dead2/cfg/`
- **Priority**: 🔴 **High** (L4D series)
- **Difficulty**: 🟢 **Low**

#### Squad
- **GSM**: `sqserver` → `/data/config-lgsm/sqserver/`
- **Server**: `Game.ini` → `/data/serverfiles/SquadGame/Saved/Config/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟢 **Low**

#### The Front
- **GSM**: `tfserver` → `/data/config-lgsm/tfserver/`
- **Server**: `Game.ini` → `/data/serverfiles/TheFront/Saved/Config/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### Zombie Panic! Source
- **GSM**: `zpsserver` → `/data/config-lgsm/zpsserver/`
- **Server**: `server.cfg` → `/data/serverfiles/zps/cfg/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

### Online Battle Arena/MOBA

#### Dota 2
- **GSM**: `dota2server` → `/data/config-lgsm/dota2server/`
- **Server**: Console commands setup
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🔴 **High** (Tournament servers)

#### League of Legends
- **GSM**: `lolserver` → `/data/config-lgsm/lolserver/`
- **Server**: Tournament configs → `/data/serverfiles/Config/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🔴 **High**

### Strategy Games

#### Age of Chivalry
- **GSM**: `aocserver` → `/data/config-lgsm/aocserver/`
- **Server**: `server.cfg` → `/data/serverfiles/cfg/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### Natural Selection 2
- **GSM**: `ns2server` → `/data/config-lgsm/ns2server/`
- **Server**: `ServerConfig.json` → `/data/serverfiles/config/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### Pirates, Vikings and Knights II
- **GSM**: `pvk2server` → `/data/config-lgsm/pvk2server/`
- **Server**: `server.cfg` → `/data/serverfiles/cfg/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### Pirates, Vikings and Knights II - cURL
- **GSM**: `pvk2curlserver` → `/data/config-lgsm/pvk2curlserver/`
- **Server**: Same as PVKII
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### Sven Co-op
- **GSM**: `svenDS` → `/data/config-lgsm/svenDS/`
- **Server**: `server.cfg` → `/data/serverfiles/svencoop/cfg/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### Synergy
- **GSM**: `synserver` → `/data/config-lgsm/synserver/`
- **Server**: `server.cfg` → `/data/serverfiles/synergy/cfg/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### Zombie Master: Reborn
- **GSM**: `zmserver` → `/data/config-lgsm/zmserver/`
- **Server**: `server.cfg` → `/data/serverfiles/zm/cfg/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

### Simulation Games

#### American Truck Simulator
- **GSM**: `amserver` → `/data/config-lgsm/amserver/`
- **Server**: `server_config.sii` → `/data/serverfiles/server_config.sii`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟡 **Medium**

#### Euro Truck Simulator 2
- **GSM**: `etserver` → `/data/config-lgsm/etserver/`
- **Server**: `server_config.sii` → `/data/serverfiles/server_config.sii`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟡 **Medium**

#### Farming Simulator 22
- **GSM**: `fsserver` → `/data/config-lgsm/fsserver/`
- **Server**: `dedicatedServerConfig.xml` → `/data/serverfiles/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟡 **Medium**

#### Farming Simulator 19
- **GSM**: `fs19server` → `/data/config-lgsm/fs19server/`
- **Server**: `dedicatedServerConfig.xml` → `/data/serverfiles/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟡 **Medium**

#### My Summer Car
- **GSM**: `mscs` → `/data/config-lgsm/mscs/`
- **Server**: `MSCServerConfig.txt` → `/data/serverfiles/Config/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### Space Engineers
- **GSM**: `seserver` → `/data/config-lgsm/seserver/`
- **Server**: `SpaceEngineers-Dedicated.cfg` → `/data/serverfiles/DedicatedServer/Config/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟡 **Medium**

#### Stormworks: Build and Rescue
- **GSM**: `swserver` → `/data/config-lgsm/swserver/`
- **Server**: `config.lua` → `/data/serverfiles/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

### Racing Games

#### Assetto Corsa
- **GSM**: `acserver` → `/data/config-lgsm/acserver/`
- **Server**: `server_cfg.ini` → `/data/serverfiles/cfg/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟢 **Low**

#### Assetto Corsa Competizione
- **GSM**: `accserver` → `/data/config-lgsm/accserver/`
- **Server**: `server_cfg.ini` → `/data/serverfiles/cfg/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟢 **Low**

#### City Car Driving
- **GSM**: `ccdrserver` → `/data/config-lgsm/ccdrserver/`
- **Server**: `ServerCFG.json` → `/data/serverfiles/Config/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### rFactor 2
- **GSM**: `rf2server` → `/data/config-lgsm/rf2server/`
- **Server**: `Dedicated.ini` → `/data/serverfiles/UserData/Multiplayer/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟡 **Medium**

### Puzzle/Platformer

#### The Forest
- **GSM**: `tfserver` → `/data/config-lgsm/tfserver/` (Same acronym as "The Front")
- **Server**: Config files → `/data/serverfiles/config/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟢 **Low**

### Horror/Survival

#### No One Survived
- **GSM**: `nosserver` → `/data/config-lgsm/nosserver/`
- **Server**: `Game.ini` → `/data/serverfiles/NoOneSurvived/Saved/Config/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### Outlaws of the Old West
- **GSM**: `ootwserver` → `/data/config-lgsm/ootwserver/`
- **Server**: Custom configs → `/data/serverfiles/Config/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### Satisfactory
- **GSM**: `sfserver` → `/data/config-lgsm/sfserver/`
- **Server**: `Game.ini` → `/data/serverfiles/FactoryGame/Saved/Config/`
- **Priority**: 🟡 **Medium**
- **Difficulty**: 🟢 **Low**

#### Sunkenland
- **GSM**: `slserver` → `/data/config-lgsm/slserver/`
- **Server**: `ServerSettings.ini` → `/data/serverfiles/Config/`
- **Priority**: 🔵 **Low**
- **Difficulty**: 🟢 **Low**

#### V Rising
- **GSM**: `vrserver` → `/data/config-lgsm/vrserver/`
- **Server**: `ServerGameSettings.json` → `/data/serverfiles/save-data/Settings/`
- **Priority**: 🔴 **High** (Popular vampire survival)
- **Difficulty**: 🟢 **Low**

### Missing Games Identified (To Be Researched)

#### Counter-Strike: Global Offensive
- **GSM**: `csgoserver` → `/data/config-lgsm/csgoserver/`
- **Server**: `server.cfg` → `/data/serverfiles/csgo/cfg/`
- **Priority**: 🔴 **High** (Extremely popular)
- **Difficulty**: 🟡 **Medium**

#### Counter-Strike 2
- **GSM**: `csserver` → `/data/config-lgsm/csserver/`
- **Server**: `server.cfg` → `/data/serverfiles/game/csgo/cfg/`
- **Priority**: 🔴 **High** (Latest CS version)
- **Difficulty**: 🟡 **Medium**

#### Rust
- **GSM**: `rustserver` → `/data/config-lgsm/rustserver/`
- **Server**: `server.cfg` → `/data/serverfiles/server/rustserver/cfg/`
- **Priority**: 🔴 **High** (Very popular survival)
- **Difficulty**: 🟡 **Medium**

#### Valheim
- **GSM**: `vhserver` → `/data/config-lgsm/vhserver/`
- **Server**: `world` → `/data/serverfiles/saves/`
- **Priority**: 🔴 **High** (Popular co-op)
- **Difficulty**: 🟢 **Low**

#### Palworld
- **GSM**: `pwserver` → `/data/config-lgsm/pwserver/`
- **Server**: `PalWorldSettings.ini` → `/data/serverfiles/Pal/Saved/Config/WindowsServer/`
- **Priority**: 🔴 **High** (New popular game)
- **Difficulty**: 🟡 **Medium**

#### Minecraft
- **GSM**: `mcserver` → `/data/config-lgsm/mcserver/`
- **Server**: `server.properties` → `/data/serverfiles/server.properties`
- **Priority**: 🔴 **High** (Most popular game server)
- **Difficulty**: 🟡 **Medium**

#### Terraria
- **GSM**: `tshockserver` → `/data/config-lgsm/tshockserver/`
- **Server**: `server.json` → `/data/serverfiles/tshock/server.json`
- **Priority**: 🟡 **Medium** (Popular indie game)
- **Difficulty**: 🟢 **Low**

#### Don't Starve Together
- **GSM**: `dstserver` → `/data/config-lgsm/dstserver/`
- **Server**: `cluster.ini` → `/data/serverfiles/Klei/save/cluster.ini`
- **Priority**: 🟡 **Medium** (Popular co-op)
- **Difficulty**: 🟢 **Low**

#### Conan Exiles
- **GSM**: `ceserver` → `/data/config-lgsm/ceserver/`
- **Server**: `Engine.ini`, `Game.ini` → `/data/serverfiles/ConanSandbox/Saved/Config/WindowsServer/`
- **Priority**: 🟡 **Medium** (Popular survival MMO)
- **Difficulty**: 🟡 **Medium**

#### Unturned
- **GSM**: `utserver` → `/data/config-lgsm/utserver/`
- **Server**: `Config.json` → `/data/serverfiles/Servers/Normal/Config.json`
- **Priority**: 🟡 **Medium** (Very popular free server)
- **Difficulty**: 🟢 **Low**

#### Team Fortress 2
- **GSM**: `tf2server` → `/data/config-lgsm/tf2server/`
- **Server**: `server.cfg` → `/data/serverfiles/tf/cfg/`
- **Priority**: 🟡 **Medium** (Valve platformer shooter)
- **Difficulty**: 🟢 **Low**

#### Killing Floor
- **GSM**: `kfserver` → `/data/config-lgsm/kfserver/`
- **Server**: `KillingFloor.ini` → `/data/serverfiles/System/`
- **Priority**: 🟡 **Medium** (KF1, predecessor to KF2)
- **Difficulty**: 🟢 **Low**

## 📊 **Implementation Statistics**

### Priority Distribution
- 🔴 **High Priority**: ARK ✅, 7DTD, Garry's Mod, Killing Floor 2, L4D2, V Rising, Rust, Palworld, Valheim (~18 games)
- 🟡 **Medium Priority**: BattleBit, ETS2, Squad, Conan Exiles, Terraria, Don't Starve Together, Unturned, TF2, Killing Floor, CSGO, CS2 (~27 games)
- 🔵 **Low Priority**: Niche/Legacy games (Age of Chivalry, Sven Co-op, etc.) (~50+ games)

### Difficulty Distribution
- 🟢 **Easy**: Most Source engine, simple config games (~60 games)
- 🟡 **Medium**: Unity/Unreal games with custom configs (~15 games)
- 🔴 **Hard**: Complex configs, tournament servers (~5 games)

### Type Distribution
- **First Person Shooter**: 20+ games
- **Survival/Horror**: 15+ games
- **Strategy**: 10+ games
- **Simulation/Racing**: 15+ games
- **MOBA/Battle Arena**: 5+ games

## 🚀 **Next Steps**

1. **Phase 1 (High Priority)**: Implement 8-12 high priority games (Minecraft ✅, ARK ✅ (COMPLETED!), **next:** Garry's Mod, Killing Floor 2, 7DTD, Rust, Palworld, Valheim)
2. **Phase 2 (Medium Priority)**: Add 15-20 medium priority games (BattleBit, ETS2, Squad, Conan Exiles, Terraria, Don't Starve Together, Unturned, TF2)
3. **Phase 3 (On Demand)**: Complete remaining 50+ games based on user demand and community requests

## 📋 **Detailed Implementation TODOs**

Cada jogo requer 5 componentes principais. Aqui estão os TODOs detalhados para alguns jogos prioritários:

### 🔴 **High Priority Games**

#### ARK: Survival Evolved
- [ ] Criar `api/v1alpha1/ark_types.go` com CRD Ark
- [ ] Implementar funções de configuração multi-arquivo em `internal/controller/utils.go`
- [ ] Criar `internal/controller/ark_controller.go`
- [ ] Gerar documentação para ARK em `_docs/ark.md`
- [ ] Adicionar exemplo YAML em `config/samples/`

#### Garry's Mod
- [ ] Criar `api/v1alpha1/gmod_types.go` com CRD Gmod
- [ ] Implementar gamemode detection em utils.go
- [ ] Criar `internal/controller/gmod_controller.go`
- [ ] Gerar documentação em `_docs/garrysmod.md`
- [ ] Adicionar exemplo YAML em configurações múltiplas

#### Killing Floor 2
- [ ] Criar `api/v1alpha1/kf2_types.go` com CRD KF2
- [ ] Implementar regras de difficulty e game modes
- [ ] Criar `internal/controller/kf2_controller.go`
- [ ] Gerar documentação em `_docs/killingfloor2.md`
- [ ] Adicionar exemplo com WaveMode settings

#### 7 Days to Die
- [ ] Criar `api/v1alpha1/sdtd_types.go` com CRD Sdtd
- [ ] Implementar horde/multi-day cycle parsing
- [ ] Criar `internal/controller/7dtd_controller.go`
- [ ] Gerar documentação em `_docs/7daystodie.md`
- [ ] Adicionar exemplo com BloodMoon settings

#### Rust
- [ ] Criar `api/v1alpha1/rust_types.go` com CRD Rust
- [ ] Implementar Oxide/uMod plugin loading em initContainer
- [ ] Criar `internal/controller/rust_controller.go`
- [ ] Gerar documentação em `_docs/rust.md`
- [ ] Adicionar exemplo com Oxide support

#### Valheim
- [ ] Criar `api/v1alpha1/valheim_types.go` com CRD Valheim
- [ ] Implementar mundo/PVP creation logic
- [ ] Criar `internal/controller/valheim_controller.go`
- [ ] Gerar documentação em `_docs/valheim.md`
- [ ] Adicionar exemplo com BepInEx support

#### ~~Minecraft~~
- [x] Criar `api/v1alpha1/minecraft_types.go` com CRD Minecraft
- [x] Implementar jarfile/java version detection
- [x] Criar `internal/controller/minecraft_controller.go`
- [x] Gerar documentação em `_docs/minecraft.md`
- [x] Adicionar exemplos para Vanilla/Paper/Forge

### 🟡 **Medium Priority Games**

#### Counter-Strike: Global Offensive
- [ ] Criar CRD CSGO com map voting configs
- [ ] Implementar GOTV/recording settings
- [ ] Criar controller com matchmaking mode handling
- [ ] Documentação completa
- [ ] Exemplos de configuração profissional

#### Euro Truck Simulator 2
- [ ] Criar CRD ETS2 com DLC/mod support
- [ ] Implementar economy sync settings
- [ ] Controller para savefile persistence
- [ ] Documentação ATS/ETS2
- [ ] Exemplos de servidor público

### 🔵 **Low Priority Games (When Demand Exists)**

#### Age of Chivalry
- [ ] Criar CRD básico
- [ ] Controller simples (sem complexidade especial)
- [ ] Documentação minimal
- [ ] Exemplo básico

## 📝 **Notes**

- All paths based on LinuxGSM documentation
- Config files may need verification from actual server installations
- Some games may require additional research for exact config paths
- Priority based on popularity and community demand
- Implementation follows the established patterns from DayZ and Project Zomboid