---
title: "Create a Lambda OCI Function with the ACK Lambda Controller"
description: "Create a Lambda Function with an OCI Image Using the ACK Lambda Controller deployed on Amazon Elastic Kubernetes Service (EKS)."
lead: "Create a Lambda Function with an OCI Image Using Elastic Kubernetes Service (EKS)."
draft: false
menu:
  docs:
    parent: "tutorials"
weight: 46
toc: true
---

The ACK service controller for Amazon Lambda lets you manage Lambda functions directly from Kubernetes.
This guide shows you how to create a Lambda function with OCI image using a single Kubernetes resource manifest.

## Setup

Although it is not necessary to use Amazon Elastic Kubernetes Service (Amazon EKS) or Amazon Elastic Container Registry (Amazon ECR) with ACK, this guide assumes that you
have access to an Amazon EKS cluster. If this is your first time creating an Amazon EKS cluster and Amazon ECR repository, see
[Amazon EKS Setup][eks-setup] and [Amazon ECR Setup](https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html). 

### Prerequisites

This guide assumes that you have:

- Created an EKS cluster with Kubernetes version 1.16 or higher.
- Have access to Amazon ECR
- AWS IAM permissions to create roles and attach policies to roles.
- Installed the following tools on the client machine used to access your Kubernetes cluster:
  - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting with AWS services.
  - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working with Kubernetes clusters.
  - [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html) - A command line tool for working with EKS clusters.
  - [Helm 3.8+](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications.
  - [Docker](https://docs.docker.com/engine/install/) - A tool to build, share, and run containers.

### Install the ACK service controller for Lambda

Log into the Helm registry that stores the ACK charts:

```bash
aws ecr-public get-login-password --region us-west-2 | helm registry login --username AWS --password-stdin public.ecr.aws
```

Deploy the ACK service controller for Amazon Lambda using the [lambda-chart Helm chart](https://gallery.ecr.aws/aws-controllers-k8s/lambda-chart). This example creates resources in the `us-west-2` region, but you can use any other region supported in AWS.

```bash
SERVICE=lambda
RELEASE_VERSION=$(curl -sL "https://api.github.com/repos/aws-controllers-k8s/${SERVICE}-controller/releases/latest" | grep '"tag_name":' | cut -d'"' -f4)
helm install --create-namespace -n ack-system oci://public.ecr.aws/aws-controllers-k8s/lambda-chart "--version=${RELEASE_VERSION}" --generate-name --set=aws.region=us-west-2
```

For a full list of available values to the Helm chart, please [review the values.yaml file](https://github.com/aws-controllers-k8s/lambda-controller/blob/main/helm/values.yaml).

### Configure IAM permissions

Once the service controller is deployed [configure the IAM permissions](https://aws-controllers-k8s.github.io/community/docs/user-docs/irsa/) for the
controller to invoke the Lambda API. For full details, please review the AWS Controllers for Kubernetes documentation
for [how to configure the IAM permissions](https://aws-controllers-k8s.github.io/community/docs/user-docs/irsa/). If you follow the examples in the documentation, use the
value of `lambda` for `SERVICE`.

## Create Lambda function handler
The Lambda [function handler](https://docs.aws.amazon.com/lambda/latest/dg/nodejs-handler.html) is the method in your function code that processes events. When your function is invoked, Lambda runs the handler method.

```bash
cat <<EOF > app.js
exports.handler = async (event) => {
    const response = {
        statusCode: 200
        body: JSON.stringify('Hello from Lambda!')
    };
    return response;
};
EOF
```

## Create and Build a Docker Image
Create a Dockerfile that will be used to build the image for our Lambda function:

```bash
cat <<EOF > Dockerfile
FROM public.ecr.aws/lambda/nodejs:14

COPY app.js package.json ./

RUN npm install

CMD [ "app.handler" ]
EOF
```
Build the Docker image in your local environment. You will need to install dependencies using `npm`:

```shell
npm init -y
docker build -t hello-world .
```
## Publish the Docker image to ECR
Publish the Docker image to an ECR repository. It's a requirement for container images to be published to the ECR repository to run Lambda OCI image functions.

```shell
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
export AWS_REGION=us-west-2

aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com
aws ecr create-repository --repository-name hello-world --image-scanning-configuration scanOnPush=true --image-tag-mutability MUTABLE
docker tag  "hello-world:latest ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/hello-world:latest"
docker push "${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/hello-world:latest"
```

## Deploy the Lambda OCI function using the ACK Lambda controller
The following example creates a manifest that contains the Lambda OCI function. It then uses `kubectl` to create the resource in Kubernetes:

```shell
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
export IMAGE_URI="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/hello-world:latest "
export FUNCTION_NAME="lambda-oci-ack"
export LAMBDA_ROLE="arn:aws:iam::${AWS_ACCOUNT_ID}:role/lambda_basic_execution"

read -r -d '' LAMBDA_MANIFEST <<EOF
apiVersion: lambda.services.k8s.aws/v1alpha1
kind: Function
metadata:
 name: $FUNCTION_NAME
 annotations:
   services.k8s.aws/region: $AWS_REGION
spec:
 name: $FUNCTION_NAME
 packageType: Image
 code:
     imageURI: $IMAGE_URI
 role: $LAMBDA_ROLE
 description: function created by ACK lambda-controller e2e tests
EOF

echo "${LAMBDA_MANIFEST}" > function.yaml

kubectl create -f function.yaml
```
You should get a confirmation that the function was created successfully.

```
function.lambda.services.k8s.aws/lambda-oci-ack created
```
To get details about the Lambda function, run the following.

```bash
kubectl describe "function/${FUNCTION_NAME}"
```

## Invoke the Lambda OCI Function
After you have verified that the Lambda OCI function is deployed correctly, you can invoke the function through the [AWS CLI](https://docs.aws.amazon.com/cli/latest/reference/lambda/index.html).

```bash
aws lambda invoke --function-name ${FUNCTION_NAME} --region us-west-2 /dev/stdout | jq
```

You will get the output as below:
```
{"statusCode":200,"body":"\"Hello from Lambda!\""} 
```

## Next steps

The ACK service controller for Amazon Lambda is based on the [Amazon Lambda API](https://docs.aws.amazon.com/lambda/latest/dg/API_Reference.html).

Refer to [API Reference](https://aws-controllers-k8s.github.io/community/reference/) for *Lambda* to find
all the supported Kubernetes custom resources and fields.

### Cleanup

You can delete your Lambda OCI function using the `kubectl delete` command:

```bash
kubectl delete -f function.yaml
```

To remove the Lambda ACK service controller, related CRDs, and namespaces, see [ACK Cleanup][cleanup].

To delete your EKS clusters, see [Amazon EKS - Deleting a cluster][cleanup-eks].

[eks-setup]: https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/deep-learning-containers-eks-setup.html
[cleanup]: ../../user-docs/cleanup/
[cleanup-eks]: https://docs.aws.amazon.com/eks/latest/userguide/delete-cluster.html

