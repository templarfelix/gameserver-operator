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

Navigate to the [gameserver-operator-infra](https://github.com/templarfelix/gameserver-operator-infra) GitHub and read the docs.

## Supported Games

The operator is capable of managing a variety of game servers supported by the LinuxGSM platform. Below is a list of
popular games that are compatible, along with links to their specific configurations:

- **DayZ** - [Configurations](/_docs/dayz.md)
- **AnotherGames** - [Open Ticket](https://github.com/templarfelix/gameserver-operator/issues/new?assignees=&labels=&projects=&template=gamerequest.md&title=)

For a complete list of supported games, visit the [LinuxGSM servers page](https://linuxgsm.com/servers/).

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

## Build Local

### Prerequisites

- go version v1.20.0+
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
> privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

> **NOTE**: Ensure that the samples has default values to test it out.

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

## Contributing

// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

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

