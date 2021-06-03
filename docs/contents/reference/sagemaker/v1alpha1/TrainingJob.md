---
resource:
  apiVersion: v1alpha1
  description: TrainingJob is the Schema for the TrainingJobs API
  group: sagemaker.services.k8s.aws
  name: TrainingJob
  names:
    kind: TrainingJob
    listKind: TrainingJobList
    plural: trainingjobs
    singular: trainingjob
  scope: Namespaced
  service: sagemaker
  spec:
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: algorithmName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: enableSageMakerMetricsTimeSeries
      required: false
      type: boolean
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
        name: regex
        required: false
        type: string
      contains_description: ''
      description: ''
      name: metricDefinitions
      required: false
      type: array
    - contains: null
      contains_description: null
      description: ''
      name: trainingImage
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingInputMode
      required: false
      type: string
    contains_description: null
    description: The registry path of the Docker image that contains the training
      algorithm and algorithm-specific metadata, including the input mode. For more
      information about algorithms provided by Amazon SageMaker, see Algorithms (https://docs.aws.amazon.com/sagemaker/latest/dg/algos.html).
      For information about providing your own algorithms, see Using Your Own Algorithms
      with Amazon SageMaker (https://docs.aws.amazon.com/sagemaker/latest/dg/your-algorithms.html).
    name: algorithmSpecification
    required: true
    type: object
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
    contains_description: null
    description: Contains information about the output location for managed spot training
      checkpoint data.
    name: checkpointConfig
    required: false
    type: object
  - contains:
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: collectionName
        required: false
        type: string
      - contains: string
        contains_description: null
        description: ''
        name: collectionParameters
        required: false
        type: object
      contains_description: ''
      description: ''
      name: collectionConfigurations
      required: false
      type: array
    - contains: string
      contains_description: null
      description: ''
      name: hookParameters
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: localPath
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: s3OutputPath
      required: false
      type: string
    contains_description: null
    description: ''
    name: debugHookConfig
    required: false
    type: object
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: instanceType
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
      name: ruleConfigurationName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: ruleEvaluatorImage
      required: false
      type: string
    - contains: string
      contains_description: null
      description: ''
      name: ruleParameters
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: s3OutputPath
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: volumeSizeInGB
      required: false
      type: integer
    contains_description: ''
    description: Configuration information for Debugger rules for debugging output
      tensors.
    name: debugRuleConfigurations
    required: false
    type: array
  - contains: null
    contains_description: null
    description: To encrypt all communications between ML compute instances in distributed
      training, choose True. Encryption provides greater security for distributed
      training, but training might take longer. How long it takes depends on the amount
      of communication between compute instances, especially if you use a deep learning
      algorithm in distributed training. For more information, see Protect Communications
      Between ML Compute Instances in a Distributed Training Job (https://docs.aws.amazon.com/sagemaker/latest/dg/train-encrypt.html).
    name: enableInterContainerTrafficEncryption
    required: false
    type: boolean
  - contains: null
    contains_description: null
    description: "To train models using managed spot training, choose True. Managed\
      \ spot training provides a fully managed and scalable infrastructure for training\
      \ machine learning models. this option is useful when training jobs can be interrupted\
      \ and when there is flexibility when the training job is run. \n The complete\
      \ and intermediate results of jobs are stored in an Amazon S3 bucket, and can\
      \ be used as a starting point to train models incrementally. Amazon SageMaker\
      \ provides metrics and logs in CloudWatch. They can be used to see when managed\
      \ spot training jobs are running, interrupted, resumed, or completed."
    name: enableManagedSpotTraining
    required: false
    type: boolean
  - contains: null
    contains_description: null
    description: Isolates the training container. No inbound or outbound network calls
      can be made, except for calls between peers within a training cluster for distributed
      training. If you enable network isolation for training jobs that are configured
      to use a VPC, Amazon SageMaker downloads and uploads customer data and model
      artifacts through the specified VPC, but the training container does not have
      network access.
    name: enableNetworkIsolation
    required: false
    type: boolean
  - contains: string
    contains_description: null
    description: The environment variables to set in the Docker container.
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
  - contains: string
    contains_description: null
    description: "Algorithm-specific parameters that influence the quality of the\
      \ model. You set hyperparameters before you start the learning process. For\
      \ a list of hyperparameters for each training algorithm provided by Amazon SageMaker,\
      \ see Algorithms (https://docs.aws.amazon.com/sagemaker/latest/dg/algos.html).\
      \ \n You can specify a maximum of 100 hyperparameters. Each hyperparameter is\
      \ a key-value pair. Each key and value is limited to 256 characters, as specified\
      \ by the Length Constraint."
    name: hyperParameters
    required: false
    type: object
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: channelName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: compressionType
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: contentType
      required: false
      type: string
    - contains:
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: directoryPath
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: fileSystemAccessMode
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: fileSystemID
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: fileSystemType
          required: false
          type: string
        contains_description: null
        description: ''
        name: fileSystemDataSource
        required: false
        type: object
      - contains:
        - contains: string
          contains_description: ''
          description: ''
          name: attributeNames
          required: false
          type: array
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
          name: s3URI
          required: false
          type: string
        contains_description: null
        description: ''
        name: s3DataSource
        required: false
        type: object
      contains_description: null
      description: ''
      name: dataSource
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: inputMode
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: recordWrapperType
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: seed
        required: false
        type: integer
      contains_description: null
      description: ''
      name: shuffleConfig
      required: false
      type: object
    contains_description: ''
    description: "An array of Channel objects. Each channel is a named input source.\
      \ InputDataConfig describes the input data and its location. \n Algorithms can\
      \ accept input data from one or more channels. For example, an algorithm might\
      \ have two channels of input data, training_data and validation_data. The configuration\
      \ for each channel provides the S3, EFS, or FSx location where the input data\
      \ is stored. It also provides information about the stored data: the MIME type,\
      \ compression method, and whether the data is wrapped in RecordIO format. \n\
      \ Depending on the input mode that the algorithm supports, Amazon SageMaker\
      \ either copies input data files from an S3 bucket to a local directory in the\
      \ Docker container, or makes it available as input streams. For example, if\
      \ you specify an EFS location, input data files will be made available as input\
      \ streams. They do not need to be downloaded."
    name: inputDataConfig
    required: false
    type: array
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
      name: s3OutputPath
      required: false
      type: string
    contains_description: null
    description: Specifies the path to the S3 location where you want to store model
      artifacts. Amazon SageMaker creates subfolders for the artifacts.
    name: outputDataConfig
    required: true
    type: object
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: profilingIntervalInMilliseconds
      required: false
      type: integer
    - contains: string
      contains_description: null
      description: ''
      name: profilingParameters
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: s3OutputPath
      required: false
      type: string
    contains_description: null
    description: ''
    name: profilerConfig
    required: false
    type: object
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: instanceType
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
      name: ruleConfigurationName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: ruleEvaluatorImage
      required: false
      type: string
    - contains: string
      contains_description: null
      description: ''
      name: ruleParameters
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: s3OutputPath
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: volumeSizeInGB
      required: false
      type: integer
    contains_description: ''
    description: Configuration information for Debugger rules for profiling system
      and framework metrics.
    name: profilerRuleConfigurations
    required: false
    type: array
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
    description: "The resources, including the ML compute instances and ML storage\
      \ volumes, to use for model training. \n ML storage volumes store model artifacts\
      \ and incremental states. Training algorithms might also use ML storage volumes\
      \ for scratch space. If you want Amazon SageMaker to use the ML storage volume\
      \ to store the training data, choose File as the TrainingInputMode in the algorithm\
      \ specification. For distributed training algorithms, specify an instance count\
      \ greater than 1."
    name: resourceConfig
    required: true
    type: object
  - contains: null
    contains_description: null
    description: "The Amazon Resource Name (ARN) of an IAM role that Amazon SageMaker\
      \ can assume to perform tasks on your behalf. \n During model training, Amazon\
      \ SageMaker needs your permission to read input data from an S3 bucket, download\
      \ a Docker image that contains training code, write model artifacts to an S3\
      \ bucket, write logs to Amazon CloudWatch Logs, and publish metrics to Amazon\
      \ CloudWatch. You grant permissions for all of these tasks to an IAM role. For\
      \ more information, see Amazon SageMaker Roles (https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-roles.html).\
      \ \n To be able to pass this role to Amazon SageMaker, the caller of this API\
      \ must have the iam:PassRole permission."
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
    - contains: null
      contains_description: null
      description: ''
      name: maxWaitTimeInSeconds
      required: false
      type: integer
    contains_description: null
    description: "Specifies a limit to how long a model training job can run. When\
      \ the job reaches the time limit, Amazon SageMaker ends the training job. Use\
      \ this API to cap model training costs. \n To stop a job, Amazon SageMaker sends\
      \ the algorithm the SIGTERM signal, which delays job termination for 120 seconds.\
      \ Algorithms can use this 120-second window to save the model artifacts, so\
      \ the results of training are not lost."
    name: stoppingCondition
    required: true
    type: object
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
      name: s3OutputPath
      required: false
      type: string
    contains_description: null
    description: ''
    name: tensorBoardOutputConfig
    required: false
    type: object
  - contains: null
    contains_description: null
    description: The name of the training job. The name must be unique within an AWS
      Region in an AWS account.
    name: trainingJobName
    required: true
    type: string
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
    description: A VpcConfig object that specifies the VPC that you want your training
      job to connect to. Control access to and from your training container by configuring
      the VPC. For more information, see Protect Training Jobs by Using an Amazon
      Virtual Private Cloud (https://docs.aws.amazon.com/sagemaker/latest/dg/train-vpc.html).
    name: vpcConfig
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
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: lastModifiedTime
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: ruleConfigurationName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: ruleEvaluationJobARN
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: ruleEvaluationStatus
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: statusDetails
      required: false
      type: string
    contains_description: ''
    description: Evaluation status of Debugger rules for debugging on a training job.
    name: debugRuleEvaluationStatuses
    required: false
    type: array
  - contains: null
    contains_description: null
    description: If the training job failed, the reason it failed.
    name: failureReason
    required: false
    type: string
  - contains: null
    contains_description: null
    description: "Provides detailed information about the state of the training job.\
      \ For detailed information on the secondary status of the training job, see\
      \ StatusMessage under SecondaryStatusTransition. \n Amazon SageMaker provides\
      \ primary statuses and secondary statuses that apply to each of them: \n InProgress\
      \ \n    * Starting - Starting the training job. \n    * Downloading - An optional\
      \ stage for algorithms that support File training    input mode. It indicates\
      \ that data is being downloaded to the ML storage    volumes. \n    * Training\
      \ - Training is in progress. \n    * Interrupted - The job stopped because the\
      \ managed spot training instances    were interrupted. \n    * Uploading - Training\
      \ is complete and the model artifacts are being uploaded    to the S3 location.\
      \ \n Completed \n    * Completed - The training job has completed. \n Failed\
      \ \n    * Failed - The training job has failed. The reason for the failure is\
      \    returned in the FailureReason field of DescribeTrainingJobResponse. \n\
      \ Stopped \n    * MaxRuntimeExceeded - The job stopped because it exceeded the\
      \ maximum    allowed runtime. \n    * MaxWaitTimeExceeded - The job stopped\
      \ because it exceeded the maximum    allowed wait time. \n    * Stopped - The\
      \ training job has stopped. \n Stopping \n    * Stopping - Stopping the training\
      \ job. \n Valid values for SecondaryStatus are subject to change. \n We no longer\
      \ support the following secondary statuses: \n    * LaunchingMLInstances \n\
      \    * PreparingTrainingStack \n    * DownloadingTrainingImage"
    name: secondaryStatus
    required: false
    type: string
  - contains: null
    contains_description: null
    description: "The status of the training job. \n Amazon SageMaker provides the\
      \ following training job statuses: \n    * InProgress - The training is in progress.\
      \ \n    * Completed - The training job has completed. \n    * Failed - The training\
      \ job has failed. To see the reason for the failure,    see the FailureReason\
      \ field in the response to a DescribeTrainingJobResponse    call. \n    * Stopping\
      \ - The training job is stopping. \n    * Stopped - The training job has stopped.\
      \ \n For more detailed information, see SecondaryStatus."
    name: trainingJobStatus
    required: false
    type: string
---
{% include "reference.md" %}
