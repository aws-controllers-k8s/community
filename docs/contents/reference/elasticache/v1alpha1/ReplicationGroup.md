---
resource:
  apiVersion: v1alpha1
  description: ReplicationGroup is the Schema for the ReplicationGroups API
  group: elasticache.services.k8s.aws
  name: ReplicationGroup
  names:
    kind: ReplicationGroup
    listKind: ReplicationGroupList
    plural: replicationgroups
    singular: replicationgroup
  scope: Namespaced
  service: elasticache
  spec:
  - contains: null
    contains_description: null
    description: ''
    name: atRestEncryptionEnabled
    required: false
    type: boolean
  - contains:
    - contains: null
      contains_description: null
      description: Key is the key within the secret
      name: key
      required: true
      type: string
    - contains: null
      contains_description: null
      description: Name is unique within a namespace to reference a secret resource.
      name: name
      required: false
      type: string
    - contains: null
      contains_description: null
      description: Namespace defines the space within which the secret name must be
        unique.
      name: namespace
      required: false
      type: string
    contains_description: null
    description: SecretKeyReference combines a k8s corev1.SecretReference with a specific
      key within the referred-to Secret
    name: authToken
    required: false
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
    name: automaticFailoverEnabled
    required: false
    type: boolean
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
  - contains: string
    contains_description: ''
    description: ''
    name: cacheSecurityGroupNames
    required: false
    type: array
  - contains: null
    contains_description: null
    description: ''
    name: cacheSubnetGroupName
    required: false
    type: string
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
  - contains: null
    contains_description: null
    description: ''
    name: kmsKeyID
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: multiAZEnabled
    required: false
    type: boolean
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
    contains_description: ''
    description: ''
    name: nodeGroupConfiguration
    required: false
    type: array
  - contains: null
    contains_description: null
    description: ''
    name: notificationTopicARN
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: numCacheClusters
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
  - contains: string
    contains_description: ''
    description: ''
    name: preferredCacheClusterAZs
    required: false
    type: array
  - contains: null
    contains_description: null
    description: ''
    name: preferredMaintenanceWindow
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: primaryClusterID
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: replicasPerNodeGroup
    required: false
    type: integer
  - contains: null
    contains_description: null
    description: ''
    name: replicationGroupDescription
    required: true
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: replicationGroupID
    required: true
    type: string
  - contains: string
    contains_description: ''
    description: ''
    name: securityGroupIDs
    required: false
    type: array
  - contains: string
    contains_description: ''
    description: ''
    name: snapshotARNs
    required: false
    type: array
  - contains: null
    contains_description: null
    description: ''
    name: snapshotName
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
    name: snapshotWindow
    required: false
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
    description: ''
    name: tags
    required: false
    type: array
  - contains: null
    contains_description: null
    description: ''
    name: transitEncryptionEnabled
    required: false
    type: boolean
  - contains: string
    contains_description: ''
    description: ''
    name: userGroupIDs
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
  - contains: string
    contains_description: ''
    description: ''
    name: allowedScaleDownModifications
    required: false
    type: array
  - contains: string
    contains_description: ''
    description: ''
    name: allowedScaleUpModifications
    required: false
    type: array
  - contains: null
    contains_description: null
    description: ''
    name: authTokenEnabled
    required: false
    type: boolean
  - contains: null
    contains_description: null
    description: ''
    name: authTokenLastModifiedDate
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: automaticFailover
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: clusterEnabled
    required: false
    type: boolean
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
      name: address
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: port
      required: false
      type: integer
    contains_description: null
    description: ''
    name: configurationEndpoint
    required: false
    type: object
  - contains: null
    contains_description: null
    description: ''
    name: description
    required: false
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: date
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: message
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: sourceIdentifier
      required: false
      type: string
    contains_description: ''
    description: ''
    name: events
    required: false
    type: array
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: globalReplicationGroupID
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: globalReplicationGroupMemberRole
      required: false
      type: string
    contains_description: null
    description: ''
    name: globalReplicationGroupInfo
    required: false
    type: object
  - contains: string
    contains_description: ''
    description: ''
    name: memberClusters
    required: false
    type: array
  - contains: string
    contains_description: ''
    description: ''
    name: memberClustersOutpostARNs
    required: false
    type: array
  - contains: null
    contains_description: null
    description: ''
    name: multiAZ
    required: false
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: nodeGroupID
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
        name: cacheNodeID
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: currentRole
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: preferredAvailabilityZone
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: preferredOutpostARN
        required: false
        type: string
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: address
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: port
          required: false
          type: integer
        contains_description: null
        description: ''
        name: readEndpoint
        required: false
        type: object
      contains_description: ''
      description: ''
      name: nodeGroupMembers
      required: false
      type: array
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: address
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: port
        required: false
        type: integer
      contains_description: null
      description: ''
      name: primaryEndpoint
      required: false
      type: object
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: address
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: port
        required: false
        type: integer
      contains_description: null
      description: ''
      name: readerEndpoint
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: slots
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: status
      required: false
      type: string
    contains_description: ''
    description: ''
    name: nodeGroups
    required: false
    type: array
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: authTokenStatus
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: automaticFailoverStatus
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: primaryClusterID
      required: false
      type: string
    - contains:
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: progressPercentage
          required: false
          type: number
        contains_description: null
        description: ''
        name: slotMigration
        required: false
        type: object
      contains_description: null
      description: ''
      name: resharding
      required: false
      type: object
    - contains:
      - contains: string
        contains_description: ''
        description: ''
        name: userGroupIDsToAdd
        required: false
        type: array
      - contains: string
        contains_description: ''
        description: ''
        name: userGroupIDsToRemove
        required: false
        type: array
      contains_description: null
      description: ''
      name: userGroups
      required: false
      type: object
    contains_description: null
    description: ''
    name: pendingModifiedValues
    required: false
    type: object
  - contains: null
    contains_description: null
    description: ''
    name: snapshottingClusterID
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: status
    required: false
    type: string
---
{% include "reference.md" %}
