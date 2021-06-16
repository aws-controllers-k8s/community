---
resource:
  apiVersion: v1alpha1
  description: "EndpointSpec defines the desired state of Endpoint. \n A hosted endpoint\
    \ for real-time inference."
  group: sagemaker.services.k8s.aws
  name: Endpoint
  names:
    kind: Endpoint
    listKind: EndpointList
    plural: endpoints
    singular: endpoint
  scope: Namespaced
  service: sagemaker
  spec:
  - contains: null
    contains_description: null
    description: The name of an endpoint configuration. For more information, see
      CreateEndpointConfig.
    name: endpointConfigName
    required: true
    type: string
  - contains: null
    contains_description: null
    description: The name of the endpoint.The name must be unique within an AWS Region
      in your AWS account. The name is case-insensitive in CreateEndpoint, but the
      case is preserved and must be matched in .
    name: endpointName
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
    description: A timestamp that shows when the endpoint was created.
    name: creationTime
    required: false
    type: string
  - contains: null
    contains_description: null
    description: "The status of the endpoint. \n    * OutOfService: Endpoint is not\
      \ available to take incoming requests. \n    * Creating: CreateEndpoint is executing.\
      \ \n    * Updating: UpdateEndpoint or UpdateEndpointWeightsAndCapacities is\
      \ executing. \n    * SystemUpdating: Endpoint is undergoing maintenance and\
      \ cannot be updated    or deleted or re-scaled until it has completed. This\
      \ maintenance operation    does not change any customer-specified values such\
      \ as VPC config, KMS    encryption, model, instance type, or instance count.\
      \ \n    * RollingBack: Endpoint fails to scale up or down or change its variant\
      \    weight and is in the process of rolling back to its previous configuration.\
      \    Once the rollback completes, endpoint returns to an InService status. \
      \   This transitional status only applies to an endpoint that has autoscaling\
      \    enabled and is undergoing variant weight or capacity changes as part of\
      \    an UpdateEndpointWeightsAndCapacities call or when the UpdateEndpointWeightsAndCapacities\
      \    operation is called explicitly. \n    * InService: Endpoint is available\
      \ to process incoming requests. \n    * Deleting: DeleteEndpoint is executing.\
      \ \n    * Failed: Endpoint could not be created, updated, or re-scaled. Use\
      \ DescribeEndpointOutput$FailureReason    for information about the failure.\
      \ DeleteEndpoint is the only operation    that can be performed on a failed\
      \ endpoint."
    name: endpointStatus
    required: false
    type: string
  - contains: null
    contains_description: null
    description: If the status of the endpoint is Failed, the reason why it failed.
    name: failureReason
    required: false
    type: string
  - contains: null
    contains_description: null
    description: Name of the Amazon SageMaker endpoint configuration.
    name: lastEndpointConfigNameForUpdate
    required: false
    type: string
  - contains: null
    contains_description: null
    description: A timestamp that shows when the endpoint was last modified.
    name: lastModifiedTime
    required: false
    type: string
  - contains: null
    contains_description: null
    description: The name of the endpoint configuration associated with this endpoint.
    name: latestEndpointConfigName
    required: false
    type: string
  - contains:
    - contains: null
      contains_description: null
      description: ''
      name: currentInstanceCount
      required: false
      type: integer
    - contains: null
      contains_description: null
      description: ''
      name: currentWeight
      required: false
      type: number
    - contains:
      - contains: null
        contains_description: null
        description: ''
        name: resolutionTime
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: resolvedImage
        required: false
        type: string
      - contains: null
        contains_description: null
        description: ''
        name: specifiedImage
        required: false
        type: string
      contains_description: "Gets the Amazon EC2 Container Registry path of the docker\
        \ image of the model that is hosted in this ProductionVariant. \n If you used\
        \ the registry/repository[:tag] form to specify the image path of the primary\
        \ container when you created the model hosted in this ProductionVariant, the\
        \ path resolves to a path of the form registry/repository[@digest]. A digest\
        \ is a hash value that identifies a specific version of an image. For information\
        \ about Amazon ECR paths, see Pulling an Image (https://docs.aws.amazon.com/AmazonECR/latest/userguide/docker-pull-ecr-image.html)\
        \ in the Amazon ECR User Guide."
      description: ''
      name: deployedImages
      required: false
      type: array
    - contains: null
      contains_description: null
      description: ''
      name: desiredInstanceCount
      required: false
      type: integer
    - contains: null
      contains_description: null
      description: ''
      name: desiredWeight
      required: false
      type: number
    - contains: null
      contains_description: null
      description: ''
      name: variantName
      required: false
      type: string
    contains_description: Describes weight and capacities for a production variant
      associated with an endpoint. If you sent a request to the UpdateEndpointWeightsAndCapacities
      API and the endpoint status is Updating, you get different desired and current
      values.
    description: An array of ProductionVariantSummary objects, one for each model
      hosted behind this endpoint.
    name: productionVariants
    required: false
    type: array
---
{% include "reference.md" %}
