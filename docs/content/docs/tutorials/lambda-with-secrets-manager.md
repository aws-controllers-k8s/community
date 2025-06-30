---
title: "Pass secrets to Lambda Function with AWS Secrets Manager"
description: "Retrieve sensitive information in a Lambda Function from AWS Secrets Manager."
lead: "Retrieve sensitive information in a Lambda Function from AWS Secrets Manager."
draft: false
menu:
  docs:
    parent: "tutorials"
weight: 43
toc: true
---

The ACK service controller for Amazon Lambda lets you manage Lambda functions directly from Kubernetes.
This guide shows you how to create a Lambda function that can retrieve sensitive data from AWS Secrets Manager.

## Setup

Although it is not necessary to use Amazon Elastic Kubernetes Service (Amazon EKS) or Amazon Elastic Container Registry (Amazon ECR) with ACK, this guide assumes that you
have access to an Amazon EKS cluster. If this is your first time creating an Amazon EKS cluster, see [Amazon EKS Setup][eks-setup]. For automated cluster creation using `eksctl`, see [Getting started with Amazon EKS - `eksctl`](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html) and create your cluster with Amazon EC2 Linux managed nodes.

## Prerequisites

This guide assumes that you have:

- Created an EKS cluster with Kubernetes version 1.16 or higher.
- AWS IAM permissions to create roles and attach policies to roles.
- Installed the following tools on the client machine used to access your Kubernetes cluster:
  - [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv1.html) - A command line tool for interacting with AWS services.
  - [kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) - A command line tool for working with Kubernetes clusters.
  - [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html) - A command line tool for working with EKS clusters.
  - [Helm 3.8+](https://helm.sh/docs/intro/install/) - A tool for installing and managing Kubernetes applications.
  - [jq](https://github.com/stedolan/jq/wiki/Installation)

### Install the Lambda ACK service controller

Log into the Helm registry that stores the ACK charts:

```bash
aws ecr-public get-login-password --region us-east-1 | helm registry login --username AWS --password-stdin public.ecr.aws
```

Deploy the ACK service controller for Amazon Lambda using the [lambda-chart Helm chart](https://gallery.ecr.aws/aws-controllers-k8s/lambda-chart). This example creates resources in the `us-west-2` region, but you can use any other region supported in AWS.

```bash
SERVICE=lambda
RELEASE_VERSION=$(curl -sL https://api.github.com/repos/aws-controllers-k8s/${SERVICE}-controller/releases/latest | jq -r '.tag_name | ltrimstr("v")')
helm install --create-namespace -n ack-system oci://public.ecr.aws/aws-controllers-k8s/lambda-chart "--version=${RELEASE_VERSION}" --generate-name --set=aws.region=us-west-2
```

For a full list of available values to the Helm chart, please [review the values.yaml file](https://github.com/aws-controllers-k8s/lambda-controller/blob/main/helm/values.yaml).

### Configure IAM permissions

Once the service controller is deployed [configure the IAM permissions](https://aws-controllers-k8s.github.io/community/docs/user-docs/irsa/) for the
controller to invoke the Lambda API. For full details, please review the AWS Controllers for Kubernetes documentation
for [how to configure the IAM permissions](https://aws-controllers-k8s.github.io/community/docs/user-docs/irsa/). If you follow the examples in the documentation, use the
value of `lambda` for `SERVICE`.

### Create Secret in Secrets Manger
To test our Lambda function's integration with AWS Secrets Manager we'll need to create a sample secret value. We can create a new secret with the aws cli.

```bash
aws secretsmanager create-secret --name test-secret --secret-string "secret value"
```

The ACK Secrets Manager service controller can also be used to create and manage secrets directly from Kubernetes. See, [Create a Secret with AWS Secrets Manager](https://aws-controllers-k8s.github.io/community/docs/tutorials/secrets-manager-example/)

### Create Lambda function handler
The Lambda function handler is the method in your function code that processes events. When your function is invoked, Lambda runs the handler method.

```bash
cat > index.mjs << 'EOF'
import http from 'http';

export const handler = async (event) => {
    try {
        const secretName = process.env.TEST_SECRET_ARN;
        const options = {
            hostname: 'localhost',
            port: 2773,
            path: `/secretsmanager/get?secretId=${secretName}`,
            headers: {
                'X-Aws-Parameters-Secrets-Token': process.env.AWS_SESSION_TOKEN
            }
        };

        const response = await new Promise((resolve, reject) => {
            http.get(options, (res) => {
                let data = '';
                res.on('data', (chunk) => { data += chunk; });
                res.on('end', () => {
                    resolve({ 
                        statusCode: res.statusCode, 
                        body: data 
                    });
                });
            }).on('error', reject);
        });

        const secret = JSON.parse(response.body).SecretString;
        console.log('Retrieved secret:', secret);

        return {
            statusCode: response.statusCode,
            body: JSON.stringify({
                message: 'Successfully retrieved secret',
                secretRetrieved: true
            })
        };
    } catch (error) {
        console.error('Error:', error);
        return {
            statusCode: 500,
            body: JSON.stringify({
                message: 'Error retrieving secret',
                error: error.message
            })
        };
    }
};
EOF
```

To package the function handler we then need to add it to a zip file.

```bash
zip -r function.zip index.mjs
```

### Create an IAM Execution Role for the Lambda function
Our Lambda function will need use an execution role that can access the secret in AWS Secrets Manager.

Create the IAM role:

```bash
read -r -d '' TRUST_RELATIONSHIP <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "lambda.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
EOF
echo "${TRUST_RELATIONSHIP}" > trust.json

ACK_LAMBDA_IAM_ROLE="ack-lambda-function"
ACK_LAMBDA_IAM_ROLE_DESCRIPTION="Role for ACK managed Lamdba function"
aws iam create-role --role-name "${ACK_LAMBDA_IAM_ROLE}" --assume-role-policy-document file://trust.json --description "${ACK_LAMBDA_IAM_ROLE_DESCRIPTION}"
ACK_LAMBDA_IAM_ROLE_ARN=$(aws iam get-role --role-name=$ACK_LAMBDA_IAM_ROLE --query Role.Arn --output text)
```

And then attach an IAM Policy that grants read access to our secret.

```bash
SECRET_ARN=$(aws secretsmanager describe-secret --secret-id test-secret | jq ".ARN")
POLICY_NAME=ack-lambda-policy
read -r -d '' POLICY <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "secretsmanager:GetSecretValue",
            "Resource": $SECRET_ARN
        }
    ]
}
EOF
echo "${POLICY}" > policy.json

POLICY_ARN=$(aws iam create-policy --policy-name $POLICY_NAME --policy-document file://policy.json | jq ".Policy.Arn" | tr -d '"')



aws iam attach-role-policy \
        --role-name $ACK_LAMBDA_IAM_ROLE \
        --policy-arn $POLICY_ARN
```

### Deploy the Lambda Function using the ACK Lambda Controller
The following example creates a manifest that contains the Lambda function with the necessary environment variable and
IAM role to read the secret from AWS Secrets Manager. In order to limit the number of calls made to AWS Secrets Manager the [AWS Parameter and Secrets Lambda extension](https://aws.amazon.com/blogs/compute/using-the-aws-parameter-and-secrets-lambda-extension-to-cache-parameters-and-secrets/) layer is applied.

```bash
BASE64_ZIP=$(cat function.zip | base64)
TEST_SECRET_ARN=$(aws secretsmanager describe-secret --secret-id test-secret | jq ".ARN")

read -r -d '' LAMBDA_MANIFEST <<EOF
apiVersion: lambda.services.k8s.aws/v1alpha1
kind: Function
metadata:
  name: sample-lambda
  annotations:
    services.k8s.aws/region: us-west-2
spec:
 name: sample-lambda
 environment:
   variables:
     TEST_SECRET_ARN: $TEST_SECRET_ARN
 packageType: Zip
 runtime: nodejs18.x
 handler: index.handler
 code:
    zipFile: $BASE64_ZIP

 role: $ACK_LAMBDA_IAM_ROLE_ARN
 description: Sample function for retrieving secrets from AWS Secrets Manager
 layers:
  - "arn:aws:lambda:us-west-2:345057560386:layer:AWS-Parameters-and-Secrets-Lambda-Extension:17"
EOF

echo "${LAMBDA_MANIFEST}" > function.yaml
```

```bash
kubectl create -f function.yaml
```

### Invoke the Lambda Function

After the Lambda function has finished deploying, you can invoke the function through the AWS CLI.

```bash
aws lambda invoke --function-name sample-lambda --region us-west-2 /dev/stdout | jq
```

You will get the output as below:

```bash
{"statusCode":200,"body":"\"Successfully retrieved secret!\""} 
```

### Cleanup

You can delete you Lambda function using the `kubectl delete` command:

```bash
kubectl delete -f function.yaml
```

The IAM role and policy can removed with the AWS CLI

```bash
aws iam detach-role-policy --role-name $ACK_LAMBDA_IAM_ROLE --policy-arn $POLICY_ARN
aws iam delete-role --role-name $ACK_LAMBDA_IAM_ROLE
aws iam delete-policy --policy-arn $POLICY_ARN
```

We can also delete our secret from AWS Secrets Manager with the AWS CLI

```bash
aws secretsmanager delete-secret --secret-id test-secret
```

To remove the Lambda ACK service controller, related CRDs, and namespaces, see [ACK Cleanup][cleanup].

To delete your EKS clusters, see [Amazon EKS - Deleting a cluster][cleanup-eks].

[eks-setup]: https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/deep-learning-containers-eks-setup.html
[cleanup]: ../../user-docs/cleanup/
[cleanup-eks]: https://docs.aws.amazon.com/eks/latest/userguide/delete-cluster.html

