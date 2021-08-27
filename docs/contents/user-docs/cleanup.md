# Cleanup 

Few custom resource definitions (CRDs) are common across services. If you have multiple controllers installed, you should not delete the common CRDs unless you are uninstalling all of the controllers.

## Uninstall ACK service controllers and CRDs

Use the following commands to uninstall ACK service controllers and delete CRDs:
```bash
export SERVICE=sagemaker
export CHART_EXPORT_PATH=/tmp/chart

# Uninstall the ACK service controller with Helm
helm uninstall -n $ACK_K8S_NAMESPACE ack-$SERVICE-controller

# Delete the CRDs
kubectl delete -f $CHART_EXPORT_PATH/ack-$SERVICE-controller/crds
```

If you have multiple controllers installed and only want to delete CRDs related to a specific resource, delete the CRDs with the the service name prefix. 

```bash
cd $CHART_EXPORT_PATH/ack-$SERVICE-controller/crds

$ ls
sagemaker.services.k8s.aws_dataqualityjobdefinitions.yaml
sagemaker.services.k8s.aws_endpointconfigs.yaml
sagemaker.services.k8s.aws_endpoints.yaml
sagemaker.services.k8s.aws_hyperparametertuningjobs.yaml
sagemaker.services.k8s.aws_modelbiasjobdefinitions.yaml
sagemaker.services.k8s.aws_modelexplainabilityjobdefinitions.yaml
sagemaker.services.k8s.aws_modelqualityjobdefinitions.yaml
sagemaker.services.k8s.aws_models.yaml
sagemaker.services.k8s.aws_monitoringschedules.yaml
sagemaker.services.k8s.aws_processingjobs.yaml
sagemaker.services.k8s.aws_trainingjobs.yaml
sagemaker.services.k8s.aws_transformjobs.yaml
services.k8s.aws_adoptedresources.yaml. # -> Common CRD across services
```
By modifying the variable values as needed, these steps can be applied for the deletion of other ACK service controllers and CRDs.

## Verify Helm charts were deleted and delete namespace

Verify that the Helm chart for your ACK service controller was deleted with the following command:
```bash
helm ls -n $ACK_K8S_NAMESPACE
```

Delete the namespace. 
```
kubectl delete namespace $ACK_K8S_NAMESPACE
```

If you used [cross account resource management][carm-docs], delete the `ConfigMap` you created. 
```bash
kubectl delete -n ack-system configmap ack-role-account-map
kubectl delete namespace production
```

[carm-docs]: https://aws-controllers-k8s.github.io/community/user-docs/cross-account-resource-management/

