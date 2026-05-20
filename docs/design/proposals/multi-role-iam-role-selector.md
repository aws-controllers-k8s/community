# Multi-Role Support via IAM Role Selector

## Background

The AWS Controllers for Kubernetes (ACK) project plays a role in bridging the gap between Kubernetes clusters and various AWS services, offering users a way to manage and integrate their AWS resources within Kubernetes environments.

When managing AWS resources, there are various use-cases which involve interacting with different resources through different IAM roles. These can include creating/managing resources in another account, delegating permissions for a subgroup of resources to an isolated role, etc.
Today, the ACK project provides functionality to achieve this through a mechanism called [Cross Account Resource Management [CARM]](../carm/cross-account-resource-management.md) (current [docs](../../../content/docs/user-docs/cross-account-resource-management.md) and [alpha features](../../../content/docs/user-docs/features.md#teamlevelcarm-and-servicelevelcarm)).

## Problem statement

While CARM supports some multi-role use-cases, its design fundamentally restricts certain use-cases and can generally be iterated and/or replaced with a more complete and extensible mechanism, with lessons learned from CARM's original design and implementation. CARM was originally designed with the use-case of solving cross-account resource management, but the problem has since evolved to general "multi-role" usage, which doesn't fit well with CARM's current design.

Some of the current issues/concerns with CARM include:

1. Configuration of CARM requires using an untyped configmap
    * This makes usage of the feature unnecessarily difficult since there is no schema/feedback about misconfiguration in the cluster
    * This also makes iteration of the feature more difficult without breaking existing usage
1. Untyped annotations are required to configure the role selection.
    * Generally, annotations can be used to inject behavior of other components/controllers into a resource using annotations, however this is less common/recommended for same-service configuration.
    * Because annotations are used for this functionality, it's easy to misconfigure this with no feedback about issues, and no support for tools that read API schemas i.e. code completion tools, etc.
    * Using annotations for configuration of IAM role usage leads to an implicit/non-obvious linking between the permissions to annotate a namespace, and which IAM role will be used, which may not be intuitive for cluster administrators.
1. IAM roles can only be mapped to entire namespaces. Different resources within a namespace cannot use different IAM roles.
1. The configuration semantics can be confusing.
    * An IAM role must be mapped to a special key "account ID", which is a string which doesn't necessarily have to even map to the corresponding role. And then resources have to be mapped with this special string, unnecessarily obfuscating what role is actually being used. The account ID in this annotation may not match the actual account ID of the role.
    * If two different roles from the same account need to be used, these semantics are even more confusing, now requiring the use of a feature-gated config of _another_ separate configmap mapping roles to other arbitrary "team id" strings.

This document aims to propose the design for a new mechanism, called IAM Role Selectors, which will be a re-design of the current CARM implementation following the scope/goals/tenants laid out below.

## Scope/Goals/Tenants

### Tenants

1. Solve the problem from the lens of "multi-role" management, not cross account management.
    * If "multi-role" management is implemented properly, then "cross-account" comes with it implicitly for free.
1. Create an extensible solution that can be easily iterated on [preferably in a non-breaking way] in the future
    * We don't have to solve for every use-case initially, but we should ensure the solution can be iterated easily to solve future use-cases as desired without major re-designs needed.

### Goals

1. Reduce/limit the usage of untyped configuration (i.e. configmaps and annotations)
    * Untyped configuration does not integrate with tooling as well, and gives little room for user feedback when misconfigured.
1. Reduce/limit the requirement of pre-defined hard-coded values (i.e. exact namespace names) in configuration.
    * Specifying a specific namespace name can be allowed, but it shouldn't be the _only_ way to configure. Allowing for something like a selector allows configuration to be much more dynamic.

### Scope/Requirements

1. Provide feature-parity with existing CARM functionality. This will enable a future possible 1:1 migration path for existing CARM users. This includes:
    * The ability to map an IAM role to all resources within a specific existing namespace
    * The ability for a cluster administrator to configure the known IAM roles/where they can be used outside of individual ACK resource config
    * The ability to scope using certain IAM roles to certain ACK services
1. Multi-role assumption/role chaining must work identically for cross-account and same-account usage
    * This simplifies the cross-account resource use-case and generally aligns it with the more generic multi-role strategy

### Out of scope

The following things are out of scope for _this_ design document. Note that these things can (and in the case of some of these, should) be done in the future, but are not part of this initial implementation/design.

* Deprecation strategy for existing CARM
  * This initial design will be implemented behind a feature flag. Depending on adoption/feedback, we may or may not choose to deprecate CARM.
* Migration documentation
  * Similar to above, although we plan to provide feature parity, we do not yet want to bring the deprecation/migration of CARM into the discussion. This work will remain open for future iteration.
* Resource-level IAM role selection
  * While we want the initial design to be extensible to allow differentiating IAM role usage on a resource-per-resource level, we want to ensure our initial implementation gets the fundamentals right. From there, we can gather feedback and decide how to further iterate.

### Personas

We should consider this design in the context of 2 personas (which may overlap depending on the user)

* Cluster Administrator
  * Someone who has permissions to modify/configure a cluster and all its contents
* End User
  * The user directly interacting with ACK (i.e. developer deploying an application with ACK resource(s) to a cluster)

## Design

### Overview

At a high level, the proposed design/solution revolves around a new cluster-scoped configuration CRD which is ultimately a selector for an IAM role.

This new CRD will be called `IAMRoleSelector`, will contain a single IAM role, and will map to different namespaces and/or resources with extended [kubernetes selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).

A cluster administrator will configure the IAM role(s) available for ACK to use, by creating an `IAMRoleSelector` for each role, and providing a corresponding selector for where that IAM role can be used.

### IAMRoleSelector API

The initial proposed API for this new CRD is as follows [using an example]:

```yaml
apiVersion: services.k8s.aws/v1alpha1
kind: IAMRoleSelector
metadata:
  name: SkyDevTeamConfig
  # Note no namespace; this is a cluster-scoped CRD
spec:
  arn: arn:aws:iam::111111111111:role/XAccountS3 # Required
  namespaceSelector: # Optional (matches all namespaces if not defined)
    names:
      - sky-dev-team
    # This label selector is AND'd with the namespace names array and is optional (only one of namespaceNames or labelSelector are required)
    labelSelector:
      # Note this namespaceLabelSelector is a standard label selector - https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/label-selector/
      matchLabels:
        my.company/dev-team: SkyTeam
      # Can also use matchExpressions from standard kubernetes label selector
  resourceTypeSelector: # Optional (matches all resource types if not defined)
    - apiVersion: ec2.services.k8s.aws/v1alpha1
    - apiVersion: s3.services.k8s.aws/v1alpha1
      kind: Bucket # Optional, matches all resources in this service if not provided
```

This provides parity with existing CARM functionality with the `names` array in the `namespaceSelector` (which specific namespace should use this role), and the `apiVersion` in the `resourceTypeSelector` (which controller should use this role).

In addition to this feature parity, this API introduces 2 new concepts:

1. Allow restriction of IAM roles by resource _kind_ in addition to service
1. Allow dynamic namespace selection by standard kubernetes label selector [selecting on the labels of a namespace]

These concepts can all be used in any combination to also provide new functionality (i.e. use a specific IAM role for all S3 buckets across the whole cluster, but not other S3 resources).

Selectors and their subcomponents are always AND'd together to avoid any confusion of accidentally including/selecting anything which was not intended, as accidentally selecting something may have worse security implications than accidentally excluding it; this follows standard conventions for [kubernetes label selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors).

#### Usage

In order to use different IAM roles, different `IAMRoleSelector` objects should be created by the cluster administrator.

Because multiple different IAMRoleSelector's can select the same resource, there can be conflicts. This will be described below.

Write access to this `IAMRoleSelector` kubernetes resource [RBAC] is what controls the ability to configure different IAM roles for ACK, and thus is the security/trust boundary for this "ACK cluster administrator" scoped action.

If an ACK resource is created and has _no_ matching IAMRoleSelector's, then the default role of the ACK controller will be used. This is the same behavior as CARM today.

### Selection Conflicts

An `IAMRoleSelector` maps 1:1 with an IAM role, and 1:N [many] with actual ACK resources.
Because of this, multiple IAMRoleSelector's may end up selecting ("mapping to") the same underlying ACK resource.

In these cases, it's ambiguous which `IAMRoleSelector` should 'win', and thus which IAM role should be used.
It is possible that multiple IAMRoleSelector's have the same IAM role which map to the same ACK resource, however we will treat these as conflicts just the same in order to simplify implementation, avoid edge cases, and clearly establish the `IAMRoleSelector` itself as the logical resource [the IAM role arn is a property of the selector].

When a conflict occurs, ACK should save a status condition onto the resource indicating the conflict, and not perform any further actions on this resource (until the conflict is resolved).

Here is an example status condition indicating a conflict:

```yaml
status:
  ...
  conditions:
  - lastTransitionTime: "2025-09-15T22:09:14Z"
    message: |-
      Cannot determine which IAMRoleSelector to use. Conflicting IAMRoleSelectors: [SkyDevTeamConfig, SkyDevDDBConfig]
    status: "True"
    type: ACK.Recoverable
  ...
```

Internally, any conflicts should be cached, so that when/if an `IAMRoleSelector` is updated/deleted, this conflict cache can be checked, and resources which previously had a conflict on this `IAMRoleSelector` can be immediately re-queued for reconciliation.

### Additional Status Updates

Aside from selection conflict status condition (described above), there are some other resource status API updates to note.

Because this new CRD is configuration that is shared between all ACK controllers, it is not inherently owned by any 1 kubernetes [ACK] controller. Because of this, the CRD will _not_ have any sort of status/other mutating actions performed by ACK. Updating the status (or any other field) of this shared resource may cause various writes to conflict between multiple running ACK controllers, and so this implementation of the CRD will _not_ have any status.

In order to provide feedback about which IAMRoleSelector was used for any individual ACK resource, we will extend the status of each ACK resource with a new field, referencing the IAMRoleSelector which was used. This will include the name of the IAMRoleSelector, as well as the kubernetes resource version of the IAMRoleSelector (to indicate which version of the resource the controller last used, in the case of updates). This is the proposed API:

```yaml
status:
  ackResourceMetadata:
    # Existing fields
    arn: arn:aws:s3:::acktestingbucket
    ownerAccountID: "123456789012"
    region: us-west-2
    # New option field/object - will only be present if an IAMRoleSelector was used for this resource
    iamRoleSelector:
      selectorName: SkyDevTeamConfig
      resourceVersion: "89424719"
  ...
```

### Assume Role Flow

The flow to assume an IAM role/cache credentials will be very similar to the CARM workflow today.

When an `IAMRoleSelector` is matched to an ACK resource, then the corresponding ACK control will perform an `sts:AssumeRole` API action on the IAM role provided by the corresponding `IAMRoleSelector` resource, and use these credentials when operating on the ACK resource.
If there are issues with the configured IAM role (i.e. failure to assume due to permission denied, malformed role ARN, etc.), then this error will be put into the status conditions of the ACK resource (not the `IAMRoleSelector`).

When ACK assumes an IAM role, it will do so with a [refreshing credential provider](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws#CredentialsCache) and save these credentials, so they can be cached and re-used as necessary with subsequent reconcile actions. This will help avoid unnecessary STS `AssumeRole` API calls.

### Resource Selection/Queuing/Updates

ACK will need to maintain a watch/cache over all namespaces to cache their labels in order to perform efficient namespace label selection when reconciling a resource.
This is [already being done by ACK](https://github.com/aws-controllers-k8s/runtime/blob/213056539c9fd6ec7a4d2d9322e3c8aba01db519/pkg/runtime/cache/namespace.go#L86), and will simply need to be modified to store labels in addition to other existing metadata.

Additionally, a watch/cache over all the IAMRoleSelector's will need to be maintained to cache all the selectors.
When reconciling any ACK resource, ACK should first perform a namespace selector and then a resource type selector check over all IAMRoleSelector's to determine if any `IAMRoleSelector` should be used.
If exactly one `IAMRoleSelector` matches, then it will check its credential cache for matching credentials for the corresponding IAM role, or otherwise create and cache new credentials for this role, and then subsequently perform the rest of the resource reconciliation with these new assume role credentials (similar to CARM today).

Additionally, whenever an `IAMRoleSelector` is created/updated, its selector should be immediately evaluated, and any matching ACK resources should be immediately re-queued for reconciliation. This will allow for changes that occur due to `IAMRoleSelector` changes to propagate to AWS resources as quickly as possible.

It is possible that an IAM role for ACK resource(s) changes when an `IAMRoleSelector` is created/updated/deleted. In these cases, the AWS resource(s) should simply be re-evaluated with the latest role, and created/updated/deleted as normal. This may involve re-creating the resource(s) with a different role in a new account/etc.

### Initial Enablement

When first implementing this feature, its usage should be locked behind a feature-gate/feature flag that must be provided when installing ACK.
This should _not_ be enabled by default to allow users to opt into its usage during initial implementation/testing (and before a proper migration plan is in place).

When enabling this feature, it will **DISABLE** CARM. This is done to avoid edge cases and maintain simplicity around conflicts between the functionality overlap of CARM and this new feature.

The ACK helm chart can put the application of this new CRD and the application of this feature flag behind the same shared value for simplicity of enablement when installing.

### Future Possible Ideas

Below lists some ideas for further enhancement of this feature which have been given some initial thought, but are not included in the initial scope for this feature.
These also serve to demonstrate how the above design is extensible, as the additional functionality described here can be added in a non-breaking way in the future.

#### Additional Session Metadata

When performing the STS `AssumeRole` API call for the IAM role provided by an `IAMRoleSelector`, additional metadata can be provided which can then be used by IAM policies to provide additional permission scoping.

For example, an [ExternalId](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole.html#API_AssumeRole_RequestParameters) can be provided, which can be used in IAM roles to scope down who can assume them (particularly useful for cross-account assumption trust).
Additionally, tags can be provided which can then be used in IAM policy conditions (`aws:PrincipalTag/{tagName}`).
These could be used in tandem with resource tags to provide additional dynamic permission scoping for multi-tenant applications, i.e. https://docs.aws.amazon.com/IAM/latest/UserGuide/access_iam-tags.html#access_iam-tags_control-principals

These parameters (external ID, session tags, etc.) could be defaulted with some values, and possibly also overrideable through config (on the `IAMRoleSelector` CRD, env vars/CLI flags, etc.).
These could also have dynamic variables which could be used in value templating similar to the functionality exposed by ACK's custom [resource tagging functionality](../../../content/docs/user-docs/ack-tags.md#configuring-default-tags).

The specifics of these options can be decided/iterated in the future, in a non-breaking way.

#### Advanced Conflict Resolution

Instead of simply throwing an error when two different IAMRoleSelector's conflict, ACK could provide some sort of functionality in how to resolve this conflict and select a particular IAMRoleSelector and/or IAM role.

This could be controlled within the resource namespace or on the ACK resource itself, allowing the end-user to decide to which role to use.
This would effectively allow a cluster administrator to 'allow' multiple different roles to be used by simply defining overlapping selectors, and deferring the actual selection of the particular role to the end-user without requiring a configuration change in the administrative `IAMRoleSelector` object itself.

This could be done through something like a new resource annotation, a new custom ACK-specific field added to all ACK resource CRDs, a new namespaced configuration CRD with a resource label selector, etc.

## Discussion

TBD
