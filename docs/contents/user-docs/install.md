# Install

In the following we walk you through installing an ACK service controller.

## Docker images

!!! note "No single ACK Docker image"
    Note that there is no single ACK Docker image. Instead, there are Docker
    images for each individual ACK service controller that manages resources
    for a particular AWS API.

Each ACK service controller is packaged into a separate container image,
published on the [`amazon/aws-controllers-k8s` DockerHub repository][0].

[0]: https://hub.docker.com/r/amazon/aws-controllers-k8s

Individual ACK service controllers are tagged with `$SERVICE-$VERSION` Docker
image tags, allowing you to download/test specific ACK service controllers. For
example, if you wanted to test the `v0.1.0` release image of the ACK service
controller for S3, you would pull the `amazon/aws-controllers-k8s:s3-v0.1.0`
image.

## Helm (recommended)

The recommended way to install an ACK service controller for Kubernetes is to
use Helm 3. Please ensure you have installed Helm 3 to your local environment
before running these steps.

Before installing an ACK service controller, ensure you have added the
AWS Controllers for Kubernetes Helm repository:

```
helm repo add ack https://aws.github.io/aws-controllers-k8s
```

Likewise, each ACK service controller has a separate Helm chart that
installs—as a Kubernetes `Deployment`—the ACK service controller, necessary
custom resource definitions (CRDs), Kubernetes RBAC manifests, and other
supporting artifacts.

You may install a particular ACK service controller using the `helm install`
CLI command:

```
helm install [--namespace $KUBERNETES_NAMESPACE] ack/ack-$SERVICE-controller
```

for example, if you wanted to install the latest ACK service controller for S3
into the "ack-system" Kubernetes namespace, you would execute:


```sh
helm install --namespace ack-system ack/ack-s3-controller
```

## Static Kubernetes manifests

If you prefer not to use Helm, you may install a service controller using
static Kubernetes manifests.

Static Kubernetes manifests that install individual service controllers are
attached as artifacts to releases of AWS Controllers for Kubernetes. Select a
release from the [list of releases][1] for AWS Controllers for Kubernetes.

[1]: https://github.com/aws/aws-controllers-k8s/releases

You will see a list of Assets for the release. One of those Assets will be
named `services/$SERVICE/all-resources.yaml`. For example, for the ACK service
controller for S3, there will be an Asset named
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
