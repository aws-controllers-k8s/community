---
title: "Create an ACK Resource"
description: "Create, Update and Delete an S3 bucket"
lead: "Create, Update and Delete an AWS Resource using ACK"
draft: false
menu:
  docs:
    parent: "getting-started"
weight: 30
toc: true
---

{{% hint type="info" title="Note" %}}
While this guide provides examples for managing S3 bucket, you can find sample
manifest files for other AWS services in `test/e2e/resources` directory of
corresponding service controller's GitHub repository. For example: Sample manifest
for ecr repository can be found [here](https://github.com/aws-controllers-k8s/ecr-controller/tree/main/test/e2e/resources)

You can find API Reference for all the services supported by ACK [here](../../../reference)
{{% /hint %}}

## Create an S3 bucket

```bash
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
export BUCKET_NAME=my-ack-s3-bucket-$AWS_ACCOUNT_ID

read -r -d '' BUCKET_MANIFEST <<EOF
apiVersion: s3.services.k8s.aws/v1alpha1
kind: Bucket
metadata:
  name: $BUCKET_NAME
spec:
  name: $BUCKET_NAME
EOF

echo "${BUCKET_MANIFEST}" > bucket.yaml

kubectl create -f bucket.yaml

kubectl describe bucket/$BUCKET_NAME
```

## Update the S3 bucket

```bash
read -r -d '' BUCKET_MANIFEST <<EOF
apiVersion: s3.services.k8s.aws/v1alpha1
kind: Bucket
metadata:
  name: $BUCKET_NAME
spec:
  name: $BUCKET_NAME
  tagging:
    tagSet:
    - key: myTagKey
      value: myTagValue
EOF

echo "${BUCKET_MANIFEST}" > bucket.yaml

kubectl apply -f bucket.yaml

kubectl describe bucket/$BUCKET_NAME
```

## Delete the S3 bucket

```bash
kubectl delete -f bucket.yaml

# verify the bucket no longer exists
kubectl get bucket/$BUCKET_NAME
```

## Next Steps

Now that you have verified ACK service controller functionality, [checkout ACK
functionality for creating resources in multiple AWS regions.](../multi-region-resource-management)
