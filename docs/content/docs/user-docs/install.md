---
title: "Installation"
description: "Installing an ACK controller"
lead: ""
draft: false
menu: 
  docs:
    parent: "installing"
weight: 10
toc: true
---

The following guide will walk you through the installation of an [ACK service controller][ack-services].

Individual ACK service controllers may be in different maintenance phases and follow separate release cadences. Please check the [project stages][proj-stages] and [maintenance phases][maint-phases] of the ACK service controllers you wish to install, including how controllers are [released and versioned][rel-ver]. Controllers in a preview maintenance phase have at least one container image and Helm chart released to a public repository. 

{{% hint title="Be mindful of maintenance phases" %}}
Check the [project stage](../../community/releases/#project-stages) and [maintenance phase](../../community/releases/#maintenance-phases) of the ACK service controller you wish to install. Be aware that controllers in a preview maintenance phase may have significant and breaking changes introduced in a future release.
{{% /hint %}}

[proj-stages]: ../../community/releases/#project-stages
[maint-phases]: ../../community/releases/#maintenance-phases
[ack-services]: ../../community/services/
[rel-ver]: ../../community/releases/#releases-and-versioning

## Install an ACK service controller with Helm (Recommended)

The recommended way to install an ACK service controller for Kubernetes is to use [Helm 3.7+][helm-3-install].

{{% hint type="warning" title="Use the correct Helm version" %}}
Helm 3.7 introduced breaking changes to this installation guide. Be sure to install a Helm version that is greater than or equal to 3.7.
{{% /hint %}}

[helm-3-install]: https://helm.sh/docs/intro/install/

Each ACK service controller has a separate Helm chart that installs the necessary supporting artifacts as a Kubernetes `Deployment`. This includes the ACK service controller, custom resource definitions (CRDs), and Kubernetes Role-Based Access Control (RBAC) manifests.

Helm charts for ACK service controllers can be found in the [ACK registry within the Amazon ECR Public Gallery][ack-ecr-gallery]. To find a Helm chart for a specific service, you can go to `gallery.ecr.aws/aws-controllers-k8s/$SERVICENAME-chart`. For example, the link to the ACK service controller Helm chart for Amazon Simple Storage Service (Amazon S3) is [`gallery.ecr.aws/aws-controllers-k8s/s3-chart`][s3-ecr-chart].

Helm charts for individual ACK service controllers are tagged with their release version. You can find charts for different releases under the `Image tags` section in the chart repository on the ECR Public Gallery.

[ack-ecr-gallery]: https://gallery.ecr.aws/aws-controllers-k8s
[s3-ecr-chart]: https://gallery.ecr.aws/aws-controllers-k8s/s3-chart

Before installing a Helm chart, you must first make the Helm chart available on the `Deployment` host. To do so, use the `helm pull` command and then extract the chart:

```bash
export HELM_EXPERIMENTAL_OCI=1
export SERVICE=s3
export RELEASE_VERSION=v0.0.1
export CHART_EXPORT_PATH=/tmp/chart
export CHART_REF=$SERVICE-chart
export CHART_REPO=public.ecr.aws/aws-controllers-k8s/$CHART_REF
export CHART_PACKAGE=$CHART_REF-$RELEASE_VERSION.tgz

mkdir -p $CHART_EXPORT_PATH

helm pull oci://$CHART_REPO --version $RELEASE_VERSION -d $CHART_EXPORT_PATH
tar xvf $CHART_EXPORT_PATH/$CHART_PACKAGE -C $CHART_EXPORT_PATH
```

Once the Helm chart is downloaded and exported, you can install a particular ACK service controller using the `helm install` command:

```bash
export ACK_K8S_NAMESPACE=ack-system

helm install --create-namespace --namespace $ACK_K8S_NAMESPACE ack-$SERVICE-controller \
    $CHART_EXPORT_PATH/$SERVICE-chart
```

The `helm install` command should return relevant installation information:

```bash
helm install --namespace $ACK_K8S_NAMESPACE ack-$SERVICE-controller $CHART_EXPORT_PATH/$SERVICE-chart
NAME: s3-chart
LAST DEPLOYED: Thu Dec 17 13:09:17 2020
NAMESPACE: ack-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

To verify that the Helm chart was installed, use the `helm list` command:

```bash
helm list --namespace $ACK_K8S_NAMESPACE -o yaml
```

The `helm list` command should return your newly-deployed Helm chart release information:

```bash
helm list --namespace $ACK_K8S_NAMESPACE -o yaml
- app_version: v0.0.1
  chart: s3-controller
  name: ack-s3-controller
  namespace: ack-system
  revision: "1"
  status: deployed
  updated: 2020-12-17 13:09:17.309002201 -0500 EST
```

## Install an ACK service controller with static Kubernetes manifests

If you prefer not to use Helm, you may install an ACK service controller using static Kubernetes manifests that are included in the source repository. 

Static Kubernetes manifests install an individual service controller as a Kubernetes `Deployment`, including the relevant Kubernetes RBAC resources. Static Kubernetes manifests are available in the `config/` directory of the associated ACK service controller's source repository.

For example, the static manifests needed to install the S3 service controller for ACK are available in the [`config/`][s3-config-dir] directory in the [S3 controller's source repository][s3-repo].

[s3-config-dir]: https://github.com/aws-controllers-k8s/s3-controller/tree/main/config
[s3-repo]: https://github.com/aws-controllers-k8s/s3-controller

## Next steps

Once you have installed your ACK service controllers, you can [configure Kubernetes and AWS permissions][authorization].

[authorization]: ../authorization/
