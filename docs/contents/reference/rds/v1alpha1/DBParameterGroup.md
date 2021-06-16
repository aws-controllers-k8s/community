---
resource:
  apiVersion: v1alpha1
  description: DBParameterGroupSpec defines the desired state of DBParameterGroup
  group: rds.services.k8s.aws
  name: DBParameterGroup
  names:
    kind: DBParameterGroup
    listKind: DBParameterGroupList
    plural: dbparametergroups
    singular: dbparametergroup
  scope: Namespaced
  service: rds
  spec:
  - contains: null
    contains_description: null
    description: The description for the DB parameter group.
    name: description
    required: true
    type: string
  - contains: null
    contains_description: null
    description: "The DB parameter group family name. A DB parameter group can be\
      \ associated with one and only one DB parameter group family, and can be applied\
      \ only to a DB instance running a database engine and engine version compatible\
      \ with that DB parameter group family. \n To list all of the available parameter\
      \ group families, use the following command: \n aws rds describe-db-engine-versions\
      \ --query \"DBEngineVersions[].DBParameterGroupFamily\" \n The output contains\
      \ duplicates."
    name: family
    required: true
    type: string
  - contains: null
    contains_description: null
    description: "The name of the DB parameter group. \n Constraints: \n    * Must\
      \ be 1 to 255 letters, numbers, or hyphens. \n    * First character must be\
      \ a letter \n    * Can't end with a hyphen or contain two consecutive hyphens\
      \ \n This value is stored as a lowercase string."
    name: name
    required: true
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: allowedValues
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: applyMethod
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: applyType
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: dataType
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: description
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: isModifiable
      required: false
      type: boolean
    - contains: null
      contains_description: null
      description: ''
      name: minimumEngineVersion
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: parameterName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: parameterValue
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: source
      required: false
      type: string
    - contains: string
      contains_description: ''
      description: ''
      name: supportedEngineModes
      required: false
      type: array
    contains_description: ''
    description: "An array of parameter names, values, and the apply method for the\
      \ parameter update. At least one parameter name, value, and apply method must\
      \ be supplied; later arguments are optional. A maximum of 20 parameters can\
      \ be modified in a single request. \n Valid Values (for the application method):\
      \ immediate | pending-reboot \n You can use the immediate value with dynamic\
      \ parameters only. You can use the pending-reboot value for both dynamic and\
      \ static parameters, and changes are applied when you reboot the DB instance\
      \ without failover."
    name: parameters
    required: false
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
    description: Tags to assign to the DB parameter group.
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
---
{% include "reference.md" %}
