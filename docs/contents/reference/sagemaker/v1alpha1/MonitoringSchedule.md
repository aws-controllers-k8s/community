---
resource:
  apiVersion: v1alpha1
  description: "MonitoringScheduleSpec defines the desired state of MonitoringSchedule.\
    \ \n A schedule for a model monitoring job. For information about model monitor,\
    \ see Amazon SageMaker Model Monitor (https://docs.aws.amazon.com/sagemaker/latest/dg/model-monitor.html)."
  group: sagemaker.services.k8s.aws
  name: MonitoringSchedule
  names:
    kind: MonitoringSchedule
    listKind: MonitoringScheduleList
    plural: monitoringschedules
    singular: monitoringschedule
  scope: Namespaced
  service: sagemaker
  spec:
  - contains:
    - contains:
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
          description: The constraints resource for a monitoring job.
          name: constraintsResource
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
          description: The statistics resource for a monitoring job.
          name: statisticsResource
          required: false
          type: object
        contains_description: null
        description: Configuration for monitoring constraints and monitoring statistics.
          These baseline resources are compared against the results of the current
          job from the series of jobs scheduled to collect data periodically.
        name: baselineConfig
        required: false
        type: object
      - contains: string
        contains_description: null
        description: ''
        name: environment
        required: false
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
          name: recordPreprocessorSourceURI
          required: false
          type: string
        contains_description: null
        description: Container image configuration object for the monitoring job.
        name: monitoringAppSpecification
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
          description: Input object for the endpoint
          name: endpointInput
          required: false
          type: object
        contains_description: The inputs for a monitoring job.
        description: ''
        name: monitoringInputs
        required: false
        type: array
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
            description: Information about where and how you want to store the results
              of a monitoring job.
            name: s3Output
            required: false
            type: object
          contains_description: The output object for a monitoring job.
          description: ''
          name: monitoringOutputs
          required: false
          type: array
        contains_description: null
        description: The output configuration for monitoring jobs.
        name: monitoringOutputConfig
        required: false
        type: object
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
          description: Configuration for the cluster used to run model monitoring
            jobs.
          name: clusterConfig
          required: false
          type: object
        contains_description: null
        description: Identifies the resources to deploy for a monitoring job.
        name: monitoringResources
        required: false
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
          description: Specifies a VPC that your training jobs and hosted models have
            access to. Control access to and from your training and model containers
            by configuring the VPC. For more information, see Protect Endpoints by
            Using an Amazon Virtual Private Cloud (https://docs.aws.amazon.com/sagemaker/latest/dg/host-vpc.html)
            and Protect Training Jobs by Using an Amazon Virtual Private Cloud (https://docs.aws.amazon.com/sagemaker/latest/dg/train-vpc.html).
          name: vpcConfig
          required: false
          type: object
        contains_description: null
        description: Networking options for a job, such as network traffic encryption
          between containers, whether to allow inbound and outbound network calls
          to and from containers, and the VPC subnets and security groups to use for
          VPC-enabled jobs.
        name: networkConfig
        required: false
        type: object
      - contains: null
        contains_description: null
        description: ''
        name: roleARN
        required: false
        type: string
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: maxRuntimeInSeconds
          required: false
          type: integer
        contains_description: null
        description: A time limit for how long the monitoring job is allowed to run
          before stopping.
        name: stoppingCondition
        required: false
        type: object
      contains_description: null
      description: Defines the monitoring job.
      name: monitoringJobDefinition
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: monitoringJobDefinitionName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: monitoringType
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: scheduleExpression
        required: false
        type: string
      contains_description: null
      description: Configuration details about the monitoring schedule.
      name: scheduleConfig
      required: false
      type: object
    contains_description: null
    description: The configuration object that specifies the monitoring schedule and
      defines the monitoring job.
    name: monitoringScheduleConfig
    required: true
    type: object
  - contains: null
    contains_description: null
    description: The name of the monitoring schedule. The name must be unique within
      an AWS Region within an AWS account.
    name: monitoringScheduleName
    required: true
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
    description: The time at which the monitoring job was created.
    name: creationTime
    required: false
    type: string
  - contains: null
    contains_description: null
    description: A string, up to one KB in size, that contains the reason a monitoring
      job failed, if it failed.
    name: failureReason
    required: false
    type: string
  - contains: null
    contains_description: null
    description: The time at which the monitoring job was last modified.
    name: lastModifiedTime
    required: false
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: creationTime
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
      name: failureReason
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: lastModifiedTime
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: monitoringExecutionStatus
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: monitoringJobDefinitionName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: monitoringScheduleName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: monitoringType
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: processingJobARN
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: scheduledTime
      required: false
      type: string
    contains_description: null
    description: Describes metadata on the last execution to run, if there was one.
    name: lastMonitoringExecutionSummary
    required: false
    type: object
  - contains: null
    contains_description: null
    description: The status of an monitoring job.
    name: monitoringScheduleStatus
    required: false
    type: string
---
{% include "reference.md" %}
