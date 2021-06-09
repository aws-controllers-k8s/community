---
resource:
  apiVersion: v1alpha1
  description: Broker is the Schema for the Brokers API
  group: mq.services.k8s.aws
  name: Broker
  names:
    kind: Broker
    listKind: BrokerList
    plural: brokers
    singular: broker
  scope: Namespaced
  service: mq
  spec:
  - contains: null
    contains_description: null
    description: ''
    name: authenticationStrategy
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: autoMinorVersionUpgrade
    required: false
    type: boolean
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: id
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: revision
      required: false
      type: integer
    contains_description: null
    description: A list of information about the configuration. Does not apply to
      RabbitMQ brokers.
    name: configuration
    required: false
    type: object
  - contains: null
    contains_description: null
    description: ''
    name: creatorRequestID
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: deploymentMode
    required: false
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: kmsKeyID
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: useAWSOwnedKey
      required: false
      type: boolean
    contains_description: null
    description: Encryption options for the broker.
    name: encryptionOptions
    required: false
    type: object
  - contains: null
    contains_description: null
    description: ''
    name: engineType
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
    name: hostInstanceType
    required: false
    type: string
  - contains:
    - contains: string
      contains_description: ''
      description: ''
      name: hosts
      required: false
      type: array
    - contains: null
      contains_description: null
      description: ''
      name: roleBase
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: roleName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: roleSearchMatching
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: roleSearchSubtree
      required: false
      type: boolean
    - contains: null
      contains_description: null
      description: ''
      name: serviceAccountPassword
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: serviceAccountUsername
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: userBase
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: userRoleName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: userSearchMatching
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: userSearchSubtree
      required: false
      type: boolean
    contains_description: null
    description: The metadata of the LDAP server used to authenticate and authorize
      connections to the broker. Currently not supported for RabbitMQ engine type.
    name: ldapServerMetadata
    required: false
    type: object
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: audit
      required: false
      type: boolean
    - contains: null
      contains_description: null
      description: ''
      name: general
      required: false
      type: boolean
    contains_description: null
    description: The list of information about logs to be enabled for the specified
      broker.
    name: logs
    required: false
    type: object
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: dayOfWeek
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: timeOfDay
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: timeZone
      required: false
      type: string
    contains_description: null
    description: The scheduled time period relative to UTC during which Amazon MQ
      begins to apply pending updates or patches to the broker.
    name: maintenanceWindowStartTime
    required: false
    type: object
  - contains: null
    contains_description: null
    description: ''
    name: name
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: publiclyAccessible
    required: false
    type: boolean
  - contains: string
    contains_description: ''
    description: ''
    name: securityGroups
    required: false
    type: array
  - contains: null
    contains_description: null
    description: ''
    name: storageType
    required: false
    type: string
  - contains: string
    contains_description: ''
    description: ''
    name: subnetIDs
    required: false
    type: array
  - contains: string
    contains_description: null
    description: ''
    name: tags
    required: false
    type: object
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: consoleAccess
      required: false
      type: boolean
    - contains: string
      contains_description: ''
      description: ''
      name: groups
      required: false
      type: array
    - contains: null
      contains_description: null
      description: ''
      name: password
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: username
      required: false
      type: string
    contains_description: A user associated with the broker.
    description: ''
    name: users
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
  - contains: null
    contains_description: null
    description: ''
    name: brokerID
    required: false
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: consoleURL
      required: false
      type: string
    - contains: string
      contains_description: ''
      description: ''
      name: endpoints
      required: false
      type: array
    - contains: null
      contains_description: null
      description: ''
      name: ipAddress
      required: false
      type: string
    contains_description: Returns information about all brokers.
    description: ''
    name: brokerInstances
    required: false
    type: array
  - contains: null
    contains_description: null
    description: ''
    name: brokerState
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
---
{% include "reference.md" %}
