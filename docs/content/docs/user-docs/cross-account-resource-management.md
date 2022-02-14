---
title: "Manage Resources In Multiple AWS Accounts"
description: "Managing resources in different AWS accounts"
lead: ""
draft: false
menu:
  docs:
    parent: "getting-started"
weight: 50
toc: true
---

ACK service controllers can manage resources in different AWS accounts. To enable and start using this feature, as an administrator, you will need to:

  1. Configure the AWS accounts where the resources will be managed
  2. Map AWS accounts with the Role ARNs that need to be assumed
  3. Annotate namespaces with AWS Account IDs

For detailed information about how ACK service controllers manage resources in multiple AWS accounts, please refer to the Cross-Account Resource Management (CARM) [design document](https://github.com/aws-controllers-k8s/community/blob/main/docs/design/proposals/carm/cross-account-resource-management.md).

{{% hint type="note" title="To use CARM, `--watch-namespace` must be empty" %}}
ACK service controllers may be started in either Cluster Mode or Namespace Mode. When a service controller is started in Namespace Mode, the `--watch-namespace` flag is supplied and the controller will *only* watch for custom resources (CRs) in that Kubernetes Namespace. Because the cross-account resource management feature requires the controller to watch for custom resources on many Kubernetes Namespaces, this feature is incompatible with the Namespace Mode of running a controller and thus the `--watch-namespace` flag must not be set (or be set to an empty string).
{{% /hint %}}

## Step 1: Configure your AWS accounts

AWS account administrators should create and configure IAM roles to allow ACK service controllers to assume roles in different AWS accounts.

To allow account A (000000000000) to create AWS S3 buckets in account B (111111111111), you can use the following commands:
```bash
# Using account B credentials
aws iam create-role --role-name s3FullAccess \
  --assume-role-policy-document '{"Version": "2012-10-17","Statement": [{ "Effect": "Allow", "Principal": {"AWS": "arn:aws:iam::000000000000:role/roleA-production"}, "Action": "sts:AssumeRole"}]}'
aws iam attach-role-policy --role-name s3FullAccess \
  --policy-arn 'arn:aws:iam::aws:policy/service-role/AmazonS3FullAccess'
```

## Step 2: Map AWS accounts to their associated role ARNs

Create a `ConfigMap` to associate each AWS Account ID with the role ARN that needs to be assumed in order to manage resources in that particular account.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: ack-role-account-map
  namespace: $ACK_SYSTEM_NAMESPACE
data:
  "111111111111": arn:aws:iam::111111111111:role/s3FullAccess
EOF
```

## Step 3: Bind accounts to namespaces

To bind AWS accounts to a specific namespace you will have to annotate the namespace with an AWS account ID. For example:
```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: production
  annotations:
    services.k8s.aws/owner-account-id: "111111111111"
EOF
```

For existing namespaces, you can run:
```bash
kubectl annotate namespace production services.k8s.aws/owner-account-id=111111111111
```

### Create resources in different AWS accounts

Next, create your custom resources (CRs) in the associated namespace.

For example, to create an S3 bucket in account B, run the following command:
```bash
cat <<EOF | kubectl apply -f -
apiVersion: s3.services.k8s.aws/v1alpha1
kind: Bucket
metadata:
  name: my-bucket
  namespace: production
spec:
  name: my-bucket
EOF
```

## Next Steps
Checkout the [RBAC and IAM permissions overview](../authorization) to understand how ACK manages authorization
