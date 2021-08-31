---
title : "AWS Controllers for Kubernetes"
description: "AWS Controllers for Kubernetes (ACK) lets you define and use AWS service resources directly from Kubernetes"
lead: ""
date: 2020-10-06T08:47:36+00:00
lastmod: 2020-10-06T08:47:36+00:00
draft: false
images: []
---

**AWS Controllers for Kubernetes (ACK)** lets you define and use AWS service
resources directly from Kubernetes. With ACK, you can take advantage of AWS
managed services for your Kubernetes applications without needing to define
resources outside of the cluster or run services that provide supporting
capabilities like databases or message queues within the cluster.

This is a fully open source project built with ❤️  by AWS. The project is
structured as a set of source code repositories containing a
[common runtime][rt], a [code generator][code-gen], some
[common testing infrastructure][test-infra] and a series of individual service
controllers that correspond to AWS services (e.g. the
[RDS service controller for ACK][rds-controller]).

[rt]: https://github.com/aws-controllers-k8s/runtime
[code-gen]: https://github.com/aws-controllers-k8s/code-generator
[test-infra]: https://github.com/aws-controllers-k8s/test-infra
[rds-controller]: https://github.com/aws-controllers-k8s/rds-controller

Individual ACK service controllers are installable as separate binaries either
manually via raw manifests or via a Helm chart that corresponds to the
individual service controller.

!!! note **IMPORTANT**
    Individual ACK service controllers may be in different
    maintenance phases and follow separate release cadences. Please read our
    documentation about [project stages][proj-stages] and
    [maintenance phases][maint-phases] fully, including how we
    [release and version][rel-ver] our controllers. Controllers in a `PREVIEW`
    maintenance phase have at least one container image and Helm chart released to
    an ECR Public repository. However, be aware that controllers in a `PREVIEW`
    maintenance phase may have significant and breaking changes introduced in a
    future release.

[proj-stages]: https://aws-controllers-k8s.github.io/community/releases/#project-stages
[maint-phases]: https://aws-controllers-k8s.github.io/community/releases/#maintenance-phases
[rel-ver]: https://aws-controllers-k8s.github.io/community/releases/#releases-and-versioning

## Background

Kubernetes applications often require a number of supporting resources like
databases, message queues, and object stores to operate. AWS provides a set of
managed services that you can use to provide these resources for your
applications, but provisioning and integrating them with Kubernetes was complex
and time consuming. ACK lets you define and consume many AWS services and
resources directly within a Kubernetes cluster. ACK gives you a unified,
operationally seamless way to manage your application and its dependencies.

## Connecting Kubernetes and AWS APIs

![A bird's eye view of ACK](images/ack-birdseye-view.png)

[ACK][gh] is a collection of [Kubernetes Custom Resource Definitions][crd]
(CRDs) and controllers which work together to extend the Kubernetes API and
create AWS resources on your cluster’s behalf.

ACK comprises a set of Kubernetes custom [controllers][controller]. Each
controller manages [custom resources][crd] representing API resources of a
single AWS service API. For example, the ACK service controller for AWS Simple
Storage Service (S3) manages custom resources that represent AWS S3 buckets.

Instead of logging into the AWS console or using the `aws` CLI tool to interact
with the AWS service API, Kubernetes users can install a controller for an AWS
service and then create, update, read and delete AWS resources using the
Kubernetes API.

This means they can use the Kubernetes API and configuration language to fully
describe both their containerized applications, using Kubernetes resources like
`Deployment` and `Service`, as well as any AWS service resources upon which
those applications depend.

Read more about [how ACK works][how-it-works].

[gh]: https://github.com/aws-controllers-k8s/community
[controller]: https://kubernetes.io/docs/reference/glossary/?fundamental=true#term-controller
[crd]: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/
[how-it-works]: https://aws-controllers-k8s.github.io/community/how-it-works/

## Getting started

Please see the list of ACK [service controllers][services] currently in one of
our [project stages][proj-stages].

You can [install][install] any of the controllers in the `RELEASED` project stage using
Helm (recommended) or manually using the raw Kubernetes manifests contained in
the individual ACK service controller's source repository.

[services]: https://aws-controllers-k8s.github.io/community/services/
[install]: https://aws-controllers-k8s.github.io/community/user-docs/install/

Once installed, Kubernetes users may apply a custom resource (CR) corresponding
to one of the resources exposed by the ACK service controller for the service.

To view the list of custom resources and each CR's schema, visit our
[reference documentation][ref-docs].

[ref-docs]: https://aws-controllers-k8s.github.io/community/reference/overview/

## Getting help

For help, please consider the following venues (in order):

* [Search open issues](https://github.com/aws/aws-controllers-k8s/issues)
* [File an issue](https://github.com/aws/aws-controllers-k8s/issues/new/choose)
* Chat with us on the `#provider-aws` channel in the [Kubernetes Slack](https://kubernetes.slack.com/) community.
