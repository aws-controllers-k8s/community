---
resource:
  apiVersion: v1alpha1
  description: "TransformJobSpec defines the desired state of TransformJob. \n A batch\
    \ transform job. For information about SageMaker batch transform, see Use Batch\
    \ Transform (https://docs.aws.amazon.com/sagemaker/latest/dg/batch-transform.html)."
  group: sagemaker.services.k8s.aws
  name: TransformJob
  names:
    kind: TransformJob
    listKind: TransformJobList
    plural: transformjobs
    singular: transformjob
  scope: Namespaced
  service: sagemaker
  spec:
  - contains: null
    contains_description: null
    description: "Specifies the number of records to include in a mini-batch for an\
      \ HTTP inference request. A record is a single unit of input data that inference\
      \ can be made on. For example, a single line in a CSV file is a record. \n To\
      \ enable the batch strategy, you must set the SplitType property to Line, RecordIO,\
      \ or TFRecord. \n To use only one record when making an HTTP invocation request\
      \ to a container, set BatchStrategy to SingleRecord and SplitType to Line. \n\
      \ To fit as many records in a mini-batch as can fit within the MaxPayloadInMB\
      \ limit, set BatchStrategy to MultiRecord and SplitType to Line."
    name: batchStrategy
    required: false
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: inputFilter
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: joinSource
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: outputFilter
      required: false
      type: string
    contains_description: null
    description: The data structure used to specify the data to be used for inference
      in a batch transform job and to associate the data that is relevant to the prediction
      results in the output. The input filter provided allows you to exclude input
      data that is not needed for inference in a batch transform job. The output filter
      provided allows you to include input data relevant to interpreting the predictions
      in the output from the job. For more information, see Associate Prediction Results
      with their Corresponding Input Records (https://docs.aws.amazon.com/sagemaker/latest/dg/batch-transform-data-processing.html).
    name: dataProcessing
    required: false
    type: object
  - contains: string
    contains_description: null
    description: The environment variables to set in the Docker container. We support
      up to 16 key and values entries in the map.
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
    description: "Associates a SageMaker job as a trial component with an experiment\
      \ and trial. Specified when you call the following APIs: \n    * CreateProcessingJob\
      \ \n    * CreateTrainingJob \n    * CreateTransformJob"
    name: experimentConfig
    required: false
    type: object
  - contains: null
    contains_description: null
    description: The maximum number of parallel requests that can be sent to each
      instance in a transform job. If MaxConcurrentTransforms is set to 0 or left
      unset, Amazon SageMaker checks the optional execution-parameters to determine
      the settings for your chosen algorithm. If the execution-parameters endpoint
      is not enabled, the default value is 1. For more information on execution-parameters,
      see How Containers Serve Requests (https://docs.aws.amazon.com/sagemaker/latest/dg/your-algorithms-batch-code.html#your-algorithms-batch-code-how-containe-serves-requests).
      For built-in algorithms, you don't need to set a value for MaxConcurrentTransforms.
    name: maxConcurrentTransforms
    required: false
    type: integer
  - contains: null
    contains_description: null
    description: "The maximum allowed size of the payload, in MB. A payload is the\
      \ data portion of a record (without metadata). The value in MaxPayloadInMB must\
      \ be greater than, or equal to, the size of a single record. To estimate the\
      \ size of a record in MB, divide the size of your dataset by the number of records.\
      \ To ensure that the records fit within the maximum payload size, we recommend\
      \ using a slightly larger value. The default value is 6 MB. \n For cases where\
      \ the payload might be arbitrarily large and is transmitted using HTTP chunked\
      \ encoding, set the value to 0. This feature works only in supported algorithms.\
      \ Currently, Amazon SageMaker built-in algorithms do not support HTTP chunked\
      \ encoding."
    name: maxPayloadInMB
    required: false
    type: integer
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: invocationsMaxRetries
      required: false
      type: integer
    - contains: null
      contains_description: null
      description: ''
      name: invocationsTimeoutInSeconds
      required: false
      type: integer
    contains_description: null
    description: Configures the timeout and maximum number of retries for processing
      a transform job invocation.
    name: modelClientConfig
    required: false
    type: object
  - contains: null
    contains_description: null
    description: The name of the model that you want to use for the transform job.
      ModelName must be the name of an existing Amazon SageMaker model within an AWS
      Region in an AWS account.
    name: modelName
    required: true
    type: string
  - contains:
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
      name: splitType
      required: false
      type: string
    contains_description: null
    description: Describes the input source and the way the transform job consumes
      it.
    name: transformInput
    required: true
    type: object
  - contains: null
    contains_description: null
    description: The name of the transform job. The name must be unique within an
      AWS Region in an AWS account.
    name: transformJobName
    required: true
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: accept
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: assembleWith
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
      name: s3OutputPath
      required: false
      type: string
    contains_description: null
    description: Describes the results of the transform job.
    name: transformOutput
    required: true
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
    contains_description: null
    description: Describes the resources, including ML instance types and ML instance
      count, to use for the transform job.
    name: transformResources
    required: true
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
    description: If the transform job failed, FailureReason describes why it failed.
      A transform job creates a log file, which includes error messages, and stores
      it as an Amazon S3 object. For more information, see Log Amazon SageMaker Events
      with Amazon CloudWatch (https://docs.aws.amazon.com/sagemaker/latest/dg/logging-cloudwatch.html).
    name: failureReason
    required: false
    type: string
  - contains: null
    contains_description: null
    description: The status of the transform job. If the transform job failed, the
      reason is returned in the FailureReason field.
    name: transformJobStatus
    required: false
    type: string
---
{% include "reference.md" %}
