# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GameServer-Operator is a Kubernetes operator built with Kubebuilder that manages LinuxGSM game servers in Kubernetes. It uses Custom Resource Definitions (CRDs) to declare game server configurations and reconciles them into Deployments, Services, PVCs, and GCP ComputeAddresses for static IPs.

**Domain**: `templarfelix.com`
**API Group**: `gameserver.templarfelix.com/v1`
**Module**: `github.com/templarfelix/gameserver-operator`

## Essential Commands

### Build and Code Generation
```bash
make generate         # Generate DeepCopy methods
make manifests        # Generate CRDs, RBAC, webhooks
make build            # Build the manager binary
make fmt              # Format code
make vet              # Run go vet
make lint             # Run golangci-lint
make lint-fix         # Run golangci-lint with auto-fix
```

### Testing
```bash
make test             # Run unit tests (excludes e2e)
make test-e2e         # Run e2e tests (creates Kind cluster)
make cleanup-test-e2e # Tear down e2e Kind cluster
```

### Local Development
```bash
make run              # Run controller locally against ~/.kube/config cluster
make install          # Install CRDs to cluster
make uninstall        # Remove CRDs from cluster
```

### Docker and Deployment
```bash
make docker-build IMG=<image>     # Build operator image
make docker-push IMG=<image>      # Push operator image
make deploy IMG=<image>           # Deploy to cluster
make undeploy                     # Remove from cluster
make build-installer IMG=<image>  # Generate dist/install.yaml
```

## Architecture

### Controller Pattern

The operator follows a **base controller + game-specific controller** pattern:

1. **Base Controller** (`internal/controller/base_controller.go`): Provides reusable reconciliation logic for:
   - PVC management with `ReconcilePVC`
   - Service creation (separate TCP/UDP services) with `ReconcileServices`
   - GCP ComputeAddress provisioning with `CreateGCPComputeAddress` and `IsGCPComputeAddressReady`
   - Helper functions: `GetSecureGameServerContainer`, `GetSecureCodeServerContainer`, `CompareDeployments`

2. **Game-Specific Controller** (e.g., `internal/controller/dayz_controller.go`): Implements reconciliation for a specific game by:
   - Calling base functions for PVC, Services, and ComputeAddress
   - Generating game-specific Deployment with init containers for configuration
   - Managing finalizers for cleanup (especially PVC preservation via `PreserveOnDelete`)

### Key Reconciliation Flow (DayzReconciler as Reference)

1. **Finalizer Management**: Add finalizer on create; on delete, remove PVC ownerRef if `PreserveOnDelete=true`, then remove finalizer
2. **ComputeAddress**: Create/verify GCP static IP using Upbound Provider GCP CRDs (`compute.gcp.upbound.io/v1beta1.Address`)
3. **PVC**: Create persistent storage with configurable size and storageClassName
4. **Deployment**:
   - **Init Container 1** (`config-writer`): Writes config files from CRD's `Config` map to `/tmp/configs` volume
   - **Init Container 2** (`config-setup`): Copies configs to `/data`, installs git if needed, runs `PostCopyCommands`
   - **Main Container** (`server`): Runs LinuxGSM game server as root (required by LinuxGSM)
   - **Sidecar Container** (`code-server`): Web-based editor on port 8080, runs as user 1000
5. **Services**: Separate LoadBalancer services for TCP (includes code-server port 8080) and UDP, with `ExternalTrafficPolicy: Local` and optional static IP from ComputeAddress

### API Types Pattern

Each game follows this structure (see `api/v1/dayz_types.go`):

```go
type <Game>Spec struct {
    Image string
    Base `json:",inline"`          // Embeds Persistence, Ports, Resources, NodeSelector, etc.
    Config <Game>Config             // map[string]string for file paths → content
    PostCopyCommands []string       // Shell commands run after config copy
}

type Base struct {
    Persistence Persistence
    Ports []corev1.ServicePort
    Resources corev1.ResourceRequirements
    NodeSelector, Tolerations, Affinity, Annotations
    EditorPassword string
}
```

### GCP Integration

- Uses Upbound Provider GCP (`github.com/upbound/provider-gcp`) for provisioning ComputeAddresses
- Requires ProviderConfig named `default` (currently hardcoded, see `base_controller.go:201`)
- Region is hardcoded to `southamerica-east1` (see `base_controller.go:251`)
- Services reference the static IP via `LoadBalancerIP` field once ComputeAddress is ready

## Adding a New Game (Follow DayZ Pattern)

The README outlines a 10-step checklist (lines 40-50). Key points:

1. **API Types**: Add `api/v1/<game>_types.go` with `<Game>Spec` embedding `Base` and `Config map[string]string`
2. **DeepCopy**: Run `make generate` to generate `zz_generated.deepcopy.go`
3. **Controller**: Add `internal/controller/<game>_controller.go` with:
   - PVC, Services, Deployment, finalizer logic
   - Init-container config writer and setup scripts
   - Call base functions for common resources
4. **Manager Wiring**: Register reconciler in `cmd/main.go` (`SetupWithManager`)
5. **CRD**: Ensure `config/crd/bases/gameserver.templarfelix.com_<plural>.yaml` exists and is listed in `config/crd/kustomization.yaml`
6. **Samples**: Add `config/samples/gameserver_v1_<game>.yaml` and reference in `config/samples/kustomization.yaml`
7. **Ports**: Define TCP/UDP ports in Spec.Ports; services are created separately for each protocol
8. **Storage**: Default size is 10G; support `PreserveOnDelete` by removing PVC ownerRef in finalizer
9. **Security**: Game container runs as root (LinuxGSM requirement); init containers and code-server run as 1000
10. **CompareDeployments**: Use helper from `utils.go` to avoid reconcile thrashing on unmanaged fields

## Important Implementation Details

### Concurrency and Conflicts
- Controllers log conflict errors (`errors.IsConflict`) and rely on requeue for retries
- Finalizer addition/removal may encounter conflicts; requeue on conflict

### Configuration Injection
- Game configs are defined as `Config map[string]string` in CRD (file path → content)
- `config-writer` init container writes these to `/tmp/configs` emptyDir
- `config-setup` init container copies from `/tmp/configs` to persistent `/data` paths
- Supports nested paths; script creates parent directories

### Security Context
- Game server container: `runAsUser: 0, runAsGroup: 0` (LinuxGSM requires root)
- Init containers: `runAsUser: 1000, runAsGroup: 1000` where applicable
- Code-server sidecar: `runAsUser: 1000, runAsGroup: 1000`
- Pod FSGroup: `1000`

### Services
- TCP and UDP ports are separated into distinct LoadBalancer services
- Code-server port 8080 is added to TCP service
- Annotation: `cloud.google.com/load-balancer-type: External`
- ExternalTrafficPolicy: `Local` (for GKE)

### Resource Defaults (utils.go)
- CPU: 500m request, 2000m limit
- Memory: 1Gi request, 4Gi limit
- Code-server: 100m CPU / 128Mi RAM request, 500m CPU / 512Mi RAM limit

### Upbound Provider GCP
- Scheme registration in `cmd/main.go:62-66` manually adds `compute.gcp.upbound.io/v1beta1` types
- RBAC marker: `+kubebuilder:rbac:groups=compute.gcp.upbound.io,resources=addresses,verbs=get;list;watch;create;update;patch;delete`
- Error handling in `base_controller.go` checks for missing CRDs, ProviderConfig, permissions

## Testing Notes

- Unit tests use `setup-envtest` with Kubernetes version extracted from `go.mod` (`ENVTEST_K8S_VERSION`)
- E2e tests create a Kind cluster named `gameserver-operator-test-e2e`
- Test files: `internal/controller/dayz_controller_test.go`, `test/e2e/e2e_test.go`

## Current Branch and Status

- Main branch: `main`
- Current branch: `1.0.0`
- Modified files: `cmd/main.go`, `internal/controller/base_controller.go`, `internal/controller/dayz_controller.go`
- Recent commits focus on GKE LoadBalancer configuration and ExternalTrafficPolicy

## Dependencies

- Go: 1.24.0
- controller-runtime: v0.22.1
- Kubernetes: v0.34.1
- Upbound Provider GCP: v1.14.0
- Crossplane runtime: v1.17.0