# ACK K8s Native Application Binding

## Problem Statement

Read the original issue:
https://github.com/aws-controllers-k8s/community/issues/740

ACK users often need to access fields from the infrastructure created using K8s
custom resources into their deployment specifications. One example of this would
be to take the [RDS DBCluster
`Status.Endpoint`](https://aws-controllers-k8s.github.io/community/reference/rds/v1alpha1/dbcluster/#status)
and inject it as an environment variable into a web application that requires a
database. Currently, users can only achieve this with two separate steps.
Firstly, the user must create the ACK resource and wait until it reconciles.
Then they must manually read the status from the resource and inject it into the
manifest of their deployment. Given this process requires a manual step, and
specific order of operations, it is not compatible with GitOps systems to create
infrastructure in the same manifests as an application.

This design aims to address two separate use cases:

1. An application team manages their own ACK resources and K8s deployments. They
   wish to dynamically link the status of an ACK resource to the environment
   variables of their deployment.
2. An infrastructure team manages the ACK resources on behalf of application
   teams. The application teams do not have read/write access to the ACK
   resource directly, but require access to the fields from them in order to
   deploy their application.

## Requirements

1. Export any number of Spec and/or Status values to application consumable
   formats (ConfigMap and Secrets)
2. Export to any namespace (for cluster-scoped controllers only)
3. The exported value should stay in sync as the resource changes
4. Resources should not need to be updated to add/remove bindings

## Existing Solutions

### Redhat Service Binding Operator

>“[Redhat Service Binding Operator] enables developers to connect their
>application to backing services with a consistent and predictable experience“ -
>https://github.com/redhat-developer/service-binding-operator


The [Redhat Service Binding
Operator](https://github.com/redhat-developer/service-binding-operator) (herein
called the SBO) uses annotations present on CRs and resources to import field
values from other resources. The main purpose of SBO is to standardise the
system for injecting secret information from database services into application
deployment environment variables. The SBO can automatically detect and bind to
services only from a limited number of [supported
operators](https://github.com/redhat-developer/service-binding-operator#known-bindable-operators). 

SBO is not suitable for this design as it can only support a limited number of
fields based on pre-configured database bindings. That is, it can handle the
case where an application developer wants to inject a PostgreSQL database
connection string into a deployment, but little more than that. SBO does not
support exporting to configmap or secrets directly.

## Proposed Solution

### ACK Binding CRD

ACK uses a common CRD to support resource adoption, implemented through a
separate (common) reconciler loop in each controller. Common CRDs allow
customers to define controller configuration for tasks that do not fit within
the spec of existing CRDs or that are too complex to fit into the annotations. 

This solution would introduce a new CRD solely for the purpose of declaring
field exports. The CRD spec would contain an array of structs, each containing:

* The source path, containing:
    * Resource API group, kind, namespace, name
* Spec/Status field path
* An output paths, containing:
    * Type: ConfigMap or Secret
    * Namespace and name
    * (Optionally) Output structure

The common field exporter reconciler running within the controller would have a
shared informer for all custom resources of this type. The reconciler would read
the spec for each binding resource, filter for any resources where the source
path matches its respective resource API group, version and kind and then read
and export the corresponding fields. 

Having this configuration as a separate CRD allows application developers to
specify their own configuration independent of the ops teams that manage the
infrastructure. Cluster operators can lock down application teams’ RBAC
permissions to only have access to read ACK resources, with an exception only to
write this new binding CRD.

As an example:

```yaml
apiVersion: services.k8s.aws/v1alpha1
kind: FieldExport
metadata:
  name: export-vpc-id
spec:
  to:
    type: ConfigMap
    namespace: my-other-namespace
    name: my-exported-configmap
  from:
    resource:
      group: ec2.services.k8s.aws
      kind: VPC
      namespace: default
      name: my-vpc
    path: ".status.vpcID"
```

## Alternate Solutions

### CR Annotations

Annotations are a generally accepted and understood way of configuring details,
particularly metadata, about resources in K8s. ACK already uses annotations
placed onto custom resources as definitions for CARM configuration. Annotations
are good for 1-to-1 bindings of information onto a resource.

The ACK controller could detect “export specifier annotations” on the resource
to signal that a given field should be exported into a given resource type. This
solution would require multiple annotations as each annotation can only hold a
single value (unless encoded as JSON, which is not user-friendly).

The annotations would need to specify, for each exported field:

* Which field should be exported
* The resource type to export (ConfigMap or Secret)
* The namespaced-name of the exported resource
* (Optionally) The structure of the exported resource

As an example:

```yaml
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: VPC
metadata:
  name: my-exported-vpc
  annotations:
    services.k8s.aws/export-field-path: ".status.vpcID"
    services.k8s.aws/export-field-type: "v1/ConfigMap"
    services.k8s.aws/export-field-name: "my-other-namespace/my-exported-configmap"
```


This solution would **not scale** for multiple exported fields. Each exported
field would need its own set of annotations, very quickly exploding the number
of annotations on the resource. Since annotations’ keys cannot collide, each
would need its own distinct set of keys as well. 

Requiring annotations also directly ties bindings of output from a custom
resource directly onto the resource definition. That is, if a developer wants to
make changes to the output bindings for a resource, they must have the
permissions to edit the resource directly. Not only does this mean that write
permissions for objects will need to be expanded to application teams, but it
also increases the odds that someone introduces changes to the spec
accidentally.

### ConfigMap/Secret Annotations

Similarly to [CR Annotations](#cr-annotations), ConfigMap annotations would
introduce a set of “export specifier annotations” that could be applied onto the
ConfigMap and Secret resources. The ACK controller would have an informer for
all ConfigMap and Secrets within the cluster and recognise when these
annotations have been attached.

This approach introduces the same set of challenges as the [CR
Annotations](#cr-annotations) examples - mainly that it would not scale for
multiple fields. However, this approach would decouple the binding from the ACK
resource, as the binding now lives on a resource type that application
developers could have uncontrolled access to. 

### CR MapToSecret Fields

Crossplane has decided to tackle this problem by introducing a new field
([`writeConnectionSecretToRef`](https://doc.crds.dev/github.com/crossplane/provider-aws/eks.aws.crossplane.io/Cluster/v1beta1@v0.23.0))
into the spec of each of their custom resources. After the user specifies a
namespaced name of a secret, the controller knows to write the “connection
secret” of the resource once it has been created. This field only exists on
resources that have a “connection secret”, which may be a connection string or
URL, such as on a database or a cluster. 

As per the requirements for this design, ACK needs to support exporting of any
spec or status field. Therefore, we would need to add to the spec of every CRD
either:

1. A new field, for every spec AND status field, of type namespaced name, OR
2. A new field that accepts a list of field paths and corresponding namespaced
   names

Option 1 would double the number of spec fields and add another field for every
status field. This would be far too many optional fields on the resource and
definitely lead to customer confusion. Therefore it should absolutely be ruled
out.

Option 2 would only add a single new spec field for every CRD. This spec field
would essentially encapsulate the information offered in the [CR
Annotations](#cr-annotations) example, however placing it into a strongly-typed
struct in the spec rather than in the weakly-typed annotation. Just like that
solution, though, it would also tie the binding of output from a custom resource
directly to the custom resource.

As an example of option 2:

```yaml
apiVersion: ec2.services.k8s.aws/v1alpha1
kind: VPC
metadata:
  name: my-exported-vpc
spec:
  cidrBlock: 10.0.0.0/16
  exportedFields:
  - path: ".status.vpcID"
    destinations:
    - type: "ConfigMap"
      namespace: "my-other-namespace"
      name: "my-exported-configmap"
```

## FAQ

**Which namespace do you need to create the `FieldExport` CRs?** For the initial
implementation, `FieldExport` CRs will need to be created in the same namespace
as the ACK resource they reference. The destination `ConfigMap` or `Secret` may
live in a different namespace (given the controller has appropriate RBAC
privileges).

Referencing a different namespace poses a security threat that application
developers could gain access to ACK resources they do not have RBAC privileges.
For example, if a secure RDS database lived in an isolated namespace, a
malicious actor could create a `FieldExport` CR to reference the connection
secrets of that database to output into a namespace in which they do have access
- giving them access to those secrets.

**How does the functionality work for namespace-scoped controllers?** Just as in
the cluster-scoped controller, namespace-scoped controllers will only be able to
export fields from ACK resources in the same namespace as the `FieldExport` CRs.
However, unlike cluster-scoped controllers, they are not able to export to
`ConfigMap` or `Secret` into different namespaces. This means all operations
will be confined into the namespace specified under the `--watch-namespace`
flag, and RBAC permissions will be adjusted to fit this accordingly.
