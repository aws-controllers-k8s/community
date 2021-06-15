---
resource:
  apiVersion: v1alpha1
  description: HyperParameterTuningJobSpec defines the desired state of HyperParameterTuningJob.
  group: sagemaker.services.k8s.aws
  name: HyperParameterTuningJob
  names:
    kind: HyperParameterTuningJob
    listKind: HyperParameterTuningJobList
    plural: hyperparametertuningjobs
    singular: hyperparametertuningjob
  scope: Namespaced
  service: sagemaker
  spec:
  - contains:
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: metricName
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: type_
        required: false
        type: string
      contains_description: null
      description: Defines the objective metric for a hyperparameter tuning job. Hyperparameter
        tuning uses the value of this metric to evaluate the training jobs it launches,
        and returns the training job that results in either the highest or lowest
        value for this metric, depending on the value you specify for the Type parameter.
      name: hyperParameterTuningJobObjective
      required: false
      type: object
    - contains:
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: name
          required: false
          type: string
        - contains: string
          contains_description: ''
          description: ''
          name: values
          required: false
          type: array
        contains_description: A list of categorical hyperparameters to tune.
        description: ''
        name: categoricalParameterRanges
        required: false
        type: array
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: maxValue
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: minValue
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: name
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: scalingType
          required: false
          type: string
        contains_description: A list of continuous hyperparameters to tune.
        description: ''
        name: continuousParameterRanges
        required: false
        type: array
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: maxValue
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: minValue
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: name
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: scalingType
          required: false
          type: string
        contains_description: For a hyperparameter of the integer type, specifies
          the range that a hyperparameter tuning job searches.
        description: ''
        name: integerParameterRanges
        required: false
        type: array
      contains_description: null
      description: "Specifies ranges of integer, continuous, and categorical hyperparameters\
        \ that a hyperparameter tuning job searches. The hyperparameter tuning job\
        \ launches training jobs with hyperparameter values within these ranges to\
        \ find the combination of values that result in the training job with the\
        \ best performance as measured by the objective metric of the hyperparameter\
        \ tuning job. \n You can specify a maximum of 20 hyperparameters that a hyperparameter\
        \ tuning job can search over. Every possible value of a categorical parameter\
        \ range counts against this limit."
      name: parameterRanges
      required: false
      type: object
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: maxNumberOfTrainingJobs
        required: false
        type: integer
      - contains: null
        contains_description: null
        description: ''
        name: maxParallelTrainingJobs
        required: false
        type: integer
      contains_description: null
      description: Specifies the maximum number of training jobs and parallel training
        jobs that a hyperparameter tuning job can launch.
      name: resourceLimits
      required: false
      type: object
    - contains: null
      contains_description: null
      description: The strategy hyperparameter tuning uses to find the best combination
        of hyperparameters for your model. Currently, the only supported value is
        Bayesian.
      name: strategy
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingJobEarlyStoppingType
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: targetObjectiveMetricValue
        required: false
        type: number
      contains_description: null
      description: The job completion criteria.
      name: tuningJobCompletionCriteria
      required: false
      type: object
    contains_description: null
    description: The HyperParameterTuningJobConfig object that describes the tuning
      job, including the search strategy, the objective metric used to evaluate training
      jobs, ranges of parameters to search, and resource limits for the tuning job.
      For more information, see How Hyperparameter Tuning Works (https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-how-it-works.html).
    name: hyperParameterTuningJobConfig
    required: true
    type: object
  - contains: null
    contains_description: null
    description: 'The name of the tuning job. This name is the prefix for the names
      of all training jobs that this tuning job launches. The name must be unique
      within the same AWS account and AWS Region. The name must have 1 to 32 characters.
      Valid characters are a-z, A-Z, 0-9, and : + = @ _ % - (hyphen). The name is
      not case sensitive.'
    name: hyperParameterTuningJobName
    required: true
    type: string
  - contains:
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: algorithmName
        required: false
        type: string
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
        contains_description: Specifies a metric that the training algorithm writes
          to stderr or stdout . Amazon SageMakerhyperparameter tuning captures all
          defined metrics. You specify one metric that a hyperparameter tuning job
          uses as its objective metric to choose the best training job.
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
      description: Specifies which training algorithm to use for training jobs that
        a hyperparameter tuning job launches and the metrics to monitor.
      name: algorithmSpecification
      required: false
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
      description: Contains information about the output location for managed spot
        training checkpoint data.
      name: checkpointConfig
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: definitionName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: enableInterContainerTrafficEncryption
      required: false
      type: boolean
    - contains: null
      contains_description: null
      description: ''
      name: enableManagedSpotTraining
      required: false
      type: boolean
    - contains: null
      contains_description: null
      description: ''
      name: enableNetworkIsolation
      required: false
      type: boolean
    - contains:
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: name
          required: false
          type: string
        - contains: string
          contains_description: ''
          description: ''
          name: values
          required: false
          type: array
        contains_description: A list of categorical hyperparameters to tune.
        description: ''
        name: categoricalParameterRanges
        required: false
        type: array
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: maxValue
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: minValue
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: name
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: scalingType
          required: false
          type: string
        contains_description: A list of continuous hyperparameters to tune.
        description: ''
        name: continuousParameterRanges
        required: false
        type: array
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: maxValue
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: minValue
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: name
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: scalingType
          required: false
          type: string
        contains_description: For a hyperparameter of the integer type, specifies
          the range that a hyperparameter tuning job searches.
        description: ''
        name: integerParameterRanges
        required: false
        type: array
      contains_description: null
      description: "Specifies ranges of integer, continuous, and categorical hyperparameters\
        \ that a hyperparameter tuning job searches. The hyperparameter tuning job\
        \ launches training jobs with hyperparameter values within these ranges to\
        \ find the combination of values that result in the training job with the\
        \ best performance as measured by the objective metric of the hyperparameter\
        \ tuning job. \n You can specify a maximum of 20 hyperparameters that a hyperparameter\
        \ tuning job can search over. Every possible value of a categorical parameter\
        \ range counts against this limit."
      name: hyperParameterRanges
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
          description: Specifies a file system data source for a channel.
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
          description: Describes the S3 data source.
          name: s3DataSource
          required: false
          type: object
        contains_description: null
        description: Describes the location of the channel data.
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
        description: "A configuration for a shuffle option for input data in a channel.\
          \ If you use S3Prefix for S3DataType, the results of the S3 key prefix matches\
          \ are shuffled. If you use ManifestFile, the order of the S3 object references\
          \ in the ManifestFile is shuffled. If you use AugmentedManifestFile, the\
          \ order of the JSON lines in the AugmentedManifestFile is shuffled. The\
          \ shuffling order is determined using the Seed value. \n For Pipe input\
          \ mode, when ShuffleConfig is specified shuffling is done at the start of\
          \ every epoch. With large datasets, this ensures that the order of the training\
          \ data is different for each epoch, and it helps reduce bias and possible\
          \ overfitting. In a multi-node training job when ShuffleConfig is combined\
          \ with S3DataDistributionType of ShardedByS3Key, the data is shuffled across\
          \ nodes so that the content sent to a particular node on the first epoch\
          \ might be sent to a different node on the second epoch."
        name: shuffleConfig
        required: false
        type: object
      contains_description: A channel is a named input source that training algorithms
        can consume.
      description: ''
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
      description: Provides information about how to store model training results
        (model artifacts).
      name: outputDataConfig
      required: false
      type: object
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
      description: Describes the resources, including ML compute instances and ML
        storage volumes, to use for model training.
      name: resourceConfig
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: roleARN
      required: false
      type: string
    - contains: string
      contains_description: null
      description: ''
      name: staticHyperParameters
      required: false
      type: object
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
      description: "Specifies a limit to how long a model training or compilation\
        \ job can run. It also specifies how long you are willing to wait for a managed\
        \ spot training job to complete. When the job reaches the time limit, Amazon\
        \ SageMaker ends the training or compilation job. Use this API to cap model\
        \ training costs. \n To stop a job, Amazon SageMaker sends the algorithm the\
        \ SIGTERM signal, which delays job termination for 120 seconds. Algorithms\
        \ can use this 120-second window to save the model artifacts, so the results\
        \ of training are not lost. \n The training algorithms provided by Amazon\
        \ SageMaker automatically save the intermediate results of a model training\
        \ job when possible. This attempt to save artifacts is only a best effort\
        \ case as model might not be in a state from which it can be saved. For example,\
        \ if training has just started, the model might not be ready to save. When\
        \ saved, this intermediate data is a valid model artifact. You can use it\
        \ to create a model with CreateModel. \n The Neural Topic Model (NTM) currently\
        \ does not support saving intermediate model artifacts. When training NTMs,\
        \ make sure that the maximum runtime is sufficient for the training job to\
        \ complete."
      name: stoppingCondition
      required: false
      type: object
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: metricName
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: type_
        required: false
        type: string
      contains_description: null
      description: Defines the objective metric for a hyperparameter tuning job. Hyperparameter
        tuning uses the value of this metric to evaluate the training jobs it launches,
        and returns the training job that results in either the highest or lowest
        value for this metric, depending on the value you specify for the Type parameter.
      name: tuningObjective
      required: false
      type: object
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
        access to. Control access to and from your training and model containers by
        configuring the VPC. For more information, see Protect Endpoints by Using
        an Amazon Virtual Private Cloud (https://docs.aws.amazon.com/sagemaker/latest/dg/host-vpc.html)
        and Protect Training Jobs by Using an Amazon Virtual Private Cloud (https://docs.aws.amazon.com/sagemaker/latest/dg/train-vpc.html).
      name: vpcConfig
      required: false
      type: object
    contains_description: null
    description: The HyperParameterTrainingJobDefinition object that describes the
      training jobs that this tuning job launches, including static hyperparameters,
      input data configuration, output data configuration, resource configuration,
      and stopping condition.
    name: trainingJobDefinition
    required: false
    type: object
  - contains:
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: algorithmName
        required: false
        type: string
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
        contains_description: Specifies a metric that the training algorithm writes
          to stderr or stdout . Amazon SageMakerhyperparameter tuning captures all
          defined metrics. You specify one metric that a hyperparameter tuning job
          uses as its objective metric to choose the best training job.
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
      description: Specifies which training algorithm to use for training jobs that
        a hyperparameter tuning job launches and the metrics to monitor.
      name: algorithmSpecification
      required: false
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
      description: Contains information about the output location for managed spot
        training checkpoint data.
      name: checkpointConfig
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: definitionName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: enableInterContainerTrafficEncryption
      required: false
      type: boolean
    - contains: null
      contains_description: null
      description: ''
      name: enableManagedSpotTraining
      required: false
      type: boolean
    - contains: null
      contains_description: null
      description: ''
      name: enableNetworkIsolation
      required: false
      type: boolean
    - contains:
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: name
          required: false
          type: string
        - contains: string
          contains_description: ''
          description: ''
          name: values
          required: false
          type: array
        contains_description: A list of categorical hyperparameters to tune.
        description: ''
        name: categoricalParameterRanges
        required: false
        type: array
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: maxValue
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: minValue
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: name
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: scalingType
          required: false
          type: string
        contains_description: A list of continuous hyperparameters to tune.
        description: ''
        name: continuousParameterRanges
        required: false
        type: array
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: maxValue
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: minValue
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: name
          required: false
          type: string
        - contains: null
          contains_description: null
          description: ''
          name: scalingType
          required: false
          type: string
        contains_description: For a hyperparameter of the integer type, specifies
          the range that a hyperparameter tuning job searches.
        description: ''
        name: integerParameterRanges
        required: false
        type: array
      contains_description: null
      description: "Specifies ranges of integer, continuous, and categorical hyperparameters\
        \ that a hyperparameter tuning job searches. The hyperparameter tuning job\
        \ launches training jobs with hyperparameter values within these ranges to\
        \ find the combination of values that result in the training job with the\
        \ best performance as measured by the objective metric of the hyperparameter\
        \ tuning job. \n You can specify a maximum of 20 hyperparameters that a hyperparameter\
        \ tuning job can search over. Every possible value of a categorical parameter\
        \ range counts against this limit."
      name: hyperParameterRanges
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
          description: Specifies a file system data source for a channel.
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
          description: Describes the S3 data source.
          name: s3DataSource
          required: false
          type: object
        contains_description: null
        description: Describes the location of the channel data.
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
        description: "A configuration for a shuffle option for input data in a channel.\
          \ If you use S3Prefix for S3DataType, the results of the S3 key prefix matches\
          \ are shuffled. If you use ManifestFile, the order of the S3 object references\
          \ in the ManifestFile is shuffled. If you use AugmentedManifestFile, the\
          \ order of the JSON lines in the AugmentedManifestFile is shuffled. The\
          \ shuffling order is determined using the Seed value. \n For Pipe input\
          \ mode, when ShuffleConfig is specified shuffling is done at the start of\
          \ every epoch. With large datasets, this ensures that the order of the training\
          \ data is different for each epoch, and it helps reduce bias and possible\
          \ overfitting. In a multi-node training job when ShuffleConfig is combined\
          \ with S3DataDistributionType of ShardedByS3Key, the data is shuffled across\
          \ nodes so that the content sent to a particular node on the first epoch\
          \ might be sent to a different node on the second epoch."
        name: shuffleConfig
        required: false
        type: object
      contains_description: A channel is a named input source that training algorithms
        can consume.
      description: ''
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
      description: Provides information about how to store model training results
        (model artifacts).
      name: outputDataConfig
      required: false
      type: object
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
      description: Describes the resources, including ML compute instances and ML
        storage volumes, to use for model training.
      name: resourceConfig
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: roleARN
      required: false
      type: string
    - contains: string
      contains_description: null
      description: ''
      name: staticHyperParameters
      required: false
      type: object
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
      description: "Specifies a limit to how long a model training or compilation\
        \ job can run. It also specifies how long you are willing to wait for a managed\
        \ spot training job to complete. When the job reaches the time limit, Amazon\
        \ SageMaker ends the training or compilation job. Use this API to cap model\
        \ training costs. \n To stop a job, Amazon SageMaker sends the algorithm the\
        \ SIGTERM signal, which delays job termination for 120 seconds. Algorithms\
        \ can use this 120-second window to save the model artifacts, so the results\
        \ of training are not lost. \n The training algorithms provided by Amazon\
        \ SageMaker automatically save the intermediate results of a model training\
        \ job when possible. This attempt to save artifacts is only a best effort\
        \ case as model might not be in a state from which it can be saved. For example,\
        \ if training has just started, the model might not be ready to save. When\
        \ saved, this intermediate data is a valid model artifact. You can use it\
        \ to create a model with CreateModel. \n The Neural Topic Model (NTM) currently\
        \ does not support saving intermediate model artifacts. When training NTMs,\
        \ make sure that the maximum runtime is sufficient for the training job to\
        \ complete."
      name: stoppingCondition
      required: false
      type: object
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: metricName
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: type_
        required: false
        type: string
      contains_description: null
      description: Defines the objective metric for a hyperparameter tuning job. Hyperparameter
        tuning uses the value of this metric to evaluate the training jobs it launches,
        and returns the training job that results in either the highest or lowest
        value for this metric, depending on the value you specify for the Type parameter.
      name: tuningObjective
      required: false
      type: object
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
        access to. Control access to and from your training and model containers by
        configuring the VPC. For more information, see Protect Endpoints by Using
        an Amazon Virtual Private Cloud (https://docs.aws.amazon.com/sagemaker/latest/dg/host-vpc.html)
        and Protect Training Jobs by Using an Amazon Virtual Private Cloud (https://docs.aws.amazon.com/sagemaker/latest/dg/train-vpc.html).
      name: vpcConfig
      required: false
      type: object
    contains_description: Defines the training jobs launched by a hyperparameter tuning
      job.
    description: A list of the HyperParameterTrainingJobDefinition objects launched
      for this tuning job.
    name: trainingJobDefinitions
    required: false
    type: array
  - contains:
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: hyperParameterTuningJobName
        required: false
        type: string
      contains_description: A previously completed or stopped hyperparameter tuning
        job to be used as a starting point for a new hyperparameter tuning job.
      description: ''
      name: parentHyperParameterTuningJobs
      required: false
      type: array
    - contains: null
      contains_description: null
      description: ''
      name: warmStartType
      required: false
      type: string
    contains_description: null
    description: "Specifies the configuration for starting the hyperparameter tuning\
      \ job using one or more previous tuning jobs as a starting point. The results\
      \ of previous tuning jobs are used to inform which combinations of hyperparameters\
      \ to search over in the new tuning job. \n All training jobs launched by the\
      \ new hyperparameter tuning job are evaluated by using the objective metric.\
      \ If you specify IDENTICAL_DATA_AND_ALGORITHM as the WarmStartType value for\
      \ the warm start configuration, the training job that performs the best in the\
      \ new tuning job is compared to the best training jobs from the parent tuning\
      \ jobs. From these, the training job that performs the best as measured by the\
      \ objective metric is returned as the overall best training job. \n All training\
      \ jobs launched by parent hyperparameter tuning jobs and the new hyperparameter\
      \ tuning jobs count against the limit of training jobs for the tuning job."
    name: warmStartConfig
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
      name: creationTime
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: failureReason
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: metricName
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: type_
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: value
        required: false
        type: number
      contains_description: null
      description: Shows the final value for the objective metric for a training job
        that was launched by a hyperparameter tuning job. You define the objective
        metric in the HyperParameterTuningJobObjective parameter of HyperParameterTuningJobConfig.
      name: finalHyperParameterTuningJobObjectiveMetric
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: objectiveStatus
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingEndTime
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingJobARN
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingJobDefinitionName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingJobName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingJobStatus
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingStartTime
      required: false
      type: string
    - contains: string
      contains_description: null
      description: ''
      name: tunedHyperParameters
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: tuningJobName
      required: false
      type: string
    contains_description: null
    description: A TrainingJobSummary object that describes the training job that
      completed with the best current HyperParameterTuningJobObjective.
    name: bestTrainingJob
    required: false
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
    description: If the tuning job failed, the reason it failed.
    name: failureReason
    required: false
    type: string
  - contains: null
    contains_description: null
    description: 'The status of the tuning job: InProgress, Completed, Failed, Stopping,
      or Stopped.'
    name: hyperParameterTuningJobStatus
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
      name: failureReason
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: metricName
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: type_
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: value
        required: false
        type: number
      contains_description: null
      description: Shows the final value for the objective metric for a training job
        that was launched by a hyperparameter tuning job. You define the objective
        metric in the HyperParameterTuningJobObjective parameter of HyperParameterTuningJobConfig.
      name: finalHyperParameterTuningJobObjectiveMetric
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: objectiveStatus
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingEndTime
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingJobARN
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingJobDefinitionName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingJobName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingJobStatus
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: trainingStartTime
      required: false
      type: string
    - contains: string
      contains_description: null
      description: ''
      name: tunedHyperParameters
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: tuningJobName
      required: false
      type: string
    contains_description: null
    description: If the hyperparameter tuning job is an warm start tuning job with
      a WarmStartType of IDENTICAL_DATA_AND_ALGORITHM, this is the TrainingJobSummary
      for the training job with the best objective metric value of all training jobs
      launched by this tuning job and all parent jobs specified for the warm start
      tuning job.
    name: overallBestTrainingJob
    required: false
    type: object
---
{% include "reference.md" %}
