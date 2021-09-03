---
title : "Getting Started"
description: "AWS Controllers for Kubernetes (ACK) lets you define and use AWS service resources directly from Kubernetes"
lead: ""
draft: false
menu: 
  docs:
    parent: "prologue"
weight: 10
toc: true
---

**AWS Controllers for Kubernetes (ACK)** lets you define and use AWS service resources directly from Kubernetes. With ACK, you can take advantage of AWS-managed services for your Kubernetes applications without needing to define resources outside of the cluster or run services that provide supporting capabilities like databases or message queues within the cluster.

* [Install ACK service controllers][install]
* [Configure permissions for authorization and access][authorization]
* [IAM Roles for Service Accounts (IRSA)][irsa]
* [Cross-Account Resource Management (CARM)][carm]
* [Cleanup][cleanup]

## Docker images

Each ACK service controller is packaged into a separate container image that is published in a public repository corresponding to an individual ACK service controller.

{{% hint title="Choose the ACK Docker image that is right for you" %}}
Note that there is no single ACK Docker image. Instead, there are Docker
images for each individual ACK service controller that manages resources
for a particular AWS API.
{{% /hint %}}

Docker images for ACK service controllers can be found in the [ACK registry within the Amazon ECR Public Gallery][ack-ecr-gallery]. To find a Docker image for a specific service, you can go to `gallery.ecr.aws/aws-controllers-k8s/$SERVICENAME-controller`. For example, the link to the ACK service controller Docker image for Amazon Simple Storage Service (Amazon S3) is [`gallery.ecr.aws/aws-controllers-k8s/s3-controller`][s3-ecr-controller].

Individual ACK service controllers are tagged with their release version. You can find image URIs for different releases under the `Image tags` section in the image repository on the ECR Public Gallery.

{{% hint title="Be sure to specify a release version" type="info" %}}
You must always specify a version tag when referencing an ACK service controller image.
{{% /hint %}}

In accordance with [best practices][no-latest-tag], we do not include `:latest` default tags for our image repositories.

## Next steps

This guide assumes that you have access to a Kubernetes cluster. You do not need to use the Amazon Elastic Kubernetes Service (Amazon EKS) to get started with ACK service controllers. If you do not yet have a Kubernetes cluster and would like to use Amazon EKS, you can visit the [Amazon EKS Setup][eks-setup] guide. 

Once you have access to a Kubernetes cluster, you can [install the ACK service controller of your choice][install]. 

[ack-ecr-gallery]: https://gallery.ecr.aws/aws-controllers-k8s
[s3-ecr-controller]: https://gallery.ecr.aws/aws-controllers-k8s/s3-controller
[no-latest-tag]: https://vsupalov.com/docker-latest-tag/
[install]: https://aws-controllers-k8s.github.io/community/user-docs/install/
[authorization]: https://aws-controllers-k8s.github.io/community/user-docs/authorization/
[irsa]: https://aws-controllers-k8s.github.io/community/user-docs/irsa/
[carm]: https://aws-controllers-k8s.github.io/community/user-docs/cross-account-resource-management/
[cleanup]: https://aws-controllers-k8s.github.io/community/user-docs/cleanup/
[eks-setup]: https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/deep-learning-containers-eks-setup.html