---
title: "Contribution Overview"
description: "Context on the contributor documentation"
lead: ""
draft: false
menu: 
  docs:
    parent: "contributor"
weight: 10
toc: true
---

This section of the docs is for contributors to the AWS Controllers for
Kubernetes (ACK) project.

If you're interested in enhancing our platform, developing on a specific
service controller or just curious how ACK is architected, you've come to the
right place.

## Project Tenets (unless you know better ones)

We follow a set of tenets in building AWS Controllers for Kubernetes.

1. **Collaborate in the Open**: Our source code is open. Our development
   methodology is open. Our testing, release and documentation processes are
   open. We are a community-driven project that strives to meet our users where
   they are.

2. **Generate Everything**: We choose to generate as much of our code as
   possible. Generated code is easier to maintain and encourages consistency.

3. **Focus on Kubernetes**: We seek ways to make the Kubernetes user experience
   as simple and consistent as possible for managing AWS resources.

4. **Run Anywhere**: ACK service controllers can be installed on any Kubernetes
   cluster, regardless of whether a user chooses to use Amazon Elastic
   Kubernetes Service (EKS).

5. **Minimize Service Dependencies**: The only AWS services that ACK service
   controllers depend on should be IAM/STS and the specific AWS service that
   the controller integrates with. We do not take a dependency on any stateful
   resource-tracking service.

Read more about our [project tenets and design principles][tenets].

[tenets]: ../tenets/

## Code Organization

ACK is a collection of source repositories containing a common runtime and type
system, a code generator and individual service controllers that manage
resources in a specific AWS API.

Learn more about how our [source code repositories are organized][code-org].

[code-org]: ../code-organization/

## API Inference

Read about [how the code generator infers][api-inference] information about a
Kubernetes Custom Resource Definitions (CRDs) from an AWS API model file.

[api-inference]: ../api-inference/

## Code Generation

The [code generation](../code-generation/) section gives you a bit of background
on how we go about automating the code generation for controllers and supporting
artifacts.

## Setting up a Development Environemnt

In the [setup](../setup/) section we walk you through setting up your local Git
environment with the repo and how advise you on how we handle contributions.

## Building an ACK Service Controller

After getting your development environment established, you will want to learn
[how to build an ACK service controller](../building-controller).

## Testing an ACK Service Controller

Last but not least, in the [testing](../testing/) section we show you how to
test ACK locally.
