---
title: "Manage EventBridge event buses and rules with the ACK EventBridge Controller"
description: "Send filtered events on a custom bus to SQS."
lead: "Create and manage EventBridge event buses and rules directly from Kubernetes"
draft: false
menu:
  docs:
    parent: "tutorials"
weight: 45
toc: true
---

EventBridge is a serverless service that uses events to connect application components together, making it easier for
you to build scalable event-driven applications. Use it to route events from sources such as home-grown applications,
AWS services, and third-party software to consumer applications across your organization. EventBridge provides a simple
and consistent way to ingest, filter, transform, and deliver events so you can build new applications quickly.

In this tutorial you will learn how to create and manage a custom EventBridge [event
bus](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-bus.html) and
[rule](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-rules.html) to filter and forward messages to an SQS
[target](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-targets.html) from an Amazon Elastic Kubernetes
(EKS) deployment.

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

### Install the ACK service controller for EventBridge

Log into the Helm registry that stores the ACK charts:
```bash
aws ecr-public get-login-password --region us-east-1 | helm registry login --username AWS --password-stdin public.ecr.aws
```

Deploy the ACK service controller for Amazon EventBridge using the [eventbridge-chart Helm chart](https://gallery.ecr.aws/aws-controllers-k8s/eventbridge-chart). Resources should be created in the `us-east-1` region:

```bash
helm install --create-namespace -n ack-system oci://public.ecr.aws/aws-controllers-k8s/eventbridge-chart --version=v0.0.3 --generate-name --set=aws.region=us-east-1
```

For a full list of available values to the Helm chart, please [review the values.yaml file](https://github.com/aws-controllers-k8s/eventbridge-controller/blob/main/helm/values.yaml).

### Configure IAM permissions

Once the service controller is deployed, you will need to [configure the IAM permissions][irsa-permissions] for the
controller to query the EventBridge API. For full details, please review the AWS Controllers for Kubernetes documentation for
[how to configure the IAM permissions][irsa-permissions]. If you follow the examples in the documentation, use the value
of `eventbridge` for `SERVICE`.

## Create an EventBridge Custom Event Bus and Rule with an SQS Target 

### Create the target SQS queue

To keep the scope of this tutorial simple, the SQS queue and IAM permissions will be created with the AWS CLI.
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
export EVENTBRIDGE_NAMESPACE=eventbridge-example
export EVENTBUS_NAME=custom-eventbus-ack
export RULE_NAME=custom-eventbus-ack-sqs-rule
export TARGET_QUEUE=custom-eventbus-ack-rule-sqs-target
```

Create the target queue.

```bash
cat <<EOF > target-queue.json
{
    "QueueName": "${TARGET_QUEUE}",
    "Attributes": {
        "Policy": "{\"Statement\":[{\"Sid\":\"EventBridgeToSqs\",\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"events.amazonaws.com\"},\"Action\":[\"sqs:SendMessage\"],\"Resource\":\"arn:aws:sqs:${AWS_REGION}:${AWS_ACCOUNT_ID}:${TARGET_QUEUE}\",\"Condition\":{\"ArnEquals\":{\"aws:SourceArn\":\"arn:aws:events:${AWS_REGION}:${AWS_ACCOUNT_ID}:rule/${EVENTBUS_NAME}/${RULE_NAME}\"}}}]}"
    }
}
EOF

aws sqs create-queue --cli-input-json file://target-queue.json
```

The output of above commands looks like

```bash
{
    "QueueUrl": "https://sqs.us-east-1.amazonaws.com/1234567890/custom-eventbus-ack-rule-sqs-target"
}
```

### Create a Custom Event Bus

Execute the following command to create the example namespace and a custom event bus.

```bash
kubectl create ns ${EVENTBRIDGE_NAMESPACE}

cat <<EOF > bus.yaml
apiVersion: eventbridge.services.k8s.aws/v1alpha1
kind: EventBus
metadata:
  name: ${EVENTBUS_NAME}
spec:
  name: ${EVENTBUS_NAME}
EOF

kubectl -n ${EVENTBRIDGE_NAMESPACE} create -f bus.yaml
```

The output of above commands looks like

```bash
namespace/eventbridge-example created
eventbus.eventbridge.services.k8s.aws/custom-eventbus-ack created
```

Verify the event bus resource is synchronized.

```bash
kubectl -n ${EVENTBRIDGE_NAMESPACE} get eventbus ${EVENTBUS_NAME}
```

The output of above commands looks like

```bash
NAME                  SYNCED   AGE
custom-eventbus-ack   True     64s
```

### Create a Rule with an SQS Target

Execute the following command to retrieve the ARN for the SQS target created above needed for the Kubernetes manifest.

```bash
export TARGET_QUEUE_ARN=$(aws --output json sqs get-queue-attributes --queue-url "https://sqs.${AWS_REGION}.amazonaws.com/${AWS_ACCOUNT_ID}/${TARGET_QUEUE}" --attribute-names QueueArn | jq -r '.Attributes.QueueArn')
```

Execute the following command to create a Kubernetes manifest for a rule, forwarding events matching the specified rule
filter criteria to the target queue. The EventBridge filter pattern will match any event received on the custom event
bus with a `detail-type` of `event.from.ack.v0`. Alternatively, the filter pattern can be omitted to forward all events
from the custom event bus.

```bash
cat <<EOF > rule.yaml
apiVersion: eventbridge.services.k8s.aws/v1alpha1
kind: Rule
metadata:
  name: $RULE_NAME
spec:
  name: $RULE_NAME
  description: "ACK EventBridge Filter Rule to SQS using event bus reference"
  eventBusRef:
    from:
      name: $EVENTBUS_NAME
  eventPattern: |
    {
      "detail-type":["event.from.ack.v0"]
    }
  targets:
    - arn: $TARGET_QUEUE_ARN
      id: sqs-rule-target
      retryPolicy:
        maximumRetryAttempts: 0 # no retries
EOF

kubectl -n ${EVENTBRIDGE_NAMESPACE} create -f rule.yaml
```

The output of above commands looks like

```bash
rule.eventbridge.services.k8s.aws/custom-eventbus-ack-sqs-rule created
```

Verify the rule resource is synchronized.

```bash
kubectl -n ${EVENTBRIDGE_NAMESPACE} get rule ${RULE_NAME}
```

The output of above commands looks like

```bash
NAME                           SYNCED   AGE
custom-eventbus-ack-sqs-rule   True     18s
```

### Verify the event filtering and forwarding is working

Execute the following command to send an event to the custom bus matching the rule filter pattern.

```bash
cat <<EOF > event.json
[
    {
        "Source": "my.aws.events.cli",
        "DetailType": "event.from.ack.v0",
        "Detail": "{\"hello-world\":\"from ACK for EventBridge\"}",
        "EventBusName": "${EVENTBUS_NAME}"
    }
]
EOF

aws events put-events --entries file://event.json
```

The output of above commands looks like

```bash
{
    "FailedEntryCount": 0,
    "Entries": [
        {
            "EventId": "ccd21ee8-339d-cabe-520d-b847c98ba2cb"
        }
    ]
}
```

Verify the message was received by the SQS queue with

```bash
aws sqs receive-message --queue-url https://sqs.${AWS_REGION}.amazonaws.com/${AWS_ACCOUNT_ID}/${TARGET_QUEUE}
```

The output of above commands looks like

```bash
{
    "Messages": [
        {
            "MessageId": "80cef2f3-ff25-4441-9217-665bb0217ec5",
            <snip>
            "Body": "{\"version\":\"0\",\"id\":\"def3d99b-806b-5d92-d036-9e0884bdc387\",\"detail-type\":\"event.from.ack.v0\",\"source\":\"my.aws.events.cli\",\"account\":\"1234567890\",\"time\":\"2023-03-22T11:22:34Z\",\"region\":\"us-east-1\",\"resources\":[],\"detail\":{\"hello-world\":\"from ACK for EventBridge\"}}"
        }
    ]
}
```

## Next steps

The ACK service controller for Amazon EventBridge is based on the [Amazon EventBridge
API](https://docs.aws.amazon.com/eventbridge/latest/APIReference/Welcome.html).

Refer to [API Reference](https://aws-controllers-k8s.github.io/community/reference/) for *EventBridge* to find all the
supported Kubernetes custom resources and fields.

### Cleanup

Remove all the resource created in this tutorial using `kubectl delete` command.

```bash
kubectl -n ${EVENTBRIDGE_NAMESPACE} delete -f rule.yaml
kubectl -n ${EVENTBRIDGE_NAMESPACE} delete -f bus.yaml
kubectl delete ns ${EVENTBRIDGE_NAMESPACE}
```

The output of delete command should look like

```bash
rule.eventbridge.services.k8s.aws "custom-eventbus-ack-sqs-rule" deleted
eventbus.eventbridge.services.k8s.aws "custom-eventbus-ack" deleted
namespace "eventbridge-example" deleted
```

Remove the manually created SQS resource.

```bash
aws sqs delete-queue --queue-url https://sqs.${AWS_REGION}.amazonaws.com/${AWS_ACCOUNT_ID}/${TARGET_QUEUE}
```

If the command executes successfully, no output is generated.

To remove the EventBridge ACK service controller, related CRDs, and namespaces, see [ACK Cleanup][cleanup].

To delete your EKS clusters, see [Amazon EKS - Deleting a cluster][cleanup-eks].

[irsa-permissions]: ../../user-docs/irsa/
[cleanup]: ../../user-docs/cleanup/
[cleanup-eks]: https://docs.aws.amazon.com/eks/latest/userguide/delete-cluster.html
