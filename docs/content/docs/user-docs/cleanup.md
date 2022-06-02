---
title: "Uninstall an ACK Controller"
description: "Uninstall an ACK Controller"
lead: ""
draft: false
menu:
  docs:
    parent: "getting-started"
weight: 100
toc: true
---

Use the `helm uninstall` command to uninstall an ACK service controller:
```bash
export SERVICE=s3

# Uninstall the ACK service controller with Helm
helm uninstall -n $ACK_SYSTEM_NAMESPACE ack-$SERVICE-controller
```

## Delete CRDs

### Delete individual CRDS

If you have multiple controllers installed and only want to delete CRDs related to a specific resource, use the `kubectl delete` command to delete the CRDs with the the service name prefix.

For example, use the following commands to delete the CRD for Amazon S3 Buckets:
```bash
export SERVICE=s3
export CHART_EXPORT_PATH=/tmp/chart

# Delete an individual CRD
kubectl delete -f $CHART_EXPORT_PATH/$SERVICE-chart/crds/s3.services.k8s.aws_buckets.yaml
```

{{% hint type="warning" title="Check for CRDs that are common across services" %}}
There are a few custom resource definitions (CRDs) that are common across services. If you have multiple controllers installed, you should not delete the common CRDs unless you are uninstalling all of the controllers.
{{% /hint %}}

### Delete all CRDs

If you are sure that you would like to delete all CRDs, use the following commands:
```bash
export SERVICE=s3
export CHART_EXPORT_PATH=/tmp/chart

# Delete all CRDs
kubectl delete -f $CHART_EXPORT_PATH/$SERVICE-chart/crds
```

## Verify Helm charts were deleted

Verify that the Helm chart for your ACK service controller was deleted with the following command:
```bash
helm ls -n $ACK_SYSTEM_NAMESPACE
```

## Delete namespaces

Delete a specified namespace with the `kubectl delete namespace` command:
```bash
kubectl delete namespace $ACK_SYSTEM_NAMESPACE
```

## Delete ConfigMap

If you used [cross account resource management][carm-docs], delete the `ConfigMap` you created.
```bash
kubectl delete -n ack-system configmap ack-role-account-map
```

[carm-docs]: ../cross-account-resource-management/

