---
resource:
  apiVersion: v1alpha1
  description: ScalingPolicy is the Schema for the ScalingPolicies API
  group: applicationautoscaling.services.k8s.aws
  name: ScalingPolicy
  names:
    kind: ScalingPolicy
    listKind: ScalingPolicyList
    plural: scalingpolicies
    singular: scalingpolicy
  scope: Namespaced
  service: applicationautoscaling
  spec:
  - contains: null
    contains_description: null
    description: ''
    name: policyName
    required: true
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: policyType
    required: false
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: resourceID
    required: true
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: scalableDimension
    required: true
    type: string
  - contains: null
    contains_description: null
    description: ''
    name: serviceNamespace
    required: true
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: adjustmentType
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: cooldown
      required: false
      type: integer
    - contains: null
      contains_description: null
      description: ''
      name: metricAggregationType
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: minAdjustmentMagnitude
      required: false
      type: integer
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: metricIntervalLowerBound
        required: false
        type: number
      - contains: null
        contains_description: null
        description: ''
        name: metricIntervalUpperBound
        required: false
        type: number
      - contains: null
        contains_description: null
        description: ''
        name: scalingAdjustment
        required: false
        type: integer
      contains_description: ''
      description: ''
      name: stepAdjustments
      required: false
      type: array
    contains_description: null
    description: ''
    name: stepScalingPolicyConfiguration
    required: false
    type: object
  - contains:
    - contains:
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: name
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
        name: dimensions
        required: false
        type: array
      - contains: null
        contains_description: null
        description: ''
        name: metricName
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: namespace
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: statistic
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: unit
        required: false
        type: string
      contains_description: null
      description: ''
      name: customizedMetricSpecification
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: disableScaleIn
      required: false
      type: boolean
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: predefinedMetricType
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: resourceLabel
        required: false
        type: string
      contains_description: null
      description: ''
      name: predefinedMetricSpecification
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: scaleInCooldown
      required: false
      type: integer
    - contains: null
      contains_description: null
      description: ''
      name: scaleOutCooldown
      required: false
      type: integer
    - contains: null
      contains_description: null
      description: ''
      name: targetValue
      required: false
      type: number
    contains_description: null
    description: ''
    name: targetTrackingScalingPolicyConfiguration
    required: false
    type: object
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
      description: ''
      name: alarmARN
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: alarmName
      required: false
      type: string
    contains_description: ''
    description: ''
    name: alarms
    required: false
    type: array
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
    name: policyARN
    required: false
    type: string
---
{% include "reference.md" %}
