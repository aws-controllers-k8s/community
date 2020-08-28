# Cross account resource management (CARM)

## The problem

> Alice is a software developer working on her company's web application. Her application is deployed to Kubernetes as a set of Pods. That application code relies on a relational database server that is provided by an Amazon RDS database instance. Unfortunately, for billing purposes, the central IT team wants this RDS database instance to be owned by a separate AWS account.
>
>Before ACK was capable of cross-account resource management, Alice would need to coordinate with the central IT team to create separate Kubernetes clusters, one tied to the AWS account her team usually used and another tied to the AWS account that RDS database resources were owned by. ACK service controllers would need to be installed into each of these separate Kubernetes clusters. Alice's team needed to have separate code repositories containing Kubernetes manifests -- one repository for the ACK custom resources owned by her team's AWS account, and another repository for the RDS database instance CRs owned by the other AWS account.
>
> With cross-account resource management, Alice only needs to have a single Kubernetes cluster and a single code repository to store Kubernetes configuration files for her ACK custom resources. This greatly simplifies Alice's life and reduces the amount of time she needs to coordinate with the central IT team.

Today administrators can manage AWS resources using ACK. However, these resources are created in the same AWS account as the Kubernetes cluster. This leads to a suboptimal experience for entreprise customers: creating resources in multiple AWS accounts requires ACK to be installed and managed in many kubernetes clusters. Considering ACK has dozens of service controllers, this means administrators must manage hundreds of AWS service controller pods across many kubernetes clusters and many AWS Accounts with scoped roles.

With Cross Account Resource Management (CARM) feature, customer(s) can install ACK on a single Kubernetes cluster which provides ability to Create/Read/Update/Delete resources in different AWS Account(s). It also allows administrators to manage permissions to access the services across accounts efficiently.

## Proposed implementation

#### Overview

The proposed solution will solve the problem of cross account management within a single kubernetes cluster. Given an AWS Owner account ID it will look up an IAM Role ARN and call **STS::AssumeRole** to pivot the AWS SDK calls in order to execute in a new AWS Account and role context.

Controllers will store AccountIDs and their associated IAM Role ARNs in a `ConfigMap` Kubernetes object. Cluster administrators will be responsible for administrating and updating the `ConfigMap`, and creating the target roles using the AWS Console/CLI. Controllers can reuse and existing IAM Role ARN if it matches the scope they need.

To prevent unauthorized users from making any CRUD operations on resources in different AWS accounts, cluster administrators will be able to set an AWS Owner Account ID as an annotation on each namespace, which will be used by ACK Controllers to override the AWS Account ID that is used to create the resources in that namespace. This will help administrators limit access to namespaces using K8s RBAC model, hence limit access to AccountIDs.

#### Design details

##### Namespace annotations

Cluster admins will be able to specify an AWS account ID in the namespace annotations in order to override any CRs account IDs created in that namespace e.g.:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: production-n1
  annotations:
    services.k8s.aws/owner-account-id: "123456789012"
```

Each time a new AWS resource object is updated, ACK service controllers will lookup the namespaces annotations and decide in which AWS account the object should be created.

In addition to the owner account id annotation, users can also specify a `services.k8s.aws/default-region` annotation. This region will be used by default in case region annotation is missing in the CRDs specs.

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: production-n1
  annotations:
    services.k8s.aws/default-region: "eu-west-1"
```

*//TODO(a-hilaly) address namespace/account-id conflict cases*

*//TODO(a-hilaly) address namespace deletion case*

##### Region determination

To determine within which region the resources should be created, ACK controllers will, in order, look for a region in the following sources: 

- CR region annotation `services.k8s.aws/region`. If provided it will override the namespace default region annotation.
- Namespace default region annotation `services.k8s.aws/default-region`
- If none of the two annotations are provided ACK will try to find a region from these sources: (see [cli user guide](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html#cli-configure-quickstart-precedence)): Controller flags, Pod IRSA environment variables...

##### Storing AWS Role ARNs

AccountIDs and their associate AWS Role ARNs will be stored in a `ConfigMap` called "ack-role-account-map". e.g:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ack-role-account-map
  namespace: default
data:
  accounts:
    "123456789012": arn:aws:iam::123456789012:root
    "454545454545": arn:aws:iam::454545454545:role/S3Access
```

In ACK runtime, the reconciler will need to lookup the `ack-role-account-map` data content to query the value of a particular key. To realize that, the controllers will keep a cached version of the ConfigMap and frequently update it when changes are made by cluster admins.

To do that we will use `SharedInformers` from the [k8s.io/client-go](https://godoc.org/k8s.io/client-go) `informer` package to watch for the `ack-role-account-map` ConfigMap updates then query for the newest version of the data. The same informer will help ACK service controllers query the `ack-role-account-map` data during startup.

##### IAM Account pivot

When a matching namespace is found, ACK service controllers will search for the IAM Role ARN associated with the AccountID in the `ack-role-account-map` ConfigMap, and pivot the session to create resources in a new AWS account.

The pivot is done by using the [aws-sdk-go](https://github.com/aws/aws-sdk-go) `session` package. `session.NewSession()` function will load the AWS related environment variables (injected by Pod IRSA) and initialize the [token provider](https://github.com/aws/aws-sdk-go/blob/master/aws/session/session.go#L200-L218) function. Later the token provider function is executed when [`AssumeRoleProvider.Retrieve`](https://github.com/aws/aws-sdk-go/blob/master/aws/credentials/stscreds/assume_role_provider.go#L329) is called and [return a set of credentials](https://github.com/aws/aws-sdk-go/blob/master/aws/credentials/stscreds/assume_role_provider.go#L357-L362) to make pivoted api calls.

```go
    // getting the role ARN to assume
    roleARN := getRoleARN()
    
    // NewSession will use POD IRSA credentials to make 
    // the STS Assume Role API.
    sess, err := session.NewSession()
    if err != nil {
        return ...
    }

    // create the credentials from AssumeRoleProvider to assume the role
    creds := stscreds.NewCredentials(sess, roleArn)

    // we can use the creds to create a new client with a pivoted session to either

    // create service client value configured for credentials
    svc := s3.New(sess, &aws.Config{Credentials: creds})

    // or store the session object in the resourceManager ...
```

##### sequence diagram

![sequence-diagram](./images/carm-sequence-diagram.png)

##### AWS Pod IRSA

ACK Service controllers will be deployed using [AWS Pod IRSA (IAM Roles for service accounts)](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html), which will associate a ServiceAccount to the controller pods and result in limiting their permissions when calling the AWS API.
Pod IRSA will limit ACK controllers API calls to the AWS service resources they are responsible for managing and provide audibility throught AWS CloudTrail.
The injected IAM Role will be used to pivot the execution with a different target IAM Role using the STS::AssumeRole call.
In addition to the injected AWS Credentials, Pod IRSA will also add `AWS_REGION` and `AWS_DEFAULT_REGION` to the controllers environment variables. When creating new resources and in case the default region isn't provided via annotations, the controllers will use these environment variables to set a default region for the AWS resources.

##### Auditability (Service controllers side)

ACK service controllers reconcile loops will log every session account pivot using zap logger values e.g:

```golang
func (r *reconciler) reconcile(req ctrlrt.Request) error {
    ...

    acctID := r.getOwnerAccountID(res)
    region := r.getRegion(res)
    accountRoleARN := getAccountIDRoleARN(...)

    r.log.WithValues(
        "account_id", acctID,
        "region", region,
        "account_arn", accountRoleARN,
        ...
    )

    ...
}
```

## Potential other solutions

- Store the IAM Role ARNs and account IDs as a kubernetes objects (new CRDs)
- Hard coded configuration loaded at the start time (need to restart the controller to update the configuration)
- Specify the role using an annotation (painful user experience, security problems)

## In scope

- Ability to have the service controller assume the same IAM role in a different account.
- Tracebility and audibility of IAM Role related events (On the controller side), by adding/updating zap logger values (see this [example](https://github.com/aws/aws-controllers-k8s/blob/mvp/pkg/runtime/reconciler.go#L84-L87))
- Unit and e2e tests with a specific ACK Controller (e.g. s3 bucket controller)
- User has ability to specify the AWS account in CR Annotations

Example of AWS Account specification:

```yaml
apiVersion: s3.services.k8s.aws
kind: Bucket
metadata:
  name: example-s3-bucket
  annotations:
    services.k8s.aws/aws-region: "eu-west-2"
    services.k8s.aws/owner-account-id: "120987654321"
spec:
  ...
```

## Out of scope

- Any service controller specific code (e.g generated code)
- CARM e2e tests for all controllers
- Determination of owner account ID (Owner Account ID will be already determined and passed to the [NewSession](aws-service-operator-k8s/pkg/runtime/session.go) function)

## Test plan

##### Unit tests

We'll need a AWS client mock and K8s client mock and then:

- Assert that CARM logic is working as expected
- Assert every important event is logged.

##### e2e tests

We will need:
- 1 Kubernetes cluster
- 2 AWS Accounts (A && B)
- 1 ACK Controller e.g S3

Testing will be similar to unit testing:
- Try to create resources using account B
- Verify that the controller called AWS Sts:AssumeRole
- Verify that the resources were created using account B
- Ensure the events are properly logged

## Discussions

Questions quoted from [Justin G. comment](https://github.com/aws/aws-controllers-k8s/pull/62#issuecomment-655670245)

> Is there an option to automatically create namespaces for each account specified

We're not thinking of adding this feature, at least in early stages. As Justin said, it might be dangerous.

> Is there also an idea of (account, team) separation without using multiple clusters?

For the moment, no.

> In a centralized ACK cluster I would see the need for developers to switch clusters frequently (ACK cluster for AWS resources, K8s cluster for each environment for pods). Do we see a way ACK can make that easier at all?

It's a little hard to imagine a solution at this stage. We might tackle this issue in the future.