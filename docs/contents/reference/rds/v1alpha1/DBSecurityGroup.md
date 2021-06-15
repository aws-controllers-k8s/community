---
resource:
  apiVersion: v1alpha1
  description: DBSecurityGroupSpec defines the desired state of DBSecurityGroup
  group: rds.services.k8s.aws
  name: DBSecurityGroup
  names:
    kind: DBSecurityGroup
    listKind: DBSecurityGroupList
    plural: dbsecuritygroups
    singular: dbsecuritygroup
  scope: Namespaced
  service: rds
  spec:
  - contains: null
    contains_description: null
    description: The description for the DB security group.
    name: description
    required: true
    type: string
  - contains: null
    contains_description: null
    description: "The name for the DB security group. This value is stored as a lowercase\
      \ string. \n Constraints: \n    * Must be 1 to 255 letters, numbers, or hyphens.\
      \ \n    * First character must be a letter \n    * Can't end with a hyphen or\
      \ contain two consecutive hyphens \n    * Must not be \"Default\" \n Example:\
      \ mysecuritygroup"
    name: name
    required: true
    type: string
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
    description: Tags to assign to the DB security group.
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
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: ec2SecurityGroupID
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: ec2SecurityGroupName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: ec2SecurityGroupOwnerID
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: status
      required: false
      type: string
    contains_description: ''
    description: Contains a list of EC2SecurityGroup elements.
    name: ec2SecurityGroups
    required: false
    type: array
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: cidrIP
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: status
      required: false
      type: string
    contains_description: ''
    description: Contains a list of IPRange elements.
    name: iPRanges
    required: false
    type: array
  - contains: null
    contains_description: null
    description: Provides the AWS ID of the owner of a specific DB security group.
    name: ownerID
    required: false
    type: string
  - contains: null
    contains_description: null
    description: Provides the VpcId of the DB security group.
    name: vpcID
    required: false
    type: string
---
{% include "reference.md" %}
