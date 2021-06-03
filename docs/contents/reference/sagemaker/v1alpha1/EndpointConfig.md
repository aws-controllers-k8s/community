---
resource:
  apiVersion: v1alpha1
  description: EndpointConfig is the Schema for the EndpointConfigs API
  group: sagemaker.services.k8s.aws
  name: EndpointConfig
  names:
    kind: EndpointConfig
    listKind: EndpointConfigList
    plural: endpointconfigs
    singular: endpointconfig
  scope: Namespaced
  service: sagemaker
  spec:
  - contains:
    - contains:
      - contains: string
        contains_description: ''
        description: ''
        name: csvContentTypes
        required: false
        type: array
      - contains: string
        contains_description: ''
        description: ''
        name: jsonContentTypes
        required: false
        type: array
      contains_description: null
      description: ''
      name: captureContentTypeHeader
      required: false
      type: object
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: captureMode
        required: false
        type: string
      contains_description: ''
      description: ''
      name: captureOptions
      required: false
      type: array
    - contains: null
      contains_description: null
      description: ''
      name: destinationS3URI
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: enableCapture
      required: false
      type: boolean
    - contains: null
      contains_description: null
      description: ''
      name: initialSamplingPercentage
      required: false
      type: integer
    - contains: null
      contains_description: null
      description: ''
      name: kmsKeyID
      required: false
      type: string
    contains_description: null
    description: ''
    name: dataCaptureConfig
    required: false
    type: object
  - contains: null
    contains_description: null
    description: The name of the endpoint configuration. You specify this name in
      a CreateEndpoint request.
    name: endpointConfigName
    required: true
    type: string
  - contains: null
    contains_description: null
    description: "The Amazon Resource Name (ARN) of a AWS Key Management Service key\
      \ that Amazon SageMaker uses to encrypt data on the storage volume attached\
      \ to the ML compute instance that hosts the endpoint. \n The KmsKeyId can be\
      \ any of the following formats: \n    * Key ID: 1234abcd-12ab-34cd-56ef-1234567890ab\
      \ \n    * Key ARN: arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab\
      \ \n    * Alias name: alias/ExampleAlias \n    * Alias name ARN: arn:aws:kms:us-west-2:111122223333:alias/ExampleAlias\
      \ \n The KMS key policy must grant permission to the IAM role that you specify\
      \ in your CreateEndpoint, UpdateEndpoint requests. For more information, refer\
      \ to the AWS Key Management Service section Using Key Policies in AWS KMS (https://docs.aws.amazon.com/kms/latest/developerguide/key-policies.html)\
      \ \n Certain Nitro-based instances include local storage, dependent on the instance\
      \ type. Local storage volumes are encrypted using a hardware module on the instance.\
      \ You can't request a KmsKeyId when using an instance type with local storage.\
      \ If any of the models that you specify in the ProductionVariants parameter\
      \ use nitro-based instances with local storage, do not specify a value for the\
      \ KmsKeyId parameter. If you specify a value for KmsKeyId when using any nitro-based\
      \ instances with local storage, the call to CreateEndpointConfig fails. \n For\
      \ a list of instance types that support local instance storage, see Instance\
      \ Store Volumes (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/InstanceStorage.html#instance-store-volumes).\
      \ \n For more information about local instance storage encryption, see SSD Instance\
      \ Store Volumes (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ssd-instance-store.html)."
    name: kmsKeyID
    required: false
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: acceleratorType
      required: false
      type: string
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: destinationS3URI
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: kmsKeyID
        required: false
        type: string
      contains_description: null
      description: ''
      name: coreDumpConfig
      required: false
      type: object
    - contains: null
      contains_description: null
      description: ''
      name: initialInstanceCount
      required: false
      type: integer
    - contains: null
      contains_description: null
      description: ''
      name: initialVariantWeight
      required: false
      type: number
    - contains: null
      contains_description: null
      description: ''
      name: instanceType
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: modelName
      required: false
      type: string
    - contains: null
      contains_description: null
      description: ''
      name: variantName
      required: false
      type: string
    contains_description: ''
    description: An list of ProductionVariant objects, one for each model that you
      want to host at this endpoint.
    name: productionVariants
    required: true
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
