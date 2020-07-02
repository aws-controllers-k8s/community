# Configure permissions

Because ACK bridges the Kubernetes and AWS APIs, before using ACK service
controllers, you will need to do some initial configuration around Kubernetes
and AWS Identity and Access Management (IAM) permissions.

## Configuring Kubernetes RBAC

As part of installation, certain Kubernetes `Role` objects will be created that
contain permissions to modify the Kubernetes custom resource (CR) objects that
the ACK service controller is responsible for.

**NOTE**: All Kubernetes CR objects managed by an ACK service controller are
Namespaced objects; that is, there are no cluster-scoped ACK-managed CRs.

By default, the following Kubernetes `Role` objects are created when installing
an ACK service controller:

* `ack.user`: a `Role` used for reading and mutating namespace-scoped custom
  resource (CR) objects that the service controller manages.
* `ack.reader`: a `Role` used for reading namespaced-scoped custom resource
  (CR) objects that the service controller manages.

When installing a service controller, if the `Role` already exists (because an
ACK controller for a different AWS service has previously been installed),
permissions to manage CRD and CR objects associated with the installed
controller's AWS service are added to the existing `Role`.

For example, if you installed the ACK service controller for AWS S3, during
that installation process, the `ack.user` `Role` would have been granted
read/write permissions to create CRs with a GroupKind of
`s3.services.k8s.aws/Bucket` within a specific Kubernetes `Namespace`.
Likewise the `ack.reader` `Role` would be been granted read permissions to view
CRs with a GroupKind of `s3.services.k8s.aws`.

If you later installed the ACK service controller for AWS SNS, the installation
process would have added permissions to the `ack.user` `Role` to read/write CR
objects of GroupKind `sns.services.k8s.aws/Topic` and added permissions to the
`ack.user` `Role` to read CR objects of GroupKind `sns.services.k8s.aws/Topic`.

If you would like to use a differently-named Kubernetes `Role` than the
defaults, you are welcome to do so by modifying the Kubernetes manifests that
are used as part of the installation process.

Once the Kubernetes `Role` objects have been created, you will want to assign
specific a Kubernetes `User` to a particular `Role`. You do this using the
typical Kubernetes `RoleBinding` object.

For example, assume you want to have the Kubernetes `User` named "Alice" have
the ability to create, read, delete and modify CRs that ACK service controllers
manage in the Kubernetes "default" `Namespace`, you would create a
`RoleBinding` that looked like this:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ack.user
  namespace: default
subjects:
- kind: User
  name: Alice
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: ack.user
  apiGroup: rbac.authorization.k8s.io
```

## Configuring AWS IAM

Since ACK service controllers bridge the Kubernetes and AWS API worlds, in
addition to configuring Kubernetes RBAC permissions, you will need to ensure
that all AWS Identity and Access Management (IAM) roles and permissions have
been properly created.

TODO

### Cross-account resource management

TODO
