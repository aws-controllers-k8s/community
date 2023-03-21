---
title: "Manage EventBridge Pipes with the ACK Pipes Controller"
description: "Forward messages between two SQS queues with a pipe."
lead: "Create and manage EventBridge Pipes directly from Kubernetes"
draft: false
menu:
  docs:
    parent: "tutorials"
weight: 45
toc: true
---

Amazon EventBridge Pipes connects sources to targets. It reduces the need for specialized knowledge and integration code
when developing event driven architectures, fostering consistency across your company’s applications. To set up a pipe,
you choose the source, add optional filtering, define optional enrichment, and choose the target for the event data.

In this tutorial you will learn how to create and manage an [EventBridge
Pipe](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-pipes.html) to forward messages between two SQS queues
from an Amazon Elastic Kubernetes (EKS) deployment.

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
- AWS IAM permissions to manages queues and send messages to a queue.
- Installed the following tools on the client machine used to access your Kubernetes cluster:
  - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting
    with AWS services.
  - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working
    with Kubernetes clusters.
  - [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html) - A command line tool for working with EKS
    clusters.
  - [Helm 3.8+](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications.
  - [jq](https://stedolan.github.io/jq/download/) to parse AWS CLI JSON output

### Install the ACK service controller for Pipes

Log into the Helm registry that stores the ACK charts:
```bash
aws ecr-public get-login-password --region us-east-1 | helm registry login --username AWS --password-stdin public.ecr.aws
```

Deploy the ACK service controller for Amazon Pipes using the [pipes-chart Helm chart](https://gallery.ecr.aws/aws-controllers-k8s/pipes-chart). Resources should be created in the `us-east-1` region:

```bash
helm install --create-namespace -n ack-system oci://public.ecr.aws/aws-controllers-k8s/pipes-chart --version=v0.0.3 --generate-name --set=aws.region=us-east-1
```

For a full list of available values to the Helm chart, please [review the values.yaml file](https://github.com/aws-controllers-k8s/pipes-controller/blob/main/helm/values.yaml).

### Configure IAM permissions

Once the service controller is deployed, you will need to [configure the IAM permissions][irsa-permissions] for the
controller to query the Pipes API. For full details, please review the AWS Controllers for Kubernetes documentation for
[how to configure the IAM permissions][irsa-permissions]. If you follow the examples in the documentation, use the value
of `pipes` for `SERVICE`.

## Create an EventBridge Pipe

### Create the source and target SQS queues

To keep the scope of this tutorial simple, the SQS queues and IAM permissions will be created with the AWS CLI.
Alternatively, the [ACK SQS
Controller](https://aws-controllers-k8s.github.io/community/docs/community/services/#amazon-sqs) and [ACK IAM
Controller](https://aws-controllers-k8s.github.io/community/docs/community/services/#amazon-iam) can be used to manage
these resources with Kubernetes.

Execute the following command to define the environment variables used throughout the example.

{{% hint type="info" title="Make sure environment variables are set" %}}
If you followed the steps in the IAM permissions section above, the required environment variables `${AWS_REGION}` and
`${AWS_ACCOUNT_ID}` are already set. Otherwise please set these variables before executing the following steps. The value for `${AWS_REGION}` must also match the `--set=aws.region` value used in the `helm install` command above.
{{% /hint %}}

```bash
export PIPE_NAME=pipes-sqs-to-sqs
export PIPE_NAMESPACE=pipes-example
export SOURCE_QUEUE=pipes-sqs-source
export TARGET_QUEUE=pipes-sqs-target
export PIPE_ROLE=pipes-sqs-to-sqs-role
export PIPE_POLICY=pipes-sqs-to-sqs-policy
```

Create the source and target queues.

```bash
aws sqs create-queue --queue-name ${SOURCE_QUEUE}
aws sqs create-queue --queue-name ${TARGET_QUEUE}
```

The output of above commands looks like

```bash
{
    "QueueUrl": "https://sqs.us-east-1.amazonaws.com/1234567890/pipes-sqs-source"
}
{
    "QueueUrl": "https://sqs.us-east-1.amazonaws.com/1234567890/pipes-sqs-target"
}
```

### Create the Pipes IAM Role

Create an IAM role for the pipe to consume messages from the source queue and send messages to the target queue.

```bash
cat <<EOF > trust.json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "pipes.amazonaws.com"
            },
            "Action": "sts:AssumeRole",
            "Condition": {
                "StringEquals": {
                    "aws:SourceAccount": "${AWS_ACCOUNT_ID}"
                }
            }
        }
    ]
}
EOF

aws iam create-role --role-name ${PIPE_ROLE} --assume-role-policy-document file://trust.json
```

The output of above commands looks like

```bash
{
    "Role": {
        "Path": "/",
        "RoleName": "pipes-sqs-to-sqs-role",
        "RoleId": "ABCDU3F4JDBEUCMGT3XBH",
        "Arn": "arn:aws:iam::1234567890:role/pipes-sqs-to-sqs-role",
        "CreateDate": "2023-03-21T13:11:59+00:00",
        "AssumeRolePolicyDocument": {
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Principal": {
                        "Service": "pipes.amazonaws.com"
                    },
                    "Action": "sts:AssumeRole",
                    "Condition": {
                        "StringEquals": {
                            "aws:SourceAccount": "1234567890"
                        }
                    }
                }
            ]
        }
    }
}
```

Attach a policy to the role to give the pipe permissions to read and send messages.

```bash
cat <<EOF > policy.json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "sqs:ReceiveMessage",
                "sqs:DeleteMessage",
                "sqs:GetQueueAttributes"
            ],
            "Resource": [
                "arn:aws:sqs:${AWS_REGION}:${AWS_ACCOUNT_ID}:${SOURCE_QUEUE}"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "sqs:SendMessage"
            ],
            "Resource": [
                "arn:aws:sqs:${AWS_REGION}:${AWS_ACCOUNT_ID}:${TARGET_QUEUE}"
            ]
        }
    ]
}
EOF

aws iam put-role-policy --role-name ${PIPE_ROLE} --policy-name ${PIPE_POLICY} --policy-document file://policy.json
```

If the command executes successfully, no output is generated.

### Create the Pipe

Execute the following command to retrieve the ARNs for the resources created above needed for the Kubernetes manifest.

```bash
export SOURCE_QUEUE_ARN=$(aws --output json sqs get-queue-attributes --queue-url "https://sqs.${AWS_REGION}.amazonaws.com/${AWS_ACCOUNT_ID}/${SOURCE_QUEUE}" --attribute-names QueueArn | jq -r '.Attributes.QueueArn')
export TARGET_QUEUE_ARN=$(aws --output json sqs get-queue-attributes --queue-url "https://sqs.${AWS_REGION}.amazonaws.com/${AWS_ACCOUNT_ID}/${TARGET_QUEUE}" --attribute-names QueueArn | jq -r '.Attributes.QueueArn')
export PIPE_ROLE_ARN=$(aws --output json iam get-role --role-name ${PIPE_ROLE} | jq -r '.Role.Arn')
```

Execute the following command to create a Kubernetes manifest for a pipe consuming messages from the source queue and
sending messages matching the filter criteria to the target queue using the above created IAM role.

The EventBridge filter pattern will match any SQS message from the source queue with a JSON-stringified body
`{\"from\":\"kubernetes\"}`. Alternatively, the filter pattern can be omitted to forward all messages from the source
queue.

```bash
kubectl create ns ${PIPE_NAMESPACE}

cat <<EOF > pipe-sqs-to-sqs.yaml
apiVersion: pipes.services.k8s.aws/v1alpha1
kind: Pipe
metadata:
  name: $PIPE_NAME
spec:
  name: $PIPE_NAME
  source: $SOURCE_QUEUE_ARN
  description: "SQS to SQS Pipe with filtering"
  sourceParameters:
    filterCriteria:
      filters:
        - pattern: "{\"body\":{\"from\":[\"kubernetes\"]}}"
    sqsQueueParameters:
      batchSize: 1
      maximumBatchingWindowInSeconds: 1
  target: $TARGET_QUEUE_ARN
  roleARN: $PIPE_ROLE_ARN
EOF

kubectl -n ${PIPE_NAMESPACE} create -f pipe-sqs-to-sqs.yaml
```

The output of above commands looks like

```bash
namespace/pipes-example created
pipe.pipes.services.k8s.aws/pipes-sqs-to-sqs created
```

### Describe Pipe Custom Resource

View the Pipe custom resource to verify it is in a `RUNNING` state.

```bash
kubectl -n $PIPE_NAMESPACE get pipe $PIPE_NAME
```

The output of above commands looks like

```bash
NAME               STATE     SYNCED   AGE
pipes-sqs-to-sqs   RUNNING   True     3m10s
```

### Verify the Pipe filtering and forwarding is working

Execute the following command to send a message to the source queue with a body matching the pipe filter pattern.

```bash
aws sqs send-message --queue-url https://sqs.${AWS_REGION}.amazonaws.com/${AWS_ACCOUNT_ID}/${SOURCE_QUEUE} --message-body "{\"from\":\"kubernetes\"}"
```

The output of above commands looks like

```bash
{
    "MD5OfMessageBody": "fde2da607356f1974691e48fa6a87157",
    "MessageId": "f4157187-0308-420c-b69b-aa439e6be7e3"
}
```

Verify the message was consumed by the pipe, the filter pattern matched and the message was received by the target queue
with

```bash
aws sqs receive-message --queue-url https://sqs.${AWS_REGION}.amazonaws.com/${AWS_ACCOUNT_ID}/${TARGET_QUEUE}
```

{{% hint type="info" title="Receive Delays" %}}
It might take some time for the Pipe to consume the message from the source and deliver it to the target queue.
If the above command does not return a message, rerun the command a couple of times with some delay in between the requests.
{{% /hint %}}

The output of above commands looks like

```bash
{
    "Messages": [
        {
            <snip>
            "MD5OfBody": "d5255184c571cca2c78e76d6eea1745d",
            "Body": "{\"messageId\":\"f4157187-0308-420c-b69b-aa439e6be7e3\",
            <snip>
            \"body\":\"{\\\"from\\\":\\\"kubernetes\\\"}\",\"attributes\":{\"ApproximateReceiveCount\":\"1\",
            <snip>
            \"eventSourceARN\":\"arn:aws:sqs:us-east-1:1234567890:pipes-sqs-source\",\"awsRegion\":\"us-east-1\"}"
        }
    ]
}
```

## Next steps

The ACK service controller for Amazon EventBridge Pipes is based on the [Amazon EventBridge Pipes
API](https://docs.aws.amazon.com/eventbridge/latest/pipes-reference/Welcome.html).

Refer to [API Reference](https://aws-controllers-k8s.github.io/community/reference/) for *Pipes* to find all the
supported Kubernetes custom resources and fields.

### Cleanup

Remove all the resource created in this tutorial using `kubectl delete` command.

```bash
kubectl -n ${QUEUE_NAMESPACE} delete -f pipe-sqs-to-sqs.yaml
```

The output of delete command should look like

```bash
pipe.pipes.services.k8s.aws "pipes-sqs-to-sqs" deleted
```

{{% hint type="info" title="Deleting Delays" %}}
It might take some time for the Pipe to be deleted as the operation is performed asynchronously in the API.
{{% /hint %}}

To remove the Pipes ACK service controller, related CRDs, and namespaces, see [ACK Cleanup][cleanup].

To delete your EKS clusters, see [Amazon EKS - Deleting a cluster][cleanup-eks].

[irsa-permissions]: ../../user-docs/irsa/
[cleanup]: ../../user-docs/cleanup/
[cleanup-eks]: https://docs.aws.amazon.com/eks/latest/userguide/delete-cluster.html
