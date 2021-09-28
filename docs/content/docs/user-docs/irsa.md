---
title: "Configure IAM Permissions"
description: "Setting up ACK with IAM Roles for Service Accounts"
lead: "Set up ACK with IAM Roles for Service Accounts"
draft: false
menu:
  docs:
    parent: "getting-started"
weight: 20
toc: true
---

[IAM Roles for Service Accounts][irsa-docs], or IRSA, is a system that automates the
provisioning and rotation of IAM temporary credentials (called a Web Identity)
that a Kubernetes `ServiceAccount` can use to call AWS APIs.

{{% hint type="info" title="TL;DR:" %}}
Instead of creating and distributing your AWS credentials to the containers or
using the Amazon EC2 instance’s role, you can associate an IAM role with a Kubernetes
service account. The applications in a Kubernetes pod container can then use an
AWS SDK or the AWS CLI to make API requests to authorized AWS services.

Quicklinks:
* [Setup IRSA for EKS cluster](./#step-1-create-an-oidc-identity-provider-for-your-cluster)
* [Setup IRSA for non-EKS cluster](https://github.com/aws/amazon-eks-pod-identity-webhook/blob/master/SELF_HOSTED_SETUP.md)

Follow the quicklink OR continue reading for more details about IRSA.
{{% /hint %}}

The primary advantage of IRSA is that Kubernetes `Pods` which use the
`ServiceAccount` associated with an IAM Role can have a reduced IAM permission
footprint than the IAM Role in use for the Kubernetes EC2 worker node (known as
the EC2 Instance Profile Role). This security concept is known as **Least
Privilege**.

For example, assume you have a broadly-scoped IAM Role with permissions to
access the Instance Metadata Service (IMDS) from the EC2 worker node. If you do
not want Kubernetes `Pods` running on that EC2 Instance to have access to IMDS,
you can create a different IAM Role with a reduced permission set and associate
this reduced-scope IAM Role with the Kubernetes `ServiceAccount` the `Pod`
uses. IRSA will ensure that a special file is injected (and rotated
periodically) into the `Pod` that contains a JSON Web Token (JWT) that
encapsulates a request for temporary credentials to assume the IAM Role with
reduced permissions.

When AWS clients or SDKs connect to an AWS API, they detect the existence of
this special token file and call the [`STS::AssumeRoleWithWebIdentity`][security-token] API
to assume the IAM Role with reduced permissions.

[IAM Roles for Service Accounts][irsa-docs] (IRSA) automates the provisioning and rotation of AWS Identity and Access Management (IAM) temporary credentials that a Kubernetes service account can use to call AWS APIs.

Instead of creating and distributing your AWS credentials to the containers or using the Amazon EC2 instance’s role, you can associate an IAM role with a Kubernetes service account. The applications in a Kubernetes pod container can then use an AWS SDK or the AWS CLI to make API requests to authorized AWS services.

By using the IRSA feature, you no longer need to provide extended permissions to the node IAM role so that pods on that node can call AWS APIs. You can scope IAM permissions to a service account, and only pods that use that service account have access to those permissions.

The following steps demonstrate how to set up IRSA on an EKS cluster while installing the ACK S3 controller using Helm charts. By modifying the variable values as needed, these steps can be applied for the installation of other ACK service controllers.

## Step 1. Create an OIDC identity provider for your cluster

Create an [OpenID Connect (OIDC) identity provider][oidc-docs] for your EKS cluster using the `eksctl utils` command:
```bash
export EKS_CLUSTER_NAME=<eks cluster name>
export AWS_REGION=<aws region id>
eksctl utils associate-iam-oidc-provider --cluster $EKS_CLUSTER_NAME --region $AWS_REGION --approve
```
For detailed instructions, refer to Amazon EKS documentation on how to [create an IAM OIDC provider for your cluster][oidc-iam-docs].

## Step 2. Create an IAM role and policy for your service account

### Create an IAM role for your ACK service controller
```bash
# Update the service name variables as needed
SERVICE="s3"
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
OIDC_PROVIDER=$(aws eks describe-cluster --name $EKS_CLUSTER_NAME --region $AWS_REGION --query "cluster.identity.oidc.issuer" --output text | sed -e "s/^https:\/\///")
ACK_K8S_NAMESPACE=ack-system

ACK_K8S_SERVICE_ACCOUNT_NAME=ack-$SERVICE-controller

read -r -d '' TRUST_RELATIONSHIP <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/${OIDC_PROVIDER}"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "${OIDC_PROVIDER}:sub": "system:serviceaccount:${ACK_K8S_NAMESPACE}:${ACK_K8S_SERVICE_ACCOUNT_NAME}"
        }
      }
    }
  ]
}
EOF
echo "${TRUST_RELATIONSHIP}" > trust.json

ACK_CONTROLLER_IAM_ROLE="ack-${SERVICE}-controller"
ACK_CONTROLLER_IAM_ROLE_DESCRIPTION='IRSA role for ACK $SERVICE controller deployment on EKS cluster using Helm charts'
aws iam create-role --role-name "${ACK_CONTROLLER_IAM_ROLE}" --assume-role-policy-document file://trust.json --description "${ACK_CONTROLLER_IAM_ROLE_DESCRIPTION}"
ACK_CONTROLLER_IAM_ROLE_ARN=$(aws iam get-role --role-name=$ACK_CONTROLLER_IAM_ROLE --query Role.Arn --output text)
```

### Attach IAM policy to the IAM role

{{% hint type="info" title="Note" %}}
The command below will attach the ACK recommended policy to the IAM role. If you
wish to use any other permissions, change `IAM_POLICY_ARN` variable
{{% /hint %}}

```bash
# This example uses a pre-existing policy for Amazon S3
# Create an IAM policy and use its ARN and update IAM_POLICY_ARN variable as needed
BASE_URL=https://raw.githubusercontent.com/aws-controllers-k8s/$SERVICE-controller/main
IAM_POLICY_URL=$BASE_URL/config/iam/recommended-policy-arn
IAM_POLICY_ARN="`wget -qO- $IAM_POLICY_URL`"
aws iam attach-role-policy \
    --role-name "${ACK_CONTROLLER_IAM_ROLE}" \
    --policy-arn "$IAM_POLICY_ARN"
```

For detailed instructions, refer to Amazon EKS documentation on [creating an IAM role and policy for your service account][iam-policy].

## Step 3. Associate an IAM role to a service account and restart deployment

If you [installed your ACK service controller using a Helm chart][install-docs], then a service account already exists on your cluster. However, it is still neccessary to associate an IAM role with the service account.

Verify that your service account exists using `kubectl describe`:
```bash
kubectl describe serviceaccount/$ACK_K8S_SERVICE_ACCOUNT_NAME -n $ACK_K8S_NAMESPACE
```
Note that the Amazon Resource Name (ARN) of the IAM role that you created is not yet set as an annotation for the service account.

Use the following commands to associate an IAM role to a service account:
```bash
# Annotate the service account with the ARN
export IRSA_ROLE_ARN=eks.amazonaws.com/role-arn=$ACK_CONTROLLER_IAM_ROLE_ARN
kubectl annotate serviceaccount -n $ACK_K8S_NAMESPACE $ACK_K8S_SERVICE_ACCOUNT_NAME $IRSA_ROLE_ARN
```

Restart ACK service controller deployment using the following commands. The restart
will update service controller pods with IRSA environment variables
```bash
# Note the deployment name for ACK service controller from following command
kubectl get deployments -n $ACK_K8S_NAMESPACE
kubectl -n $ACK_K8S_NAMESPACE rollout restart deployment <ACK deployment name>
```

## Step 4: Verify successful setup

When AWS clients or SDKs connect to an AWS API, they detect an [AssumeRoleWithWebIdentity][security-token] security token to assume the IAM role.

Verify that the `AWS_WEB_IDENTITY_TOKEN_FILE` and `AWS_ROLE_ARN` environment variables exist for your Kubernetes pod using the following commands:
```bash
kubectl get pods -n $ACK_K8S_NAMESPACE
kubectl describe pod -n $ACK_K8S_NAMESPACE <NAME> | grep "^\s*AWS_"
```
The output should contain following two lines:
```bash
AWS_ROLE_ARN=arn:aws:iam::<AWS_ACCOUNT_ID>:role/<IAM_ROLE_NAME>
AWS_WEB_IDENTITY_TOKEN_FILE=/var/run/secrets/eks.amazonaws.com/serviceaccount/token
```

## Next Steps

Now that ACK service controller is setup successfully with AWS permissions, let's
validate by [creating a S3 bucket](../resource-crud)

[irsa-docs]: https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
[security-token]: https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRoleWithWebIdentity.html
[oidc-iam-docs]: https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html
[iam-policy]: https://docs.aws.amazon.com/eks/latest/userguide/create-service-account-iam-policy-and-role.html
[iam-service-account]: https://docs.aws.amazon.com/eks/latest/userguide/specify-service-account-role.html
[install-docs]: ../install/
[s3-helm-values]: https://github.com/aws-controllers-k8s/s3-controller/blob/main/helm/values.yaml
[oidc-docs]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html

