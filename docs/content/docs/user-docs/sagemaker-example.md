---
title: "Example: Machine learning with ACK service controllers and EKS"
description: "Train, deploy, and scale a machine learning model with the Amazon SageMaker and Application Auto Scaling ACK service controllers"
lead: ""
draft: false
menu: 
  docs:
    parent: "installing"
weight: 40
toc: true
---

The SageMaker ACK service controller makes it easier for machine learning developers and data scientists who use Kubernetes as their control plane to train, tune, and deploy machine learning models in Amazon SageMaker without logging into the SageMaker console. The Application Auto Scaling ACK service controller can then be used to scale a SageMaker endpoint.  

The following steps will guide you through the setup and use of the Amazon SageMaker ACK service controller for training and deploying machine learning models, and the Amazon Application Auto Scaling ACK service controller to dynamically scale your hosted model.

## Step 1: Setup

Although it is not necessary to use Amazon Elastic Kubernetes Service (Amazon EKS) with ACK, this guide assumes that you have access to an Amazon EKS cluster. For automated cluster creation using `eksctl`, see [Create an Amazon EKS Cluster](https://docs.aws.amazon.com/eks/latest/userguide/create-cluster.html). If this is your first time creating an Amazon EKS cluster, see [Getting started with Amazon EKS](https://docs.aws.amazon.com/eks/latest/userguide/getting-started.html).

### Prerequisites

This guide assumes that you have:
  - Created an EKS cluster with Kubernetes version 1.16 or higher. 
  - [AWS IAM][AWS-IAM] permissions to create roles and attach policies to roles.
  - Installed the following tools on the client machine used to access your Kubernetes cluster:
    - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working with Kubernetes clusters. 
    - [helm](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications.
    - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting with AWS services. 
    - [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html) - A command line tool for working with EKS clusters.
    - [yq](https://mikefarah.gitbook.io/yq) - A command line tool for YAML processing.
[AWS-IAM]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html

### Create an IAM role and attach IAM policy

Check to make sure that you are connected to an Amazon EKS cluster. 

```bash
export CLUSTER_NAME=<CLUSTER_NAME>
export AWS_DEFAULT_REGION=<CLUSTER_REGION>
aws eks update-kubeconfig --name $CLUSTER_NAME --region $AWS_DEFAULT_REGION
kubectl config get-contexts
# Ensure cluster has compute
kubectl get nodes
```

Before you can deploy your SageMaker service controller using an IAM role, associate an OpenID Connect (OIDC) provider with your role to authenticate with the IAM service.

```bash
eksctl utils associate-iam-oidc-provider --cluster ${CLUSTER_NAME} \
--region ${AWS_DEFAULT_REGION} --approve
```

Get the OIDC ID. 

```bash
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
OIDC_PROVIDER_URL=$(aws eks describe-cluster --name $CLUSTER_NAME --region $AWS_DEFAULT_REGION \
--query "cluster.identity.oidc.issuer" --output text | cut -c9-)
```

In your working directory, create a file named trust.json and insert the following trust relationship code block into the document:

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

Run the `iam create-role`command to create an IAM role with the trust relationship you just defined in `trust.json`. This IAM role enables the Amazon EKS cluster to get and refresh credentials from IAM.

```bash
OIDC_ROLE_NAME=ack-controller-role-$CLUSTER_NAME
aws --region $AWS_DEFAULT_REGION iam create-role --role-name $OIDC_ROLE_NAME --assume-role-policy-document file://trust.json
```

Attach the AmazonSageMakerFullAccess Policy to the IAM Role. 

```bash
aws --region $AWS_DEFAULT_REGION iam attach-role-policy --role-name $OIDC_ROLE_NAME --policy-arn arn:aws:iam::aws:policy/AmazonSageMakerFullAccess
export IAM_ROLE_ARN_FOR_IRSA=$(aws --region $AWS_DEFAULT_REGION iam get-role --role-name $OIDC_ROLE_NAME --output text --query 'Role.Arn')
echo $IAM_ROLE_ARN_FOR_IRSA
```

Take note of the `IAM_ROLE_ARN_FOR_IRSA` value printed in the previous step. You will pass this value to the Kubernetes service account used by the ACK service controller. 

For more information on authorization and access for ACK service controllers, including detailes regarding recommended IAM policies, see [Configure Permissions][configure-permissions].

### Install the SageMaker ACK service controller

Make the SageMaker Helm chart available on the `Deployment` host with the following commands:

```bash
export HELM_EXPERIMENTAL_OCI=1
export SERVICE=sagemaker
export RELEASE_VERSION=v0.0.4
export CHART_EXPORT_PATH=/tmp/chart
export CHART_REPO=public.ecr.aws/aws-controllers-k8s/$SERVICE-chart
export CHART_REF=$CHART_REPO:$RELEASE_VERSION
export ACK_K8S_NAMESPACE=ack-system

mkdir -p $CHART_EXPORT_PATH
helm chart pull $CHART_REF
helm chart list
helm chart export $CHART_REF --destination $CHART_EXPORT_PATH
```
Choose either a cluster-scoped or namespace-scoped `Deployment`.

  - Cluster-scoped `Deployment`
    - ```bash
      # Update values in helm chart
      cd $CHART_EXPORT_PATH/$SERVICE-chart
      yq e '.aws.region = env(AWS_DEFAULT_REGION)' -i values.yaml
      yq e '.aws.account_id = env(AWS_ACCOUNT_ID)' -i values.yaml
      yq e '.serviceAccount.annotations."eks.amazonaws.com/role-arn" = env(IAM_ROLE_ARN_FOR_IRSA)' -i values.yaml
      cd -
      ```
  - Namespace-scoped `Deployment`
    - The controller will watch for the resources in the Helm chart release namespace. In this guide, that value is set from the $ACK_K8S_NAMESPACE variable defined above.
    - ```bash
      # Update values in helm chart
      cd $CHART_EXPORT_PATH/$SERVICE-chart
      yq e '.aws.region = env(AWS_DEFAULT_REGION)' -i values.yaml
      yq e '.aws.account_id = env(AWS_ACCOUNT_ID)' -i values.yaml
      yq e '.serviceAccount.annotations."eks.amazonaws.com/role-arn" = env(IAM_ROLE_ARN_FOR_IRSA)' -i values.yaml
      yq e '.watchNamespace" = env(WATCH_NAMESPACE)' -i values.yaml
      cd -
      ```

Install the relevant custom resource definitions (CRDs) for the SageMaker ACK service controller. 
```bash
kubectl apply -f $CHART_EXPORT_PATH/$SERVICE-chart/crds
```

Create a namespace and install the SageMaker ACK service controller with the Helm chart. 
```bash
helm install -n $ACK_K8S_NAMESPACE --create-namespace --skip-crds ack-$SERVICE-controller \
 $CHART_EXPORT_PATH/$SERVICE-chart
```

Verify that the CRDs and Helm charts were deployed with the following commands:
```bash
kubectl get crds
kubectl get pods -n $ACK_K8S_NAMESPACE
```

### Install the Application Auto Scaling ACK service controller

Repeat the installation steps above to install the Application Auto Scaling ACK service controller. Be sure to specify the correct service name and release version. 

```bash
export HELM_EXPERIMENTAL_OCI=1
export SERVICE=applicationautoscaling
export RELEASE_VERSION=v0.0.2
export CHART_EXPORT_PATH=/tmp/chart
export CHART_REPO=public.ecr.aws/aws-controllers-k8s/$SERVICE-chart
export CHART_REF=$CHART_REPO:$RELEASE_VERSION

mkdir -p $CHART_EXPORT_PATH
helm chart pull $CHART_REF
helm chart list
helm chart export $CHART_REF --destination $CHART_EXPORT_PATH
```

For more information on installing ACK service controllers, see [Installation][install].
[install]: https://aws-controllers-k8s.github.io/community/docs/user-docs/install/

## Step 2: Training and deployment

### Prepare your data

For training a model with SageMaker, we will need an S3 bucket to store a dataset and model training artifacts. For this example, we will use the [UCI Abalone dataset](https://archive.ics.uci.edu/ml/datasets/Abalone), which is already processed and available on S3.  

First, create a variable for the S3 bucket:

```bash
export SAGEMAKER_BUCKET=ack-sagemaker-bucket-$RANDOM_VAR
```

Then, create a file named `create-bucket.sh` and insert the following code block:

```bash
printf '
#!/usr/bin/env bash
# create bucket
if [[ $SERVICE_REGION != "us-east-1" ]]; then
  aws s3api create-bucket --bucket "$SAGEMAKER_BUCKET" --region "$SERVICE_REGION" --create-bucket-configuration LocationConstraint="$SERVICE_REGION"
else
  aws s3api create-bucket --bucket "$SAGEMAKER_BUCKET" --region "$SERVICE_REGION"
fi
# sync dataset
aws s3 sync s3://sagemaker-sample-files/datasets/tabular/uci_abalone/train s3://"$SAGEMAKER_BUCKET"/datasets/tabular/uci_abalone/train
aws s3 sync s3://sagemaker-sample-files/datasets/tabular/uci_abalone/validation s3://"$SAGEMAKER_BUCKET"/datasets/tabular/uci_abalone/validation
' > ./create-bucket.sh
```

Run the `create-bucket.sh` script to create an S3 bucket and copy the UCI Abalone dataset into your new bucket.

```bash
chmod +x create-bucket.sh
./create-bucket.sh
```

### Train an XGBoost model

The SageMaker training job that we execute will need an IAM role to access Amazon S3 and Amazon SageMaker. Run the following commands to create a SageMaker execution IAM role that will be used by SageMaker to access the appropriate AWS resources:

```bash
export SAGEMAKER_EXECUTION_ROLE_NAME=ack-sagemaker-execution-role-$RANDOM_VAR

TRUST="{ \"Version\": \"2012-10-17\", \"Statement\": [ { \"Effect\": \"Allow\", \"Principal\": { \"Service\": \"sagemaker.amazonaws.com\" }, \"Action\": \"sts:AssumeRole\" } ] }"
aws iam create-role --role-name ${SAGEMAKER_EXECUTION_ROLE_NAME} --assume-role-policy-document "$TRUST"
aws iam attach-role-policy --role-name ${SAGEMAKER_EXECUTION_ROLE_NAME} --policy-arn arn:aws:iam::aws:policy/AmazonSageMakerFullAccess
aws iam attach-role-policy --role-name ${SAGEMAKER_EXECUTION_ROLE_NAME} --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess

SAGEMAKER_EXECUTION_ROLE_ARN=$(aws iam get-role --role-name ${SAGEMAKER_EXECUTION_ROLE_NAME} --output text --query 'Role.Arn')

echo $SAGEMAKER_EXECUTION_ROLE_ARN
```

{{% hint type="info" title="Take note of the ARN" %}}
Take note of the execution role's Amazon Resource Name (ARN) to use in the specifications below.
{{% /hint %}}

Create a `training.yaml` file to specify the parameters for your SageMaker training job. You will need to specify your SageMaker training job name, any relevant hyperperameters, and the location of your training and validation data. You can also use this document to specify which Amazon Elastic Container Registry (ECR) image to use for training. 

```bash
export JOB_NAME=ack-xgboost-training-job-$RANDOM_VAR
```

{{% hint type="info" title="Change XGBoost image URI based on region" %}}
**IMPORTANT**: If your `SERVICE_REGION` is **not** **us-east-1**, you must change the `XGBOOST_IMAGE` URI in `training.yaml`. To find your region-specific XGBoost image URI, choose your region in the [SageMaker Docker Registry Paths page](https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html), and then select **XGBoost (algorithm)**. For this example, use version 1.2-1.
{{% /hint %}}

```bash
# Give your SageMaker training job a unique name
export JOB_NAME=ack-xgboost-training-job-$RANDOM_VAR

# Be sure to specify your region-specific XGBoost image URI
export XGBOOST_IMAGE=683313688378.dkr.ecr.us-east-1.amazonaws.com/sagemaker-xgboost:1.2-1

printf '
apiVersion: sagemaker.services.k8s.aws/v1alpha1
kind: TrainingJob
metadata:
  name: '$JOB_NAME'
spec:
  # Name that will appear in SageMaker console
  trainingJobName: '$JOB_NAME'
  hyperParameters: 
    max_depth: "5"
    gamma: "4"
    eta: "0.2"
    min_child_weight: "6"
    subsample: "0.7"
    objective: "reg:linear"
    num_round: "50"
    verbosity: "2"
  algorithmSpecification:
    trainingImage: '$XGBOOST_IMAGE'
    trainingInputMode: File
  roleARN: '$SAGEMAKER_EXECUTION_ROLE_ARN'
  outputDataConfig:
    # The output path of our model
    s3OutputPath: s3://'$SAGEMAKER_BUCKET'
  resourceConfig:
    instanceCount: 1
    instanceType: ml.m4.xlarge
    volumeSizeInGB: 5
  stoppingCondition:
    maxRuntimeInSeconds: 3600
  inputDataConfig:
    - channelName: train
      dataSource:
        s3DataSource:
          s3DataType: S3Prefix
          # The input path of our train data 
          s3URI: s3://'$SAGEMAKER_BUCKET'/datasets/tabular/uci_abalone/train/abalone.train
          s3DataDistributionType: FullyReplicated
      contentType: text/libsvm
      compressionType: None
    - channelName: validation
      dataSource:
        s3DataSource:
          s3DataType: S3Prefix
          # The input path of our validation data 
          s3URI: s3://'$SAGEMAKER_BUCKET'/datasets/tabular/uci_abalone/validation/abalone.validation
          s3DataDistributionType: FullyReplicated
      contentType: text/libsvm
      compressionType: None 
' > ./training.yaml
```

Use your `training.yaml` file to create a SageMaker training job:

```bash
kubectl apply -f training.yaml
```

After applying your `training.yaml` file, you should see that your training job was successfully created:

```bash
trainingjob.sagemaker.services.k8s.aws/ack-xgboost-training-job-7420 created
```

You can watch the status of the training job with the following command:

```bash
kubectl get trainingjob.sagemaker --watch
```

It will a take a few minutes for `TRAININGJOBSTATUS` to be `Completed`.

```bash
NAME                            SECONDARYSTATUS   TRAININGJOBSTATUS
ack-xgboost-training-job-7420   Starting          InProgress
ack-xgboost-training-job-7420   Downloading       InProgress
ack-xgboost-training-job-7420   Training          InProgress
ack-xgboost-training-job-7420   Completed         Completed
```

### Deploy an XGBoost model

To deploy the XGBoost model, you need to provide a unique name for the model, the endpoint configuration, and the endpoint itself.  

```bash
export MODEL_NAME=ack-xgboost-model-$RANDOM_VAR
export ENDPOINT_CONFIG_NAME=ack-xgboost-endpoint-config-$RANDOM_VAR
export ENDPOINT_NAME=ack-xgboost-endpoint-$RANDOM_VAR
```

In this example, we will deploy our XGboost model on a `c5.large` instance type. Create a `deploy.yaml` file to specify the parameters for your SageMaker model deployment.

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
    instanceType: ml.c5.large
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

Apply your `deploy.yaml` file to deploy your SageMaker endpoint.

```bash
kubectl apply -f deploy.yaml
```

After deploying your SageMaker endpoint, you should see that the model, endpoint configuration, and endpoint were successfully created:

```bash
model.sagemaker.services.k8s.aws/ack-xgboost-model-7420 created
endpointconfig.sagemaker.services.k8s.aws/ack-xgboost-endpoint-config-7420 created
endpoint.sagemaker.services.k8s.aws/ack-xgboost-endpoint-7420 created
```

You can get the status of your model, endpoint configuration, and endpoint with the `kubectl describe` command. 

```bash
kubectl describe models.sagemaker | yq e .Status -
kubectl describe endpointconfigs.sagemaker | yq e .Status -
kubectl describe endpoints.sagemaker | yq e .Status -
```

Deploying the endpoint may take some time. You can watch the deployment process using the following command:

```bash
kubectl get endpoints.sagemaker --watch
```

After some time, the `ENDPOINTSTATUS` will change to `InService`, which indicates that the deployed endpoint is ready for use.

```bash
NAME                        ENDPOINTSTATUS
ack-xgboost-endpoint-7420   Creating         
ack-xgboost-endpoint-7420   InService        
```

## Step 3: Cross-region resource management

If you are using resources across different regions, you can override the default region of a given ACK service controller. ACK service controllers will first look for a region in the following order:

1. The region annotation `services.k8s.aws/region` on the resource. If provided, this will override the namespace default region annotation.
2. The namespace default region annotation `services.k8s.aws/default-region`.
3. Controller flags, such as the `aws.region` variable in a given Helm chart
4. Kubernetes Pod IRSA environment variables

For example, the ACK service controller default region is `us-west-2`. If you want to create a resource in `us-east-1`, use one of the following options to override the default region

  - Option 1: Region annotation
    - Add the `services.k8s.aws/region` annotation while creating the resource. For example:
    ```yaml
      apiVersion: sagemaker.services.k8s.aws/v1alpha1
      kind: TrainingJob
      metadata:
        name: ack-sample-tainingjob
        annotations:
          services.k8s.aws/region: us-east-1
      spec:
        trainingJobName: ack-sample-tainingjob
        roleARN: <sagemaker_execution_role_arn>
        ...
      ```

  - Option 2: Namespace default region annotation 
    - To bind a region to a specific namespace, you will have to annotate the namespace with `services.k8s.aws/default-region` annotation. For example:
    ```yaml
      apiVersion: v1
      kind: Namespace
      metadata:
        name: production
        annotations:
          services.k8s.aws/default-region: us-east-1
    ```
    - For existing namespaces, you can run:
      - ```bash
        kubectl annotate namespace production services.k8s.aws/default-region=us-east-1
        ```
    - You can also create the resource in the same namespace:
      - ```yaml
        apiVersion: sagemaker.services.k8s.aws/v1alpha1
        kind: TrainingJob
        metadata:
          name: ack-sample-trainingjob
          namespace: production
        spec:
          trainingJobName: ack-sample-trainingjob
          roleARN: <sagemaker_execution_role_arn>
          ...
        ```

If you are interested in managing resources across accounts as well as regions, see [Cross-Account Resource Management][cross-account].

## Step 4: Application Auto Scaling

SageMaker ACK service controllers support CRDs for automatic scaling (using [ScalableTarget](https://aws-controllers-k8s.github.io/community/reference/applicationautoscaling/v1alpha1/ScalableTarget/) and [ScalingPolicy](https://aws-controllers-k8s.github.io/community/reference/applicationautoscaling/v1alpha1/ScalingPolicy/)) for your hosted models. The following CRDs will adjust the number of instances provisioned for a model in response to changes in the  `SageMakerVariantInvocationsPerInstancetracking` metric, which is the average number of times per minute that each instance for a variant is invoked. The minimum number of instances provisioned is one and scaling can automatically provision up to 20. 

Create a `scale-enpoint.yaml` file to configure and apply your Application Auto Scaling CRDs. 

```yaml
printf '
apiVersion: applicationautoscaling.services.k8s.aws/v1alpha1
kind: ScalableTarget
metadata:
  name: ack-scalable-target-predfined
spec:
  maxCapacity: 20
  minCapacity: 1
  resourceID: endpoint/'$ENDPOINT_NAME'/variant/AllTraffic
  scalableDimension: "sagemaker:variant:DesiredInstanceCount"
  serviceNamespace: sagemaker
---
apiVersion: applicationautoscaling.services.k8s.aws/v1alpha1
kind: ScalingPolicy
metadata:
  name: ack-scaling-policy-predefined
spec:
  policyName: ack-scaling-policy-predefined
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
 ' > ./scale-endpoint.yaml
```

Apply your `scale-enpoint.yaml` file:

```bash
kubectl apply -f scale-endpoint.yaml
```

After applying your autoscaling CRDs, you should see the following output:

```bash
scalabletarget.applicationautoscaling.services.k8s.aws/ack-scalable-target-predfined created
scalingpolicy.applicationautoscaling.services.k8s.aws/ack-scaling-policy-predefined created
```

You can verify that the `ScalingPolicy` was created with the `kubectl describe` command.

```bash
kubectl describe scalingpolicy.applicationautoscaling | yq e .Status -
```

The status should look similar to the following:

```bash
Status:
  Ack Resource Metadata:
    Arn:               arn:aws:autoscaling:us-east-1:1234567890:scalingPolicy:b33d12b8-aa81-4cb8-855e-c7b6dcb9d6e7:resource/SageMaker/endpoint/ack-xgboost-endpoint/variant/AllTraffic:policyName/ack-scaling-policy-predefined
    Owner Account ID:  1234567890
  Alarms:
    Alarm ARN:   arn:aws:cloudwatch:us-east-1:1234567890:alarm:TargetTracking-endpoint/ack-xgboost-endpoint/variant/AllTraffic-AlarmHigh-966b8232-a9b9-467d-99f3-95436f5c0383
    Alarm Name:  TargetTracking-endpoint/ack-xgboost-endpoint/variant/AllTraffic-AlarmHigh-966b8232-a9b9-467d-99f3-95436f5c0383
    Alarm ARN:   arn:aws:cloudwatch:us-east-1:1234567890:alarm:TargetTracking-endpoint/ack-xgboost-endpoint/variant/AllTraffic-AlarmLow-71e39f85-1afb-401d-9703-b788cdc10a93
    Alarm Name:  TargetTracking-endpoint/ack-xgboost-endpoint/variant/AllTraffic-AlarmLow-71e39f85-1afb-401d-9703-b788cdc10a93
```

## Next Steps 
For more information on the SageMaker ACK service controller, see the [SageMaker controller samples repository][sagemaker-samples]. To learn more about Application Auto Scaling on a SageMaker endpoint, see the [Application Auto Scaling controller samples repository][application-autoscaling-samples]. 

To uninstall the SageMaker or Application Auto Scaling ACK service controllers, see [cleanup][cleanup].

[configure-permissions]: https://aws-controllers-k8s.github.io/community/docs/user-docs/authorization/
[cross-account]: https://aws-controllers-k8s.github.io/community/docs/user-docs/cross-account-resource-management/
[sagemaker-samples]: https://github.com/aws-controllers-k8s/sagemaker-controller/tree/main/samples
[application-autoscaling-samples]: https://github.com/aws-controllers-k8s/applicationautoscaling-controller/tree/main/samples/hosting-autoscaling-on-sagemaker
[cleanup]: https://aws-controllers-k8s.github.io/community/docs/user-docs/cleanup/