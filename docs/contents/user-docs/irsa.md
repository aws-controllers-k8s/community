# Setting up ACK with IAM Roles for Service Accounts

[IAM Roles for Service Accounts][0], or IRSA, is a system that automates the
provisioning and rotation of IAM temporary credentials (called a Web Identity)
that a Kubernetes `ServiceAccount` can use to call AWS APIs.

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
this special token file and call the [`STS::AssumeRoleWithWebIdentity`][2] API
to assume the IAM Role with reduced permissions.

!!! note "EKS is not required to use IRSA"

    Note that you do *not* need to be using the Amazon EKS service in order to
    use IRSA. There are [instructions][1] on the
    amazon-eks-pod-identity-webhook repository for setting up IRSA on your own
    Kubernetes installation.

## 

[0]: https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
[1]: https://github.com/aws/amazon-eks-pod-identity-webhook/blob/master/SELF_HOSTED_SETUP.md
[2]: https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRoleWithWebIdentity.html
