# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**IMPORTANT NOTE: This entire project is in English.** All code comments, commit messages, documentation, and communication must be in English. No Portuguese or other languages are allowed in this codebase.

## Development Commands

**Build and Test:**
- `make build` - Build manager binary
- `make test` - Run unit tests
- `make lint` - Run golangci-lint
- `make lint-fix` - Run golangci-lint with fixes
- `make test-e2e` - Run end-to-end tests (requires Kind cluster)

**Deployment:**
- `make install` - Install CRDs
- `make uninstall` - Remove CRDs
- `make deploy IMG=templarfelix/gameserver-operator:latest` - Deploy controller
- `make undeploy` - Remove controller
- `make docker-build docker-push IMG=templarfelix/gameserver-operator:latest` - Build and push image

**Development Workflow:**
- `make run` - Run controller locally against current kubeconfig
- `make manifests` - Generate CRD manifests
- `make generate` - Generate deep copy methods

## Code Architecture

This is a Kubernetes operator built with Kubebuilder/Operator SDK for managing LinuxGSM game servers.

**Key Components:**
- **API Types**: `api/v1alpha1/` - Custom Resource Definitions (Dayz, ProjectZomboid)
- **Controllers**: `internal/controller/` - Reconciliation logic for each game type
- **Base Controller**: `internal/controller/base_controller.go` - Shared utilities for PVC and Service management
- **Game Configs**: `games/setup.go` - Game-specific setup logic

**Supported Games:**
- DayZ (`Dayz` CRD)
- Project Zomboid (`ProjectZomboid` CRD)

**Custom Resource Structure:**
Each game CRD includes:
- Base configuration (persistence, resources, ports)
- Game-specific configuration
- Status conditions

**Persistence Configuration:**

```yaml
persistence:
  storageConfig:
    size: 10G                    # (default: "10G") Storage size
    storageClassName: standard   # Storage class name (optional)
  preserveOnDelete: false        # Preserve PVC when CR is deleted (default: false)
```

**Reconciliation Pattern:**
Controllers follow standard Kubernetes operator pattern:
1. Reconcile PersistentVolumeClaim for game data (respects preserveOnDelete)
2. Reconcile Services (TCP/UDP separated)
3. Manage game server Pod deployment with dynamically configured ports based on CRD spec

## CRD Configuration Pattern

**IMPORTANT**: All game CRDs must follow this standardized configuration pattern going forward. This pattern was established after refactoring ARK, Minecraft, DayZ, and Project Zomboid implementations.

### CRD Structure Requirements

Each game CRD must have the following structure:

```go
type GameXSpec struct {
    //+kubebuilder:default="gameservermanagers/gameserver:gamename"
    Image string `json:"image"`

    Base `json:",inline"`  // Embedded base configuration

    Config GameXConfig `json:"config,omitempty"`
}

type GameXConfig struct {
    // Game-specific server configuration
    Game GameXServerConfig `json:"game,omitempty"`

    // LinuxGSM specific configuration
    GSM ArkGSMConfig `json:"gsm,omitempty"`
}
```

### Configuration Generation Pattern

Controllers must implement these functions:

1. **`generateGameXConfigData(instance *gameserverv1alpha1.GameX) map[string]string`**
   - Returns a map of filename → content for all config files
   - Handles custom vs default GSM config
   - Calls game-specific config generators

2. **`generateGameXServerConfig(settings *gameserverv1alpha1.GameXServerConfig) string`**
   - Converts CRD spec fields to game server config file format
   - Handles all supported config options
   - Uses appropriate formatting (key=value, key="value", etc.)

3. **`generateGameXGSMConfig() string`**
   - Provides default LinuxGSM configuration
   - Includes all required fields: servicename, appid, ports

### Key Implementation Requirements

1. **No Generic String Fields**: Never use `Server string` or `GSM string` fields in CRDs
2. **Structured Configuration**: Embed configuration details in CRD types with proper validation
3. **Type Safety**: Use appropriate Go types (int32, bool, custom enums) instead of strings
4. **Default Values**: Use kubebuilder defaults extensively
5. **Validation**: Include validation annotations where appropriate
6. **Documentation**: Document every config field with comments

### Controller Configuration Pattern

In the controller's `Reconcile` method, replace string-based config with:

```go
configMapName := instance.Name + "-configmap"
configData := r.generateGameXConfigData(instance)
if err := ReconcileConfigMap(ctx, r.Client, instance, configMapName, configData); err != nil {
    return reconcile.Result{}, err
}
```

### Base Configuration

All CRDs embed the `Base` struct which provides:
- Persistence configuration (PVC size, storage class, preserve on delete)
- Resource requirements (CPU/memory limits)
- Port configurations (dynamic service creation)
- Node selector, tolerations, affinity

### Example Configuration Output

For a DayZ server, this generates:
- `dayzserver.server.cfg` - Game server settings from CRD
- `dayzserver.cfg` - LinuxGSM configuration (custom or default)

**Benefits of this pattern:**
- Type safety and validation
- IDE autocompletion and documentation
- Consistent API across all games
- Easy to extend with new config options
- Reduced runtime errors from misconfiguration

**Development Notes:**
- Uses controller-runtime v0.16.3
- Requires Go 1.20+
- Follows standard Kubebuilder project structure
- Includes RBAC configurations in `config/rbac/`
- Sample configurations in `config/samples/`
- Dynamic port configuration: Game server ports are automatically configured based on CRD port specifications
- Robust reconciliation: Controller handles deployment updates and resource reconciliation reliably

**Testing:**
- Unit tests with envtest
- End-to-end tests require Kind cluster
- Linting with golangci-lint