# Configure permissions for authorization and access

There are two different Role-Based Access Control (RBAC) systems needed for ACK service controller authorization: Kubernetes RBAC and AWS IAM. 

[Kubernetes RBAC][0] governs a Kubernetes user's ability to read or write Kubernetes resources, while [AWS Identity and Access Management][2] (IAM) policies govern the ability of an AWS IAM role to read or write AWS resources.

[0]: https://kubernetes.io/docs/reference/access-authn-authz/authorization/

!!! note "These two RBAC systems do not overlap"
    The Kubernetes user that makes a Kubernetes API call with `kubectl` has no association with an IAM role. Instead, the IAM role is associated with the [service account][1] that runs the ACK service controller's pod.

[1]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/

Refer to the following diagram for more details on running a Kubernetes API server with RBAC authorization mode enabled.

![Authorization in ACK](../images/authorization.png)

[2]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/

You will need to configure Kubernetes RBAC and AWS IAM permissions before using ACK service controllers.

## Step 1: Configure Kubernetes RBAC

As part of installation, Kubernetes roles are automatically created. These roles contain permissions to modify the Kubernetes custom resources (CRs) that the ACK service controller is responsible for.

!!! note "Resources are namespace-scoped"
    All Kubernetes CRs managed by an ACK service controller are namespace-scoped resources. There are no cluster-scoped ACK-managed CRs.

By default, the following Kubernetes role resources are created when installing an ACK service controller:

* `ack-$SERVICE-writer`: a `Role` used for reading and mutating namespace-scoped CRs that the ACK service controller manages.
* `ack-$SERVICE-reader`: a `Role` used for reading namespaced-scoped CRs that the ACK service controller manages.

If you already have an ACK service controller installed, the Kubernetes role might already exist. In this case, permissions to manage custom resource definitions (CRDs) and CRs associated with a newly-installed ACK service controller are added to the existing role.

To rename a default Kubernetes role, modify the Kubernetes manifests that are used as part of the installation process. 

### Bind a Kubernetes user to a Kubernetes role

Once the Kubernetes roles have been created, you can assign a specific Kubernetes user to a particular role with the `kubectl create rolebinding` command. 

```bash
kubectl create rolebinding alice-ack-s3-writer --role ack-s3-writer --namespace testing --user alice
kubectl create rolebinding alice-ack-sns--reader --role ack-sns-reader --namespace production --user alice
```

You can check the permissions of a particular Kubernetes user with the `kubectl auth can-i` command.
```
kubectl auth can-i create buckets --namespace default
```

## Step 2: Configure AWS IAM

After configuring Kubernetes RBAC permissions, you need to create the neccessary AWS IAM roles and policies. 

The IAM role needs the correct [IAM policies][2] for a given ACK service controller. For example, the ACK service controller for AWS S3 needs read and write permission for S3 Buckets. It is recommended that the IAM policy gives only enough access to properly manage the resources needed for a specific AWS service.

Apply the IAM policy to the IAM role with the `aws iam attach-role-policy` command: 

```bash
aws iam attach-role-policy \
    --role-name $IAM_ROLE \
    --policy-arn $POLICY_ARN
```

If you haven't yet created an IAM role, see the user documentation on how to [create an IAM role for your ACK service controller][irsa-docs].

[2]: https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies.html
[irsa-docs]: https://aws-controllers-k8s.github.io/community/user-docs/irsa/#create-an-iam-role-for-your-ack-service-controller