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


## Understanding ACK Controller Conditons


ACK controllers use conditions to indicate the state of custom resources and their corresponding AWS service resources. These conditions are exposed in the `Status.Conditions` collection of each custom resource.

### Condition Types

#### ACK.Adopted

Indicates that an adopted resource custom resource has been successfully reconciled and the target has been created.

* **True**: Resource has been successfully adopted
* **False**: Resource adoption failed
* **Unknown**: Resource adoption status cannot be determined

#### ACK.ResourceSynced

Indicates whether the state of the resource in the backend AWS service is in sync with the ACK service controller.

* **True**: Resource is fully synced
* **False**: Resource is out of sync
* **Unknown**: Sync status cannot be determined

#### ACK.Terminal

Indicates that the custom resource Spec needs to be updated before any further sync can occur.

* **True**: Resource is in terminal state
* **False**: Resource is not in terminal state
* **Unknown**: Terminal state cannot be determined

Possible Causes:
* Invalid arguments in input YAML
* Resource creation failed in AWS

#### ACK.Recoverable

Indicates errors that may be resolved without updating the custom resource spec.

* **True**: Error is recoverable
* **False**: Error is not recoverable
* **Unknown**: Recovery status cannot be determined

Possible Causes:
* Transient AWS service unavailability
* Access denied exceptions requiring credential updates

#### ACK.Advisory

Indicates advisory information present in the resource.

* **True**: Advisory condition exists
* **False**: No advisory condition
* **Unknown**: Advisory status cannot be determined

Possible Causes:
* Attempting to modify an immutable field after resource creation

#### ACK.LateInitialized

Indicates the status of late initialization of fields.

* **True**: Late initialization completed
* **False**: Late initialization in progress
* Not present: No late initialization needed

#### ACK.ReferencesResolved

Indicates whether all AWSResourceReference type references have been resolved.

* **True**: All references resolved
* **False**: Reference resolution failed
* **Unknown**: Resolution status cannot be determined
* Not present: No references to resolve

## Next Steps

Now that you have verified ACK service controller functionality, [checkout ACK
functionality for creating resources in multiple AWS regions.](../multi-region-resource-management)