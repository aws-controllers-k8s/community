---
title : "FAQ"
description: "Frequently asked questions"
lead: ""
draft: false
menu: 
  docs:
    parent: "discussion"
weight: 20
toc: true
---

## Service Broker

{{% hint type="success" title="Question" %}}
Does ACK replace the [service broker](https://svc-cat.io/)?
{{% /hint %}}

{{% hint type="answer" title="Answer" %}}
For the time being, people using the service broker should continue to use it and we're coordinating with the maintainers to provide a unified solution.

The service broker project is also an AWS activity that, with the general shift of focus in the community from service broker to operators, can be considered less actively developed. There are a certain things around application lifecycle management that the service broker currently covers and which are at this juncture not yet covered by the scope of ACK, however we expect in the mid to long run that these two projects converge. We had AWS-internal discussions with the team that maintains the service broker and we're on the same page concerning a unified solution.

We appreciate input and advice concerning features that are currently covered by the service broker only, for example bind/unbind or cataloging and looking forward to learn from the community how they are using service broker so that we can take this into account.
{{% /hint %}}

## Cluster API

{{% hint type="success" title="Question" %}}
Does the planned ACK service controller for EKS replace [Kubernetes Cluster API](https://github.com/kubernetes-sigs/cluster-api)?
{{% /hint %}}

{{% hint type="answer" title="Answer" %}}
No, the ACK service controller for EKS does not replace Kubernetes Cluster API.
Cluster API does a lot of really cool things and is designed to be a generic way to create Kubernetes clusters that run anywhere.
It makes some different design decisions with that goal in mind.
Some differences include:

- Cluster API is treated as your source of truth for all infrastructure.
This means things like the cluster autoscaler need to be configured to use cluster api instead of AWS cloud provider.
- Generic Kubernetes clusters rely on running more services in the cluster and not services from AWS.
Things like metrics and logging will likely need to run inside Kubernetes instead of using services like CloudWatch.
- IAM permission for Cluster-API Provider AWS (CAPA) need to be more broad than the ACK service controller for EKS because CAPA is responsible for provisioning everything needed for the cluster (VPC, gateway, etc).
You don't need to run all of the ACK controllers if all you want is a way to provision an EKS cluster.
You can pick and choose which ACK controllers you want to deploy.
- With the EKS ACK controller you will get all of the configuration flexibility of the EKS API including things like managed node groups and fargate.
This is because the ACK service controller for EKS is built directly from the EKS API spec and not abstracted to be a general Kubernetes cluster.
{{% /hint %}}

## cdk8s

{{% hint type="success" title="Question" %}}
How does ACK relate to [cdk8s](https://cdk8s.io/)?
{{% /hint %}}

{{% hint type="answer" title="Answer" %}}
cdk8s is an open-source software development framework for defining Kubernetes applications and reusable abstractions using familiar programming languages and rich object-oriented APIs.
You can use cdk8s to create any resource inside a Kubernetes cluster.
This includes [Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) (CRs).

All of the ACK controllers watch for specific CRs and you can generate those resources using cdk8s or any Kubernetes tooling.
The two projects complement each other.
cdk8s can create the Kubernetes resources and ACK uses those resources to create the AWS infrastructure.
{{% /hint %}}

## Troubleshooting

{{% hint type="success" title="Question" %}}
Why are my AWS resources sometimes not being fully deleted when trying to delete via `kubectl delete ... --cascade=foreground ...` (or via ArgoCD uninstalling my Helm chart)?
{{% /hint %}}

{{% hint type="answer" title="Answer" %}}
There is a [known issue with foreground cascading deletion](https://github.com/aws-controllers-k8s/community/issues/1759) in the ACK runtime that potentially impacts all controllers.

Until the above issue is resolved, you should use [background cascading deletion](https://kubernetes.io/docs/tasks/administer-cluster/use-cascading-deletion/#use-background-cascading-deletion) (the default behavior of `kubectl`) to delete resources.
{{% /hint %}}

{{% hint type="success" title="Question" %}}
Why am I seeing `Error: manifest does not contain minimum number of descriptors (2), descriptors found: 1` when trying to install the Helm chart?
{{% /hint %}}

{{% hint type="answer" title="Answer" %}}
[Helm 3.7](https://github.com/helm/helm/releases/tag/v3.7.0) included backward compatibility breaking changes to the manifest format of Helm charts stored in OCI chart repositories. Any images built using Helm <3.7 are not compatible with the latest version of the Helm CLI. This can be solved by using latest version of the chart. Use Helm version 3.7 or above with the latest version of the charts.
{{% /hint %}}

## Contributing

{{% hint type="success" title="Question" %}}
Where and how can I help?
{{% /hint %}}

{{% hint type="answer" title="Answer" %}}
Excellent question and we're super excited that you're interested in ACK.
For now, if you're a developer, you can check out the [contributor docs](../../contributor-docs/overview/).
{{% /hint %}}
