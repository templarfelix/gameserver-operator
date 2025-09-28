# Minecraft GameServer Operator Config

## Linux GSM Minecraft Config

https://github.com/GameServerManagers/LinuxGSM/blob/master/lgsm/config-default/config-lgsm/mcserver/_default.cfg

### Minecraft Server Types

- **Vanilla**: Standard Minecraft Java Edition server
- **Spigot**: High-performance fork of CraftBukkit
- **Paper**: High-performance fork of Spigot with more features
- **Bukkit**: Plugin platform for Minecraft
- **Forge**: Modding platform for Minecraft
- **Fabric**: Lightweight modding platform

### LinuxGSM Minecraft Configuration

For vanilla Minecraft server:
```bash
steamuser=""
steampass=""

# Optional: Custom JVM arguments
startparameters="-Xms${jvmheap} -Xmx${jvmheap}"

# Server details
servicename="mcserver"
appid="0"
```

## Minecraft Server Properties

The server.properties file is automatically generated from CRD configuration. Full reference: https://minecraft.fandom.com/wiki/Server.properties

### Common Configuration Options

| Property | Default | Description |
|----------|---------|-------------|
| server-port | 25565 | Server port |
| max-players | 20 | Maximum players |
| motd | "" | Message of the day |
| difficulty | normal | Game difficulty |
| gamemode | survival | Game mode |
| pvp | true | Enable PvP |
| online-mode | true | Authentication mode |
| view-distance | 10 | Render distance |
| simulation-distance | 10 | Simulation distance |

### JVM Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| maxHeapSize | 2G | Maximum Java heap |
| minHeapSize | 1G | Minimum Java heap |
| extraArgs | "" | Additional JVM flags |

## Kubernetes Minecraft CRD

```yaml
apiVersion: gameserver.templarfelix.com/v1alpha1
kind: Minecraft
metadata:
  labels:
    app.kubernetes.io/name: minecraft
    app.kubernetes.io/instance: minecraft-sample
    app.kubernetes.io/part-of: gameserver-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: gameserver-operator
  name: minecraft-sample
spec:
  # Container image (default: gameservermanagers/gameserver:minecraft)
  image: gameservermanagers/gameserver:minecraft

  # Persistence configuration
  persistence:
    storageConfig:
      size: 10G
    preserveOnDelete: false

  # Resource requirements (auto-adjusted based on JVM config)
  resources:
    limits:
      cpu: 2000m
      memory: 4Gi
    requests:
      cpu: 500m
      memory: 1Gi

  # Ports configuration
  ports:
    - name: minecraft-tcp
      port: 25565
      targetPort: 25565
      protocol: TCP
    - name: minecraft-udp
      port: 25565
      targetPort: 25565
      protocol: UDP

  # Load balancer configuration (optional)
  # loadBalancerIP: your-public-ip-address

  # Code-server editor password (required for VS Code access)
  editorPassword: your-editor-password

  # Node selection (optional)
  # nodeSelector:
  #   disktype: ssd
  # tolerations:
  # - key: "dedicated"
  #   operator: "Equal"
  #   value: "gameserver"
  #   effect: "NoSchedule"
  # affinity:
  #   nodeAffinity:
  #     requiredDuringSchedulingIgnoredDuringExecution:
  #       nodeSelectorTerms:
  #       - matchExpressions:
  #         - key: "kubernetes.io/arch"
  #           operator: "In"
  #           values: ["amd64"]

  # Minecraft specific configuration
  config:
    # Server type
    serverType: vanilla

    # Server properties configuration
    serverProperties:
      # Server settings
      serverPort: 25565
      maxPlayers: 50
      motd: "Welcome to My Minecraft Server!"

      # Game settings
      difficulty: hard
      gameMode: survival
      pvp: true
      onlineMode: true

      # Performance settings
      viewDistance: 12
      simulationDistance: 10

      # World settings
      allowNether: true
      allowEnd: true
      levelSeed: "minecraft-server"

      # Security settings
      enforceWhitelist: false
      enableCommandBlocks: false

    # JVM configuration
    jvm:
      maxHeapSize: "4G"
      minHeapSize: "1G"
      extraArgs: "-XX:+UseG1GC -XX:+UnlockExperimentalVMOptions"

    # Plugins configuration (for Spigot/Paper servers)
    plugins:
      - name: "EssentialsX"
        url: "https://dev.bukkit.org/download"
      - name: "WorldEdit"
        spigetId: 1911

    # LinuxGSM configuration
    gsm: |
      ## LinuxGSM Configuration
      ### https://github.com/GameServerManagers/LinuxGSM/blob/master/lgsm/config-default/config-lgsm/mcserver/_default.cfg

      # Steam credentials (usually not needed for Minecraft)
      steamuser=""
      steampass=""

      # Custom JVM arguments
      startparameters="-Xms${jvmheap} -Xmx${jvmheap}"

      # Server details
      servicename="mcserver"
      appid="0"
```

## Server Types Configuration

### Vanilla Minecraft Server

```yaml
spec:
  config:
    serverType: vanilla
    serverProperties:
      serverPort: 25565
      maxPlayers: 20
      difficulty: normal
      gameMode: survival
```

### Paper/Spigot Server with Plugins

```yaml
spec:
  config:
    serverType: paper
    serverProperties:
      serverPort: 25565
      maxPlayers: 100
    plugins:
      - name: "EssentialsX"
        spigetId: 9089
        enable: true
      - name: "WorldGuard"
        spigetId: 1183
        enable: true
    jvm:
      maxHeapSize: "8G"
      extraArgs: "-XX:+UseG1GC -XX:G1HeapRegionSize=32M"
```

### Forge Server with Mods

```yaml
spec:
  config:
    serverType: forge
    serverProperties:
      serverPort: 25565
      difficulty: hard
    jvm:
      maxHeapSize: "6G"
      minHeapSize: "2G"
      extraArgs: "-Dfml.readTimeout=90 -Dfml.queryResultLimit=90"
```

### Fabric Server

```yaml
spec:
  config:
    serverType: fabric
    serverProperties:
      serverPort: 25565
    jvm:
      maxHeapSize: "4G"
```

## Resource Requirements

### JVM Memory Mapping

| JVM Heap | Kubernetes Memory |
|----------|-------------------|
| 1G | Requests: 1.5Gi, Limits: 2.5Gi |
| 2G (default) | Requests: 3Gi, Limits: 4Gi |
| 4G | Requests: 6Gi, Limits: 8Gi |
| 8G | Requests: 12Gi, Limits: 16Gi |

## Network Configuration

### Required Ports

| Port | Protocol | Purpose |
|------|----------|---------|
| 25565 | TCP/UDP | Minecraft game port |

### Optional Services

- **Code Server**: 8080 - Web-based editor access
- **Load Balancer**: External IP for public access

## Performance Tuning

### JVM Optimization

```yaml
spec:
  config:
    jvm:
      maxHeapSize: "8G"
      extraArgs: >
        -XX:+UseG1GC -XX:+UnlockExperimentalVMOptions
        -XX:G1NewSizePercent=20 -XX:G1ReservePercent=20
        -XX:MaxGCPauseMillis=50 -XX:G1HeapRegionSize=32M
```

### Server Performance Settings

```yaml
spec:
  config:
    serverProperties:
      viewDistance: 8
      simulationDistance: 6
      maxPlayers: 50
      # Reduce entity tracking
      entityBroadcastRangePercentage: 75
      # Tick improvements
      rateLimit: 10
```

## Plugin Management

### Spiget Integration

Plugins can be automatically downloaded from Spiget.org:

```yaml
plugins:
  - name: "EssentialsX"
    spigetId: 9089
    version: "2.21.0"
    enable: true
```

### Custom Plugins

For plugins not on Spiget, provide direct URLs:

```yaml
plugins:
  - name: "CustomPlugin"
    url: "https://example.com/CustomPlugin.jar"
```

## Security Considerations

### Authentication

By default, servers use `online-mode: true` which requires Minecraft Java Edition authentication.

For offline/simplified authentication:
```yaml
serverProperties:
  onlineMode: false
```

### Resource Limits

Always set appropriate resource limits to prevent resource exhaustion:

```yaml
resources:
  limits:
    cpu: 4000m
    memory: 8Gi
  requests:
    cpu: 1000m
    memory: 4Gi
```

## Troubleshooting

### Common Issues

1. **Server not starting**: Check JVM memory settings
2. **Connection refused**: Verify port configuration
3. **Plugins not loading**: Ensure server type supports plugins
4. **Performance issues**: Monitor JVM heap usage
5. **Authentication errors**: Check online-mode setting

### Logs Location

Server logs are available in:
- **Container logs**: `kubectl logs deployment/minecraft-sample-deployment`
- **Persistent storage**: `/data/serverfiles/logs/`
- **LinuxGSM logs**: `/data/lgsm/logs/`

## Advanced Configuration

### Environment Variables

LinuxGSM supports additional environment variables in the GSM config:

```yaml
gsm: |
  # Custom environment variables
  JAVAVER="17"
  MAXPLAYERS="100"
  # Custom JVM settings
  JVMHEAP="8G"
```

### Startup Parameters

Custom startup parameters can be defined:

```yaml
gsm: |
  startparameters="-Xms4G -Xmx8G --nogui --forceUpgrade"
```

## More Information

- **Minecraft Wiki**: https://minecraft.fandom.com/wiki/Server.properties
- **LinuxGSM Minecraft**: https://linuxgsm.com/lgsm/mcserver/
- **Spigot Plugin Repository**: https://www.spigotmc.org/resources/
- **Spiget API**: https://spiget.org/ (for plugin downloads)
- **Fabric Mods**: https://modrinth.com/mods
- **Forge Mods**: https://www.curseforge.com/minecraft/mc-mods