---
title: "Create a Secret with AWS Secrets Manager"
description: "Use ACK secretsmanger-controller to create and manage secrets directly from Kubernetes. "
lead: "Use ACK secretsmanger-controller to create and manage secrets directly from Kubernetes."
draft: false
menu:
  docs:
    parent: "tutorials"
weight: 43
toc: true
---

The ACK service controller for AWS Secrets Manager lets you create secrets directly from Kubernetes.
This guide shows you how to create a new secret in AWS Secrets Manager using a reference to a Kubernetes Secret.

## Setup

Although it is not necessary to use Amazon Elastic Kubernetes Service (Amazon EKS) or Amazon Elastic Container Registry (Amazon ECR) with ACK, this guide assumes that you
have access to an Amazon EKS cluster. If this is your first time creating an Amazon EKS cluster, see
[Amazon EKS Setup][eks-setup]. For automated cluster creation using `eksctl`, see [Getting started with Amazon EKS - `eksctl`](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html) and create your cluster with Amazon EC2 Linux managed nodes.

## Prerequisites

This guide assumes that you have:

- Created an EKS cluster with Kubernetes version 1.16 or higher.
- AWS IAM permissions to create roles and attach policies to roles.
- Installed the following tools on the client machine used to access your Kubernetes cluster:
  - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting with AWS services.
  - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working with Kubernetes clusters.
  - [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html) - A command line tool for working with EKS clusters.
  - [Helm 3.8+](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications.

### Install the Secrets Manager ACK service controller

Log into the Helm registry that stores the ACK charts:

```bash
aws ecr-public get-login-password --region us-east-1 | helm registry login --username AWS --password-stdin public.ecr.aws
```

Deploy the ACK service controller for AWS Secrets Manager using the [secretsmanager-chart Helm chart](https://gallery.ecr.aws/aws-controllers-k8s/secretsmanager-chart). This example creates resources in the `us-west-2` region, but you can use any other region supported in AWS.

```bash
SERVICE=secretsmanager
RELEASE_VERSION=$(curl -sL https://api.github.com/repos/aws-controllers-k8s/${SERVICE}-controller/releases/latest | jq -r '.tag_name | ltrimstr("v")')
helm install --create-namespace -n ack-system oci://public.ecr.aws/aws-controllers-k8s/secretsmanager-chart "--version=${RELEASE_VERSION}" --generate-name --set=aws.region=us-west-2
```

For a full list of available values to the Helm chart, please [review the values.yaml file](https://github.com/aws-controllers-k8s/secretsmanager-controller/blob/main/helm/values.yaml).

### Configure IAM permissions

Once the service controller is deployed [configure the IAM permissions](https://aws-controllers-k8s.github.io/community/docs/user-docs/irsa/) for the
controller to invoke the Secrets Manager API. For full details, please review the AWS Controllers for Kubernetes documentation
for [how to configure the IAM permissions](https://aws-controllers-k8s.github.io/community/docs/user-docs/irsa/). If you follow the examples in the documentation, use the
value of `secretsmanager` for `SERVICE`.

### Create Kubernetes Secret

Now that the ACK secretsmanager-controller is setup we'll need to create a Kubernetes Secret. 

```bash
cat <<EOF > secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-k8s-secret
  namespace: default
type: Opaque
data:
  secret: UzNjcjN0UGFzc3cwcmQ=  # Base64 encoded "S3cr3tPassw0rd"
EOF
```

```bash
kubectl apply -f secret.yaml
```

### Create ACK Secret
Finally, we'll create an ACK Secret referencing the Kubernetes Secret we just created. 

```bash
cat <<EOF > aws-secret.yaml
apiVersion: secretsmanager.services.k8s.aws/v1alpha1
kind: Secret
metadata:
  name: my-aws-secret
spec:
  name: sample-aws-secret
  description: "A sample secret created for demonstration"
  secretString:
    name: my-k8s-secret
    namespace: default
    key: secret
EOF
```

```bash
kubectl apply -f aws-secret.yaml
```

You can verify that the secret was created with the AWS CLI.

```bash
aws secretsmanager describe-secret sample-aws-secret
```

### Cleanup
You can delete your ACK and Kubernetes Secrets using the `kubectl delete` command:

```bash
kubectl delete -f secret.yaml && kubectl delete -f aws-secret.yaml
```

To remove the Secrets Manager ACK service controller, related CRDs, and namespaces, see [ACK Cleanup][cleanup].

To delete your EKS clusters, see [Amazon EKS - Deleting a cluster][cleanup-eks].

[eks-setup]: https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/deep-learning-containers-eks-setup.html
[cleanup]: ../../user-docs/cleanup/
[cleanup-eks]: https://docs.aws.amazon.com/eks/latest/userguide/delete-cluster.html






