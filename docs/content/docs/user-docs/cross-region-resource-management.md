---
title: "Cross-Region Resource Management"
description: "Managing resources across different AWS regions"
lead: ""
draft: false
menu: 
  docs:
    parent: "installing"
weight: 50
toc: true
---

If you are using resources across different regions, you can override the default region of a given ACK service controller. ACK service controllers will first look for a region in the following order:

  1. The region annotation `services.k8s.aws/region` on the resource. If provided, this will override the namespace default region annotation.
  2. The namespace default region annotation `services.k8s.aws/default-region`.
  3. Controller flags, such as the `aws.region` variable in a given Helm chart
  4. Kubernetes pod [IRSA](https://aws-controllers-k8s.github.io/community/docs/user-docs/irsa/) environment variables

For example, the ACK service controller default region is `us-west-2`. If you want to create a resource in `us-east-1`, use one of the following options to override the default region

Option 1: Region annotation

Add the `services.k8s.aws/region` annotation while creating the resource. For example:

```bash
  apiVersion: sagemaker.services.k8s.aws/v1alpha1kind: TrainingJobmetadata:name: ack-sample-tainingjobannotations:services.k8s.aws/region: us-east-1spec:trainingJobName: ack-sample-tainingjobroleARN: <sagemaker_execution_role_arn>...
```

Option 2: Namespace default region annotation

To bind a region to a specific namespace, you will have to annotate the namespace with the `services.k8s.aws/default-region` annotation. For example:

```bash
apiVersion: v1kind: Namespacemetadata:name: productionannotations:services.k8s.aws/default-region: us-east-1
```

For existing namespaces, you can run:

```bash
kubectl annotate namespace production services.k8s.aws/default-region=us-east-1
```

You can also create the resource in the same namespace:

```bash
apiVersion: sagemaker.services.k8s.aws/v1alpha1kind: TrainingJobmetadata:name: ack-sample-trainingjobnamespace: productionspec:trainingJobName: ack-sample-trainingjobroleARN: <sagemaker_execution_role_arn>...
```