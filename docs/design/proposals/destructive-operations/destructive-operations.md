# Semantics of destructive operations

## Summary

This design document stems from the initial discussion on the [original Github issue #82](https://github.com/aws-controllers-k8s/community/issues/82).

When deleting an ACK custom resource, the underlying AWS resource is first deleted. Users should expect this to be the default behaviour, as creating and updating CR similarly have the same respective actions in AWS. However users have requested that some resources should continue to live on after the ACK CR has been deleted. Resources with stateful data, such as S3 buckets, RDS instance or EBS volumes, may want to be retained so that they may be migrated between ACK installations or off ACK entirely without requiring they be backed up, deleted and then restored.

Similar to the CDK[`RemovalPolicy` type](https://docs.aws.amazon.com/cdk/api/v2//docs/aws-cdk-lib.RemovalPolicy.html#members) and the native [K8s PV reclaim policy](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#reclaiming), ACK should support the ability to leave the AWS resource alive while still deleting the custom resource object. 

## AWS Removal Policy

### Requirements

1. Multiple levels of granularity over resource protection (controller-wide, namespace-wide and resource-wide)
2. Protection should be defined per-service (if namespace-wide)
3. Users are able to define either of:
    1. The AWS resource should be deleted along with the CR
    2. The AWS resource is left in-tact after the CR is deleted

### Investigation

A number of K8s resources already have similar functionality:

* Helm uses an annotation: `helm.sh/resource-policy: keep`
* Google Config Connector uses an annotation: `cnrm.cloud.google.com/deletion-policy: abandon`
* Skaffold uses a label: `skaffold.dev/cleanup: "false"`
* K8s PV use a spec field: `persistentVolumeReclaimPolicy``: Retain`


Existing AWS products have their own variants:

* AWS CDK uses a spec field: `removalPolicy: RemovalPolicy.RETAIN`
* CloudFormation uses a metadata field on the resource: `DeletionPolicy: Retain`

CloudFormation and CDK both provide the ability to snapshot given resources before deleting them (mostly databases). While this is outside of the scope for these requirements, it’s not outside the scope of ACK in the future.

## Proposed Option

The following proposed option will introduce:

* Two new annotations, one service-wide and one ACK-wide, to be applied on namespace and ACK resources
* One new optional command line argument for the controller binary

(Requirements 1) Following the current CARM design, which also requires the same levels of granularity, it is consistent to mark the policies in the annotation metadata for the namespace and resources, and to create a new command line argument on the controller for the controller-wide policies. The controller would place precedence on the resource policy, then the namespace policy and finally the controller policy - the annotation on the object with the highest precedence will set the policy for the given resource.

(Requirements 2) The full annotation path needs to consist of an annotation prefix (typically a URL) before the annotation name. One variation should concern every ACK controller regardless of service, and one should discriminate for controllers only concerned with a single service. Cross-service annotations, finalizers and CRDS are already configured with the `services.k8s.aws/` prefix. For this option, prepending the service identifier to the beginning would identify that the annotation applies only to the given service, eg. `s3.services.k8s.aws/` . Hence, there would be two prefixes:

* `services.k8s.aws/`: All ACK controllers will read the policy of this namespace/resource
* `<service>.services.k8s.aws/`: Only the <service> ACK controller will read the policy of this namespace (taking precedence over the former option)


The name of this annotation could be either:

* Follow the AWS variations (`RemovalPolicy` or `DeletionPolicy`) which would put it inline with expectations of current CDK and CloudFormation users
* Follow existing K8s variations (`ResourcePolicy` or `ReclaimPolicy`) which would put it inline with expectations of current K8s users

Users familiar with K8s would find the existing K8s variations more recognizable but may incorrectly assume that the annotation has the same values and behaves identically to native K8s resources. Therefore, it may be better to introduce the AWS terminology into the annotation - as the proposed functionality will behave identically to CDK and CloudFormation. Considering this annotation takes effect after the user calls “Delete”, I propose using `DeletionPolicy` as the annotation name.

(Requirement 3) In order to account for a future possibility of adding additional policy options, the value of this annotation should be an enumeration. Considering the name already references the existing CloudFormation `DeletionPolicy` field, the values should similarly follow those values:

* `Delete` - deletes the AWS resource alongside the CR
* `Retain` - retains the AWS resource, but deletes the CR

The default value for this annotation will be `Delete` as it is in-line with the existing ACK functionality of deleting the underlying AWS resource.

#### Examples

Definition at the controller - all resources reconciled by this controller will use the “retain” policy (unless overridden by an annotation):

```yaml
containers:
    - command:
    - ./bin/controller
    args:
    - --aws-account-id
    - "$(AWS_ACCOUNT_ID)"
    - --aws-region
    - "$(AWS_REGION)"
    - --enable-development-logging
    - "$(ACK_ENABLE_DEVELOPMENT_LOGGING)"
    - --log-level
    - "$(ACK_LOG_LEVEL)"
    - --resource-tags
    - "$(ACK_RESOURCE_TAGS)"
+  - --deletion-policy
+  - "retain"
```

Definition at the namespace (protecting all ACK resources across any service):

```yaml
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    services.k8s.aws/deletion-policy: retain
  name: my-retained-namespace
```

Definition at the namespace (protecting only S3 ACK resources):

```yaml
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    s3.services.k8s.aws/deletion-policy: retain
  name: my-retained-namespace
```

Definition at the resource:

```yaml
apiVersion: s3.services.k8s.aws/v1alpha1
kind: Bucket
metadata:
  name: test-bucket
  annotations:
    s3.services.k8s.aws/deletion-policy: retain
spec:
  name: retained-bucket
```

## Backwards Compatibility and Risks

The proposed solution introduces two new annotations and an optional command line argument. The default behaviour for any resource without these annotations, or for the controller binary without this optional command, is backwards compatible with all current ACK controllers. There are no backwards compatibility issues with this solution.

TODO: Identify possible risks with this proposal

