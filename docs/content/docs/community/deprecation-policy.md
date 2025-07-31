---
title: "Deprecation Policy"
description: "ACK deprecation policy and current deprecations dashboard"
lead: "Understanding ACK's deprecation process and timeline"
draft: false
menu:
  docs:
    parent: "introduction"
weight: 15
toc: true
---

## Deprecation Policy

The AWS Controllers for Kubernetes (ACK) project follows a structured deprecation process to ensure users have adequate time to migrate to new features.

### Deprecation Timeline

ACK follows a structured 3-month deprecation process to ensure users have adequate time for migration:

| Phase | Duration | Description | Actions |
|-------|----------|-------------|---------|
| **Announcement** | Month 1 | Official deprecation announcement | • Public announcement via GitHub issue<br>• Documentation updated with deprecation notices<br>• Migration guidance provided<br> |
| **Deprecation** | Months 1-3 | Feature marked deprecated with warnings | • Feature turned off by default<br>• Alternative features promoted to higher stability<br>• Community migration support provided |
| **Removal** | End of Month 3 | Complete feature removal | • Feature completely removed from codebase<br>• CRDs and related resources deleted<br>• Alternative features reach production readiness<br>• Final migration deadline |

## Current Deprecations

{{% hint type="warning" title="Deprecation Notice" %}}
The following features are **deprecated** and will be **removed on October 31, 2024**:
{{% /hint %}}

| Feature | Status | Replacement |
|---------|---------|-------------|
| **FieldExport CRD** | Deprecated<br>**Removal: Oct 31, 2024** | [KRO (Kubernetes Resource Orchestrator)](https://kro.run/docs/overview/) |
| **AdoptedResource CRD** | Deprecated<br>**Removal: Oct 31, 2024** | [ResourceAdoption via annotations](../user-docs/features/#resourceadoption) |


### Feature Graduations

| Feature | Current Status | Next Milestone | Info |
|---------|---------------|----------------|---------|
| **ResourceAdoption** | Beta (July 2025) | **GA: October 2025** | **Replacement for AdoptedResource CRD** 
| **ReadOnlyResources** | Beta (July 2025) | **GA: October 2025** | -


## Getting Help

If you need assistance with migration:

1. Check the [ACK Community Discussions](https://github.com/aws-controllers-k8s/community/discussions)
2. File an issue in the [ACK Community repository](https://github.com/aws-controllers-k8s/community/issues)

