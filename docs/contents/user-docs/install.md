# Install

In the following we walk you through installing an AWS service controller.

## Helm (recommended)

The recommended way to install an AWS service controller for Kubernetes is to
use Helm 3.

Before installing an AWS service controller, ensure you have added the
AWS Controllers for Kubernetes Helm repository:

```
helm repo add ack https://aws.github.io/aws-controllers-k8s
```

Each AWS service controller is packaged into a separate container image, published on a public AWS Elastic Container Registry repository. Likewise,
each AWS service controller has a separate Helm chart that installs—as a
Kubernetes `Deployment`—the AWS service controller, necessary custom resource
definitions (CRDs), Kubernetes RBAC manifests, and other supporting artifacts.

You may install a particular AWS service controller using the `helm install`
CLI command:

```
helm install [--namespace $KUBERNETES_NAMESPACE] ack/$SERVICE_ALIAS
```

for example, if you wanted to install the AWS S3 service controller into the
"ack-system" Kubernetes namespace, you would execute:


```sh
helm install --namespace ack-system ack/s3
```

## Static Kubernetes manifests

If you prefer not to use Helm, you may install a service controller using
static Kubernetes manifests.

Static Kubernetes manifests that install individual service controllers are
attached as artifacts to releases of AWS Controllers for Kubernetes. Select a
release from the [list of
releases](https://github.com/aws/aws-controllers-k8s/releases) for AWS
Controllers for Kubernetes.

You will see a list of Assets for the release. One of those Assets will be
named `services/$SERVICE_ALIAS/all-resources.yaml`. For example, for the AWS S3
service controller, there will be an Asset named
`services/s3/all-resources.yaml` attached to the release. Click on the link to
download the YAML file. This YAML file may be fed to `kubectl apply -f`
directly to install the service controller, any CRDs that it manages, and all
necessary Kubernetes RBAC manifests.

For example:

```sh
kubectl apply -f https://github.com/aws/aws-controllers-k8s/releases/download/v0.0.1/services/s3/all-resources.yaml
```

Once you've installed one or more ACK service controllers, make sure to
[configure permissions](../authorization#configure-permissions), next.
