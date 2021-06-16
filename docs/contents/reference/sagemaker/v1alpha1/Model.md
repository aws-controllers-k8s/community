---
resource:
  apiVersion: v1alpha1
  description: ModelSpec defines the desired state of Model.
  group: sagemaker.services.k8s.aws
  name: Model
  names:
    kind: Model
    listKind: ModelList
    plural: models
    singular: model
  scope: Namespaced
  service: sagemaker
  spec:
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: containerHostname
      required: false
      type: string
    - contains: string
      contains_description: null
      description: ''
      name: environment
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: image
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: repositoryAccessMode
        required: false
        type: string
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: repositoryCredentialsProviderARN
          required: false
          type: string
        contains_description: null
        description: Specifies an authentication configuration for the private docker
          registry where your model image is hosted. Specify a value for this property
          only if you specified Vpc as the value for the RepositoryAccessMode field
          of the ImageConfig object that you passed to a call to CreateModel and the
          private Docker registry where the model image is hosted requires authentication.
        name: repositoryAuthConfig
        required: false
        type: object
      contains_description: null
      description: Specifies whether the model container is in Amazon ECR or a private
        Docker registry accessible from your Amazon Virtual Private Cloud (VPC).
      name: imageConfig
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: mode
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: modelDataURL
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: modelPackageName
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: modelCacheSetting
        required: false
        type: string
      contains_description: null
      description: Specifies additional configuration for hosting multi-model endpoints.
      name: multiModelConfig
      required: false
      type: object
    contains_description: Describes the container, as part of model definition.
    description: Specifies the containers in the inference pipeline.
    name: containers
    required: false
    type: array
  - contains: null
    contains_description: null
    description: Isolates the model container. No inbound or outbound network calls
      can be made to or from the model container.
    name: enableNetworkIsolation
    required: false
    type: boolean
  - contains: null
    contains_description: null
    description: "The Amazon Resource Name (ARN) of the IAM role that Amazon SageMaker\
      \ can assume to access model artifacts and docker image for deployment on ML\
      \ compute instances or for batch transform jobs. Deploying on ML compute instances\
      \ is part of model hosting. For more information, see Amazon SageMaker Roles\
      \ (https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-roles.html). \n\
      \ To be able to pass this role to Amazon SageMaker, the caller of this API must\
      \ have the iam:PassRole permission."
    name: executionRoleARN
    required: true
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: mode
      required: false
      type: string
    contains_description: null
    description: Specifies details of how containers in a multi-container endpoint
      are called.
    name: inferenceExecutionConfig
    required: false
    type: object
  - contains: null
    contains_description: null
    description: The name of the new model.
    name: modelName
    required: true
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: containerHostname
      required: false
      type: string
    - contains: string
      contains_description: null
      description: ''
      name: environment
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: image
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: repositoryAccessMode
        required: false
        type: string
      - contains:
        - contains: null
          contains_description: null
          description: ''
          name: repositoryCredentialsProviderARN
          required: false
          type: string
        contains_description: null
        description: Specifies an authentication configuration for the private docker
          registry where your model image is hosted. Specify a value for this property
          only if you specified Vpc as the value for the RepositoryAccessMode field
          of the ImageConfig object that you passed to a call to CreateModel and the
          private Docker registry where the model image is hosted requires authentication.
        name: repositoryAuthConfig
        required: false
        type: object
      contains_description: null
      description: Specifies whether the model container is in Amazon ECR or a private
        Docker registry accessible from your Amazon Virtual Private Cloud (VPC).
      name: imageConfig
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: mode
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: modelDataURL
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: modelPackageName
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: modelCacheSetting
        required: false
        type: string
      contains_description: null
      description: Specifies additional configuration for hosting multi-model endpoints.
      name: multiModelConfig
      required: false
      type: object
    contains_description: null
    description: The location of the primary docker image containing inference code,
      associated artifacts, and custom environment map that the inference code uses
      when the model is deployed for predictions.
    name: primaryContainer
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
    description: A VpcConfig object that specifies the VPC that you want your model
      to connect to. Control access to and from your model container by configuring
      the VPC. VpcConfig is used in hosting services and in batch transform. For more
      information, see Protect Endpoints by Using an Amazon Virtual Private Cloud
      (https://docs.aws.amazon.com/sagemaker/latest/dg/host-vpc.html) and Protect
      Data in Batch Transform Jobs by Using an Amazon Virtual Private Cloud (https://docs.aws.amazon.com/sagemaker/latest/dg/batch-vpc.html).
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
---
{% include "reference.md" %}
