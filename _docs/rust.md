# Rust GameServer Operator Config

## LinuxGSM Rust config

Reference (default LinuxGSM cfg):
- https://github.com/GameServerManagers/LinuxGSM/blob/master/lgsm/config-default/config-lgsm/rustserver/_default.cfg

Useful docs:
- Rust server ports and RCON: https://linuxgsm.com/servers/rustserver/
- SteamCMD login guidance: https://docs.linuxgsm.com/steamcmd

### Steam credentials
- Create a dedicated Steam account for the operator (recommended) and consider disabling Steam Guard for automation.
- Provide the credentials via environment or LinuxGSM config as per LinuxGSM documentation.

## Kubernetes Custom Resource (Rust)

The operator manages Rust servers via a Rust custom resource. A minimal working example is below; adjust ports and resources to your needs.

```yaml
apiVersion: gameserver.templarfelix.com/v1alpha1
kind: Rust
metadata:
  name: rust-sample
spec:
  image: gameservermanagers/gameserver:rust
  # Network ports
  ports:
    - name: game
      port: 28015
      protocol: UDP
      targetPort: 28015
    - name: rcon
      port: 28016
      protocol: TCP
      targetPort: 28016
  # Optional LoadBalancer IP (leave empty to let cloud assign one)
  # loadBalancerIP: ""

  # Persistent data volume
  persistence:
    storageConfig:
      size: 10G
      # storageClassName: ""
    preserveOnDelete: false

  # Resource requests/limits
  resources:
    requests:
      cpu: "500m"
      memory: "1Gi"
    limits:
      cpu: "2000m"
      memory: "4Gi"

  # Web code editor (code-server) password
  editorPassword: "changeme"

  # Optional scheduling knobs
  # nodeSelector: {}
  # tolerations: []
  # affinity: {}

  # Optional: inline config files written into the container filesystem
  # Keys are absolute file paths inside the container. The init containers
  # will place them accordingly under /data.
  # config:
  #   /data/config-lgsm/rustserver/rustserver.cfg: |
  #     ip="0.0.0.0"
  #     port="28015"
  #     rconport="28016"
```

Notes:
- The controller automatically exposes a separate TCP Service that includes the code-server port 8080 for web editing. You only declare game/RCON ports above; 8080 is added automatically.
- All game data lives under /data (PVC). Set preserveOnDelete=true if you want the data to survive CR deletion. The controller then removes the PVC ownerReference during finalization so it won’t be garbage-collected.
- Security context defaults: the game container runs as root (LinuxGSM requirement); init containers handling file copies run as uid/gid 1000 where applicable; code-server runs as uid/gid 1000.

## File paths and configuration
- LinuxGSM config (typical): /data/config-lgsm/rustserver/rustserver.cfg
- Game server files live under: /data/serverfiles
- You can supply additional files (e.g., oxide plugins/configs) via the `spec.config` map using absolute paths under /data/serverfiles or /data/config-lgsm.

## Validation status
- Rust support follows the same reconciliation pattern as DayZ and Project Zomboid: PVC, Services (TCP+UDP), Deployment with init containers for config writing, and a finalizer honoring preserveOnDelete.
