# GameServer-Operator ( LinuxGSM )

Kubernetes operator for running LinuxGSM game servers.

## Thanks TO <a href="https://www.templarfelix.com"><img src="https://media.giphy.com/media/hvRJCLFzcasrR4ia7z/giphy.gif" width="25px"></a>

- **LinuxGSM** - [Visit LinuxGSM Site](https://linuxgsm.com/)
- **Operator SDK** - [Visit OperatorSDK Site](https://sdk.operatorframework.io)
- **Steam** - [Visit Steam Site](https://store.steampowered.com)

## Description

The `GameServer-Operator` is a Kubernetes project designed to facilitate the deployment and management of LinuxGSM game
servers in a Kubernetes environment. With this operator, users can easily scale their game servers, automate updates,
and maintain desired configurations through Custom Resource Definitions (CRDs). This project aims to simplify the
complexity of managing game servers by providing a robust and scalable solution for gaming communities and service
providers.

## Need help to deploy kubernetes clusters?

Navigate to the [gameserver-operator-infra](https://github.com/templarfelix/gameserver-operator-infra) GitHub project and read the docs.

## Supported Games

The operator is capable of managing a variety of game servers supported by the LinuxGSM platform. Below is a list of
popular games that are compatible, along with links to their specific configurations:

- **DayZ** - [Configurations](/_docs/dayz.md)
- **Project Zomboid** - [Configurations](/_docs/projectzomboid.md)
- **Rust** - [Configurations](/_docs/rust.md)
- **AnotherGames** - [Open Ticket](https://github.com/templarfelix/gameserver-operator/issues/new?assignees=&labels=&projects=&template=gamerequest.md&title=)

Full backlog to implement (from LinuxGSM): [GAMES_TO_IMPLEMENT.md](/_docs/GAMES_TO_IMPLEMENT.md)

For a complete list of upstream supported games, visit the [LinuxGSM servers page](https://linuxgsm.com/servers/).

## Implementation notes and validation

- DayZ CRD and controller are validated 100% functional in this repository. Use DayZ as the reference when onboarding new games.
- Checklist when adding a new game (follow DayZ pattern):
  1. API types: add types_<game>.go under api/v1alpha1/game with Spec embedding Base and Config map[string]string.
  2. DeepCopy: run `make generate` or provide temporary zz_generated deepcopy until controller-gen is wired.
  3. Controller: add internal/controller/game/controller_<game>.go implementing PVC, Services, Deployment, finalizer, and init-container config writer.
  4. Manager: wire the reconciler in cmd/main.go (SetupWithManager).
  5. CRD: include CRD YAML in config/crd/bases (gameserver.templarfelix.com_<plural>.yaml) and reference it in config/crd/kustomization.yaml.
  6. Samples: add config/samples/gameserver_v1alpha1_<game>.yaml and ensure config/samples/kustomization.yaml lists it.
  7. Ports: ensure Spec.Ports matches the game’s required TCP/UDP ports; Services are created separately for TCP and UDP with code-server added to TCP.
  8. Storage: verify Persistence defaults and PreserveOnDelete behavior; PVC ownerRef removed when preserveOnDelete=true.
  9. Security: keep game container running as root (LinuxGSM requires), init containers run as 1000 where applicable; code-server runs as 1000.
  10. CompareDeployments: rely on helper to avoid reconcile thrash; don’t mutate unmanaged fields.

## Steam Configuration

To configure Steam credentials for game server authentication, please follow the official LinuxGSM documentation:

**SteamCMD Setup:** https://docs.linuxgsm.com/steamcmd

This is required for the operator to authenticate with Steam and download/update game server files.

## Getting Started

## Install

Examples

### Kustomize

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - github.com/templarfelix/gameserver-operator/config/default?ref=main
```

### ArgoCD

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: gameserver-operator
  namespace: argocd  # This should be the namespace where Argo CD is installed
spec:
  project: default  # The Argo CD project, 'default' unless you've created others
  source:
    repoURL: 'https://github.com/templarfelix/gameserver-operator.git'
    targetRevision: 'main'
    path: 'config/default'
  destination:
    server: 'https://kubernetes.default.svc'  # URL of the Kubernetes API server
    namespace: 'default'  # The namespace in Kubernetes where to deploy the application
  syncPolicy:
    automated:  # Optional: enable automatic sync
      selfHeal: true
      prune: true  # This will prune resources that are not in git anymore
```


## Getting Started

### Prerequisites
- go version v1.24.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=templarfelix/gameserver-operator:latest
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=templarfelix/gameserver-operator:latest
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=templarfelix/gameserver-operator:latest
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/templarfelix/gameserver-operator/master/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
operator-sdk edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
