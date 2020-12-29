# API Inference

This document discusses how ACK introspects an AWS API model file and
determines which `CustomResourceDefinition`s (CRDs) to construct and what the
structure of those CRDs look like.

## The Kubernetes Resource Model

The [Kubernetes Resource Model][krm] (KRM) is a set of [standards][api-stds]
and naming conventions that govern how an [`Object`][object] may be created and
updated.

[krm]: https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/
[api-stds]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
[object]: https://kubernetes.io/docs/reference/glossary/?all=true#term-object

An `Object` includes some metadata about the object -- a
[`GroupVersionKind`][gvk] (GVK), a `Name`, a `Namespace`, and zero or more `Labels`
and `Annotations`.

[gvk]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#resources

In addition to this metadata, each `Object` has a `Spec` field which is a
struct that contains the **desired** state of the `Object`. `Objects` are
typically denoted using YAML, like so:

```yaml
apiVersion: s3.services.k8s.aws/v1alpha1
kind: Bucket
metadata:
  name: my-amazing-bucket
  annotations:
    pronounced-as: boo-kay
spec:
  name: my-amazing-bucket
```

!!! note "Manifests"
    The YAML files containing an object definition like above are typically
    called [**manifests**][manifest].

[manifest]: https://kubernetes.io/docs/reference/glossary/?fundamental=true#term-manifest

Above, the `Object` has a GVK of "s3.services.k8s.aws/v1alpha1:Bucket" with an
**internal-to-Kubernetes** `Name` of "my-amazing-bucket" and a single
`Annotation` key/value pair "pronounced-as: boo-kay".

The `Spec` field is a structure containing desired state fields about this
Bucket. You can see here that there is a `Spec.Name` field representing the
Bucket name that will be passed to the S3 CreateBucket API as the name of the
Bucket. Note that the `Metadata.Name` field value is the same as the
`Spec.Name` field value here, but there's nothing mandatory about this.

When a Kubernetes user creates an `Object`, typically by passing some YAML to
the `kubectl create` or `kubectl apply` CLI command, the Kubernetes API server
reads the manifest and determines whether the supplied contents are valid.

In order to determine if a manifest is valid, the Kubernetes API server must
look up the **definition** of the specified `GroupVersionKind`. For all of the
resources that ACK is concerned about, what this means is that the Kubernetes
API server will search for the [`CustomResourceDefinition`][crd] (CRD) matching
the `GroupVersionKind`.

[crd]: https://kubernetes.io/docs/reference/glossary/?fundamental=true#term-CustomResourceDefinition

This CRD describes the fields that comprise `Object`s of that particular
`GroupVersionKind` -- called `CustomResources` (CRs).

In the next sections we discuss:

* how ACK determines what will become a CRD
* how ACK determines the fields that go into each CRD's `Spec` and `Status`

## Which things become ACK Resources?

As mentioned in the [code generation documentation][codegen], ACK reads AWS API
model files when generating its API types and controller implementations. These
model files are JSON files contain some important information about the
structure of the AWS service API, including a set of *Operation* definitions
(commonly called "Actions" in the official AWS API documentation) and a set of
*Shape* definitions.

[codegen]: https://aws.github.io/aws-controllers-k8s/dev-docs/code-generation/

Some AWS APIs have dozens (hundreds even!) of Operations exposed by the API.
Consider EC2's API. It has over **400 separate Actions**. Out of all those
Operations, how are we to tell which ones refer to something that we can model
as a Kubernetes `CustomResource`?

Well, we could look at the EC2 API's list of Operations and manually decide
which ones seem "resource-y". Operations like "AdvertiseByoipCidr" and
"AcceptTransitGatewayVpcAttachment" don't seem very "resource-y". Operations
like "CreateKeyPair" and "DeleteKeyPair", however, do seem like they would
match a resource called "KeyPair".

And this is actually how ACK decides what is a `CustomResource` and what isn't.

It uses a simple heuristic: *look through the list of Operations in the API
model file and filter out the ones that start with the string "Create". If what
comes after the word "Create" describes a singular noun, then we create a
`CustomResource` of that `Kind`*.

It really is that simple.

## How is an ACK Resource Defined?

Let's take a look at the [CRD for ACK's S3 Bucket][ack-bucket-crd] (the
`s3.services.k8s.aws/Bucket` `GroupKind` (GK)) (snipped slightly for brevity):

[ack-bucket-crd]: https://github.com/aws/aws-controllers-k8s/blob/df6183acdc5b9b8508ea2fc8ec8c39fd19301ac6/services/s3/config/crd/bases/s3.services.k8s.aws_buckets.yaml

```yaml
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: buckets.s3.services.k8s.aws
spec:
  group: s3.services.k8s.aws
  names:
    kind: Bucket
  scope: Namespaced
  versions:
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        description: Bucket is the Schema for the Buckets API
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            description: BucketSpec defines the desired state of Bucket
            properties:
              acl:
                type: string
              createBucketConfiguration:
                properties:
                  locationConstraint:
                    type: string
                type: object
              grantFullControl:
                type: string
              grantRead:
                type: string
              grantReadACP:
                type: string
              grantWrite:
                type: string
              grantWriteACP:
                type: string
              name:
                type: string
              objectLockEnabledForBucket:
                type: boolean
            required:
            - name
            type: object
          status:
            description: BucketStatus defines the observed state of Bucket
            properties:
              ackResourceMetadata:
                properties:
                  arn:
                    type: string
                  ownerAccountID:
                    type: string
                required:
                - ownerAccountID
                type: object
              conditions:
                items:
                  properties:
                    lastTransitionTime:
                      format: date-time
                      type: string
                    message:
                      type: string
                    reason:
                      type: string
                    status:
                      type: string
                    type:
                      type: string
                  required:
                  - status
                  - type
                  type: object
                type: array
              location:
                type: string
            required:
            - ackResourceMetadata
            - conditions
            type: object
        type: object
```

The above YAML representation of a `CustomResourceDefinition` (CRD) is actually
generated from a set of Go type definitions. These Go type definitions live in
each ACK service's `services/$SERVICE/apis/$VERSION` directory.

This section of our documentation discusses how we create those Go type
definitions.

!!! note "controller-gen crd"
    The OpenAPIv3 Validating Schema shown above is created by the
    [`controller-gen crd`][cg] CLI command and is a convenient human-readable
    representation of the `CustomResourceDefinition`.

[cg]: https://book.kubebuilder.io/reference/controller-gen.html

The Bucket CR's `Spec` field is defined above as containing a set of fields --
"acl", "createBucketConfiguration", "name", etc. Each field has a JSONSchema
type that corresponds with the Go type from the associated field member.

You will also notice that in addition to the definition of a `Spec` field,
there is also the definition of a `Status` field for the Bucket CRs. Above,
this `Status` contains fields that represent the "observed" state of the Bucket
CRs. The above shows three fields in the Bucket's `Status`:
`ackResourceMetadata`, `conditions` and `location`.

You might be wondering how the ACK code generator determined which fields go
into the Bucket's `Spec` and which fields go into the Bucket's `Status`?

Well, it's definitely not a manual process. Everything in ACK is code-generated
and discovered by inspecting the AWS API model files.

!!! note "what are AWS API model files?"
    The AWS API model files are JSON files that contain information about a
    particular AWS service API's Actions and Shapes. We consume the model files
    [distributed][aws-sdk-go-model-files] in the `aws-sdk-go` project. (Look
    for the `api-2.json` files in the linked service-specific directories...)

[aws-sdk-go-model-files]: https://github.com/aws/aws-sdk-go/tree/master/models/apis

Let's take a look at a tiny bit of the [AWS S3 API model file][s3-api-file] and
you can start to see how we identify the things that go into the `Spec` and
`Status` fields.

[s3-api-file]: https://github.com/aws/aws-controllers-k8s/blob/main/pkg/generate/testdata/models/apis/s3/0000-00-00/api-2.json

```json
{
  "metadata":{
    "serviceId":"S3",
  },
  "operations":{
    "CreateBucket":{
      "name":"CreateBucket",
      "http":{
        "method":"PUT",
        "requestUri":"/{Bucket}"
      },
      "input":{"shape":"CreateBucketRequest"},
      "output":{"shape":"CreateBucketOutput"},
    },
  },
  "shapes":{
    "BucketCannedACL":{
      "type":"string",
      "enum":[
        "private",
        "public-read",
        "public-read-write",
        "authenticated-read"
      ]
    },
    "BucketName":{"type":"string"},
    "CreateBucketConfiguration":{
      "type":"structure",
      "members":{
        "LocationConstraint":{"shape":"BucketLocationConstraint"}
      }
    },
    "CreateBucketOutput":{
      "type":"structure",
      "members":{
        "Location":{
          "shape":"Location",
        }
      }
    },
    "CreateBucketRequest":{
      "type":"structure",
      "required":["Bucket"],
      "members":{
        "ACL":{
          "shape":"BucketCannedACL",
        },
        "Bucket":{
          "shape":"BucketName",
        },
        "CreateBucketConfiguration":{
          "shape":"CreateBucketConfiguration",
        },
        "GrantFullControl":{
          "shape":"GrantFullControl",
        },
        "GrantRead":{
          "shape":"GrantRead",
        },
        "GrantReadACP":{
          "shape":"GrantReadACP",
        },
        "GrantWrite":{
          "shape":"GrantWrite",
        },
        "GrantWriteACP":{
          "shape":"GrantWriteACP",
        },
        "ObjectLockEnabledForBucket":{
          "shape":"ObjectLockEnabledForBucket",
        }
      },
    },
  }
}
```

As mentioned above, we determine what things in an API are
`CustomResourceDefinition`s by looking for `Operation`s that begin with the
string "Create" and where the remainder of the `Operation` name refers to a
*singular* noun.

For the S3 API, there happens to be only a single Operation that begins with
the string "Create", and it happens to be "[CreateBucket][s3-create-bucket]".
And since "Bucket" refers to a singular noun, that is the
`CustomResourceDefinition` that is identified by the ACK code generator.

[s3-create-bucket]: https://docs.aws.amazon.com/AmazonS3/latest/API/API_CreateBucket.html

The ACK code generator writes a file [`apis/v1alpha1/bucket.go`][bucket-go]
that contains a `BucketSpec` struct definition, a `BucketStatus` struct
definition and a `Bucket` struct definition that ties the Spec and Status
together into our CRD.

[bucket-go]: https://github.com/aws/aws-controllers-k8s/blob/a10e9fc4f201129765260fa4f6751a6c9421bc31/services/s3/apis/v1alpha1/bucket.go

In determining the structure of the `s3.services.k8s.aws/Bucket` CRD, the ACK
code generator inspects the `Shape`s referred to by the "input" and "output"
members of the "CreateBucket" `Operation`: "CreateBucketRequest" and
"CreateBucketOutput" respectively.

### Determining the Spec fields

For the `BucketSpec` fields, we grab members of the `Input` shape. The
[generated Go type definition][spec-code] for the `BucketSpec` ends up looking
like this:

[spec-code]: https://github.com/aws/aws-controllers-k8s/blob/a10e9fc4f201129765260fa4f6751a6c9421bc31/services/s3/apis/v1alpha1/bucket.go#L23-L35

```go
// BucketSpec defines the desired state of Bucket
type BucketSpec struct {
	ACL                       *string                    `json:"acl,omitempty"`
	CreateBucketConfiguration *CreateBucketConfiguration `json:"createBucketConfiguration,omitempty"`
	GrantFullControl          *string                    `json:"grantFullControl,omitempty"`
	GrantRead                 *string                    `json:"grantRead,omitempty"`
	GrantReadACP              *string                    `json:"grantReadACP,omitempty"`
	GrantWrite                *string                    `json:"grantWrite,omitempty"`
	GrantWriteACP             *string                    `json:"grantWriteACP,omitempty"`
	// +kubebuilder:validation:Required
	Name                       *string `json:"name"`
	ObjectLockEnabledForBucket *bool   `json:"objectLockEnabledForBucket,omitempty"`
}
```

Let's take a closer look at the `BucketSpec` fields.

The `ACL`, `GrantFullControl`, `GrantRead`, `GrantReadACP`, `GrantWrite` and
`GrantWriteACP` fields are simple `*string` types. However, if we look at the
`CreateBucketRequest` Shape definition in the API model file, we see that these
fields actually are differently-named Shapes, not `*string`. Why is this? Well,
the ACK code generator "flattens" some Shapes when it notices that a named
Shape is just an alias for a simple scalar type (like `*string`).

!!! "why `*string`?"
    The astute reader may be wondering why the Go type for string fields is
    `*string` and not `string`. The reason for this lies in `aws-sdk-go`. All
    types for all Shape members are pointer types, even when the underlying
    data type is a simple scalar type like `bool` or `int`. Yes, even when
    the field is required...

Note that even though the `ACL` field has a Shape of `BucketCannedACL`, that
Shape is actually just a `string` with a set of enumerated values. Enumerated
values are collected and written out by the ACK code generator into an
[`apis/v1alpha1/enums.go`][enums-go] file, with content like this:

[enums-go]: https://github.com/aws/aws-controllers-k8s/blob/a10e9fc4f201129765260fa4f6751a6c9421bc31/services/s3/apis/v1alpha1/enums.go

```go
type BucketCannedACL string

const (
	BucketCannedACL_private            BucketCannedACL = "private"
	BucketCannedACL_public_read        BucketCannedACL = "public-read"
	BucketCannedACL_public_read_write  BucketCannedACL = "public-read-write"
	BucketCannedACL_authenticated_read BucketCannedACL = "authenticated-read"
)
```

The `CreateBucketConfiguration` field is of type `*CreateBucketConfiguration`.
All this means is that the field refers to a nested struct. All struct type
definitions for CRD Spec or Status field members are placed by the ACK code
generator into a [`apis/v1alpha1/types.go`][types-go] file.

[types-go]: https://github.com/aws/aws-controllers-k8s/blob/a10e9fc4f201129765260fa4f6751a6c9421bc31/services/s3/apis/v1alpha1/types.go

Here is a [snippet][cbc-def] of that file that contains the type definition for
the `CreateBucketConfiguration` struct:

[cbc-def]: https://github.com/aws/aws-controllers-k8s/blob/a10e9fc4f201129765260fa4f6751a6c9421bc31/services/s3/apis/v1alpha1/types.go#L36-L38

```go
type CreateBucketConfiguration struct {
	LocationConstraint *string `json:"locationConstraint,omitempty"`
}
```

Now, the `Name` field in the `BucketSpec` struct seems out of place, no? There
is no "Name" member of the `CreateBucketRequest` Shape, so why is there a
`Name` field in `BucketSpec`?

Well, this is an example of ACK's code generator using some special
instructions contained in something called the `generator.yaml` (or "generator
config") for the S3 service controller.

Each service in the `services/` directory can have a `generator.yaml` file that
contains overrides and special instructions for how to interpret and transform
parts of the service's API.

Here is part of the [S3 service's `generator.yaml`][s3-gen-yaml] file:

[s3-gen-yaml]: https://github.com/aws/aws-controllers-k8s/blob/a10e9fc4f201129765260fa4f6751a6c9421bc31/services/s3/generator.yaml

```yaml
resources:
  Bucket:
    renames:
      operations:
        CreateBucket:
          input_fields:
            Bucket: Name
```

As you can see, the generator config for the ACK S3 service controller is
renaming the `CreateBucket` Operation's Input Shape `Bucket` field to `Name`.
We do this for some APIs to add a little consistency and a more
Kubernetes-native experience for the CRDs. In Kubernetes, there is a
`Metadata.Name` (internal Kubernetes name) and there is typically a `Spec.Name`
field which refers to the **external** Name of the resource. So, in order to
align the `s3.services.k8s.aws/Bucket`'s definition to be more Kubernetes-like,
we rename the `Bucket` field to `Name`.

We do this renaming for other things that produce a bit of a
"[stutter][stutter]", as well as where the name of a field does not conform to
Go exported name constraints or [naming best practices][go-naming].

[stutter]: https://github.com/aws/aws-sdk-go/blob/master/private/model/api/legacy_stutter.go
[go-naming]: https://golang.org/doc/effective_go.html#names

### Determining the Status fields

Remember that fields in a CR's `Status` struct are not mutable by normal
Kubernetes users. Instead, these fields represent the latest observed state of
a resource (instead of the *desired* state of that resource which is
represented by fields in the CR's `Spec` struct).

The ACK code generator takes the members of the Create `Operation`'s `Output`
shape and puts those fields into the CR's `Status` struct.

We assume that fields in the `Output` that have the same name as fields in the
`Input` shape for the Create `Operation` refer to the resource field that was
set in the `Spec` field and therefore **are only interested in fields in the
`Output` that are not in the `Input`**.

Looking at the `BucketSpec` struct definition that was generated after
processing the S3 API model file, we find [this][bucket-status]:

[bucket-status]: https://github.com/aws/aws-controllers-k8s/blob/a10e9fc4f201129765260fa4f6751a6c9421bc31/services/s3/apis/v1alpha1/bucket.go#L37-L49

```go
// BucketStatus defines the observed state of Bucket
type BucketStatus struct {
	// All CRs managed by ACK have a common `Status.ACKResourceMetadata` member
	// that is used to contain resource sync state, account ownership,
	// constructed ARN for the resource
	ACKResourceMetadata *ackv1alpha1.ResourceMetadata `json:"ackResourceMetadata"`
	// All CRS managed by ACK have a common `Status.Conditions` member that
	// contains a collection of `ackv1alpha1.Condition` objects that describe
	// the various terminal states of the CR and its backend AWS service API
	// resource
	Conditions []*ackv1alpha1.Condition `json:"conditions"`
	Location   *string                  `json:"location,omitempty"`
}
```

Let's discuss each of the fields shown above.

First, the `ACKResourceMetadata` field is included in **every ACK CRD's Status
field**. It is a pointer to a [`ackv1alpha1.ResourceMetadata`][ack-rm] struct.
This struct contains some standard and important pieces of information about
the resource, including the AWS Resource Name (ARN) and the Owner AWS Account
ID.

[ack-rm]: https://github.com/aws/aws-controllers-k8s/blob/a10e9fc4f201129765260fa4f6751a6c9421bc31/apis/core/v1alpha1/resource_metadata.go#L16-L33

The ARN is a globally-unique identifier for the resource in AWS. The Owner AWS
Account ID is the 12-digit AWS account ID that is billed for the resource.

!!! note "cross-account resource management"
    The Owner AWS Account ID for a resource [may be different][carm] from the
    AWS Account ID of the IAM Role that the ACK service controller is executing
    under.

[carm]: https://aws.github.io/aws-controllers-k8s/user-docs/authorization/#create-resource-in-different-aws-accounts

The `Conditions` field is also included in every ACK CRD's Status field. It is
a slice of pointers to [`ackv1alpha1.Condition`][ack-cond] structs. The
`Condition` struct is responsible for conveying information about the latest
observed sync state of a resource, including any terminal condition states that
cause the resource to be "unsyncable".

[ack-cond]: https://github.com/aws/aws-controllers-k8s/blob/a10e9fc4f201129765260fa4f6751a6c9421bc31/apis/core/v1alpha1/conditions.go#L37-L54

Next is the `Location` field. This field gets its definition from the S3
`CreateBucketOutput.Location` field:

```json
    "CreateBucketOutput":{
      "type":"structure",
      "members":{
        "Location":{
          "shape":"Location",
        }
      }
    },
```
