---
title: "Adopting Existing AWS Resources"
description: "Adopting Existing AWS Resources"
lead: "Adopting existing AWS resources that were created using other tools"
draft: false
menu:
  docs:
    parent: "getting-started"
weight: 66
toc: true
---

The ACK controllers are intended to manage the complete lifecycle of an AWS
resource, from creation through deletion. However, you may already be
managing those resources using other tools - such as CloudFormation or
Terraform. Migrating to ACK could be time-consuming to redeclare all resources
as YAML, or even cause you to lose the state of the application if parts of the
system are recreated. The ACK `AdoptedResource` custom resource was designed to
help migrate these AWS resources to be managed by an ACK controller in your Kubernetes
cluster without having to define the full YAML spec or needing to delete and
re-create the resource.

To adopt an AWS resource, create an `AdoptedResource` custom
resource that specifies the unique identifier for the AWS resource and a target
K8s object. After applying this custom resource to the cluster, the ACK
controller will describe the AWS resource and create the associated ACK resource
inside the cluster - with a complete spec and status. The ACK controller will
then treat the newly-created ACK resource like any other.

All ACK controllers ship with the same `AdoptedResource` CRD. Every controller
contains the logic for adopting resources from its particular service. That is,
the S3 controller understands how to adopt all S3 resources. If you don't have a
particular service controller installed, and try to adopt a resource from that
service, the `AdoptedResource` CR will have no effect.

## Spec

The full spec for the `AdoptedResource` CRD is located [in the API
reference][api-ref]. The spec contains two parts: the AWS resource reference and
the Kubernetes target.

The AWS resource reference requires the unique identifier for the object, either
as an ARN or as the name or ID of the object. Which of these is required depends
on the service and the particular resource. You can find which field is required
by finding the unique identifier field used by the `Describe*` or `List*` API
calls for that resource.

The Kubernetes target requires the `group` and `kind` - these identify from
which service and resource you wish to adopt. For example, to adopt an S3
bucket, you would specify a `group` of `s3.services.k8s.aws` and a `kind` of
`Bucket`. The Kubernetes target also allows you to override the `metadata` for
the object that is created. By default, all resources created through an
`AdoptedResource` will have the same `metadata.name` as the `AdoptedResource`
that created it. 

[api-ref]: https://aws-controllers-k8s.github.io/community/reference/common/v1alpha1/adoptedresource/

### Example

Below is an example of adopting an S3 bucket, overriding the name and namespace
of the target K8s object and adding a label.

```yaml
apiVersion: services.k8s.aws/v1alpha1
kind: AdoptedResource
metadata:
  name: adopt-my-existing-bucket
spec:  
  aws:
    nameOrID: example-bucket
  kubernetes:
    group: s3.services.k8s.aws
    kind: Bucket
    metadata:
      name: my-existing-bucket
      namespace: default
      labels:
        app: foo
```

### Additional Keys

Some AWS resources cannot be defined using only a single unique identifier. For
APIs where we need to provide multiple identifiers, the `AdoptedResource` spec
contains a field called `aws.additionalKeys` which allows for any number of
arbitrary key-value pairs required to define the multiple identifier keys. When
adopting a resource with multiple identifiers, provide the *most specific*
identifier in the `nameOrID` field. Then for each additional identifier, set the
name of the key in `additionalKeys` to be the name in the ACK spec or status for
that field, and the value to be the actual identifier value.

For example, the [`Integration`][apigw-integration] resource in AWS API Gateway
V2 is uniquely identified by both an `integrationID` and an `apiID`. The API
requires both of these features to [describe an integration][integ-describe]. In
this case, we would provide the `integrationID` for the `nameOrID` field - since
it is unique for every API Gateway v2 `API` object. Then to provide the `apiID`,
we add a key of `apiID` in the `additionalKeys` and the value as the API ID for
the resource we want to adopt. The full spec of the `AdoptedResource` would look
like the following:

```yaml
apiVersion: services.k8s.aws/v1alpha1
kind: AdoptedResource
metadata:
  name: adopt-my-existing-integration
spec:  
  aws:
    nameOrID: integration-id-123456789012
    additionalKeys:
      apiID: api-id-123456789012
  kubernetes:
    group: apigatewayv2.services.k8s.aws
    kind: Integration
```

[apigw-integration]:
    https://aws-controllers-k8s.github.io/community/reference/apigatewayv2/v1alpha1/integration/#spec
[integ-describe]: https://docs.aws.amazon.com/sdk-for-go/api/service/apigatewayv2/#GetIntegrationInput