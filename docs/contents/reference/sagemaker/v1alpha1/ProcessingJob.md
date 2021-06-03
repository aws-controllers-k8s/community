---
resource:
  apiVersion: v1alpha1
  description: ProcessingJob is the Schema for the ProcessingJobs API
  group: sagemaker.services.k8s.aws
  name: ProcessingJob
  names:
    kind: ProcessingJob
    listKind: ProcessingJobList
    plural: processingjobs
    singular: processingjob
  scope: Namespaced
  service: sagemaker
  spec:
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
    contains_description: null
    description: Configures the processing job to run a specified Docker container
      image.
    name: appSpecification
    required: true
    type: object
  - contains: string
    contains_description: null
    description: The environment variables to set in the Docker container. Up to 100
      key and values entries in the map are supported.
    name: environment
    required: false
    type: object
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: experimentName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trialComponentDisplayName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trialName
      required: false
      type: string
    contains_description: null
    description: ''
    name: experimentConfig
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
      description: ''
      name: vpcConfig
      required: false
      type: object
    contains_description: null
    description: Networking options for a processing job, such as whether to allow
      inbound and outbound network calls to and from processing containers, and the
      VPC subnets and security groups to use for VPC-enabled processing jobs.
    name: networkConfig
    required: false
    type: object
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: appManaged
      required: false
      type: boolean
    - contains:
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: catalog
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: database
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
          name: outputCompression
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: outputFormat
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: outputS3URI
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: queryString
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: workGroup
          required: false
          type: string
        contains_description: null
        description: ''
        name: athenaDatasetDefinition
        required: false
        type: object
      - contains: null
        contains_description: null
        description: ''
        name: dataDistributionType
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: inputMode
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: localPath
        required: false
        type: string
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: clusterID
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: clusterRoleARN
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: database
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: dbUser
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
          name: outputCompression
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: outputFormat
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: outputS3URI
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: queryString
          required: false
          type: string
        contains_description: null
        description: ''
        name: redshiftDatasetDefinition
        required: false
        type: object
      contains_description: null
      description: ''
      name: datasetDefinition
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: inputName
      required: false
      type: string
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
        name: s3CompressionType
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: s3DataDistributionType
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: s3DataType
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
        name: s3URI
        required: false
        type: string
      contains_description: null
      description: ''
      name: s3Input
      required: false
      type: object
    contains_description: ''
    description: An array of inputs configuring the data to download into the processing
      container.
    name: processingInputs
    required: false
    type: array
  - contains: null
    contains_description: null
    description: The name of the processing job. The name must be unique within an
      AWS Region in the AWS account.
    name: processingJobName
    required: true
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: kmsKeyID
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: appManaged
        required: false
        type: boolean
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: featureGroupName
          required: false
          type: string
        contains_description: null
        description: ''
        name: featureStoreOutput
        required: false
        type: object
      - contains: null
        contains_description: null
        description: ''
        name: outputName
        required: false
        type: string
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
      name: outputs
      required: false
      type: array
    contains_description: null
    description: Output configuration for the processing job.
    name: processingOutputConfig
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
      description: ''
      name: clusterConfig
      required: false
      type: object
    contains_description: null
    description: Identifies the resources, ML compute instances, and ML storage volumes
      to deploy for a processing job. In distributed training, you specify more than
      one instance.
    name: processingResources
    required: true
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
    description: The time limit for how long the processing job is allowed to run.
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
  - contains: null
    contains_description: null
    description: A string, up to one KB in size, that contains the reason a processing
      job failed, if it failed.
    name: failureReason
    required: false
    type: string
  - contains: null
    contains_description: null
    description: Provides the status of a processing job.
    name: processingJobStatus
    required: false
    type: string
---
{% include "reference.md" %}
