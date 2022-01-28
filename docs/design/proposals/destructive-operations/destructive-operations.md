# Semantics of destructive operations

## Summary

This design document stems from the initial discussion on the [original Github issue #82](https://github.com/aws-controllers-k8s/community/issues/82).

ACK currently treats all resources with the same amount of protection for deletion. That is, if you delete an ACK custom resource, the underlying AWS resource is deleted as well. While this might be the expected behaviour for the vast majority of resources, it also increases the chance that an AWS resource is accidentally deleted - through prematurely or accidentally deleting the ACK CR. Some sort of deletion protection should be introduced to limit the odds of this happening, requiring the user to unset a property manually before the controller will continue with deletion.

Another consideration of destructive operations is that some resources should continue to live on after the ACK CR has been deleted. Similar to the CDK[`RemovalPolicy` type](https://docs.aws.amazon.com/cdk/api/v2//docs/aws-cdk-lib.RemovalPolicy.html#members) and the native [K8s PV reclaim policy](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#reclaiming), ACK should support the ability to leave the AWS resource alive while still deleting the custom resource object. Users typically use these removal policy options when working with stateful data, such as S3, EBS and RDS resources, in cases where they would like to migrate the management of the resources from one system to another without having to recreate the data.

This design document proposes different mechanisms to address each of the following concerns:

* Protection against deletion
* Control over AWS removal policy

## Delete Protection

### Requirements

1. Multiple levels of granularity over resource protection (namespace-wide and resource-wide)
2. Protection should be defined per-service (if namespace-wide)
3. Controller must strictly check for existence of resource protection before any deletion action
4. The controller must not edit any of the resource protection marks

### Investigation

Kubernetes finalizers are the standard way to protect a resource from deletion. A finalizer placed on a resource prohibits it from being deleted out of etcd until it is removed - either through code or by manually editing the resource. ACK already places finalizers on every resource, however these are used to guard the custom resource object being deleted before the the underlying AWS resource. Finalizers, however, do not apply in a cascading effect to resources below them ie. a finalizer placed on a namespace will not protect the deletion of a custom resource within it. This violates requirement 1, as there is only resource-wide granularity over resource protection. 

Finalizers do not fit with requirement 3. If the user places a finalizer on an ACK resource, then issues a deletion of that resource, the ACK controller will immediately delete the AWS resource and then remove its own finalizer, regardless of any other finalizer on the object. The current ACK controller does not look for other finalizers on the object before running its delete logic. Only the K8s control plane uses finalizers to protect it from deletion in the key-value store, it does not limit the interaction of custom controllers.

There is on ongoing issue regarding similar functionality in the [Kubernetes Github repository](https://github.com/kubernetes/kubernetes/issues/10179).
There is a feature request for the same function in the [Azure service operator repository](https://github.com/Azure/azure-service-operator/issues/1633). 

## AWS Removal Policy

### Requirements

5. Multiple levels of granularity over resource protection (controller-wide, namespace-wide and resource-wide)
6. Protection should be defined per-service (if namespace-wide)
7. Users are able to define either of:
    - The AWS resource should be deleted along with the CR
    - The AWS resource is left in-tact after the CR is deleted

### Investigation

A number of K8s resources already have similar functionality:

* Helm uses an annotation: `helm.sh/resource-policy: keep`
* Google Config Connector uses an annotation: `cnrm.cloud.google.com/deletion-policy: abandon`
* Skaffold uses a label: `skaffold.dev/cleanup: "false"`
* K8s PV use a spec field: `persistentVolumeReclaimPolicy`: Retain`


Existing AWS products have their own variants:

* AWS CDK uses a spec field: `removalPolicy: RemovalPolicy.RETAIN`
* CloudFormation uses a metadata field on the resource: `DeletionPolicy: Retain`

CloudFormation and CDK both provide the ability to snapshot given resources before deleting them (mostly databases). While this is outside of the scope for these requirements, it’s not outside the scope of ACK in the future.

## Proposed Option

The following proposed options will introduce:

* Two new annotations (each with a service-wide and an ACK-wide variation) for namespace and resource metadata fields
* One new optional command line argument for the controller binary

(Requirements 1 & 5) Following the current CARM design, which also requires the same levels of granularity, it is consistent to mark the policies in the annotation metadata for the namespace and resources, and to create a new command line argument on the controller for the controller-wide policies. The controller would place precedence on the resource policy, then the namespace policy and finally the controller policy - the annotation on the object with the highest precedence will set the policy for the given resource.

(Requirements 2 & 6) The full annotation path needs to consist of an annotation prefix (typically a URL) before the annotation name. One variation should concern every ACK controller regardless of service, and one should discriminate for controllers only concerned with a single service. Cross-service annotations, finalizers and CRDS are already configured with the `services.k8s.aws/` prefix. For this option, prepending the service identifier to the beginning would identify that the annotation applies only to the given service, eg. `s3.services.k8s.aws/` . Hence, there would be two prefixes:

* `services.k8s.aws/`: All ACK controllers will read the policy of this namespace/resource
* `<service>.services.k8s.aws/`: Only the <service> ACK controller will read the policy of this namespace (taking precedence over the former option)

### AWS Removal Policy

The name of this annotation could be either:

* Follow the AWS variations (`RemovalPolicy` or `DeletionPolicy`) which would put it inline with expectations of current CDK and CloudFormation users
* Follow existing K8s variations (`ResourcePolicy` or `ReclaimPolicy`) which would put it inline with expectations of current K8s users

Users familiar with K8s would find the existing K8s variations more recognizable but may incorrectly assume that the annotation has the same values and behaves identically to native K8s resources. Therefore, it may be better to introduce the AWS terminology into the annotation - as the proposed functionality will behave identically to CDK and CloudFormation. Considering this annotation takes effect after the user calls “Delete”, I propose using `DeletionPolicy` as the annotation name.

(Requirement 7) In order to account for a future possibility of adding additional policy options, the value of this annotation should be an enumeration. Considering the name already references the existing CloudFormation `DeletionPolicy` field, the values should similarly follow those values:

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

### Delete Protection

Proposed annotation names:

* `no-delete`
* `block-delete` / `prevent-delete` / `lock-delete`
* `delete-lock`

Acceptable values for this annotation will only need to be boolean `true/false` . The annotation set to `true` will indicate to the controller that it, or any ACK resources within its scope, should not be deleted. The default value for this annotation (when the annotation is omitted) will be `false` - the resource can be deleted.

When a user attempts to delete a resource, a validating webhook will read through each of the annotations on the resource and the namespace. If the premature deletion annotation is defined on either of those, if the one with the higher precedence has the value set to `true`, the validating webhook will reject the delete request with a corresponding error message. 

#### Examples

Definition at the namespace (protecting all ACK resources across any service):

```yaml
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    services.k8s.aws/prevent-delete: true
  name: my-safe-namespace
```

Definition at the namespace (protecting only S3 ACK resources):

```yaml
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    s3.services.k8s.aws/prevent-delete: true
  name: my-safe-for-s3-namespace
```

Definition at the resource:

```yaml
apiVersion: s3.services.k8s.aws/v1alpha1
kind: Bucket
metadata:
  name: test-bucket
  annotations:
    s3.services.k8s.aws/prevent-delete: true
spec:
  name: safe-bucket
```

## Backwards Compatibility and Risks

The proposed solution introduces two new annotations and an optional command line argument. The default behaviour for any resource without these annotations, or for the controller binary without this optional command, is backwards compatible with all current ACK controllers. There are no backwards compatibility issues with this solution.

TODO: Identify possible risks with this proposal