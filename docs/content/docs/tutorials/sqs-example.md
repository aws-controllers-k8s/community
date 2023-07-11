---
title: "Manage SQS queues with the ACK SQS Controller"
description: "Create an SQS queue from an Amazon Elastic Kubernetes Service (EKS) deployment."
lead: "Create and manage an SQS queue directly from Kubernetes"
draft: false
menu:
  docs:
    parent: "tutorials"
weight: 45
toc: true
---

Amazon Simple Queue Service (SQS) is a fully managed message queuing service for microservices, distributed systems, and
serverless applications. SQS lets you send, store, and receive messages between software components
without losing messages or requiring other services to be available.

In this tutorial you will learn how to create and manage [SQS](https://aws.amazon.com/sqs) queues from an Amazon Elastic
Kubernetes (EKS) deployment.

## Setup

Although it is not necessary to use Amazon Elastic Kubernetes Service (Amazon EKS) with ACK, this guide assumes that you
have access to an Amazon EKS cluster. If this is your first time creating an Amazon EKS cluster, see [Amazon EKS
Setup](https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/deep-learning-containers-eks-setup.html).
For automated cluster creation using `eksctl`, see [Getting started with Amazon EKS -
`eksctl`](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html) and create your cluster with
Amazon EC2 Linux managed nodes.

### Prerequisites

This guide assumes that you have:

- Created an EKS cluster with Kubernetes version 1.24 or higher.
- AWS IAM permissions to create roles and attach policies to roles.
- AWS IAM permissions to send messages to a queue.
- Installed the following tools on the client machine used to access your Kubernetes cluster:
  - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting
    with AWS services.
  - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working
    with Kubernetes clusters.
  - [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html) - A command line tool for working with EKS
    clusters.
  - [Helm 3.8+](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications.

### Install the ACK service controller for SQS

> **_NOTE:_** This guide assumes you're using `us-east-1` as the region where the ACK controller will be deployed, as well as the Amazon SQS resource. If you want to create the object in another resource, simply change the region name to your region of choice.

Log into the Helm registry that stores the ACK charts:
```bash
aws ecr-public get-login-password --region us-east-1 | helm registry login --username AWS --password-stdin public.ecr.aws
```

Deploy the ACK service controller for Amazon SQS using the [sqs-chart Helm chart](https://gallery.ecr.aws/aws-controllers-k8s/sqs-chart). If you're looking to deploy resources to multiple regions, please refer to the [Manage Resources In Multiple Regions]([url](https://aws-controllers-k8s.github.io/community/docs/user-docs/multi-region-resource-management/)) documentation.

```bash
CONTROLLER_REGION=us-east-1
helm install --create-namespace -n ack-system oci://public.ecr.aws/aws-controllers-k8s/sqs-chart --version=1.0.4 --generate-name --set=aws.region=$CONTROLLER_REGION
```

For a full list of available values to the Helm chart, please [review the values.yaml file](https://github.com/aws-controllers-k8s/sqs-controller/blob/main/helm/values.yaml).

### Configure IAM permissions

Once the service controller is deployed, you will need to [configure the IAM permissions][irsa-permissions] for the
controller to query the SQS API. For full details, please review the AWS Controllers for Kubernetes documentation for
[how to configure the IAM permissions][irsa-permissions]. If you follow the examples in the documentation, use the value
of `sqs` for `SERVICE`.

## Create an SQS Queue

Execute the following command to create a manifest for a basic SQS queue, with an inline policy with `SendMessage`
permissions for the account owner, and submit this manifest to EKS cluster using kubectl.

{{% hint type="info" title="Make sure environment variables are set" %}}
If you followed the steps in the IAM permissions section above, the required environment variables `${AWS_REGION}` and
`${AWS_ACCOUNT_ID}` are already set. Otherwise please set these variables before executing the following steps. The value for `${AWS_REGION}` must also match the `--set=aws.region` value used in the `helm install` command above.
{{% /hint %}}

```bash
QUEUE_NAMESPACE=sqs-example
QUEUE_NAME=basic-sqs

kubectl create ns ${QUEUE_NAMESPACE}

cat <<EOF > basic-sqs-queue.yaml
apiVersion: sqs.services.k8s.aws/v1alpha1
kind: Queue
metadata:
  name: ${QUEUE_NAME}
  annotations:
    services.k8s.aws/region: ${AWS_REGION}
spec:
  queueName: ${QUEUE_NAME}
  policy: |
    {
      "Statement": [{
        "Sid": "__owner_statement",
        "Effect": "Allow",
        "Principal": {
          "AWS": "${AWS_ACCOUNT_ID}"
        },
        "Action": "sqs:SendMessage",
        "Resource": "arn:aws:sqs:${AWS_REGION}:${AWS_ACCOUNT_ID}:${QUEUE_NAME}"
      }]
    }
EOF

kubectl -n ${QUEUE_NAMESPACE} create -f basic-sqs-queue.yaml
```

The output of above commands looks like

```
namespace/sqs-example created
queue.sqs.services.k8s.aws/basic-sqs created
```

## Describe SQS Custom Resource

View the SQS custom resource to retrieve the `Queue URL` in the `Status` field

```bash
kubectl -n $QUEUE_NAMESPACE describe queue $QUEUE_NAME
```

The output of above commands looks like

```bash
Name:         basic-sqs
Namespace:    sqs-example
<snip>
Status:
  Ack Resource Metadata:
    Arn:               arn:aws:sqs:us-east-1:1234567890:basic-sqs
    Owner Account ID:  1234567890
    Region:            us-east-1
  Conditions:
    Last Transition Time:  2023-02-22T13:31:43Z
    Message:               Resource synced successfully
    Reason:                
    Status:                True
    Type:                  ACK.ResourceSynced
  Queue URL:               https://sqs.us-east-1.amazonaws.com/1234567890/basic-sqs
Events:                    <none>
```

Copy and set the Queue URL as an environment variable

```bash
QUEUE_URL=$(kubectl -n $QUEUE_NAMESPACE get queues/basic-sqs -o jsonpath='{.status.queueURL}')
```

## Send a Message

Execute the following command to send a message to the queue

```bash
aws sqs send-message --queue-url ${QUEUE_URL} --message-body "hello from ACK"
```

The output of above commands looks like

```
{
    "MD5OfMessageBody": "51e9ec3a483ba8b3159bc5fddcbbf36a",
    "MessageId": "281d7695-b066-4a50-853e-1b7c6c65f4a9"
}
```

Verify the message was received with

```bash
aws sqs receive-message --queue-url ${QUEUE_URL}
```

The output of above commands looks like

```
{
    "Messages": [
        {
            "MessageId": "281d7695-b066-4a50-853e-1b7c6c65f4a9",
            "ReceiptHandle": "ABCDeFZQxPfbAI201bRkdHZvRWeJUVSFfm2eL/T91L23ltB9nmf0dcx3ALQHz2WsXZhAbThZR+Ns5rX42+OjySNG6pi9Iu/SRZCVuuMzSBXeTrnLo8JjK3h9KE3uUkWirINgXd4fgVR2/C7feI3lCUhMOVhhYhec8ej5EDorL85Ay1IwZ43WYUQ1bIschP6xDvfzHk6vCi3kCXz6ZvPsNH3kTxp1gEvpQsaL/cq+aIZt/d1VVFsHtExbEk32iK1bo39tyA1A3Q7pT2WMowYh6MrfYdHoBw7PxJueGgx9MIQhQge2E+g6rKzGpFN9oPzPx59gu8n8n7Or6oncNM57pESD2LdzWTYjmS5H+Aw74qJ/gAMBIDNVuFt4Wl/5BvJHUTpOSAdi+Jekdbm3+AegzX8qyA==",
            "MD5OfBody": "51e9ec3a483ba8b3159bc5fddcbbf36a",
            "Body": "hello from ACK"
        }
    ]
```

## Next steps

The ACK service controller for Amazon SQS is based on the [Amazon SQS
API](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/Welcome.html).

Refer to [API Reference](https://aws-controllers-k8s.github.io/community/reference/) for *SQS* to find all the supported
Kubernetes custom resources and fields.

### Cleanup

Remove all the resource created in this tutorial using `kubectl delete` command.

```bash
kubectl -n ${QUEUE_NAMESPACE} delete -f basic-sqs-queue.yaml
```

The output of delete command should look like

```bash
queue.sqs.services.k8s.aws "basic-sqs" deleted
```

To remove the SQS ACK service controller, related CRDs, and namespaces, see [ACK Cleanup][cleanup].

To delete your EKS clusters, see [Amazon EKS - Deleting a cluster][cleanup-eks].

[irsa-permissions]: ../../user-docs/irsa/
[cleanup]: ../../user-docs/cleanup/
[cleanup-eks]: https://docs.aws.amazon.com/eks/latest/userguide/delete-cluster.html
