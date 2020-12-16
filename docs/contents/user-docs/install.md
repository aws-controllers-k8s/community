# Install

In the following we walk you through installing an ACK service controller.

## Docker images

!!! note "No single ACK Docker image"
    Note that there is no single ACK Docker image. Instead, there are Docker
    images for each individual ACK service controller that manages resources
    for a particular AWS API.

Each ACK service controller is packaged into a separate container image,
published on a [public ECR repository][controller-repo].

[controller-repo]: https://gallery.ecr.aws/aws-controllers-k8s/controller

Individual ACK service controllers are tagged with `$SERVICE-$VERSION` Docker
image tags, allowing you to download/test specific ACK service controllers. For
example, if you wanted to test the `v0.1.0` release image of the ACK service
controller for S3, you would pull the
`public.ecr.aws/aws-controllers-k8s/controller:s3-v0.1.0` image.

!!! note "No 'latest' tag"
    It is [not good practice][no-latest-tag] to rely on a `:latest` default
    image tag. There are actually no images tagged with a `:latest` tag in our
    image repositories. You should always specify a `$SERVICE-$VERSION` tag
    when referencing an ACK service controller image.

[no-latest-tag]: https://vsupalov.com/docker-latest-tag/

## Helm (recommended)

The recommended way to install an ACK service controller for Kubernetes is to
use Helm 3. Please ensure you have [installed Helm 3][install-helm] to your
local environment before running these steps.

[install-helm]: https://helm.sh/docs/intro/install/

Each ACK service controller has a separate Helm chart that installs—as a
Kubernetes `Deployment`—the ACK service controller, necessary custom resource
definitions (CRDs), Kubernetes RBAC manifests, and other supporting artifacts.

To view the Helm charts available for installation, check the ECR public
repository for the [ACK Helm charts][charts-repo]. Click on the "Image tags"
tab and take a note of the Helm chart tag for the service controller and
version you wish to install.

[charts-repo]: https://gallery.ecr.aws/aws-controllers-k8s/chart

Before installing a Helm chart, you must first make the Helm chart available on
the deployment host. To do so, use the `helm chart pull` and `helm chart
export` commands:

```bash
export HELM_EXPERIMENTAL_OCI=1
export SERVICE=s3
export RELEASE_VERSION=v0.0.1
export CHART_EXPORT_PATH=/tmp/chart
export CHART_REPO=public.ecr.aws/aws-controllers-k8s/chart
export CHART_REF=$CHART_REPO:$SERVICE-$RELEASE_VERSION

mkdir -p $CHART_EXPORT_PATH

helm chart pull $CHART_REF
helm chart export $CHART_REF --destination $CHART_EXPORT_PATH
```

You then install a particular ACK service controller using the `helm install`
CLI command:

```bash
export ACK_K8S_NAMESPACE=ack-system

kubectl create namespace $ACK_K8S_NAMESPACE

helm install --namespace $ACK_K8S_NAMESPACE ack-$SERVICE-controller \
    $CHART_EXPORT_PATH/ack-$SERVICE-controller
```

You will see the Helm chart installed:

```
$ helm install --namespace $ACK_K8S_NAMESPACE ack-$SERVICE-controller $CHART_EXPORT_PATH/ack-$SERVICE-controller
NAME: ack-s3-controller
LAST DEPLOYED: Thu Dec 17 13:09:17 2020
NAMESPACE: ack-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

You may then verify the Helm chart was installed using the `helm list` command:

```bash
helm list --namespace $ACK_K8S_NAMESPACE -o yaml
```

you should see your newly-deployed Helm chart release:

```
$ helm list --namespace $ACK_K8S_NAMESPACE -o yaml
- app_version: v0.0.1
  chart: ack-s3-controller-v0.0.1
  name: ack-s3-controller
  namespace: ack-system
  revision: "1"
  status: deployed
  updated: 2020-12-17 13:09:17.309002201 -0500 EST
```

## Static Kubernetes manifests

If you prefer not to use Helm, you may install a service controller using
static Kubernetes manifests.

Static Kubernetes manifests that install individual service controllers are
attached as artifacts to releases of AWS Controllers for Kubernetes. Select a
release from the [list of releases][1] for AWS Controllers for Kubernetes.

[1]: https://github.com/aws/aws-controllers-k8s/releases

TODO(jaypipes)
