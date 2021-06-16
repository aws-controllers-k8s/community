---
services:
- name: apigatewayv2
  resources:
  - apiVersion: v1alpha1
    name: Model
  - apiVersion: v1alpha1
    name: Deployment
  - apiVersion: v1alpha1
    name: RouteResponse
  - apiVersion: v1alpha1
    name: Authorizer
  - apiVersion: v1alpha1
    name: VPCLink
  - apiVersion: v1alpha1
    name: IntegrationResponse
  - apiVersion: v1alpha1
    name: DomainName
  - apiVersion: v1alpha1
    name: Stage
  - apiVersion: v1alpha1
    name: Route
  - apiVersion: v1alpha1
    name: API
  - apiVersion: v1alpha1
    name: Integration
  - apiVersion: v1alpha1
    name: APIMapping
- name: sagemaker
  resources:
  - apiVersion: v1alpha1
    name: EndpointConfig
  - apiVersion: v1alpha1
    name: ModelQualityJobDefinition
  - apiVersion: v1alpha1
    name: DataQualityJobDefinition
  - apiVersion: v1alpha1
    name: TransformJob
  - apiVersion: v1alpha1
    name: ModelBiasJobDefinition
  - apiVersion: v1alpha1
    name: Endpoint
  - apiVersion: v1alpha1
    name: MonitoringSchedule
  - apiVersion: v1alpha1
    name: Model
  - apiVersion: v1alpha1
    name: ProcessingJob
  - apiVersion: v1alpha1
    name: ModelExplainabilityJobDefinition
  - apiVersion: v1alpha1
    name: HyperParameterTuningJob
  - apiVersion: v1alpha1
    name: TrainingJob
- name: applicationautoscaling
  resources:
  - apiVersion: v1alpha1
    name: ScheduledAction
  - apiVersion: v1alpha1
    name: ScalingPolicy
  - apiVersion: v1alpha1
    name: ScalableTarget
- name: elasticache
  resources:
  - apiVersion: v1alpha1
    name: CacheParameterGroup
  - apiVersion: v1alpha1
    name: ReplicationGroup
  - apiVersion: v1alpha1
    name: Snapshot
  - apiVersion: v1alpha1
    name: CacheSubnetGroup
- name: mq
  resources:
  - apiVersion: v1alpha1
    name: Broker
- name: rds
  resources:
  - apiVersion: v1alpha1
    name: DBSubnetGroup
  - apiVersion: v1alpha1
    name: DBParameterGroup
  - apiVersion: v1alpha1
    name: DBSecurityGroup
- name: s3
  resources:
  - apiVersion: v1alpha1
    name: Bucket
---
{% include "reference_overview.md" %}
