---
resource:
  apiVersion: v1alpha1
  description: ModelQualityJobDefinition is the Schema for the ModelQualityJobDefinitions
    API
  group: sagemaker.services.k8s.aws
  name: ModelQualityJobDefinition
  names:
    kind: ModelQualityJobDefinition
    listKind: ModelQualityJobDefinitionList
    plural: modelqualityjobdefinitions
    singular: modelqualityjobdefinition
  scope: Namespaced
  service: sagemaker
  spec:
  - contains: null
    contains_description: null
    description: The name of the monitoring job definition.
    name: jobDefinitionName
    required: true
    type: string
  - contains:
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: instanceCount
        required: false
        type: integer
      - contains: null
        contains_description: null
        description: ''
        name: instanceType
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: volumeKMSKeyID
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: volumeSizeInGB
        required: false
        type: integer
      contains_description: null
      description: ''
      name: clusterConfig
      required: false
      type: object
    contains_description: null
    description: ''
    name: jobResources
    required: true
    type: object
  - contains:
    - contains: string
      contains_description: ''
      description: ''
      name: containerArguments
      required: false
      type: array
    - contains: string
      contains_description: ''
      description: ''
      name: containerEntrypoint
      required: false
      type: array
    - contains: string
      contains_description: null
      description: ''
      name: environment
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: imageURI
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: postAnalyticsProcessorSourceURI
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: problemType
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: recordPreprocessorSourceURI
      required: false
      type: string
    contains_description: null
    description: The container that runs the monitoring job.
    name: modelQualityAppSpecification
    required: true
    type: object
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: baseliningJobName
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: s3URI
        required: false
        type: string
      contains_description: null
      description: ''
      name: constraintsResource
      required: false
      type: object
    contains_description: null
    description: Specifies the constraints and baselines for the monitoring job.
    name: modelQualityBaselineConfig
    required: false
    type: object
  - contains:
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: endTimeOffset
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: endpointName
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: featuresAttribute
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: inferenceAttribute
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: localPath
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: probabilityAttribute
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: probabilityThresholdAttribute
        required: false
        type: number
      - contains: null
        contains_description: null
        description: ''
        name: s3DataDistributionType
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: s3InputMode
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: startTimeOffset
        required: false
        type: string
      contains_description: null
      description: ''
      name: endpointInput
      required: false
      type: object
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: s3URI
        required: false
        type: string
      contains_description: null
      description: ''
      name: groundTruthS3Input
      required: false
      type: object
    contains_description: null
    description: A list of the inputs that are monitored. Currently endpoints are
      supported.
    name: modelQualityJobInput
    required: true
    type: object
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: kmsKeyID
      required: false
      type: string
    - contains:
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: localPath
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: s3URI
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: s3UploadMode
          required: false
          type: string
        contains_description: null
        description: ''
        name: s3Output
        required: false
        type: object
      contains_description: ''
      description: ''
      name: monitoringOutputs
      required: false
      type: array
    contains_description: null
    description: ''
    name: modelQualityJobOutputConfig
    required: true
    type: object
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: enableInterContainerTrafficEncryption
      required: false
      type: boolean
    - contains: null
      contains_description: null
      description: ''
      name: enableNetworkIsolation
      required: false
      type: boolean
    - contains:
      - contains: string
        contains_description: ''
        description: ''
        name: securityGroupIDs
        required: false
        type: array
      - contains: string
        contains_description: ''
        description: ''
        name: subnets
        required: false
        type: array
      contains_description: null
      description: ''
      name: vpcConfig
      required: false
      type: object
    contains_description: null
    description: Specifies the network configuration for the monitoring job.
    name: networkConfig
    required: false
    type: object
  - contains: null
    contains_description: null
    description: The Amazon Resource Name (ARN) of an IAM role that Amazon SageMaker
      can assume to perform tasks on your behalf.
    name: roleARN
    required: true
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: maxRuntimeInSeconds
      required: false
      type: integer
    contains_description: null
    description: ''
    name: stoppingCondition
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
