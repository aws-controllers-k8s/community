---
title: "Copy a resource field into a ConfigMap or Secret"
description: "Using the FieldExport custom resource"
lead: "Exporting a Spec or Status field using the FieldExport resource"
draft: false
menu:
  docs:
    parent: "getting-started"
weight: 65
toc: true
---

ACK controllers are intended to manage your AWS infrastructure using Kubernetes
custom resources. Their responsibilities end after managing the lifecycle of
your AWS resource and do not extend into binding to applications running in the
Kubernetes data plane. The ACK `FieldExport` custom resource was designed to
bridge the gap between managing the control plane of your ACK resources and
using the properties of those resources in your application.

The `FieldExport` resource configures an ACK controller to export any spec or
status field from an ACK resource into a Kubernetes `ConfigMap` or `Secret`.
These fields are automatically updated when any field value changes. You are
then able to mount the `ConfigMap` or `Secret` onto your Kubernetes Pods as
environment variables that can ingest those values. 

`FieldExport` is included by default in every ACK controller installation and
can be used to reference any field within the `Spec` or `Status` of any ACK
resource.

## Using a FieldExport

The `Spec` and `Status` fields of the `FieldExport` custom resource definition
are available in the [API reference][spec-reference]. For this example, we will
be creating an [S3 Bucket][bucket-spec] and exporting the `Status.Location`
field into a `ConfigMap`.

```yaml
apiVersion: s3.services.k8s.aws/v1alpha1
kind: Bucket
metadata:
  name: application-user-data
spec:
  name: doc-example-bucket
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: application-user-data-cm
data: {}
---
apiVersion: services.k8s.aws/v1alpha1
kind: FieldExport
metadata:
  name: export-user-data-bucket
spec:  
  to:
    name: application-user-data-cm # Matches the ConfigMap we created above
    kind: configmap
  from:
    path: ".status.location"
    resource:
      group: s3.services.k8s.aws
      kind: Bucket
      name: application-user-data
```

Applying this manifest to the cluster will:
1. Create a new S3 bucket called `doc-example-bucket`
1. Create a `ConfigMap` called `application-user-data-cm`
1. Create a `FieldExport` called `export-user-data-bucket` that will export the
   `.status.location` path from the bucket into the ConfigMap

After the reconciler has created the bucket, the `application-user-data-cm`
`ConfigMap` looks like the following:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: application-user-data-cm
  namespace: default
data:
  default.export-user-data-bucket: http://doc-example-bucket.s3.amazonaws.com/
```

The `ConfigMap` data contains a new key-value pair. The key is the namespace and
name of the `FieldExport` that created it, and the value is the resolved value
from the resourc. This value can then be included as an environment variable in
a pod like so:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-application
spec:
  containers:
  - name: field-export-demo-container
    image: k8s.gcr.io/busybox
    command: [ "/bin/sh", "-c", "env" ]
    env:
    - name: USER_DATA_BUCKET_LOCATION
      valueFrom:
        configMapKeyRef:
          name: application-user-data-cm # The ConfigMap that we created earlier
          key: "default.export-user-data-bucket"
```

Looking at the container logs, you can see the `USER_DATA_BUCKET_LOCATION`
environment is set with the value from the `ConfigMap`:
```bash
USER_DATA_BUCKET_LOCATION=http://doc-example-bucket.s3.amazonaws.com/
```

{{% hint type="warning" title="`FieldExport` RBAC permissions" %}}
The ACK controller will fetch the source path from the ACK resource assuming
its service account has the RBAC permissions to read that type of resource. If a
user has the privileges to create a `FieldExport` resource, it is possible that
they can create one which fetches fields from a resource they do not have RBAC
permissions to read directly. This could potentially expose that resource's
properties to the unprivileged user.

To mitigate this problem, the ACK controller will only export fields from
resources that exist in the same namespace as the `FieldExport` resource
requesting it. {{% /hint %}}

[spec-reference]: ../../../reference/common/v1alpha1/fieldexport/
[bucket-spec]: ../../../reference/s3/v1alpha1/fieldexport/bucket/#spec