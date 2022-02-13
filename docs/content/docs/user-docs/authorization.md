---
title: "Permissions Overview"
description: "Configuring RBAC and IAM for ACK"
lead: "Overview of RBAC and IAM for authorization and access"
draft: false
menu:
  docs:
    parent: "getting-started"
weight: 60
toc: true
---

There are two different Role-Based Access Control (RBAC) systems needed for ACK service controller authorization: Kubernetes RBAC and AWS IAM.

{{% hint type="info" title="Note" %}}
This guide is only informative. You do not need to execute any commands from this page.

Kubernetes RBAC permissions below are already handled when you install ACK service
controller using [Helm chart or static manifests](../install)

AWS IAM permissions are handled using the IAM role created during [IRSA setup](../irsa)

{{% /hint %}}

[Kubernetes RBAC][k8s-rbac] governs a Kubernetes user's ability to read or write Kubernetes resources, while [AWS Identity and Access Management][aws-iam] (IAM) policies govern the ability of an AWS IAM role to read or write AWS resources.

[k8s-rbac]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
[aws-iam]: https://docs.aws.amazon.com/IAM/latest/UserGuide/access.html

{{% hint type="info" title="These two RBAC systems to not overlap" %}}
The Kubernetes user that makes a Kubernetes API call with `kubectl` has no
association with an IAM role. Instead, the IAM role is associated with the
[service account](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)
that runs the ACK service controller's pod.
{{% /hint %}}

Refer to the following diagram for more details on running a Kubernetes API server with RBAC authorization mode enabled.

![Authorization in ACK](../images/authorization.png)

You will need to configure Kubernetes RBAC and AWS IAM permissions before using ACK service controllers.

## Kubernetes RBAC

### Permissions required for the ACK service controller

ACK service controllers may be started in either *Cluster Mode* or *Namespace
Mode*. Controllers started in Cluster Mode watch for new, updated and deleted
custom resources (CRs) in all Kubernetes `Namespaces`. Conversely, controllers
started in Namespace Mode only watch for CRs in a single Kubernetes `Namespace`
identified by the `--watch-namespace` flag.

#### Namespace Mode

When a service controller is started in Namespace Mode, the `--watch-namespace`
flag is supplied and the controller will *only* watch for custom resources
(CRs) in that Kubernetes Namespace.

Controllers started in Namespace Mode require that the Kubernetes `Service
Account` associated with the controller's `Deployment` have a `Role` with
permissions to create, update/patch, delete, read, list and watch ACK custom
resources matching the associated AWS service in the specific Kubernetes
`Namespace` identified by the `--watch-namespace` flag.

{{% hint type="info" title="The `installScope: namespace` Helm Chart value" %}}
If you are installing an ACK service controller via the associated Helm Chart,
you can simplify a Namespace Mode installation by setting the `installScope`
value to `namespace`. This will cause the Helm Chart to install a
namespace-scoped `RoleBinding` with the necessary permissions the controller
needs to create, update, read, list and watch the ACK custom resources managed
by the controller.
{{% /hint %}}

### Cluster Mode

When a service controller is started in Cluster Mode, the `--watch-namespace`
flag is not supplied and the controller will watch for ACK custom resources
(CRs) across *all* Kubernetes `Namespaces`.

Controllers started in Cluster Mode require that the Kubernetes `Service
Account` associated with the controller's `Deployment` have a `ClusterRole`
with permissions to create, update/patch, delete, read, list and watch ACK
custom resources matching the associated AWS service in *all* Kubernetes
`Namespaces`.

To support cross-account resource management, controllers started in Cluster
Mode require that the Kubernetes `Service Account` associated with the
controller's `Deployment` have a `ClusterRole` with permissions to read, list
and watch *all* `Namespace` objects.

Additionally, the `ClusterRole` will need permissions to read `ConfigMap`
resources in a specific Kubernetes `Namespace` identified by the environment
variable `ACK_SYSTEM_NAMESPACE`, defaulting to `ack-system`.

{{% hint type="info" title="Cross-account resource management requires Cluster Mode" %}}
If you plan to use an ACK service controller to manage resources across many
AWS accounts (cross-account resource management, or CARM), you *must* start the
controller in Cluster Mode.
{{% /hint %}}

### Permission to read `Secret` objects

Some ACK service controllers will replace plain-text values for some resource
fields with the value of Kubernetes `Secret` keys.

For controllers started in Namespace Mode, the `Role` must have permissions to
read `Secret` objects in the Kubernetes `Namespace` identified by the
`--watch-namespace` flag.

For controllers started in Cluster Mode, the `ClusterRole` must have
permissions to read `Secret` resources in *any Kubernetes `Namespace` within
which ACK custom resources may be launched*.

### Roles for reading and writing ACK custom resources

As part of installation, Kubernetes `Role` resources are automatically created. These roles contain permissions to modify the Kubernetes custom resources (CRs) that the ACK service controller is responsible for.

{{% hint type="info" title="ACK resources are namespace-scoped" %}}
All Kubernetes CRs managed by an ACK service controller are namespace-scoped resources. There are no cluster-scoped ACK-managed CRs.
{{% /hint %}}

By default, the following Kubernetes `Role` resources are created when installing an ACK service controller:

* `ack-$SERVICE-writer`: a `Role` used for reading and mutating namespace-scoped CRs that the ACK service controller manages.
* `ack-$SERVICE-reader`: a `Role` used for reading namespaced-scoped CRs that the ACK service controller manages.

For example, installing the ACK service controller for AWS S3 creates the `ack-s3-writer` and `ack-s3-reader` roles, both with a `GroupKind` of `s3.services.k8s.aws/Bucket` within a specific Kubernetes `Namespace`.

### Bind a Kubernetes user to a Kubernetes role

Once the Kubernetes `Role` resources have been created, you can assign a specific Kubernetes `User` to a particular `Role` with the `kubectl create rolebinding` command.

```bash
kubectl create rolebinding alice-ack-s3-writer --role ack-s3-writer --namespace testing --user alice
kubectl create rolebinding alice-ack-sns-reader --role ack-sns-reader --namespace production --user alice
```

You can check the permissions of a particular Kubernetes `User` with the `kubectl auth can-i` command.
```
kubectl auth can-i create buckets --namespace default
```

## AWS IAM permissions for ACK controller

The IAM role needs the correct [IAM policies][aws-iam] for a given ACK service controller. For example, the ACK service controller for AWS S3 needs read and write permission for S3 Buckets. It is recommended that the IAM policy gives only enough access to properly manage the resources needed for a specific AWS service.

To use the recommended IAM policy for a given ACK service controller, refer to the `recommended-policy-arn` file in the `config/iam/` folder within that service's public repository. This document contains the AWS Resource Name (ARN) of the recommended managed policy for a specific service. For example, the [recommended IAM policy ARN for AWS S3][s3-recommended-arn] is: `arn:aws:iam::aws:policy/AmazonS3FullAccess`.

[s3-recommended-arn]: https://github.com/aws-controllers-k8s/s3-controller/tree/main/config/iam

Some services may need an additional inline policy. For example, the service controller may require `iam:PassRole` permission in order to pass an execution role that will be assumed by the AWS service. If applicable, resources for additional recommended policies will be located in the `recommended-inline-policy` file within the `config/iam` folder of a given ACK service controller's public repository. This inline policy is applied along with the managed policies when creating the role.

If you have not yet created an IAM role, see the user documentation on how to [create an IAM role for your ACK service controller][irsa-docs].

[irsa-docs]: ../irsa/#create-an-iam-role-for-your-ack-service-controller
