# IAM Roles for Service Accounts (IRSA)

[IAM Roles for Service Accounts][0] (IRSA) automates the provisioning and rotation of AWS Identity and Access Management (IAM) temporary credentials that a Kubernetes service account can use to call AWS APIs.

Instead of creating and distributing your AWS credentials to the containers or using the Amazon EC2 instance’s role, you can associate an IAM role with a Kubernetes service account. The applications in a Kubernetes pod container can then use an AWS SDK or the AWS CLI to make API requests to authorized AWS services.

By using the IRSA feature, you no longer need to provide extended permissions to the node IAM role so that pods on that node can call AWS APIs. You can scope IAM permissions to a service account, and only pods that use that service account have access to those permissions. 

!!! note "EKS is not required to use IRSA"
    You do not need to use the Amazon EKS service in order to use IRSA. You can [set up IRSA on your own Kubernetes installation][1].

The following steps demonstrate how to set up IRSA on an EKS cluster while installing the ACK ElastiCache controller using Helm charts. By modifying the variable values as needed, these steps can be applied for the installation of other ACK service controllers.

## Step 1. Create an OIDC identity provider for your cluster

Create an [OpenID Connect (OIDC) identity provider][9] for your EKS cluster using the `eksctl utils` command:
```
export EKS_CLUSTER_NAME=<eks cluster name>
eksctl utils associate-iam-oidc-provider --cluster $EKS_CLUSTER_NAME --approve
```
For detailed instructions, refer to Amazon EKS documentation on how to [create an IAM OIDC provider for your cluster][3].

## Step 2. Create an IAM role and policy for your service account

### Create an IAM role for your ACK service controller
```
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
OIDC_PROVIDER=$(aws eks describe-cluster --name $EKS_CLUSTER_NAME --region <eks cluster region> --query "cluster.identity.oidc.issuer" --output text | sed -e "s/^https:\/\///")
ACK_K8S_NAMESPACE=ack-system
ACK_K8S_SERVICE_ACCOUNT_NAME=ack-elasticache-controller

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

# Update the service name variables as needed
AWS_SERVICE_NAME='elasticache'
ACK_CONTROLLER_IAM_ROLE="ack-${AWS_SERVICE_NAME}-controller"
ACK_CONTROLLER_IAM_ROLE_DESCRIPTION='IRSA role for ACK $AWS_SERVICE_NAME controller deployment on EKS cluster using Helm charts'
aws iam create-role --role-name "${ACK_CONTROLLER_IAM_ROLE}" --assume-role-policy-document file://trust.json --description "${ACK_CONTROLLER_IAM_ROLE_DESCRIPTION}"
```
Take note of the Amazon Resource Name (ARN) that is associated with your new ACK service controller IAM role. You will need this ARN for later steps. 

### Attach an IAM policy to the IAM role
```
# This example uses a pre-existing policy for ElastiCache
# Create an IAM policy and use its ARN and update IAM_POLICY_ARN variable as needed
IAM_POLICY_ARN='arn:aws:iam::aws:policy/AmazonElastiCacheFullAccess'
aws iam attach-role-policy \
    --role-name "${ACK_CONTROLLER_IAM_ROLE}" \
    --policy-arn "$IAM_POLICY_ARN"
```

For detailed instructions, refer to Amazon EKS documentation on [creating an IAM role and policy for your service account][4].

## Step 3. Associate an IAM role to a service account

If you [installed your ACK service controller using a Helm chart][7], then a service account already exists on your cluster. However, it is still neccessary to associate an IAM role with the service account. 

Verify that your service account exists using `kubectl describe`:
```
kubectl describe serviceaccount/$ACK_K8S_SERVICE_ACCOUNT_NAME -n $ACK_K8S_NAMESPACE
```
Note that the Amazon Resource Name (ARN) of the IAM role that you created is not yet set as an annotation for the service account. 

Use the following commands to associate an IAM role to a service account:
```
# Annotate the service account with the ARN
ISRA_ROLE_ARN=eks.amazonaws.com/role-arn=<IAM Role Arn>
kubectl annotate serviceaccount -n $ACK_K8S_NAMESPACE $ACK_K8S_SERVICE_ACCOUNT_NAME $ISRA_ROLE_ARN
```

If you have not done so already, update the AWS region of your ACK service controller using the following commands:
```
# Update your controller with the correct AWS region. Example: us-east-1
AWS_SERVICE_NAME='elasticache'
ACK_K8S_RELEASE_NAME=ack-$AWS_SERVICE_NAME-controller
AWS_REGION=<aws region id>
kubectl -n $ACK_K8S_NAMESPACE set env deployment/$ACK_K8S_RELEASE_NAME \
    AWS_REGION="$AWS_ACCOUNT_ID"
```

### Edit Helm charts directly on your local file system

If you did not already install your ACK service controller using Helm charts, you can directly edit the Helm chart YAML files on your local file system. 

Update the [`values.yaml` file][8] and set the value for `aws.region` and `serviceAccount.annotations`.

```
# Update variables as needed for your ACK service controller
ACK_CONTROLLER_HELM_CHARTS_DIR=<directory containing Helm chart for ACK service controller>
AWS_SERVICE_NAME='elasticache'
ACK_K8S_NAMESPACE=ack-system
ACK_K8S_RELEASE_NAME=ack-$AWS_SERVICE_NAME-controller

kubectl create namespace "$ACK_K8S_NAMESPACE"
cd "$ACK_CONTROLLER_HELM_CHARTS_DIR"

# Dry run and view the output
helm install --debug --dry-run --namespace "$ACK_K8S_NAMESPACE" "$ACK_K8S_RELEASE_NAME" .
# Install the ACK service controller on your cluster
helm install --namespace "$ACK_K8S_NAMESPACE" "$ACK_K8S_RELEASE_NAME" .
```

Verify that the service account has been created on your cluster and that its annotations include the IAM role for your ACK service controller.
```
kubectl describe serviceaccount/$ACK_K8S_SERVICE_ACCOUNT_NAME -n $ACK_K8S_NAMESPACE
```

For detailed instructions, refer to Amazon EKS documentation on how to [associate an IAM role to a service account ][5].

## Step 4: Verify successful setup

When AWS clients or SDKs connect to an AWS API, they detect an [AssumeRoleWithWebIdentity][2] security token to assume the IAM role. 

Verify that the `AWS_WEB_IDENTITY_TOKEN_FILE` and `AWS_ROLE_ARN` environment variables exist for your Kubernetes pod using the following commands:
```
kubectl get pods -A
kubectl exec -n <NAMESPACE> <NAME> env | grep AWS
```
The output should look similar to the following:
```
AWS_VPC_K8S_CNI_LOGLEVEL=DEBUG
AWS_ROLE_ARN=arn:aws:iam::<AWS_ACCOUNT_ID>:role/<IAM_ROLE_NAME>
AWS_WEB_IDENTITY_TOKEN_FILE=/var/run/secrets/eks.amazonaws.com/serviceaccount/token
```

[0]: https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
[1]: https://github.com/aws/amazon-eks-pod-identity-webhook/blob/master/SELF_HOSTED_SETUP.md
[2]: https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRoleWithWebIdentity.html
[3]: https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html
[4]: https://docs.aws.amazon.com/eks/latest/userguide/create-service-account-iam-policy-and-role.html
[5]: https://docs.aws.amazon.com/eks/latest/userguide/specify-service-account-role.html
[6]: https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html#installing-eksctl
[7]: https://aws-controllers-k8s.github.io/community/user-docs/install/
[8]: https://github.com/aws-controllers-k8s/elasticache-controller/blob/main/helm/values.yaml
[9]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html

