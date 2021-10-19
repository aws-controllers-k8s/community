---
title: "Project Tenets"
description: "Our project tenets and design principles"
lead: ""
draft: false
menu:
  docs:
    parent: "contributor"
weight: 15
toc: true
---

We follow a set of tenets in building AWS Controllers for Kubernetes.

## Collaborate in the Open

When given a choice between keeping something hidden or making something open,
we default to open.

All of our [source code][source] is open.

Our development methodology is open.

Our [testing][testing], [release][release] and [documentation][docs] processes
are open.

Our [continuous integration system][ci] is open.

We are a community-driven project that strives to meet our users where they
are. Come join our [community meeting][comm-meet] on Zoom.

[source]: https://github.com/aws-controllers-k8s/
[testing]: https://github.com/aws-controllers-k8s/test-infra
[release]: ../../community/releases/
[docs]: https://github.com/aws-controllers-k8s/community/tree/main/docs
[ci]: https://prow.ack.aws.dev/
[comm-meet]: https://github.com/aws-controllers-k8s/community/#community-meeting

## Generate Everything

We choose to generate as much of our code as possible.

While we recognize that the differences and peculiarities of AWS service APIs
will naturally require some implementation code to be hand-written, we look for
patterns in AWS service APIs and enhance our code generator to handle these
patterns.

Generated code is easier to maintain and encourages consistency.

## Focus on Kubernetes

The ACK code generator produces controller implementations that include a
[common ACK runtime][rt]. This common runtime builds on top of the Kubernetes
upstream [controller-runtime][ctrl-rt] framework and provides a common
reconciliation loop that processes events receive from the Kubernetes API
server representing create, modify or delete operations for a custom resource.
By building ACK controllers with a common ACK runtime, we encourage consistent
behaviour in how controllers handle these custom resources.

We seek ways to make the **Kubernetes user experience** as simple and
consistent as possible for managing AWS resources. This means that the ACK code
generator enables service teams to rename fields for a resource, inject custom
code into a controller and instruct the controller implementation to handle
resources in ways that smooth over the inconsistencies between AWS service
APIs.

[rt]: https://github.com/aws-controllers-k8s/runtime
[ctrl-rt]: https://github.com/kubernetes-sigs/controller-runtime/

## Run Anywhere

ACK service controllers can be installed on any Kubernetes cluster, regardless
of whether a user chooses to use Amazon Elastic Kubernetes Service (EKS).

## Minimize Service Dependencies

The only AWS services that ACK service controllers depend on should be IAM/STS
and the specific AWS service that the controller integrates with.

We do not take a dependency on any stateful resource-tracking service,
including AWS CloudFormation, the AWS Cloud Control API, or Terraform.

ACK service controllers communicate directly with the AWS service API that the
controller is built for. The `s3-controller` speaks the S3 API. The
`ec2-controller` speaks the EC2 API.
