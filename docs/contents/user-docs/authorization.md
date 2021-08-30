# Configure permissions for authorization and access

There are two different Role-Based Access Control (RBAC) systems needed for ACK service controller authorization: Kubernetes RBAC and AWS IAM. 

[Kubernetes RBAC][k8s-rbac] governs a Kubernetes user's ability to read or write Kubernetes resources, while [AWS Identity and Access Management][aws-iam] (IAM) policies govern the ability of an AWS IAM role to read or write AWS resources.

[k8s-rbac]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
[aws-iam]: https://docs.aws.amazon.com/IAM/latest/UserGuide/access.html

!!! note "These two RBAC systems do not overlap"
    The Kubernetes user that makes a Kubernetes API call with `kubectl` has no association with an IAM role. Instead, the IAM role is associated with the [service account][k8s-service-account] that runs the ACK service controller's pod.

[k8s-service-account]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/

Refer to the following diagram for more details on running a Kubernetes API server with RBAC authorization mode enabled.

![Authorization in ACK](../images/authorization.png)

You will need to configure Kubernetes RBAC and AWS IAM permissions before using ACK service controllers.

## Step 1: Configure Kubernetes RBAC

As part of installation, Kubernetes `Role` resources are automatically created. These roles contain permissions to modify the Kubernetes custom resources (CRs) that the ACK service controller is responsible for.

!!! note "ACK resources are namespace-scoped"
    All Kubernetes CRs managed by an ACK service controller are namespace-scoped resources. There are no cluster-scoped ACK-managed CRs.

By default, the following Kubernetes `Role` resources are created when installing an ACK service controller:

* `ack-$SERVICE-writer`: a `Role` used for reading and mutating namespace-scoped CRs that the ACK service controller manages.
* `ack-$SERVICE-reader`: a `Role` used for reading namespaced-scoped CRs that the ACK service controller manages.

For example, installing the ACK service controller for AWS S3 creates the `ack-s3-writer` and `ack-s3-reader` roles, both with a `GroupKind` of `s3.services.k8s.aws/Bucket` within a specific Kubernetes `Namespace`.

### Bind a Kubernetes user to a Kubernetes role

Once the Kubernetes `Role` resources have been created, you can assign a specific Kubernetes `User` to a particular `Role` with the `kubectl create rolebinding` command. 

```bash
kubectl create rolebinding alice-ack-s3-writer --role ack-s3-writer --namespace testing --user alice
kubectl create rolebinding alice-ack-sns-reader --role ack-sns-reader --namespace production --user alice
```

You can check the permissions of a particular Kubernetes `User` with the `kubectl auth can-i` command.
```
kubectl auth can-i create buckets --namespace default
```

## Step 2: Configure AWS IAM

After configuring Kubernetes RBAC permissions, you need to create the necessary AWS IAM roles and policies. 

The IAM role needs the correct [IAM policies][aws-iam] for a given ACK service controller. For example, the ACK service controller for AWS S3 needs read and write permission for S3 Buckets. It is recommended that the IAM policy gives only enough access to properly manage the resources needed for a specific AWS service.

To use the recommended IAM policy for a given ACK service controller, refer to the `recommended-policy-arn` file in the `config/iam/` folder within that service's public repository. This document contains the AWS Resource Name (ARN) of the recommended managed policy for a specific service. For example, the [recommended IAM policy ARN for AWS S3][s3-recommended-arn] is: `arn:aws:iam::aws:policy/AmazonS3FullAccess`.

[s3-recommended-arn]: https://github.com/aws-controllers-k8s/s3-controller/tree/main/config/iam

Apply the IAM policy to the IAM role with the `aws iam attach-role-policy` command: 

```bash
SERVICE=s3
BASE_URL=https://github.com/aws-controllers-k8s/$SERVICE-controller/blob/main
POLICY_URL=$BASE_URL/config/iam/recommended-policy-arn
POLICY_ARN="`wget -qO- $POLICY_URL`"
aws iam attach-role-policy \
    --role-name $IAM_ROLE \
    --policy-arn $POLICY_ARN
```

Some services may need an additional inline policy. For example, the service controller may require `iam:PassRole` permission in order to pass an execution role that will be assumed by the AWS service. If applicable, resources for additional recommended policies will be located in the `additional-policy` file within the `config/iam` folder of a given ACK service controller's public repository. You can apply this policy to an IAM role by replacing the `POLICY_URL` variable in the script above. 

If you have not yet created an IAM role, see the user documentation on how to [create an IAM role for your ACK service controller][irsa-docs].

[irsa-docs]: https://aws-controllers-k8s.github.io/community/user-docs/irsa/#create-an-iam-role-for-your-ack-service-controller