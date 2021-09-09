---
title: "Machine learning with the ACK SageMaker controller"
description: "Train a machine learning model with the ACK service controller for Amazon SageMaker using Amazon Elastic Kubernetes Service"
lead: ""
draft: false
menu: 
  docs:
    parent: "tutorials"
weight: 40
toc: true
---

The SageMaker ACK service controller makes it easier for machine learning developers and data scientists who use Kubernetes as their control plane to train, tune, and deploy machine learning models in Amazon SageMaker without logging into the SageMaker console. 

The following steps will guide you through the setup and use of the Amazon SageMaker ACK service controller for training a machine learning model.

## Setup

Although it is not necessary to use Amazon Elastic Kubernetes Service (Amazon EKS) with ACK, this guide assumes that you have access to an Amazon EKS cluster. For automated cluster creation using `eksctl`, see [Getting started with Amazon EKS - `eksctl`](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html) and create your cluster with Amazon EC2 Linux managed nodes. If this is your first time creating an Amazon EKS cluster, see [Amazon EKS Setup](https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/deep-learning-containers-eks-setup.html).

### Prerequisites

This guide assumes that you have:
  - Created an EKS cluster with Kubernetes version 1.16 or higher. 
  - AWS IAM permissions to create roles and attach policies to roles.
  - Installed the following tools on the client machine used to access your Kubernetes cluster:
    - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working with Kubernetes clusters. 
    - [Helm](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications.
    - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting with AWS services. 
    - [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html) - A command line tool for working with EKS clusters.
    - [yq](https://mikefarah.gitbook.io/yq) - A command line tool for YAML processing.
      - Linux
        ```
        sudo wget https://github.com/mikefarah/yq/releases/download/v4.9.8/yq_linux_amd64 -O /usr/bin/yq
        sudo chmod +x /usr/bin/yq
        ```
      - Mac
        ```
        brew install yq
        ```

### Configure IAM permissions

Create an IAM role and attach an IAM policy to that role to ensure that your SageMaker service controller has access to the appropriate AWS resources. First, check to make sure that you are connected to an Amazon EKS cluster. 

```bash
export CLUSTER_NAME=<CLUSTER_NAME>
export AWS_DEFAULT_REGION=<CLUSTER_REGION>
aws eks update-kubeconfig --name $CLUSTER_NAME --region $AWS_DEFAULT_REGION
kubectl config get-contexts
# Ensure cluster has compute
kubectl get nodes
```

Before you can deploy your SageMaker service controller using an IAM role, associate an OpenID Connect (OIDC) provider with your IAM role to authenticate your cluster with the IAM service.

```bash
eksctl utils associate-iam-oidc-provider --cluster ${CLUSTER_NAME} \
--region ${AWS_DEFAULT_REGION} --approve
```

Get the following OIDC information for future reference:

```bash
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
export OIDC_PROVIDER_URL=$(aws eks describe-cluster --name $CLUSTER_NAME --region $AWS_DEFAULT_REGION \
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

Run the `iam create-role` command to create an IAM role with the trust relationship you just defined in `trust.json`. This IAM role enables the Amazon EKS cluster to get and refresh credentials from IAM.

```bash
export OIDC_ROLE_NAME=ack-controller-role-$CLUSTER_NAME
aws --region $AWS_DEFAULT_REGION iam create-role --role-name $OIDC_ROLE_NAME --assume-role-policy-document file://trust.json
```

Attach the AmazonSageMakerFullAccess Policy to the IAM Role to ensure that your SageMaker service controller has access to the appropriate resources. 

```bash
aws --region $AWS_DEFAULT_REGION iam attach-role-policy --role-name $OIDC_ROLE_NAME --policy-arn arn:aws:iam::aws:policy/AmazonSageMakerFullAccess
export IAM_ROLE_ARN_FOR_IRSA=$(aws --region $AWS_DEFAULT_REGION iam get-role --role-name $OIDC_ROLE_NAME --output text --query 'Role.Arn')
echo $IAM_ROLE_ARN_FOR_IRSA
```

For more information on authorization and access for ACK service controllers, including details regarding recommended IAM policies, see [Configure Permissions][configure-permissions].

### Install the SageMaker ACK service controller

Get the SageMaker Helm chart and make it available on the `Deployment` host with the following commands:

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

Update the Helm chart values for a cluster-scoped `Deployment`. 

```bash
# Update the following values in the Helm chart
cd $CHART_EXPORT_PATH/$SERVICE-chart
yq e '.aws.region = env(AWS_DEFAULT_REGION)' -i values.yaml
yq e '.aws.account_id = env(AWS_ACCOUNT_ID)' -i values.yaml
yq e '.serviceAccount.annotations."eks.amazonaws.com/role-arn" = env(IAM_ROLE_ARN_FOR_IRSA)' -i values.yaml
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

## Train an XGBoost model

### Prepare your data

For training a model with SageMaker, we will need an S3 bucket to store the dataset and model training artifacts. For this example, we will use [MNIST](http://yann.lecun.com/exdb/mnist/) data stored in [LIBSVM](https://www.csie.ntu.edu.tw/~cjlin/libsvm/) format.

First, create a variable for the S3 bucket:

```bash
export RANDOM_VAR=$RANDOM
# Specify your service region
export SERVICE_REGION=us-east-1
export SAGEMAKER_BUCKET=ack-sagemaker-bucket-$RANDOM_VAR
```

Then, create a file named `create-bucket.sh` with the following code block:

```bash
printf '
#!/usr/bin/env bash
# Create the S3 bucket
if [[ $SERVICE_REGION != "us-east-1" ]]; then
  aws s3api create-bucket --bucket "$SAGEMAKER_BUCKET" --region "$SERVICE_REGION" --create-bucket-configuration LocationConstraint="$SERVICE_REGION"
else
  aws s3api create-bucket --bucket "$SAGEMAKER_BUCKET" --region "$SERVICE_REGION"
' > ./create-bucket.sh
```

Run the `create-bucket.sh` script to create an S3 bucket.

```bash
chmod +x create-bucket.sh
./create-bucket.sh
```

Copy the MNIST data into your S3 bucket.
```bash
wget https://raw.githubusercontent.com/aws-controllers-k8s/sagemaker-controller/main/samples/training/s3_sample_data.py
python3 s3_sample_data.py $SAGEMAKER_BUCKET
```

### Configure permissions for your training job

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

### Create a SageMaker training job 

Give your SageMaker training job a unique name: 

```bash
export JOB_NAME=ack-xgboost-training-job-$RANDOM_VAR
```

Specify your region-specific XGBoost image URI: 

```bash
export XGBOOST_IMAGE=683313688378.dkr.ecr.us-east-1.amazonaws.com/sagemaker-xgboost:1.2-1
```
{{% hint type="info" title="Change XGBoost image URI based on region" %}}
**IMPORTANT**: If your `SERVICE_REGION` is not `us-east-1`, you must change the `XGBOOST_IMAGE` URI. To find your region-specific XGBoost image URI, choose your region in the [SageMaker Docker Registry Paths page](https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html), and then select **XGBoost (algorithm)**. For this example, use version 1.2-1.
{{% /hint %}}

Next, create a `training.yaml` file to specify the parameters for your SageMaker training job. This file specifies your SageMaker training job name, any relevant hyperperameters, and the location of your training and validation data. You can also use this document to specify which Amazon Elastic Container Registry (ECR) image to use for training. 

```yaml
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
    objective: "multi:softmax"
    num_class: "10"
    num_round: "10"
  algorithmSpecification:
    # The URL and tag of your ECR container
    # If you are not on us-west-1 you can find an imageURI here https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
    trainingImage: '$XGBOOST_IMAGE'
    trainingInputMode: File
  # A role with SageMaker and S3 access
  # example arn:aws:iam::1234567890:role/service-role/AmazonSageMaker-ExecutionRole
  roleARN: '$SAGEMAKER_EXECUTION_ROLE_ARN' 
  outputDataConfig:
    # The output path of our model
    s3OutputPath: s3://'$SAGEMAKER_BUCKET' 
  resourceConfig:
    instanceCount: 1
    instanceType: ml.m4.xlarge
    volumeSizeInGB: 5
  stoppingCondition:
    maxRuntimeInSeconds: 86400
  inputDataConfig:
    - channelName: train
      dataSource:
        s3DataSource:
          s3DataType: S3Prefix
          # The input path of our train data 
          s3URI: s3://'$SAGEMAKER_BUCKET'/sagemaker/xgboost/train
          s3DataDistributionType: FullyReplicated
      contentType: text/libsvm
      compressionType: None
    - channelName: validation
      dataSource:
        s3DataSource:
          s3DataType: S3Prefix
          # The input path of our validation data 
          s3URI: s3://'$SAGEMAKER_BUCKET'/sagemaker/xgboost/validation
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

If your training job failed, look to the `status.failureReason` after running the following command:
```bash
kubectl describe trainingjobs $JOB_NAME
```

## Next Steps 

For more examples on how to use the SageMaker ACK service controller, see the [SageMaker controller samples repository][sagemaker-samples]. 

### Cleanup

You can delete your SageMaker training job with the `kubectl delete` command.
```bash
kubectl delete trainingjob $JOB_NAME
```

To remove the SageMaker ACK service controller, related CRDs, and namespaces see [ACK Cleanup][cleanup].

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

To delete your EKS clusters, pods, or nodes, see [Amazon EKS Setup - Cleanup][cleanup-eks].  

[configure-permissions]: https://aws-controllers-k8s.github.io/community/docs/user-docs/authorization/
[sagemaker-samples]: https://github.com/aws-controllers-k8s/sagemaker-controller/tree/main/samples
[cleanup]: https://aws-controllers-k8s.github.io/community/docs/user-docs/cleanup/
[cleanup-eks]: https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/deep-learning-containers-eks-setup.html#deep-learning-containers-eks-setup-cleanup