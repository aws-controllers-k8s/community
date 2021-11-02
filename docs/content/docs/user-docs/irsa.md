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
# Download the recommended managed and inline policies and apply them to the
# newly created IRSA role
BASE_URL=https://raw.githubusercontent.com/aws-controllers-k8s/${SERVICE}-controller/main
POLICY_ARN_URL=${BASE_URL}/config/iam/recommended-policy-arn
POLICY_ARN_STRINGS="$(wget -qO- ${POLICY_ARN_URL})"

INLINE_POLICY_URL=${BASE_URL}/config/iam/recommended-inline-policy
INLINE_POLICY="$(wget -qO- ${INLINE_POLICY_URL})"

while IFS= read -r POLICY_ARN; do
    echo -n "Attaching $POLICY_ARN ... "
    aws iam attach-role-policy \
        --role-name "${ACK_CONTROLLER_IAM_ROLE}" \
        --policy-arn "${POLICY_ARN}"
    echo "ok."
done <<< "$POLICY_ARN_STRINGS"

if [ ! -z "$INLINE_POLICY" ]; then
    echo -n "Putting inline policy ... "
    aws iam put-role-policy \
        --role-name "${ACK_CONTROLLER_IAM_ROLE}" \
        --policy-name "ack-recommended-policy" \
        --policy-document "$INLINE_POLICY"
    echo "ok."
fi
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


## OpenShift single AWS account pre-installation 
### Summary
When ACK service controllers are installed via OperatorHub, a cluster administrator will need to perform the following pre-installation steps to provide the controller any credentials and authentication context it needs to interact with the AWS API.

Rather than setting up a `ServiceAccount` like in the EKS instructions above, you need to use IAM users and policies. You will then set the required authentication credentials inside a `ConfigMap` and a `Secret`.

The following directions will use the Elasticache controller as an example, but the instructions should apply to any ACK controller. Just make sure to appropriately name any values that include `elasticache` in them.

### Step 1: Create a user and enable programmatic access

Create a user with the `aws` CLI (named `ack-elasticache-service-controller` in our example):
```bash
aws iam create-user --user-name ack-elasticache-service-controller
```

Enable programmatic access for the user you just created:
```bash
aws iam create-access-key --user-name ack-elasticache-service-controller
```

You should see output with important credentials:
```json
{
    "AccessKey": {
        "UserName": "ack-elasticache-service-controller",
        "AccessKeyId": "00000000000000000000",
        "Status": "Active",
        "SecretAccessKey": "abcdefghIJKLMNOPQRSTUVWXYZabcefghijklMNO",
        "CreateDate": "2021-09-30T19:54:38+00:00"
    }
}
```

This is the user that will end up representing our ACK service controller, which means these are the credentials we’ll eventually pass to our controller. Save or note `AccessKeyId` and `SecretAccessKey` for use in a later step.

### Step 2: Give the user permissions by applying an access policy

{{% hint type="info" title="Note on permissions" %}}
AWS best practice is to provision permissions to groups and then bind specific users to those groups. In this example, we are directly applying permissions to a user.
{{% /hint %}}

Each service controller repository provides a recommended policy ARN for use with the controller. For an example, see the recommended policy for [Elasticache here](https://github.com/aws-controllers-k8s/elasticache-controller/blob/main/config/iam/recommended-policy-arn).

Attach the recommended policy to the user we created in the previous step:
```bash
aws iam attach-user-policy \
    --user-name ack-elasticache-service-controller \
    --policy-arn 'arn:aws:iam::aws:policy/AmazonElastiCacheFullAccess'
```

Since the previous command has no output, you can verify that the policy applied properly:
```bash
aws iam list-attached-user-policies --user-name ack-elasticache-service-controller
```

### Step 3: Create the default ACK namespace

Create the namespace for any ACK controllers you might install. The controllers as they are packaged in OperatorHub and OLM expect the namespace to be `ack-system`.
```bash
oc new-project ack-system
```

### Step 4: Create required `ConfigMap` and `Secret` in OpenShift

Enter the `ack-system` namespace. Create a file, `config.txt`, with the following variables, leaving `ACK_WATCH_NAMESPACE` blank so the controller can properly watch all namespaces, and change any other values to suit your needs:

```bash
ACK_ENABLE_DEVELOPMENT_LOGGING=true
ACK_LOG_LEVEL=debug
ACK_WATCH_NAMESPACE=
AWS_REGION=us-west-2
ACK_RESOURCE_TAGS=hellofromocp
```

Now use `config.txt` to create a `ConfigMap` in your OpenShift cluster:
```bash
oc create configmap \
--namespace ack-system \
--from-env-file=config.txt ack-user-config
```

Save another file, `secrets.txt`, with the following authentication values, which you should have saved from earlier when you created your user's access keys:
```bash
AWS_ACCESS_KEY_ID=00000000000000000000
AWS_SECRET_ACCESS_KEY=abcdefghIJKLMNOPQRSTUVWXYZabcefghijklMNO
```

Use `secrets.txt` to create a `Secret` in your OpenShift cluster:
```bash
oc create secret generic \
--namespace ack-system \
--from-env-file=secrets.txt ack-user-secrets
```

{{% hint type="warning" title="Warning" %}}
If you change the name of either the `ConfigMap` or the `Secret` from the values given above, i.e. `ack-user-config` and `ack-user-secrets`, then installations from OperatorHub will not function properly. The Deployment for the controller is preconfigured for these key values.
{{% /hint %}}

### Step 5: Install the controller

Now you can follow the instructions for [installing the controller using OperatorHub](../install/#install-an-ack-service-controller-with-operatorhub-in-red-hat-openshift).



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

