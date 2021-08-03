# Install

In the following we walk you through installing an ACK service controller.

!!! note **IMPORTANT**
    Individual ACK service controllers may be in different
    maintenance phases and follow separate release cadences. Please read our
    documentation about [project stages][proj-stages] and
    [maintenance phases][maint-phases] fully, including how we
    [release and version][rel-ver] our controllers. Controllers in a `PREVIEW`
    maintenance phase have at least one container image and Helm chart released to
    an ECR Public repository. However, be aware that controllers in a `PREVIEW`
    maintenance phase may have significant and breaking changes introduced in a
    future release.

[proj-stages]: https://aws-controllers-k8s.github.io/community/releases/#project-stages
[maint-phases]: https://aws-controllers-k8s.github.io/community/releases/#maintenance-phases
[rel-ver]: https://aws-controllers-k8s.github.io/community/releases/#releases-and-versioning

## Docker images

!!! note "No single ACK Docker image"
    Note that there is no single ACK Docker image. Instead, there are Docker
    images for each individual ACK service controller that manages resources
    for a particular AWS API.

Each ACK service controller is packaged into a separate container image,
published in a public ECR repository that corresponds to that ACK service
controller.

Controller images can be found in an ECR Public repository following the naming
scheme `public.ecr.aws/aws-controllers-k8s/{SERVICE}-controller`. For example,
you can find the Docker images for different releases of the RDS service
controller for ACK in the `public.ecr.aws/aws-controllers-k8s/rds-controller`
repository.

There is an ECR Public Gallery link for the controller image repository at a
similarly-schemed link
`https://gallery.ecr.aws/aws-controllers-k8s/$SERVICE-controller`. For
instance, to view the available Docker image releases of the RDS service
controller for ACK, visit
[https://gallery.ecr.aws/aws-controllers-k8s/rds-controller][rds-ecr-gallery].

[rds-ecr-gallery]: https://gallery.ecr.aws/aws-controllers-k8s/rds-controller

Individual ACK service controllers are tagged with a Semantic Version release
tag, for example `v0.5.0`.

!!! note "No 'latest' tag"
    It is [not good practice][no-latest-tag] to rely on a `:latest` default
    image tag. There are actually no images tagged with a `:latest` tag in our
    image repositories. You should always specify a Semantic Version tag
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

There is a separate ECR Public repository that contains the Helm charts for a
specific ACK service controller. The ECR Public repository for Helm charts
follows the naming scheme `public.ecr.aws/aws-controllers-k8s/$SERVICE-chart`.
For example, you can find the Helm charts for different releases of the RDS
service controller for ACK in the
`public.ecr.aws/aws-controllers-k8s/rds-chart` repository.

There is an ECR Public Gallery link for the Helm chart repository at a
similarly-schemed link
`https://gallery.ecr.aws/aws-controllers-k8s/$SERVICE-chart`. For
instance, to view the available Helm charts that install releases of the RDS service
controller for ACK, visit
[https://gallery.ecr.aws/aws-controllers-k8s/rds-chart][rds-chart-ecr-gallery].

[rds-chart-ecr-gallery]: https://gallery.ecr.aws/aws-controllers-k8s/rds-chart

Before installing a Helm chart, you must first make the Helm chart available on
the deployment host. To do so, use the `helm chart pull` and `helm chart
export` commands:

```bash
export HELM_EXPERIMENTAL_OCI=1
export SERVICE=s3
export RELEASE_VERSION=v0.0.1
export CHART_EXPORT_PATH=/tmp/chart
export CHART_REPO=public.ecr.aws/aws-controllers-k8s/$SERVICE-chart
export CHART_REF=$CHART_REPO:$RELEASE_VERSION

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
  chart: s3-controller
  name: ack-s3-controller
  namespace: ack-system
  revision: "1"
  status: deployed
  updated: 2020-12-17 13:09:17.309002201 -0500 EST
```

## Static Kubernetes manifests

If you prefer not to use Helm, you may install a service controller using
static Kubernetes manifests that are included in the source repository for an
ACK service controller.

Static Kubernetes manifests that install the individual service controller as a
Kubernetes `Deployment`, along with the relevant Kubernetes RBAC resources are
available in the `config/` directory of the associated ACK service controller's
source repository.

For example, for the static manifests that will install the RDS service
controller for ACK, check out the [`config/`][rds-config-dir] directory of the
[RDS controller's source repo][rds-repo].

[rds-config-dir]: https://github.com/aws-controllers-k8s/rds-controller/tree/main/config
[rds-repo]: https://github.com/aws-controllers-k8s/rds-controller

## Next Steps

Once finished installing an ACK service controller, read about how
[Authorization and Access Control][authorization] works.

[authorization]: https://aws-controllers-k8s.github.io/community/user-docs/authorization/
