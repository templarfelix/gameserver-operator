# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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
- Base configuration (storage, resources, ports)
- Game-specific configuration
- Status conditions

**Reconciliation Pattern:**
Controllers follow standard Kubernetes operator pattern:
1. Reconcile PersistentVolumeClaim for game data
2. Reconcile Services (TCP/UDP separated)
3. Manage game server Pod deployment

**Development Notes:**
- Uses controller-runtime v0.16.3
- Requires Go 1.20+
- Follows standard Kubebuilder project structure
- Includes RBAC configurations in `config/rbac/`
- Sample configurations in `config/samples/`

**Testing:**
- Unit tests with envtest
- End-to-end tests require Kind cluster
- Linting with golangci-lint