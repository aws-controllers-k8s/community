---
resource:
  apiVersion: v1alpha1
  description: "BucketSpec defines the desired state of Bucket. \n In terms of implementation,\
    \ a Bucket is a resource. An Amazon S3 bucket name is globally unique, and the\
    \ namespace is shared by all AWS accounts."
  group: s3.services.k8s.aws
  name: Bucket
  names:
    kind: Bucket
    listKind: BucketList
    plural: buckets
    singular: bucket
  scope: Namespaced
  service: s3
  spec:
  - contains: null
    contains_description: null
    description: The canned ACL to apply to the bucket.
    name: acl
    required: false
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: locationConstraint
      required: false
      type: string
    contains_description: null
    description: The configuration information for the bucket.
    name: createBucketConfiguration
    required: false
    type: object
  - contains: null
    contains_description: null
    description: Allows grantee the read, write, read ACP, and write ACP permissions
      on the bucket.
    name: grantFullControl
    required: false
    type: string
  - contains: null
    contains_description: null
    description: Allows grantee to list the objects in the bucket.
    name: grantRead
    required: false
    type: string
  - contains: null
    contains_description: null
    description: Allows grantee to read the bucket ACL.
    name: grantReadACP
    required: false
    type: string
  - contains: null
    contains_description: null
    description: Allows grantee to create, overwrite, and delete any object in the
      bucket.
    name: grantWrite
    required: false
    type: string
  - contains: null
    contains_description: null
    description: Allows grantee to write the ACL for the applicable bucket.
    name: grantWriteACP
    required: false
    type: string
  - contains: null
    contains_description: null
    description: The name of the bucket to create.
    name: name
    required: true
    type: string
  - contains: null
    contains_description: null
    description: Specifies whether you want S3 Object Lock to be enabled for the new
      bucket.
    name: objectLockEnabledForBucket
    required: false
    type: boolean
  status:
  - contains:
    - contains: null
      contains_description: null
      description: 'ARN is the Amazon Resource Name for the resource. This is a globally-unique
        identifier and is set only by the ACK service controller once the controller
        has orchestrated the creation of the resource OR when it has verified that
        an "adopted" resource (a resource where the ARN annotation was set by the
        Kubernetes user on the CR) exists and matches the supplied CR''s Spec field
        values. TODO(vijat@): Find a better strategy for resources that do not have
        ARN in CreateOutputResponse https://github.com/aws/aws-controllers-k8s/issues/270'
      name: arn
      required: false
      type: string
    - contains: null
      contains_description: null
      description: OwnerAccountID is the AWS Account ID of the account that owns the
        backend AWS service API resource.
      name: ownerAccountID
      required: true
      type: string
    contains_description: null
    description: All CRs managed by ACK have a common `Status.ACKResourceMetadata`
      member that is used to contain resource sync state, account ownership, constructed
      ARN for the resource
    name: ackResourceMetadata
    required: true
    type: object
  - contains:
    - contains: null
      contains_description: null
      description: Last time the condition transitioned from one status to another.
      name: lastTransitionTime
      required: false
      type: string
    - contains: null
      contains_description: null
      description: A human readable message indicating details about the transition.
      name: message
      required: false
      type: string
    - contains: null
      contains_description: null
      description: The reason for the condition's last transition.
      name: reason
      required: false
      type: string
    - contains: null
      contains_description: null
      description: Status of the condition, one of True, False, Unknown.
      name: status
      required: false
      type: string
    - contains: null
      contains_description: null
      description: Type is the type of the Condition
      name: type
      required: false
      type: string
    contains_description: Condition is the common struct used by all CRDs managed
      by ACK service controllers to indicate terminal states  of the CR and its backend
      AWS service API resource
    description: All CRS managed by ACK have a common `Status.Conditions` member that
      contains a collection of `ackv1alpha1.Condition` objects that describe the various
      terminal states of the CR and its backend AWS service API resource
    name: conditions
    required: true
    type: array
  - contains: null
    contains_description: null
    description: Specifies the Region where the bucket will be created. If you are
      creating a bucket on the US East (N. Virginia) Region (us-east-1), you do not
      need to specify the location.
    name: location
    required: false
    type: string
---
{% include "reference.md" %}
