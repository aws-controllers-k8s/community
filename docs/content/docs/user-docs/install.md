---
title: "Install an ACK Controller"
description: "Install an ACK Controller"
lead: ""
draft: false
menu:
  docs:
    parent: "getting-started"
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

The recommended way to install an ACK service controller for Kubernetes is to use [Helm 3.8+][helm-3-install].

[helm-3-install]: https://helm.sh/docs/intro/install/

Each ACK service controller has a separate Helm chart that installs the necessary supporting artifacts as a Kubernetes `Deployment`. This includes the ACK service controller, custom resource definitions (CRDs), and Kubernetes Role-Based Access Control (RBAC) manifests.

Helm charts for ACK service controllers can be found in the [ACK registry within the Amazon ECR Public Gallery][ack-ecr-gallery]. To find a Helm chart for a specific service, you can go to `gallery.ecr.aws/aws-controllers-k8s/$SERVICENAME-chart`. For example, the link to the ACK service controller Helm chart for Amazon Simple Storage Service (Amazon S3) is [`gallery.ecr.aws/aws-controllers-k8s/s3-chart`][s3-ecr-chart].

Helm charts for individual ACK service controllers are tagged with their release version. You can find charts for different releases under the `Image tags` section in the chart repository on the ECR Public Gallery.

[ack-ecr-gallery]: https://gallery.ecr.aws/aws-controllers-k8s
[s3-ecr-chart]: https://gallery.ecr.aws/aws-controllers-k8s/s3-chart

Before installing a Helm chart, you can query the controller repository to find the latest release tag. This tag will correspond with a version of the Helm chart and a controller image. Then, you can use the Helm CLI to log into the ECR public Helm registry and install the chart.

```bash
export SERVICE=s3
export RELEASE_VERSION=$(curl -sL https://api.github.com/repos/aws-controllers-k8s/$SERVICE-controller/releases/latest | grep '"tag_name":' | cut -d'"' -f4)
export ACK_SYSTEM_NAMESPACE=ack-system
export AWS_REGION=us-west-2

aws ecr-public get-login-password --region us-east-1 | helm registry login --username AWS --password-stdin public.ecr.aws
helm install --create-namespace -n $ACK_SYSTEM_NAMESPACE ack-$SERVICE-controller \
  oci://public.ecr.aws/aws-controllers-k8s/$SERVICE-chart --version=$RELEASE_VERSION --set=aws.region=$AWS_REGION
```

{{% hint type="info" title="Specify a release version" %}}
The commands above download the latest version of the S3 controller. To select a
different version, change the `RELEASE_VERSION` variable and execute the commands again.
{{% /hint %}}

{{% hint type="info" title="Specify your target service region" %}}
The commands above set the target service region of the S3 controller to `us-west-2`. Be sure to specify your target service region in the `AWS_REGION` variable. This will be the *default* AWS region in which resources will be created by the ACK service controller. Note that a single ACK service controller can manage the lifecycle of resources in multiple AWS regions: simply add the `services.k8s.aws/region=$REGION` annotation to your resource. Alternately, you can add the `services.k8s.aws/region=$REGION` annotation to a Kubernetes `Namespace` and any resource launched in that `Namespace` will be created in that region by default.
{{% /hint %}}

The `helm install` command should return relevant installation information:

```bash
NAME: ack-s3-controller
LAST DEPLOYED: Thu Jun 16 19:30:16 2022
NAMESPACE: ack-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES: ...
```

To verify that the Helm chart was installed, use the `helm list` command:

```bash
helm list --namespace $ACK_SYSTEM_NAMESPACE -o yaml
```

The `helm list` command should return your newly-deployed Helm chart release information:

```bash
app_version: 0.1.1
chart: s3-chart-0.1.1
name: ack-s3-controller
namespace: ack-system
revision: "1"
status: deployed
updated: 2022-06-16 19:30:16.260012435 +0000 UTC
```

{{% hint type="important" title="NOTE" %}}
The `s3-controller` should be installed now, but it is NOT yet fully functional.
ACK controllers need access to AWS IAM credentials to manage AWS resources.
See [Next Steps](#Next-steps) for configuring AWS IAM credentials for ACK controller.
{{% /hint %}}

## Install an ACK service controller with static Kubernetes manifests

If you prefer not to use Helm, you may install an ACK service controller using static Kubernetes manifests that are included in the source repository.

Static Kubernetes manifests install an individual service controller as a Kubernetes `Deployment`, including the relevant Kubernetes RBAC resources. Static Kubernetes manifests are available in the `config/` directory of the associated ACK service controller's source repository.

For example, the static manifests needed to install the S3 service controller for ACK are available in the [`config/`][s3-config-dir] directory in the [S3 controller's source repository][s3-repo].

[s3-config-dir]: https://github.com/aws-controllers-k8s/s3-controller/tree/main/config
[s3-repo]: https://github.com/aws-controllers-k8s/s3-controller


## Install an ACK service controller with OperatorHub in Red Hat OpenShift

{{% hint type="important" title="Note" %}}
Please reference the [Openshift](../openshift) page for necessary pre-installation steps before installing any ACK controllers.
{{% /hint %}}

Navigate to the __Catalog -> OperatorHub__ page in the OpenShift web console and then search for the ACK service controller operator you wish to install. Click __Install__ and ensure you use the __All Namespaces__ install mode, if prompted.

For more information, see the official documentation for [installing Operators into an OpenShift cluster](https://docs.openshift.com/container-platform/4.9/operators/user/olm-installing-operators-in-namespace.html).

{{% hint type="info" title="Note" %}}
Since authentication setup is required before installing an ACK operator into OpenShift as explained in the OpenShift pre-installation steps, you do not need to set up authentication after installation as suggested below in "Next steps."
{{% /hint %}}

## Next steps

Once you have installed your ACK service controllers, you can
[create an IAM role to provide AWS access][irsa].

And learn the different ways that
[AWS credentials can be supplied][authentication] to the ACK controller.

[irsa]: ../irsa/
[authentication]: ../authentication/
