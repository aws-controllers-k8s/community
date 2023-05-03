---
title: "Run Spark jobs using the ACK EMR on EKS controller"
description: "ACK service controller for EMR on EKS enables customers to run spark jobs on EKS clusters"
lead: "Run Spark jobs using ACK service controller for EMR on EKS."
draft: false
menu:
  docs:
    parent: "tutorials"
weight: 47
toc: true
---
Using ACK service controller for EMR on EKS, customers have the ability to define and run EMR jobs directly from their Kubernetes clusters. EMR on EKS manages the lifecycle of these jobs and it is [3.5 times faster than open-source Spark](https://aws.amazon.com/blogs/big-data/amazon-emr-on-amazon-eks-provides-up-to-61-lower-costs-and-up-to-68-performance-improvement-for-spark-workloads/) because it uses highly optimized EMR runtime  

To get started, you can download the EMR on EKS controller image from [Amazon ECR](https://gallery.ecr.aws/aws-controllers-k8s/emrcontainers-controller) and run Spark jobs in minutes. ACK service controller for EMR on EKS is **generally available**. To learn more about EMR on EKS, visit our [documentation](https://docs.aws.amazon.com/emr/latest/EMR-on-EKS-DevelopmentGuide/emr-eks.html).

## Installation steps
Here are the steps involved for installing EMR on EKS controller.
1. [Install EKS cluster](#install-eks-cluster)
  - [Create IAM Identity mapping](#create-iam-identity-mapping)
2. [Install emrcontainers-controller](#install-emrcontainers-controller-in-your-eks-cluster)
  - [Configure IRSA for emr on eks controller](#configure-irsa-for-emr-on-eks-controller)
3. [Create EMR VirtualCluster](#create-emr-virtualcluster)
  - [Create EMR Job Execution Role & configure IRSA](#create-job-execution-role)
4. [Run a sample job](#run-a-sample-spark-job)

#### Prereqs
Install these tools before proceeding:
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2-linux.html)
- `kubectl` - [the Kubernetes CLI](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)
- `eksctl` - [the CLI for AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html)
- `yq` - [YAML processor](https://github.com/mikefarah/yq)
- `helm` - [the package manager for Kubernetes](https://helm.sh/docs/intro/install/)

Configure AWS CLI with sufficient permissions to install EKS cluster. Please see [documentation](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html) for further guidance

## Install EKS cluster

You can either create an EKS cluster or re-use existing one. Below listed are steps for creating new EKS cluster. Let's export environment variables that are needed for the EMR on EKS cluster setup. Please copy and paste commands into terminal for faster provisioning
```
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
export EKS_CLUSTER_NAME="ack-emr-eks"
export AWS_REGION="us-west-2"
```
We'll use eksctl to install EKS cluster.
```
eksctl create cluster -f - << EOF
---
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: ${EKS_CLUSTER_NAME}
  region: ${AWS_REGION}
  version: "1.23"

managedNodeGroups:
  - instanceType: m5.xlarge
    name: ${EKS_CLUSTER_NAME}-ng
    desiredCapacity: 2

iam:
  withOIDC: true
EOF
```
#### Create IAM Identity mapping
We need to create emr-containers identity in EKS cluster so that EMR service has proper RBAC permissions needed to  run and manage Spark jobs
```
export EMR_NAMESPACE=emr-ns
echo "creating namespace for $SERVICE"
kubectl create ns $EMR_NAMESPACE

echo "creating iamidentitymapping using eksctl"
eksctl create iamidentitymapping \
   --cluster $EKS_CLUSTER_NAME \
   --namespace $EMR_NAMESPACE \
   --service-name "emr-containers"
```
**Expected outcome**
```
2022-08-26 09:07:42 [ℹ]  created "emr-ns:Role.rbac.authorization.k8s.io/emr-containers"
2022-08-26 09:07:42 [ℹ]  created "emr-ns:RoleBinding.rbac.authorization.k8s.io/emr-containers"
2022-08-26 09:07:42 [ℹ]  adding identity "arn:aws:iam::012345678910:role/AWSServiceRoleForAmazonEMRContainers" to auth ConfigMap
```
## Install emrcontainers-controller in your EKS cluster
Now we can go ahead and install EMR on EKS controller. First, let's export environment variables needed for setup
```
export SERVICE=emrcontainers
export RELEASE_VERSION=$(curl -sL https://api.github.com/repos/aws-controllers-k8s/$SERVICE-controller/releases/latest | grep '"tag_name":' | cut -d'"' -f4)
export ACK_SYSTEM_NAMESPACE=ack-system
```
We cam use Helm for the installation
```
echo "installing ack-$SERVICE-controller"
aws ecr-public get-login-password --region us-east-1 | helm registry login --username AWS --password-stdin public.ecr.aws
helm install --create-namespace -n $ACK_SYSTEM_NAMESPACE ack-$SERVICE-controller \
  oci://public.ecr.aws/aws-controllers-k8s/$SERVICE-chart --version=$RELEASE_VERSION --set=aws.region=$AWS_REGION
```
**Expected outcome**
```
NAME: ack-emrcontainers-controller
LAST DEPLOYED: Fri Aug 26 09:05:08 2022
NAMESPACE: ack-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
emrcontainers-chart has been installed.
This chart deploys "public.ecr.aws/aws-controllers-k8s/emrcontainers-controller:0.0.6".

Check its status by running:
  kubectl --namespace ack-system get pods -l "app.kubernetes.io/instance=ack-emrcontainers-controller"

You are now able to create Amazon EMR on EKS (EMRContainers) resources!
```
#### Configure IRSA for emr on eks controller
Once the controller is deployed, you need to setup IAM permissions for the controller so that it can create and manage resources using EMR, S3 and other API's. We will use [IAM Roles for Service Account](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html) to secure this IAM role so that only EMR on EKS controller service account can assume the permissions assigned.

Please follow [how to configure IAM permissions](https://aws-controllers-k8s.github.io/community/docs/user-docs/irsa/) for IRSA setup. Make sure to change the value for **`SERVICE`** to **`emrcontainers`**

After completing all the steps, please validate annotation for service account before proceeding.
```
# validate annotation
kubectl get pods -n $ACK_SYSTEM_NAMESPACE
CONTROLLER_POD_NAME=$(kubectl get pods -n $ACK_SYSTEM_NAMESPACE --selector=app.kubernetes.io/name=emrcontainers-chart -o jsonpath='{.items..metadata.name}')
kubectl describe pod -n $ACK_SYSTEM_NAMESPACE $CONTROLLER_POD_NAME | grep "^\s*AWS_"
```
**Expected outcome**
```
AWS_REGION:                      us-west-2
AWS_ENDPOINT_URL:                
AWS_ROLE_ARN:                    arn:aws:iam::012345678910:role/ack-emrcontainers-controller
AWS_WEB_IDENTITY_TOKEN_FILE:     /var/run/secrets/eks.amazonaws.com/serviceaccount/token (http://eks.amazonaws.com/serviceaccount/token)
```

## Create EMR VirtualCluster
We can now create EMR Virtual Cluster. An EMR Virtual Cluster is mapped to a Kubernetes namespace. EMR uses virtual clusters to run jobs and host endpoints.
```
cat << EOF > virtualcluster.yaml
---
apiVersion: emrcontainers.services.k8s.aws/v1alpha1
kind: VirtualCluster
metadata:
  name: my-ack-vc
spec:
  name: my-ack-vc
  containerProvider:
    id: $EKS_CLUSTER_NAME
    type_: EKS
    info:
      eksInfo:
        namespace: emr-ns
EOF
```
Let's create a virtualcluster
```
envsubst < virtualcluster.yaml | kubectl apply -f -
kubectl describe virtualclusters
```
**Expected outcome**
```
Name:         my-ack-vc
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  emrcontainers.services.k8s.aws/v1alpha1
Kind:         VirtualCluster
...
Status:
  Ack Resource Metadata:
    Arn:               arn:aws:emr-containers:us-west-2:012345678910:/virtualclusters/dxnqujbxexzri28ph1wspbxo0
    Owner Account ID:  012345678910
    Region:            us-west-2
  Conditions:
    Last Transition Time:  2022-08-26T17:21:26Z
    Message:               Resource synced successfully
    Reason:                
    Status:                True
    Type:                  ACK.ResourceSynced
  Id:                      dxnqujbxexzri28ph1wspbxo0
Events:                    <none>
```

#### Create Job Execution Role
In order to run sample spark job, we need to create EMR Job Execution Role. This Role will have IAM permissions such as S3, CloudWatch Logs for running your job. We will use IRSA to secure this job role

```
ACK_JOB_EXECUTION_ROLE="ack-${SERVICE}-jobexecution-role"
ACK_JOB_EXECUTION_IAM_ROLE_DESCRIPTION="IRSA role for ACK ${SERVICE} Job Execution"

cat <<EOF > job_trust.json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "",
            "Effect": "Allow",
            "Principal": {
                "Service": "ec2.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
EOF
aws iam create-role --role-name "${ACK_JOB_EXECUTION_ROLE}" \
    --assume-role-policy-document file://job_trust.json  \
    --description "${ACK_JOB_EXECUTION_IAM_ROLE_DESCRIPTION}"

export ACK_JOB_EXECUTION_ROLE_ARN=$(aws iam get-role --role-name=$ACK_JOB_EXECUTION_ROLE --query Role.Arn --output text)    
```
```
cat <<EOF > job_policy.json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:ListBucket"
            ],
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket",
                "s3:GetObject*"
            ],
            "Resource": [
                "arn:aws:s3:::tripdata",
                "arn:aws:s3:::tripdata/*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "logs:PutLogEvents",
                "logs:CreateLogStream",
                "logs:DescribeLogGroups",
                "logs:DescribeLogStreams"
            ],
            "Resource": [
                "arn:aws:logs:*:*:*"
            ]
        }
    ]
}
EOF
echo "Creating ACK-${SERVICE}-JobExecution-POLICY"
aws iam create-policy   \
  --policy-name ack-${SERVICE}-jobexecution-policy \
  --policy-document file://job_policy.json

echo -n "Attaching IAM policy ..."
aws iam attach-role-policy \
  --role-name "${ACK_JOB_EXECUTION_ROLE}" \
  --policy-arn "arn:aws:iam::${AWS_ACCOUNT_ID}:policy/ack-${SERVICE}-jobexecution-policy"

aws emr-containers update-role-trust-policy \
  --cluster-name ${EKS_CLUSTER_NAME} \
  --namespace ${EMR_NAMESPACE} \
  --role-name ${ACK_JOB_EXECUTION_ROLE}
```
## Run a Sample Spark Job

Before running a sample job, let's create CloudWatch Logs and an S3 bucket to store EMR on EKS logs
```
export RANDOM_ID1=$(LC_ALL=C tr -dc a-z0-9 </dev/urandom | head -c 8)

aws logs create-log-group --log-group-name=/emr-on-eks-logs/$EKS_CLUSTER_NAME
aws s3 mb s3://$EKS_CLUSTER_NAME-$RANDOM_ID1
```

Now let's submit sample spark job
```
echo "checking if VirtualCluster Status is "True""
VC=$(kubectl get virtualcluster -o jsonpath='{.items..metadata.name}')
kubectl describe virtualcluster/$VC | yq e '.Status.Conditions.Status'

export RANDOM_ID2=$(LC_ALL=C tr -dc a-z0-9 </dev/urandom | head -c 8)

cat << EOF > jobrun.yaml
---
apiVersion: emrcontainers.services.k8s.aws/v1alpha1
kind: JobRun
metadata:
  name: my-ack-jobrun-${RANDOM_ID2}
spec:
  name: my-ack-jobrun-${RANDOM_ID2}
  virtualClusterRef:
    from:
      name: my-ack-vc
  executionRoleARN: "${ACK_JOB_EXECUTION_ROLE_ARN}"
  releaseLabel: "emr-6.7.0-latest"
  jobDriver:
    sparkSubmitJobDriver:
      entryPoint: "local:///usr/lib/spark/examples/src/main/python/pi.py"
      entryPointArguments:
      sparkSubmitParameters: "--conf spark.executor.instances=2 --conf spark.executor.memory=1G --conf spark.executor.cores=1 --conf spark.driver.cores=1"
  configurationOverrides: |
    ApplicationConfiguration: null
    MonitoringConfiguration:
      CloudWatchMonitoringConfiguration:
        LogGroupName: /emr-on-eks-logs/$EKS_CLUSTER_NAME
        LogStreamNamePrefix: pi-job
      S3MonitoringConfiguration:
        LogUri: s3://$EKS_CLUSTER_NAME-$RANDOM_ID1   
EOF
```
```
echo "running sample job"
kubectl apply -f jobrun.yaml
kubectl describe jobruns
```
**Expected outcome**
```
Name:         my-ack-jobrun-t2rpcpks
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  emrcontainers.services.k8s.aws/v1alpha1
Kind:         JobRun
...
Status:
  Ack Resource Metadata:
    Arn:               arn:aws:emr-containers:us-west-2:012345678910:/virtualclusters/dxnqujbxexzri28ph1wspbxo0/jobruns/000000030mrd934cdqc
    Owner Account ID:  012345678910
    Region:            us-west-2
  Conditions:
    Last Transition Time:  2022-08-26T18:29:12Z
    Message:               Resource synced successfully
    Reason:                
    Status:                True
    Type:                  ACK.ResourceSynced
  Id:                      000000030mrd934cdqc
Events:                    <none>
```

## Cleanup
Simply run these commands to cleanup your environment
```
# delete all custom resources
kubectl delete -f virtualcluster.yaml
kubectl delete -f jobrun.yaml
# note: you cannot delete jobruns until virtualcluster its mapped to is deleted

# uninstall emrcontainers controller
helm delete ack-$SERVICE-controller -n $ACK_SYSTEM_NAMESPACE

# delete namespace
kubectl delete ns $ACK_SYSTEM_NAMESPACE
kubectl delete ns $EMR_NAMESPACE

# delete aws resources
aws logs delete-log-group --log-group-name=/emr-on-eks-logs/$EKS_CLUSTER_NAME
aws s3 rm s3://$EKS_CLUSTER_NAME-$RANDOM_ID1 --recursive
aws s3 rb s3://$EKS_CLUSTER_NAME-$RANDOM_ID1 

# delete EKS cluster
eksctl delete cluster --name "${EKS_CLUSTER_NAME}"
```
## Limitations

* You cannot delete a JobRun unless its in **error** state. There is no delete-job-run API for deleting jobs (for good reason). However, if your JobRun goes into error state, you can run `kubectl delete jobrun/<job-run-name>` to cancel the job.

## Troubleshooting

* If you run into issues creating VirtualCluster or JobRuns, check EMR on EKS controller logs for troubleshooting
```
CONTROLLER_POD=$(kubectl get pod -n ${ACK_SYSTEM_NAMESPACE} -o jsonpath='{.items..metadata.name}')
kubectl logs ${CONTROLLER_POD} -n ${ACK_SYSTEM_NAMESPACE}
```
* You can enable debug logs for EMR on EKS controller if you are unable to determine cause of the error. You need to change values for `enable-development-logging` to `true` and `--log-level` to `debug`
```
CONTROLLER_DEPLOYMENT=$(kubectl get deploy -n ${ACK_SYSTEM_NAMESPACE} -o jsonpath='{.items..metadata.name}')
kubectl edit deploy/${CONTROLLER_DEPLOYMENT} -n ${ACK_SYSTEM_NAMESPACE}
```
This is how your values should look after changes are applied.
```
        - --aws-region
        - $(AWS_REGION)
        - --aws-endpoint-url
        - $(AWS_ENDPOINT_URL)
        - --enable-development-logging
        - "true"
        - --log-level
        - debug
        - --resource-tags
        - $(ACK_RESOURCE_TAGS)
```
* If you run into any issue, please create [Github issue](https://github.com/aws-controllers-k8s/community/issues). Click **New issue** and select the type of issue, add `[emr-containers] <highlevel overview>` under title, and add enough details so that we can reproduce and provide a response
