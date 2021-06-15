---
resource:
  apiVersion: v1alpha1
  description: SnapshotSpec defines the desired state of Snapshot
  group: elasticache.services.k8s.aws
  name: Snapshot
  names:
    kind: Snapshot
    listKind: SnapshotList
    plural: snapshots
    singular: snapshot
  scope: Namespaced
  service: elasticache
  spec:
  - contains: null
    contains_description: null
    description: ''
    name: cacheClusterID
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: kmsKeyID
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: replicationGroupID
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: snapshotName
    required: true
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: sourceSnapshotName
    required: false
    type: string
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
  - contains: null
    contains_description: null
    description: ''
    name: autoMinorVersionUpgrade
    required: false
    type: boolean
  - contains: null
    contains_description: null
    description: ''
    name: automaticFailover
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: cacheClusterCreateTime
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: cacheNodeType
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: cacheParameterGroupName
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: cacheSubnetGroupName
    required: false
    type: string
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
    description: ''
    name: engine
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: engineVersion
    required: false
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: cacheClusterID
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: cacheNodeCreateTime
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: cacheNodeID
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: cacheSize
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: nodeGroupID
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: primaryAvailabilityZone
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: primaryOutpostARN
        required: false
        type: string
      - contains: string
        contains_description: ''
        description: ''
        name: replicaAvailabilityZones
        required: false
        type: array
      - contains: null
        contains_description: null
        description: ''
        name: replicaCount
        required: false
        type: integer
      - contains: string
        contains_description: ''
        description: ''
        name: replicaOutpostARNs
        required: false
        type: array
      - contains: null
        contains_description: null
        description: ''
        name: slots
        required: false
        type: string
      contains_description: null
      description: ''
      name: nodeGroupConfiguration
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: nodeGroupID
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: snapshotCreateTime
      required: false
      type: string
    contains_description: ''
    description: ''
    name: nodeSnapshots
    required: false
    type: array
  - contains: null
    contains_description: null
    description: ''
    name: numCacheNodes
    required: false
    type: integer
  - contains: null
    contains_description: null
    description: ''
    name: numNodeGroups
    required: false
    type: integer
  - contains: null
    contains_description: null
    description: ''
    name: port
    required: false
    type: integer
  - contains: null
    contains_description: null
    description: ''
    name: preferredAvailabilityZone
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: preferredMaintenanceWindow
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: preferredOutpostARN
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: replicationGroupDescription
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: snapshotRetentionLimit
    required: false
    type: integer
  - contains: null
    contains_description: null
    description: ''
    name: snapshotSource
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: snapshotStatus
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: snapshotWindow
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: topicARN
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: vpcID
    required: false
    type: string
---
{% include "reference.md" %}
