# ARK: Survival Evolved GameServer Operator Config

## Linux GSM ARK Config

https://github.com/GameServerManagers/LinuxGSM/blob/master/lgsm/config-default/config-lgsm/arkserver/_default.cfg

### ARK Server Maps

Supported maps in LinuxGSM:
- **TheIsland** (Default)
- **TheCenter**
- **ScorchedEarth**
- **Aberration**
- **Extinction**
- **TheIslandII**
- **Fjordur**
- **LostIsland**
- **Valguero**
- **CrystalIsles**
- **Genesis**
- **Genesis2**
- **Ragnarok**

## LinuxGSM ARK Configuration

For standard ARK server:
```bash
steamuser=""
steampass=""

# Optional: Custom startup parameters
startparameters="TheIsland?listen"

# Server details
servicename="arkserver"
appid="376030"
```

## ARK Configuration Files

ARK uses two main configuration files:

### GameUserSettings.ini
Contains basic server settings, general configuration, and player settings. Located at: `/ShooterGame/Saved/Config/WindowsServer/GameUserSettings.ini`

### Game.ini
Contains advanced game settings, breeding parameters, resource rates, and server tweaks. Located at: `/ShooterGame/Saved/Config/WindowsServer/Game.ini`

## Game Configuration Structure

### Server Settings (GameUserSettings.ini)

| Setting | Default | Description |
|---------|---------|-------------|
| ServerName | "" | Server display name |
| ServerDescription | "" | Server description |
| MaxPlayers | 20 | Maximum player count |
| Difficulty | 1.0 | Difficulty level (0.2-1.0) |
| ServerPort | 7777 | Server port |
| RCONPort | 27020 | RCON port |
| EnableRCON | true | Enable RCON |
| PVP | true | Allow PVP |
| Crossplay | true | Enable cross-platform play |
| AdminPassword | "adminpassword" | Admin password |
| ModServiceType | "1" | Mod service type |

### Server Performance Settings

| Setting | Default | Description |
|---------|---------|-------------|
| NightTimeSpeed | 1.0 | Night time speed multiplier |
| DayCycleSpeed | 1.0 | Day cycle speed multiplier |
| AutoSavePeriodMinutes | 15 | Auto-save interval |
| StructureDamageMultiplier | 1.0 | Structure damage multiplier |
| PlayerDamageMultiplier | 1.0 | Player damage multiplier |
| DinoDamageMultiplier | 1.0 | Dino damage multiplier |
| DinoCountMultiplier | 1.0 | Dino count multiplier |
| DinoFoodDrainMultiplier | 1.0 | Dino food drain multiplier |

### Breeding Configuration (Game.ini)

| Setting | Default | Description |
|---------|---------|-------------|
| MatingIntervalMultiplier | 1.0 | Time between mating attempts |
| EggHatchSpeedScale | 1.0 | Egg hatch speed |
| BabyMaturationSpeedScale | 1.0 | Baby growth speed |
| ImprintPeriodMultiplier | 1.0 | Impression period multiplier |
| SingleBabyGestation | false | Single baby per gestation |

### Resource Rates (Game.ini)

| Resource | Multiplier | Description |
|----------|------------|-------------|
| MetalHarvestQuantity | 1.0 | Metal gathering rate |
| WoodHarvestQuantity | 1.0 | Wood gathering rate |
| StoneHarvestQuantity | 1.0 | Stone gathering rate |
| ThatchHarvestQuantity | 1.0 | Thatch gathering rate |
| FlintHarvestQuantity | 1.0 | Flint gathering rate |
| CrystalHarvestQuantity | 1.0 | Crystal gathering rate |
| OilHarvestQuantity | 1.0 | Oil gathering rate |
| ObsidianHarvestQuantity | 1.0 | Obsidian gathering rate |

## Kubernetes ARK CRD

```yaml
apiVersion: gameserver.templarfelix.com/v1alpha1
kind: Ark
metadata:
  labels:
    app.kubernetes.io/name: ark
    app.kubernetes.io/instance: ark-sample
    app.kubernetes.io/part-of: gameserver-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: gameserver-operator
  name: ark-sample
spec:
  # Container image (default: gameservermanagers/gameserver:ark)
  image: gameservermanagers/gameserver:ark

  # Persistence configuration
  persistence:
    storageConfig:
      size: 50G  # ARK servers need significant storage
    preserveOnDelete: false

  # Resource requirements - ARK is resource intensive
  resources:
    limits:
      cpu: 4000m
      memory: 16Gi
    requests:
      cpu: 2000m
      memory: 8Gi

  # Ports configuration
  ports:
    - name: ark-udp
      port: 7777
      targetPort: 7777
      protocol: UDP
    - name: ark-query-udp
      port: 27015
      targetPort: 27015
      protocol: UDP
    - name: ark-rcon-tcp
      port: 27020
      targetPort: 27020
      protocol: TCP

  # Load balancer IP configuration (optional)
  # loadBalancerIP: your-public-ip-address

  # Code server editor password (required for VS Code access)
  editorPassword: your-editor-password

  # Node selection for high-performance or GPU workloads
  # nodeSelector:
  #   disktype: highperformance
  # tolerations:
  # - key: "gameserver"
  #   operator: "Equal"
  #   value: "ark"
  #   effect: "NoSchedule"
  # affinity:
  #   nodeAffinity:
  #     requiredDuringSchedulingIgnoredDuringExecution:
  #       nodeSelectorTerms:
  #       - matchExpressions:
  #         - key: "workload"
  #           operator: "In"
  #           values: ["gaming"]

  # ARK specific configuration
  config:
    # GameUserSettings.ini configuration
    gameUserSettings:
      # Admin configuration
      adminPassword: "your-admin-password"

      # Server basic settings
      serverSettings:
        serverName: "My ARK Server"
        serverDescription: "A wonderful ARK: Survival Evolved server"
        maxPlayers: 50
        difficulty: 1.0

      # Server ports
      serverPort: 7777
      rconPort: 27020
      enableRcon: true

      # Game settings
      serverMap: "TheIsland"
      pvp: true
      crossplay: true

      # Performance and balance
      serverSettings:
        nightTimeSpeed: 1.0
        dayCycleSpeedScale: 1.0
        playerDamageMultiplier: 1.0
        structureDamageMultiplier: 1.0
        tameDinoDamageMultiplier: 1.0
        wildDinoDamageMultiplier: 1.0
        dinoCountMultiplier: 1.0
        dinoFoodDrainMultiplier: 1.0
        autoSavePeriodMinutes: 15

      # Mod configuration
      modInstaller: true
      mods: "731604991,924933745,719928795"  # SurvivalPlus, Better Dinos, Upgrade to Raptors

    # Game.ini advanced configuration
    game:
      # Harvesting rates
      harvesting:
        damageMultiplier: 1.0
        metalHarvestQuantity: 1.0
        woodHarvestQuantity: 1.0
        stoneHarvestQuantity: 2.0
        thatchHarvestQuantity: 1.0
        flintHarvestQuantity: 1.0
        crystalHarvestQuantity: 1.0
        oilHarvestQuantity: 1.0
        obsidianHarvestQuantity: 1.0

      # Breeding configuration
      breeding:
        matingIntervalMultiplier: 1.0
        eggHatchSpeedScale: 1.0
        babyMaturationSpeedScale: 1.0
        imprintPeriodMultiplier: 1.0
        singleBabyGestation: false

      # Difficulty settings
      difficulty:
        overrideOfficialDifficulty: 1.0
        difficultyLevel: 1.0
        maxDifficulty: false
        useSinglePlayerSettings: false

      # Custom rates and multipliers
      customRates:
        tamedDinoLevelMultiplier: 0.5
        wildDinoLevelMultiplier: 0.5
        xpMultiplier: 2.0
        disablePvPAutoBalance: false

    # LinuxGSM configuration
    gsm: |
      ## LinuxGSM Configuration for ARK
      ### https://github.com/GameServerManagers/LinuxGSM/blob/master/lgsm/config-default/config-lgsm/arkserver/_default.cfg

      # Steam credentials (optional)
      steamuser=""
      steampass=""

      # ARK specific startup parameters
      startparameters="TheIsland?listen"

      # Server details
      servicename="arkserver"
      appid="376030"
```

## Configuration Examples

### PVE Server with Single Player Settings

```yaml
spec:
  config:
    gameUserSettings:
      serverSettings:
        serverName: "PVE ARK Server - Single Player Harvesting"
        maxPlayers: 10
        difficulty: 0.5
      serverMap: "TheCenter"
      pvp: false
    game:
      harvesting:
        useSinglePlayerDamage: true
      difficulty:
        useSinglePlayerSettings: true
```

### High Difficulty PVP Server

```yaml
spec:
  config:
    gameUserSettings:
      serverSettings:
        serverName: "Hardcore ARK PVP"
        difficulty: 1.0
        nightTimeSpeed: 1.5
        playerDamageMultiplier: 1.2
        dinoCountMultiplier: 1.5
      serverMap: "Ragnarok"
      pvp: true
    game:
      harvesting:
        damageMultiplier: 1.5
      difficulty:
        difficultyLevel: 1.0
        maxDifficulty: true
      customRates:
        tamedDinoLevelMultiplier: 1.0
        wildDinoLevelMultiplier: 2.0
        xpMultiplier: 0.5
```

### Modded Server

```yaml
spec:
  config:
    gameUserSettings:
      modInstaller: true
      mods: "731604991,924933745,819858520"  # SurvivalPlus, Better Dinos, Awesome Teleporters
      serverSettings:
        serverName: "Modded ARK Server"
    gsm: |
      # RCON is required for modded servers
      rconpassword="your-rcon-password"
      enablequery="yes"
```

### Large Population Server

```yaml
spec:
  config:
    gameUserSettings:
      serverSettings:
        maxPlayers: 100
        structureDamageMultiplier: 0.8
        playerResistance: 0.9
        autoSavePeriodMinutes: 10
      serverMap: "TheIsland"
    game:
      breeding:
        matingIntervalMultiplier: 2.0
      harvesting:
        damageMultiplier: 0.8
      customRates:
        dinoCountMultiplier: 0.7
```

## Mod Management

### Installing Mods

Mods are specified by their Steam Workshop ID in the `mods` field:

```yaml
spec:
  config:
    gameUserSettings:
      modInstaller: true
      mods: "731604991,924933745,719928795"
```

### Popular Mod Categories

- **Quality of Life**: Platforms Plus, Awesome Teleporters
- **Dino Enhancements**: Better Dinos, Dino Storage
- **Building**: Building Plus, Structure Plus
- **Stack Sizes**: Stack Plus, Master Stack
- **Aesthetics**: Dino Colored Textures, Better Chairs

## Performance Optimization

### Resource Requirements

| Server Size | CPU Cores | Memory | Storage |
|-------------|-----------|--------|---------|
| 4 Players | 2 cores | 8GB | 20GB |
| 10 Players | 4 cores | 16GB | 30GB |
| 50 Players | 8 cores | 32GB | 100GB |
| 100+ Players | 16+ cores | 64GB+ | 200GB+ |

### JVM-like Memory Settings

ARK doesn't use JVM, but you can limit memory usage through resource limits:

```yaml
resources:
  limits:
    cpu: "4000m"
    memory: "16Gi"
  requests:
    cpu: "2000m"
    memory: "8Gi"
```

### Tick Rate and Performance

External performance settings are managed through config files:

```yaml
spec:
  config:
    game:
      difficulty:
        dontUseDifficulty: true
      customRates:
        disablePvPAutoBalance: true
        disableStructurePlacementCollision: false
```

## Networking Configuration

### Required Ports

| Port | Protocol | Purpose | Firewall |
|------|----------|---------|----------|
| 7777 | UDP | Game server port | Open |
| 7778 | UDP | Additional connections | Open |
| 27015 | UDP | Steam query port | Open |
| 27020 | TCP | RCON port | Restricted |

### Port Configuration

```yaml
ports:
  - name: ark-game
    port: 7777
    targetPort: 7777
    protocol: UDP
  - name: ark-game2
    port: 7778
    targetPort: 7778
    protocol: UDP
  - name: ark-query
    port: 27015
    targetPort: 27015
    protocol: UDP
  - name: ark-rcon
    port: 27020
    targetPort: 27020
    protocol: TCP
```

## Backup and Recovery

### Automatic Backups

ARK automatically saves every 15 minutes by default. Configure the interval:

```yaml
spec:
  config:
    gameUserSettings:
      serverSettings:
        autoSavePeriodMinutes: 15
```

### Backup Strategy

- **World Data**: `/data/serverfiles/ShooterGame/Saved/SavedArks/`
- **Config Files**: `/data/serverfiles/ShooterGame/Saved/Config/`
- **Cluster Data**: For cross-server clusters

### Volume Persistence

Important directories to persist:
```yaml
persistence:
  storageConfig:
    size: 100G  # Include space for world saves and mods
  preserveOnDelete: true  # Keep world data
```

## Troubleshooting

### Common Issues

1. **Server not appearing in browser**
   - Check Steam ports (27015-27020)
   - Verify firewall rules
   - Check server name settings

2. **Mods not loading**
   - Verify Mod Service Type is set to "1"
   - Check mod IDs are correct
   - Enable mod installer

3. **Performance issues**
   - Check CPU and memory usage
   - Adjust player count
   - Review harvesting multipliers

4. **RCON connection fails**
   - Verify RCON port is open
   - Check RCON password
   - Enable RCON in config

### Diagnostic Commands

```bash
# Check server logs
kubectl logs -f deployment/ark-sample-deployment

# Access code server for live config editing
kubectl port-forward deployment/ark-sample-deployment 8080:8080

# Check server query status
nmap -sU -p 27015 your-server-ip
```

## Advanced Features

### Crossplay Configuration

```yaml
spec:
  config:
    gameUserSettings:
      crossplay: true
```

### Custom Difficulty

```yaml
spec:
  config:
    game:
      difficulty:
        overrideOfficialDifficulty: 1.5
        difficultyLevel: 1.5
        maxDifficulty: true
```

### Cluster Configuration

For cross-server ecosystems:

```yaml
# Cluster ID must match across servers
clusterId: "MyCluster"

# Shared cluster directory (mount external volume)
clusterPath: "/cluster/shared"
```

### Battle Eye Anti-Cheat

ARK uses BattleEye by default to prevent cheating. Server admins can manage bans through RCON.

## More Information

- **Steam Workshop**: https://steamcommunity.com/app/346110/workshop/
- **ARK Wiki**: https://survivetheark.com/
- **Server Configuration**: https://survivetheark.com/index.php?/forums/topic/278-ini-configuration-files-server-tweak-guide/
- **Modding Guide**: https://survivetheark.com/index.php?/forums/forum/209-dedicated-server-modding/