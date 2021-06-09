---
resource:
  apiVersion: v1alpha1
  description: DBSubnetGroup is the Schema for the DBSubnetGroups API
  group: rds.services.k8s.aws
  name: DBSubnetGroup
  names:
    kind: DBSubnetGroup
    listKind: DBSubnetGroupList
    plural: dbsubnetgroups
    singular: dbsubnetgroup
  scope: Namespaced
  service: rds
  spec:
  - contains: null
    contains_description: null
    description: The description for the DB subnet group.
    name: description
    required: true
    type: string
  - contains: null
    contains_description: null
    description: "The name for the DB subnet group. This value is stored as a lowercase\
      \ string. \n Constraints: Must contain no more than 255 letters, numbers, periods,\
      \ underscores, spaces, or hyphens. Must not be default. \n Example: mySubnetgroup"
    name: name
    required: true
    type: string
  - contains: string
    contains_description: ''
    description: The EC2 Subnet IDs for the DB subnet group.
    name: subnetIDs
    required: true
    type: array
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: key
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: value
      required: false
      type: string
    contains_description: ''
    description: Tags to assign to the DB subnet group.
    name: tags
    required: false
    type: array
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
    description: Provides the status of the DB subnet group.
    name: subnetGroupStatus
    required: false
    type: string
  - contains:
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: name
        required: false
        type: string
      contains_description: null
      description: ''
      name: subnetAvailabilityZone
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: subnetIdentifier
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: arn
        required: false
        type: string
      contains_description: null
      description: ''
      name: subnetOutpost
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: subnetStatus
      required: false
      type: string
    contains_description: ''
    description: Contains a list of Subnet elements.
    name: subnets
    required: false
    type: array
  - contains: null
    contains_description: null
    description: Provides the VpcId of the DB subnet group.
    name: vpcID
    required: false
    type: string
---
{% include "reference.md" %}
