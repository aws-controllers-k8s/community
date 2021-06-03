---
services:
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
- name: s3
  resources:
  - apiVersion: v1alpha1
    name: Bucket
---
{% include "reference_overview.md" %}