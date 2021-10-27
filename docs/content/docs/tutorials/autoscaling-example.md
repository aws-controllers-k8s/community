---
title: "Scale SageMaker Workloads with Application Auto Scaling"
description: "Scale a SageMaker endpoint with the ACK Application Auto Scaling service controller"
lead: "Scale a SageMaker endpoint with the ACK Application Auto Scaling service controller"
draft: false
menu: 
  docs:
    parent: "tutorials"
weight: 41
toc: true
---

The Application Auto Scaling ACK service controller makes it easier for developers to automatically scale resources for individual AWS services. Application Auto Scaling allows you to configure automatic scaling for resources such as Amazon SageMaker endpoint variants. 

In this tutorial, we will use the Application Auto Scaling ACK service controller in conjunction with the SageMaker ACK service controller to automatically scale a deployed machine learning model. 

## Setup

Although it is not necessary to use Amazon Elastic Kubernetes Service (Amazon EKS) with ACK, this guide assumes that you have access to an Amazon EKS cluster. If this is your first time creating an Amazon EKS cluster, see [Amazon EKS Setup](https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/deep-learning-containers-eks-setup.html). For automated cluster creation using `eksctl`, see [Getting started with Amazon EKS - `eksctl`](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html) and create your cluster with Amazon EC2 Linux managed nodes.

This guide also assumes that you have a trained machine learning model that you are ready to dynamically scale with the Application Auto Scaling ACK service controller. To train a machine learning model using the SageMaker ACK service controller, see [Machine Learning with the ACK Service Controller](../sagemaker-example/) and return to this guide when you have successfully completed a SageMaker training job. 

### Prerequisites

This guide assumes that you have:
  - Created an EKS cluster with Kubernetes version 1.16 or higher. 
  - AWS IAM permissions to create roles and attach policies to roles.
  - A trained machine learning model that you want to scale dynamically. 
  - Installed the following tools on the client machine used to access your Kubernetes cluster:
    - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting with AWS services. 
    - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working with Kubernetes clusters. 
    - [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html) - A command line tool for working with EKS clusters.
    - [yq](https://mikefarah.gitbook.io/yq) - A command line tool for YAML processing. (For Linux environments, use the [`wget` plain binary installation](https://mikefarah.gitbook.io/yq/#wget))
    - [Helm 3.7+](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications.

### Configure IAM permissions

Create an IAM role and attach an IAM policy to that role to ensure that your Application Auto Scaling service controller has access to the appropriate AWS resources. First, check to make sure that you are connected to an Amazon EKS cluster. 

```bash
export CLUSTER_NAME=<CLUSTER_NAME>
export SERVICE_REGION=<CLUSTER_REGION>
aws eks update-kubeconfig --name $CLUSTER_NAME --region $SERVICE_REGION
kubectl config current-context
kubectl get nodes
```

Before you can deploy your ACK service controllers using an IAM role, associate an OpenID Connect (OIDC) provider with your IAM role to authenticate your cluster with the IAM service.

```bash
eksctl utils associate-iam-oidc-provider --cluster ${CLUSTER_NAME} \
--region ${SERVICE_REGION} --approve
```

Get the following OIDC information for future reference:

```bash
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
export OIDC_PROVIDER_URL=$(aws eks describe-cluster --name $CLUSTER_NAME --region $SERVICE_REGION \
--query "cluster.identity.oidc.issuer" --output text | cut -c9-)
```

In your working directory, create a file named `trust.json` using the following trust relationship code block:

```bash
printf '{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::'$AWS_ACCOUNT_ID':oidc-provider/'$OIDC_PROVIDER_URL'"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "'$OIDC_PROVIDER_URL':aud": "sts.amazonaws.com",
          "'$OIDC_PROVIDER_URL':sub": [
            "system:serviceaccount:ack-system:ack-sagemaker-controller",
            "system:serviceaccount:ack-system:ack-applicationautoscaling-controller"
          ]
        }
      }
    }
  ]
}
' > ./trust.json
```

Updating an Application Auto Scaling Scalable Target requires additional permissions. Create a file named `pass_role_policy.json` to create the policy required for the IAM role.

```bash
printf '{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "*"
    }
  ]
}
' > ./pass_role_policy.json
```

Run the `iam create-role` command to create an IAM role with the trust relationship you just defined in `trust.json`. This IAM role enables the Amazon EKS cluster to get and refresh credentials from IAM.

```bash
export OIDC_ROLE_NAME=ack-controller-role-$CLUSTER_NAME
aws --region $SERVICE_REGION iam create-role --role-name $OIDC_ROLE_NAME --assume-role-policy-document file://trust.json
```

Attach the AmazonSageMakerFullAccess Policy to the IAM Role to ensure that your SageMaker service controller has access to the appropriate resources. 

```bash
aws --region $SERVICE_REGION iam attach-role-policy --role-name $OIDC_ROLE_NAME --policy-arn arn:aws:iam::aws:policy/AmazonSageMakerFullAccess
```

Attach the `iam:PassRole` policy required for updating an Application Auto Scaling Scalable Target. 
```bash
aws iam put-role-policy --role-name $OIDC_ROLE_NAME --policy-name "iam-pass-role-policy" --policy-document file://pass_role_policy.json
```

Get the following IAM role information for future reference:
```bash
export IAM_ROLE_ARN_FOR_IRSA=$(aws --region $SERVICE_REGION iam get-role --role-name $OIDC_ROLE_NAME --output text --query 'Role.Arn')
echo $IAM_ROLE_ARN_FOR_IRSA
```

For more information on authorization and access for ACK service controllers, including details regarding recommended IAM policies, see [Configure Permissions][configure-permissions].

### Install the Application Auto Scaling ACK service controller

Get the Application Auto Scaling Helm chart and make it available on the client machine with the following commands:

```bash
export HELM_EXPERIMENTAL_OCI=1
export SERVICE=applicationautoscaling
export RELEASE_VERSION=v0.2.0
export CHART_EXPORT_PATH=/tmp/chart
export CHART_REF=$SERVICE-chart
export CHART_REPO=public.ecr.aws/aws-controllers-k8s/$CHART_REF
export CHART_PACKAGE=$CHART_REF-$RELEASE_VERSION.tgz

mkdir -p $CHART_EXPORT_PATH

helm pull oci://$CHART_REPO --version $RELEASE_VERSION -d $CHART_EXPORT_PATH
tar xvf $CHART_EXPORT_PATH/$CHART_PACKAGE -C $CHART_EXPORT_PATH
```

Update the Helm chart values for a cluster-scoped installation. 

```bash
# Update the following values in the Helm chart
cd $CHART_EXPORT_PATH/$SERVICE-chart
yq e '.aws.region = env(SERVICE_REGION)' -i values.yaml
yq e '.serviceAccount.annotations."eks.amazonaws.com/role-arn" = env(IAM_ROLE_ARN_FOR_IRSA)' -i values.yaml
cd -
```

Install the relevant custom resource definitions (CRDs) for the Application Auto Scaling ACK service controller. 

```bash
kubectl apply -f $CHART_EXPORT_PATH/$SERVICE-chart/crds
```

Create a namespace and install the Application Auto Scaling ACK service controller with the Helm chart. 

```bash
export ACK_K8S_NAMESPACE=ack-system
helm install -n $ACK_K8S_NAMESPACE --create-namespace --skip-crds ack-$SERVICE-controller \
 $CHART_EXPORT_PATH/$SERVICE-chart
 ```

Verify that the CRDs and Helm charts were deployed with the following commands:
```bash
kubectl get pods -A | grep applicationautoscaling
kubectl get crd | grep applicationautoscaling
```

To scale a SageMaker endpoint variant with the Application Auto Scaling ACK service controller, you will also need the SageMaker ACK service controller. For step-by-step installation instructions see [Install the SageMaker ACK Service Controller](../sagemaker-example/#install-the-sagemaker-ack-service-controller).

## Deploy a SageMaker endpoint

You can apply an autoscaling policy to an existing SageMaker endpoint or you can create one using the SageMaker ACK service controller. For this tutorial, we will create an endpoint and deploy a model based on the SageMaker training job created in the [ACK SageMaker Controller tutorial](../sagemaker-example/). For more information on this model, see [Train an XGBoost Model](../sagemaker-example/#train-an-xgboost-model). 

```bash
export RANDOM_VAR=$RANDOM
export MODEL_NAME=ack-xgboost-model-$RANDOM_VAR
export ENDPOINT_CONFIG_NAME=ack-xgboost-endpoint-config-$RANDOM_VAR
export ENDPOINT_NAME=ack-xgboost-endpoint-$RANDOM_VAR
```

Deploy the model on an `ml.t2.medium` instance using the following `deploy.yaml` file. This file uses the SageMaker ACK service controller to create a model, an endpoint configuration, and an endpoint. To use your own model, change the `modelDataURL` and `executionRoleARN` values. 
```bash
printf '
apiVersion: sagemaker.services.k8s.aws/v1alpha1
kind: Model
metadata:
  name: '$MODEL_NAME'
spec:
  modelName: '$MODEL_NAME'
  primaryContainer:
    containerHostname: xgboost
    # The source of the model data
    modelDataURL: s3://'$SAGEMAKER_BUCKET'/'$JOB_NAME'/output/model.tar.gz
    image: '$XGBOOST_IMAGE'
  executionRoleARN: '$SAGEMAKER_EXECUTION_ROLE_ARN'
---
apiVersion: sagemaker.services.k8s.aws/v1alpha1
kind: EndpointConfig
metadata:
  name: '$ENDPOINT_CONFIG_NAME'
spec:
  endpointConfigName: '$ENDPOINT_CONFIG_NAME'
  productionVariants:
  - modelName: '$MODEL_NAME'
    variantName: AllTraffic
    instanceType: ml.t2.medium
    initialInstanceCount: 1
---
apiVersion: sagemaker.services.k8s.aws/v1alpha1
kind: Endpoint
metadata:
  name: '$ENDPOINT_NAME'
spec:
  endpointName: '$ENDPOINT_NAME'
  endpointConfigName: '$ENDPOINT_CONFIG_NAME'
' > ./deploy.yaml
```

Deploy the endpoint by applying the `deploy.yaml` file. 
```bash
kubectl apply -f deploy.yaml
```

After applying the `deploy.yaml` file, you should see that the model, endpoint configuration, and endpoint were successfully created.
```bash
model.sagemaker.services.k8s.aws/ack-xgboost-model-7420 created
endpointconfig.sagemaker.services.k8s.aws/ack-xgboost-endpoint-config-7420 created
endpoint.sagemaker.services.k8s.aws/ack-xgboost-endpoint-7420 created
```

Watch the process with the `kubectl get` command. Deploying the endpoint may take some time. 
```bash
kubectl get endpoints.sagemaker --watch
```

The endpoint status will be `InService` when the endpoint is successfully deployed and ready for use.
```bash
NAME                        ENDPOINTSTATUS
ack-xgboost-endpoint-7420   Creating         
ack-xgboost-endpoint-7420   InService    
```

## Automatically scale your SageMaker endpoint

Scale your SageMaker endpoint using the Application Auto Scaling [`ScalableTarget`](https://aws-controllers-k8s.github.io/community/reference/applicationautoscaling/v1alpha1/scalabletarget/) and [`ScalingPolicy`](https://aws-controllers-k8s.github.io/community/reference/applicationautoscaling/v1alpha1/scalingpolicy/) resources.

### Create a scalable target

Create a scalable target with the `scalable-target.yaml` file. The following file designates that a specified SageMaker endpoint variant can automatically scale to up to three instances. 

```bash
printf '
apiVersion: applicationautoscaling.services.k8s.aws/v1alpha1
kind: ScalableTarget
metadata:
  name: ack-tutorial-endpoint-scalable-target
spec:
  maxCapacity: 3
  minCapacity: 1
  resourceID: endpoint/'$ENDPOINT_NAME'/variant/AllTraffic
  scalableDimension: "sagemaker:variant:DesiredInstanceCount"
  serviceNamespace: sagemaker
 ' > ./scalable-target.yaml
 ```

 Apply your `scalable-target.yaml` file:
```bash
kubectl apply -f scalable-target.yaml
```

After applying your scalable target, you should see the following output:
```bash
scalabletarget.applicationautoscaling.services.k8s.aws/ack-tutorial-endpoint-scalable-target created
```

You can verify that the `ScalableTarget` was created with the `kubectl describe` command.
```bash
kubectl describe scalabletarget.applicationautoscaling | yq e .Status -
```

### Create a scaling policy

Create a scaling policy with the `scaling-policy.yaml` file. The following file creates a target tracking scaling policy that scales a specified SageMaker endpoint based on the number of variant invocations per instance. The scaling policy adds or removes capacity as required to keep this number close to the target value of 60. 

```bash
printf '
apiVersion: applicationautoscaling.services.k8s.aws/v1alpha1
kind: ScalingPolicy
metadata:
  name: ack-tutorial-endpoint-scaling-policy
spec:
  policyName: ack-tutorial-endpoint-scaling-policy
  policyType: TargetTrackingScaling
  resourceID: endpoint/'$ENDPOINT_NAME'/variant/AllTraffic
  scalableDimension: "sagemaker:variant:DesiredInstanceCount"
  serviceNamespace: sagemaker
  targetTrackingScalingPolicyConfiguration:
    targetValue: 60
    scaleInCooldown: 700
    scaleOutCooldown: 300
    predefinedMetricSpecification:
        predefinedMetricType: SageMakerVariantInvocationsPerInstance
 ' > ./scaling-policy.yaml
 ```

 Apply your `scaling-policy.yaml` file:
```bash
kubectl apply -f scaling-policy.yaml
```

After applying your scaling policy, you should see the following output:
```bash
scalingpolicy.applicationautoscaling.services.k8s.aws/ack-tutorial-endpoint-scaling-policy created
```

You can verify that the `ScalingPolicy` was created with the `kubectl describe` command.
```bash
kubectl describe scalingpolicy.applicationautoscaling | yq e .Status -
```

## Next steps 

To learn more about Application Auto Scaling on a SageMaker endpoint, see the [Application Auto Scaling controller samples](https://github.com/aws-controllers-k8s/applicationautoscaling-controller/tree/main/samples/hosting-autoscaling-on-sagemaker) repository.

### Updates

To update the `ScalableTarget` and `ScalingPolicy` parameters after the resources are created, make any changes to the `scalable-target.yaml` or `scaling-policy.yaml` files and reapply them with `kubectl apply`. 
```bash
kubectl apply -f scalable-target.yaml
kubectl apply -f scaling-policy.yaml.yaml
```

### Cleanup

You can delete your training jobs, endpoints, scalable targets, and scaling policies with the `kubectl delete` command.
```bash
kubectl delete -f training.yaml
kubectl delete -f deploy.yaml
kubectl delete -f scalable-target.yaml
kubectl delete -f scaling-policy.yaml
```

To remove the SageMaker and Application Auto Scaling ACK service controllers, related CRDs, and namespaces see [ACK Cleanup][cleanup].

It is recommended to delete any additional resources such as S3 buckets, IAM roles, and IAM policies when you no longer need them. You can delete these resources with the following commands or directly in the AWS console.

```bash
# Delete S3 bucket
aws s3 rb s3://$SAGEMAKER_BUCKET --force

# Delete SageMaker execution role
aws iam detach-role-policy --role-name $SAGEMAKER_EXECUTION_ROLE_NAME --policy-arn arn:aws:iam::aws:policy/AmazonSageMakerFullAccess
aws iam detach-role-policy --role-name $SAGEMAKER_EXECUTION_ROLE_NAME --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess
aws iam delete-role --role-name $SAGEMAKER_EXECUTION_ROLE_NAME

# Delete IAM role created for IRSA
aws iam detach-role-policy --role-name $OIDC_ROLE_NAME --policy-arn arn:aws:iam::aws:policy/AmazonSageMakerFullAccess
aws iam delete-role --role-name $OIDC_ROLE_NAME
```

To delete your EKS clusters, see [Amazon EKS - Deleting a cluster][cleanup-eks].  

[configure-permissions]: ../../user-docs/authorization/
[cleanup]: ../../user-docs/cleanup/
[cleanup-eks]: https://docs.aws.amazon.com/eks/latest/userguide/delete-cluster.html