# Getting started With ACK service controllers

**AWS Controllers for Kubernetes (ACK)** makes it simple to build scalable and highly-available Kubernetes applications that utilize AWS services. The following sections describe how to work with ACK service controllers. 

* [Install ACK Service Controllers][install]
* [Configure Permissions for Authorization and Access][authorization]
* [IAM Roles for Service Accounts][irsa]
* [Cross-Account Resource Management][carm]
* [Cleanup][cleanup]

## Prerequisites 

To install an ACK service controller, you need the following: 

1. (Optional) An Amazon Elastic Kubernetes Service (Amazon EKS) cluster. If you haven't set up an Amazon EKS cluster, visit the [Amazon EKS Setup][eks-setup] guide. 
2. IAM permissions to create roles and attach policies to roles.
3. The following tools installed on the client machine used to access your Kubernetes cluster: 
    * [AWS CLI][aws-cli-install] - A command line tool for interacting with AWS services
    * [`eksctl`][eksctl-install] - A command line tool for creating and managing clusters on EKS
    * [`kubectl`][kubectl-install] - A command line tool for working with Kubernetes clusters
    * [Helm 3][helm-3-install] - (Optional) A tool for installing and managing Kubernetes applications

[eks-setup]: https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/deep-learning-containers-eks-setup.html
[aws-cli-install]: https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html
[eksctl-install]: https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html
[kubectl-install]: https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html
[helm-3-install]: https://helm.sh/docs/intro/install/

## Docker images

Each ACK service controller is packaged into a separate container image that is published in a public repository corresponding to an individual ACK service controller.

!!! note "Choose the ACK Docker image that is right for you"
    Note that there is no single ACK Docker image. Instead, there are Docker
    images for each individual ACK service controller that manages resources
    for a particular AWS API.

Docker images for ACK service controllers can be found in the [ACK registry within the Amazon ECR Public Gallery][ack-ecr-gallery]. To find a Docker image for a specific service, you can go to `gallery.ecr.aws/aws-controllers-k8s/$SERVICENAME-controller`. For example, the link to the ACK service controller Docker image for Amazon Simple Storage Service (Amazon S3) is [`gallery.ecr.aws/aws-controllers-k8s/s3-controller`][s3-ecr-controller].

Individual ACK service controllers are tagged with their release version. You can find image URIs for different releases under the `Image tags` section in the image repository on the ECR Public Gallery.

!!! note "Be sure to specify release version"
    You must always specify a version tag when referencing an ACK service controller image.

In accordance with [best practices][no-latest-tag], we do not include `:latest` default tags for our image repositories.

[ack-ecr-gallery]: https://gallery.ecr.aws/aws-controllers-k8s
[s3-ecr-controller]: https://gallery.ecr.aws/aws-controllers-k8s/s3-controller
[no-latest-tag]: https://vsupalov.com/docker-latest-tag/
[install]: https://aws-controllers-k8s.github.io/community/user-docs/install/
[authorization]: https://aws-controllers-k8s.github.io/community/user-docs/authorization-and-access/
[irsa]: https://aws-controllers-k8s.github.io/community/user-docs/irsa/
[carm]: https://aws-controllers-k8s.github.io/community/user-docs/carm/
[cleanup]: https://aws-controllers-k8s.github.io/community/user-docs/cleanup/
