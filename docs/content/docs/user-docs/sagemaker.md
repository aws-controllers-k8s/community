# Example: Machine learning with Amazon SageMaker ACK service controller  

## Train and deploy a machine learning model with Amazon SageMaker ACK service controller 

The SageMaker ACK service controller makes it easier for machine learning developers and data scientists who use Kubernetes as their control plane to train, tune, and deploy machine learning models in Amazon SageMaker without logging into the SageMaker console.

The following sections will guide you to install SageMaker and Application Autoscaling controllers.

## Step 1: Prerequisites

This guide assumes that you’ve the following prerequisites:
  - Installed the following tools on the client machine used to access your Kubernetes cluster:
    - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working with Kubernetes clusters. 
    - [helm](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications
    - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting with AWS services. 
    - [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html) - A command line tool for working with EKS clusters that automates many individual tasks.
    - [yq](https://mikefarah.gitbook.io/yq) - command-line YAML processor.
      - Linux
        ```
        sudo wget https://github.com/mikefarah/yq/releases/download/v4.9.8/yq_linux_amd64 -O /usr/bin/yq
        sudo chmod +x /usr/bin/yq
        ```
      - Mac
        ```
        brew install yq
        ```
  - Have IAM permissions to create roles and attach policies to roles.
  - Created an EKS cluster on which to run the controllers. It should be Kubernetes version 1.16+. For automated cluster creation using eksctl, see [Create an Amazon EKS Cluster](https://docs.aws.amazon.com/eks/latest/userguide/create-cluster.html) and select eksctl option.

## Step 2: Set Up IAM Role-based Authentication

### 2.1 Ensure you are connected to EKS Cluster

```sh
export CLUSTER_NAME=<CLUSTER_NAME>
export AWS_DEFAULT_REGION=<CLUSTER_REGION>

aws eks update-kubeconfig --name $CLUSTER_NAME --region $AWS_DEFAULT_REGION

kubectl config get-contexts
# Ensure cluster has compute
kubectl get nodes
```

### 2.1 Setup IRSA for controller pod

Before you can deploy your operator using an IAM role, associate an OpenID Connect (OIDC) provider with your role to authenticate with the IAM service.

#### 2.1.1 Create an OpenID Connect Provider for Your Cluster

```sh
eksctl utils associate-iam-oidc-provider --cluster ${CLUSTER_NAME} \
--region ${AWS_DEFAULT_REGION} --approve
```

Get the OIDC ID
```sh
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
OIDC_PROVIDER_URL=$(aws eks describe-cluster --name $CLUSTER_NAME --region $AWS_DEFAULT_REGION \
--query "cluster.identity.oidc.issuer" --output text | cut -c9-)
```

#### 2.1.2 Create an IAM Role

Create a file named trust.json and insert the following trust relationship code block required for IAM role into it.
```sh
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


Run the following command to create a role with the trust relationship defined in `trust.json`. This role enables the Amazon EKS cluster to get and refresh credentials from IAM.

```sh
OIDC_ROLE_NAME=ack-controller-role-$CLUSTER_NAME

aws --region $AWS_DEFAULT_REGION iam create-role --role-name $OIDC_ROLE_NAME --assume-role-policy-document file://trust.json

# Attach the AmazonSageMakerFullAccess Policy to the Role
aws --region $AWS_DEFAULT_REGION iam attach-role-policy --role-name $OIDC_ROLE_NAME --policy-arn arn:aws:iam::aws:policy/AmazonSageMakerFullAccess
export IAM_ROLE_ARN_FOR_IRSA=$(aws --region $AWS_DEFAULT_REGION iam get-role --role-name $OIDC_ROLE_NAME --output text --query 'Role.Arn')
echo $IAM_ROLE_ARN_FOR_IRSA
```

Take note of IAM_ROLE_ARN_FOR_IRSA printed in the previous step; you will pass this value to the service account used by the controller.

## 3.0 Install Controllers

### 3.1 Install SageMaker Controller

#### 3.1.1 Download helm chart

```sh
export HELM_EXPERIMENTAL_OCI=1
export SERVICE=sagemaker
export RELEASE_VERSION=v0.0.3
export CHART_EXPORT_PATH=/tmp/chart
export CHART_REPO=public.ecr.aws/aws-controllers-k8s/$SERVICE-chart
export CHART_REF=$CHART_REPO:$RELEASE_VERSION

mkdir -p $CHART_EXPORT_PATH
helm chart pull $CHART_REF
helm chart list
helm chart export $CHART_REF --destination $CHART_EXPORT_PATH
```

#### 3.1.2 Choose one of the two options for deployment

  - [Option 1] Cluster scoped deployment
    - ```sh
      # Update values in helm chart
      cd $CHART_EXPORT_PATH/$SERVICE-chart
      yq e '.aws.region = env(AWS_DEFAULT_REGION)' -i values.yaml
      yq e '.aws.account_id = env(AWS_ACCOUNT_ID)' -i values.yaml
      yq e '.serviceAccount.annotations."eks.amazonaws.com/role-arn" = env(IAM_ROLE_ARN_FOR_IRSA)' -i values.yaml
      cd -
      ```
  - [Option 2] Namespace scoped deployment
    - Specify the namespace to listen to
      ```sh
      export WATCH_NAMESPACE=<NAMESPACE_TO_LISTEN_TO>
      ```
    - ```sh
      # Update values in helm chart
      cd $CHART_EXPORT_PATH/$SERVICE-chart
      yq e '.aws.region = env(AWS_DEFAULT_REGION)' -i values.yaml
      yq e '.aws.account_id = env(AWS_ACCOUNT_ID)' -i values.yaml
      yq e '.serviceAccount.annotations."eks.amazonaws.com/role-arn" = env(IAM_ROLE_ARN_FOR_IRSA)' -i values.yaml
      yq e '.watchNamespace" = env(WATCH_NAMESPACE)' -i values.yaml
      cd -
      ```
#### 3.1.3 Install Controller

Install CRDs
```sh
kubectl apply -f $CHART_EXPORT_PATH/$SERVICE-chart/crds
```

Create a namespace and install the helm chart
```sh
export ACK_K8S_NAMESPACE=ack-system
helm install -n $ACK_K8S_NAMESPACE --create-namespace --skip-crds ack-$SERVICE-controller \
 $CHART_EXPORT_PATH/$SERVICE-chart
```

Verify CRDs and helm charts were deployed
```sh
kubectl get crds

kubectl get pods -n $ACK_K8S_NAMESPACE
```

Jump to Section 4.0 if you only wish to install SageMaker controller
### 3.2 Install ApplicationAutoscaling Controller

##### 3.2.1 Download helm chart

```sh
export HELM_EXPERIMENTAL_OCI=1
export SERVICE=applicationautoscaling
export RELEASE_VERSION=v0.0.1
export CHART_EXPORT_PATH=/tmp/chart
export CHART_REPO=public.ecr.aws/aws-controllers-k8s/$SERVICE-chart
export CHART_REF=$CHART_REPO:$RELEASE_VERSION

mkdir -p $CHART_EXPORT_PATH
helm chart pull $CHART_REF
helm chart list
helm chart export $CHART_REF --destination $CHART_EXPORT_PATH
```

##### 3.2.2 Choose one of the two options for deployment

Run the steps in section [3.1.2](#312-choose-one-of-the-two-options-for-deployment)

##### 3.2.3 Install Controller

Run the steps in section [3.1.3](#313-install-controller)

## 4.0 Samples

### 4.1 SageMaker samples

Head over to the [samples directory](/samples) and follow the README to create resources. 

### 4.2 Application-autoscaling samples

Head over to the [samples directory in application-autoscaling controller repository](https://github.com/aws-controllers-k8s/applicationautoscaling-controller/tree/main/samples/hosting-autoscaling-on-sagemaker) and follow the README to create resources. 

Note: these samples will only work if you installed application autoscaling controller in [Section 3.2](#32-applicationautoscaling)

## 5.0 Cross Region Resource Management

To determine which region the resources should be created, ACK controllers will, in order, look for a region in the following sources:

1. Region annotation services.k8s.aws/region on the resource. If provided it will override the namespace default region annotation.
2. Namespace default region annotation services.k8s.aws/default-region
3. If none of the two annotations are provided ACK will try to find a region from these sources: 
    1. Controller flags i.e. aws.region in helm charts
    2. Pod IRSA environment variables

for example, the controller default region is us-west-2 (3.a/3.b) and you want to create resource in us-east-1. Use the one of the following options to override the default region

  - [Option 1] Region annotation sample
    - Add the `services.k8s.aws/region` annotation while creating the resource. For example:
    - ```yaml
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

  - [Option 2] Namespace default region annotation sample
    - To bind a region to a specific Namespace you will have to annotate the Namespace with `services.k8s.aws/default-region` annotation. For example:
    - ```yaml
      apiVersion: v1
      kind: Namespace
      metadata:
        name: production
        annotations:
          services.k8s.aws/default-region: us-east-1
      ```
    - For existing namespaces you can also run:
      - ```sh
        kubectl annotate namespace production services.k8s.aws/default-region=us-east-1
        ```

    - Create the resource in the same namespace
      - ```yaml
        apiVersion: sagemaker.services.k8s.aws/v1alpha1
        kind: TrainingJob
        metadata:
          name: ack-sample-tainingjob
          namespace: production
        spec:
          trainingJobName: ack-sample-tainingjob
          roleARN: <sagemaker_execution_role_arn>
          ...
        ```

## 6.0 Cross Account Resource Management

ACK service controllers can manage resources in different AWS accounts. To enable and start using this feature, you will need to:

1. Configure your AWS accounts, where the resources will be managed.
2. Create a ConfigMap to map AWS accounts with the Role ARNs that needs to be assumed
3. Annotate namespaces with AWS Account IDs

For detailed information about how ACK service controllers manage resource in multiple AWS accounts, please refer to [CARM](https://github.com/aws/aws-controllers-k8s/blob/main/docs/design/proposals/carm/cross-account-resource-management.md) design document.

### 6.1 Setting up AWS accounts

AWS Account administrators should create/configure IAM roles to allow ACK service controllers to assume Roles in different AWS accounts.
For example, to allow account A (000000000000) to create resources in account B (111111111111) and you have configured the controller to use `arn:aws:iam::000000000000:role/roleA-production` role

Using account A credentials
```sh
export POLICY="{\"Version\":\"2012-10-17\",\"Statement\":{\"Effect\":\"Allow\",\"Action\":\"sts:AssumeRole\",\"Resource\":\"*\"}}"
aws iam put-role-policy --role-name roleA-production \
  --policy-name sts-assumerole --policy-document "$POLICY"
```

Using account B credentials
```sh
export CARM_ROLE_NAME=SagemakerCrossAccountAccess
export TRUST="{ \"Version\": \"2012-10-17\", \"Statement\": [ { \"Effect\": \"Allow\", \"Principal\": { \"AWS\": \"arn:aws:iam::000000000000:role/roleA-production\" }, \"Action\": \"sts:AssumeRole\" } ] }"
aws iam create-role --role-name ${CARM_ROLE_NAME} \
  --assume-role-policy-document "$TRUST"
aws iam attach-role-policy --role-name ${CARM_ROLE_NAME} \
  --policy-arn arn:aws:iam::aws:policy/AmazonSageMakerFullAccess
```

### 6.2 Map AWS Accounts with their associated Role ARNs

Create a ConfigMap named ack-role-account-map in the namespace controller is installed. This ConfigMap will be used to associate each AWS Account ID with the role ARN that needs be assumed, in order to manage resources in that particular account. For example:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ack-role-account-map
  namespace: ack-system
data:
  "111111111111": arn:aws:iam::111111111111:role/SagemakerCrossAccountAccess
```

### 6.3 Bind accounts to namespaces

To bind AWS accounts to a specific Namespace you will have to annotate the Namespace with an AWS Account ID. For example:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: production
  annotations:
    services.k8s.aws/owner-account-id: 111111111111
```

For existing namespaces you can also run:
```sh
kubectl annotate namespace production services.k8s.aws/owner-account-id=111111111111
```
### 6.4 Create resource in different AWS account

Now to create resources in account B, you will have to create your resources in the associated Namespace. For example:

```yaml
apiVersion: sagemaker.services.k8s.aws/v1alpha1
kind: TrainingJob
metadata:
  name: ack-sample-tainingjob
  namespace: production
spec:
  trainingJobName: ack-sample-tainingjob
  roleARN: <sagemaker_execution_role_arn>
  ...
```

## 8.0 Adopt Resources

ACK controller provides to provide the ability to “adopt” resources that were not originally created by ACK service controller. Given the user configures the controller with permissions which has access to existing resource, the controller will be able to determine the current specification and status of the AWS resource and reconcile said resource as if the ACK controller had originally created it.

Sample:
```yaml
apiVersion: services.k8s.aws/v1alpha1
kind: AdoptedResource
metadata:
  name: adopt-endpoint-sample
spec:  
  aws:
    # resource to adopt, not created by ACK
    nameOrID: xgboost-endpoint
  kubernetes:
    group: sagemaker.services.k8s.aws
    kind: Endpoint
    metadata:
      # target K8s CR name
      name: xgboost-endpoint
```
Save the above to a file name adopt-endpoint-sample.yaml.

Submit the CR
```sh
kubectl apply -f adopt-endpoint-sample.yaml
```

Check for `ACK.Adopted` condition to be true under `status.conditions`

```sh
kubectl describe adoptedresource adopt-endpoint-sample
```

Output should look similar to this:
```yaml
---
kind: AdoptedResource
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"apiVersion":"services.k8s.aws/v1alpha1","kind":"AdoptedResource","metadata":{"annotations":{},"name":"xgboost-endpoint","namespace":"default"},"spec":{"aws":{"nameOrID":"xgboost-endpoint"},"kubernetes":{"group":"sagemaker.services.k8s.aws","kind":"Endpoint","metadata":{"name":"xgboost-endpoint"}}}}'
  creationTimestamp: '2021-04-27T02:49:14Z'
  finalizers:
  - finalizers.services.k8s.aws/AdoptedResource
  generation: 1
  name: adopt-endpoint-sample
  namespace: default
  resourceVersion: '12669876'
  selfLink: "/apis/services.k8s.aws/v1alpha1/namespaces/default/adoptedresources/adopt-endpoint-sample"
  uid: 35f8fa92-29dd-4040-9d0d-0b07bbd7ca0b
spec:
  aws:
    nameOrID: xgboost-endpoint
  kubernetes:
    group: sagemaker.services.k8s.aws
    kind: Endpoint
    metadata:
      name: xgboost-endpoint
status:
  conditions:
  - status: 'True'
    type: ACK.Adopted
```

Check resource exists in cluster
```sh
kubectl describe endpoints.sagemaker xgboost-endpoint
```

Note: This feature is not enabled in release applicationautoscaling:v0.0.1 for application autoscaling
## 9.0 Cleanup 

Few crds are common across services like `services.k8s.aws_adoptedresources.yaml`. If you have multiple controllers installed, you should be not delete the common CRDs unless you are uninstalling all the controllers.

### 9.1 Uninstall SageMaker controller and crds

```sh
export SERVICE=sagemaker
# Uninstall the Helm Chart
helm uninstall -n $ACK_K8S_NAMESPACE ack-$SERVICE-controller

# Delete the CRDs
cd $CHART_EXPORT_PATH/$SERVICE-chart/crds

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

Choose either of the options below to delete CRDs
  - [Option 1] If you have multiple controllers installed and want to delete CRDs only related to sagemaker resources
    - ```
      kubectl delete -f <CRDs which have the prefix applicationautoscaling.>
      ```
  - [Option 2] If you want to delete all CRDs
    - ```
      kubectl delete -f $CHART_EXPORT_PATH/$SERVICE-chart/crds
      ```

### 9.2 Uninstall applicationautoscaling controller and CRDs

Skip this section if you only installed SageMaker controller

```sh
export SERVICE=applicationautoscaling
# Uninstall the Helm Chart
helm uninstall -n $ACK_K8S_NAMESPACE ack-$SERVICE-controller

# Delete the CRDs
cd $CHART_EXPORT_PATH/$SERVICE-chart/crds
$ ls
applicationautoscaling.services.k8s.aws_scalabletargets.yaml
applicationautoscaling.services.k8s.aws_scalingpolicies.yaml
services.k8s.aws_adoptedresources.yaml # -> Common CRD across services
```

Choose either of the options below to delete CRDs
  - [Option 1] If you have multiple controllers installed and want to delete CRDs only related to applicationautoscaling resources
    - ```
      kubectl delete -f <CRDs which have the prefix applicationautoscaling.>
      ```
  - [Option 2] If you want to delete all CRDs
    - ```
      kubectl delete -f $CHART_EXPORT_PATH/$SERVICE-chart/crds
      ```

### 9.3 Verify charts were deleted
```sh
helm ls -n $ACK_K8S_NAMESPACE

# Delete the namespace
kubectl delete namespace $ACK_K8S_NAMESPACE
```

\[Optional\] If you used cross account resource management
```sh

kubectl delete -n ack-system configmap ack-role-account-map
kubectl delete namespace production
```