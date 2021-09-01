---
title : "FAQ"
description: "Frequently asked questions"
lead: ""
date: 2020-10-06T08:47:36+00:00
lastmod: 2020-10-06T08:47:36+00:00
draft: false
images: []
menu: 
  docs:
    parent: "Discussion"
weight: 20
toc: true
---

## Service Broker

!!! question "Question"
    Does ACK replace the [service broker](https://svc-cat.io/)?

!!! quote "Answer"
    For the time being, people using the service broker should continue to use it and we're coordinating with the maintainers to provide a unified solution.

    The service broker project is also an AWS activity that, with the general shift of focus in the community from service broker to operators, can be considered less actively developed. There are a certain things around application lifecycle management that the service broker currently covers and which are at this juncture not yet covered by the scope of ACK, however we expect in the mid to long run that these two projects converge. We had AWS-internal discussions with the team that maintains the service broker and we're on the same page concerning a unified solution.

    We appreciate input and advice concerning features that are currently covered by the service broker only, for example bind/unbind or cataloging and looking forward to learn from the community how they are using service broker so that we can take this into account.

## Cluster API

!!! question "Question"
    Does the planned ACK service controller for EKS replace [Kubernetes Cluster API](https://github.com/kubernetes-sigs/cluster-api)?

!!! quote "Answer"
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

## cdk8s

!!! question "Question"
    How does ACK relate to [cdk8s](https://cdk8s.io/)?

!!! quote "Answer"
    cdk8s is an open-source software development framework for defining Kubernetes applications and reusable abstractions using familiar programming languages and rich object-oriented APIs.
    You can use cdk8s to create any resource inside a Kubernetes cluster.
    This includes [Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) (CRs).
    
    All of the ACK controllers watch for specific CRs and you can generate those resources using cdk8s or any Kubernetes tooling.
    The two projects complement each other.
    cdk8s can create the Kubernetes resources and ACK uses those resources to create the AWS infrastructure.

## Contributing

!!! question "Question"
    Where and how can I help?

!!! quote "Answer"
    Excellent question and we're super excited that you're interested in ACK.
    For now, if you're a developer, you can check out the [contributor docs](../../dev-docs/overview/).