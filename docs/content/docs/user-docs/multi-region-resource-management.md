---
title: "Multi-Region Resource Management"
description: "Managing resources in multiple AWS regions"
lead: ""
draft: false
menu: 
  docs:
    parent: "installing"
weight: 50
toc: true
---

You can manage resources in multiple AWS regions using a single ACK service controller. To determine the AWS region for a given resource, the ACK service controller looks for region information in the following order:

  1. The region annotation `services.k8s.aws/region` on the resource.
  2. The region annotation `services.k8s.aws/region` on the resource's namespace.
  3. The `--aws-region` controller flag. This flag may be set using the `aws.region` Helm chart variable.
  4. Kubernetes pod [IRSA](https://aws-controllers-k8s.github.io/community/docs/user-docs/irsa/) environment variables.

For example, the `--aws-region` ACK service controller flag is `us-west-2`. If you want to create a resource in `us-east-1`, use one of the following options to override the default region.

## Option 1: Region annotation

Add the `services.k8s.aws/region` annotation while creating the resource. For example:

```yaml {linenos=table,hl_lines=["5-6"],linenostart=27}
apiVersion: s3.services.k8s.aws/v1alpha1
kind: Bucket
metadata:
  name: my-bucket
  annotations:
    services.k8s.aws/region: us-east-1
spec:
  name: my-bucket
  ...
```

## Option 2: Namespace default region annotation

To bind a region to a specific namespace, you will have to annotate the namespace with the `services.k8s.aws/default-region` annotation.

{{% hint type="info" title="Namespace-scoped deployment does not support this option" %}}
Use this solution for multi-region resource management on cluster-scoped deployments.
{{% /hint %}}

```yaml {linenos=table,hl_lines=["5-6"],linenostart=47}
apiVersion: v1
kind: Namespace
metadata:
  name: production
  annotations:
    services.k8s.aws/default-region: us-east-1
```

For existing namespaces, you can run:

```bash
kubectl annotate namespace production services.k8s.aws/default-region=us-east-1
```

You can also create the resource in the same namespace:

```yaml
apiVersion: s3.services.k8s.aws/v1alpha1
kind: Bucket
metadata:
  name: my-bucket
  namespace: production
spec:
  name: my-bucket
  ...
```