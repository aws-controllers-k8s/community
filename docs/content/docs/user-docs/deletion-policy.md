---
title: "Retain AWS Resources after Deletion"
description: "Using the ACK deletion policy configuration"
lead: "Using the ACK deletion policy configuration"
draft: false
menu:
  docs:
    parent: "getting-started"
weight: 70
toc: true
---

The ACK controllers are designed to create, update and delete AWS resources
following the lifecycle of their respective Kubernetes custom resources. As a
result, when deleting an ACK resource, the underlying AWS resource is first
deleted before deleting its Kubernetes custom resource. This behavior is
expected so that users can delete AWS resources using the same Kubernetes APIs
as they used to create them.

There are some cases where a user wants to leave the underlying AWS resource
intact, but still delete the resource from Kubernetes. For example, migrating
stateful data resources (like S3 buckets or RDS database instances) between
Kubernetes installations or removing a resource from the control of an ACK
controller without deleting the resource altogether.

All ACK controllers support "deletion policy" configuration, which lets the
controller know which resources should be deleted from AWS (or left untouched)
before deleting their K8s resources. This configuration can be defined at any of
the following levels (with increasing order of precedence):
- Within the controller command-line using the `--deletion-policy` argument
- Within a `Namespace` annotation as
  `{service}.services.k8s.aws/deletion-policy`
- Within an ACK resource annotation as `services.k8s.aws/deletion-policy`

Each of these configuration options supports the following values:
- `delete` - **(Default)** Deletes the resource from AWS before deleting it from
  K8s
- `retain` - Keeps the AWS resource intact before deleting it from K8s

## Configuring the deletion policy

### Using Helm values

To set a controller-wide deletion policy, which will apply to all ACK resources
owned by the ACK controller, you can set the `deletionPolicy` Helm chart value.

For example, to retain all AWS resources when installing the Helm chart through
the Helm CLI: `helm install ... --set=deletionPolicy=retain`

### For all resources within a Namespace

To set the deletion policy for all resources within a namespace (only for a
single service), you can add an annotation to the `Namespace` object itself.

For example, to set all S3 resources to be retained within the namespace:
```yaml
apiVersion: v1
kind: Namespace
metadata:
 annotations:
   s3.services.k8s.aws/deletion-policy: retain
 name: retain-s3-namespace
```

### For a single ACK resource

If you want to just retain a single specific resource, you can override the
default behavior by setting an annotation directly onto the resource.

For example, to retain a specific S3 bucket:
```yaml
apiVersion: s3.services.k8s.aws/v1alpha1
kind: Bucket
metadata:
  name: retained-bucket
  annotations:
    services.k8s.aws/deletion-policy: retain
spec:
  name: my-retained-bucket
```

*Note: The key for annotating a single resource is not the same as the key when
annotating a namespace. You do not need to provide the name of the service as a
suffix for a single resource.*