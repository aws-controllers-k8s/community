# AWS Controllers for Kubernetes

This repo contains a set of Kubernetes
[controllers](https://kubernetes.io/docs/reference/glossary/?fundamental=true#term-controller).
Each controller manages [custom
resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
that represent API resources in a single AWS service API. For example, the
service controller for AWS Simple Storage Service (S3) manages custom resources
that represent AWS S3 Buckets.

Instead of logging into the AWS console or using the `aws` CLI tool to interact
with the AWS service API, Kubernetes users may install a controller for an AWS
service and then create, update, read and delete resources using the Kubernetes
API. This means they can use the Kubernetes API to fully describe both their
containerized applications (using Kubernetes resources like `Deployment` and
`Service`) as well as any AWS managed services upon which those applications
depend.

## Installation

You have a number of choices when installing an AWS service controller:

* Helm (recommended)
* Static Kubernetes manifests

### Helm (recommended)

The recommended way to install an AWS service controller for Kubernetes is to
use Helm.

Before installing an AWS service controller, first ensure you have added the
AWS Controllers for Kubernetes Helm repository:

```sh
helm repo add ack https://aws.github.io/aws-service-operator-k8s
```

Each AWS service controller is packaged into a separate container image
(published on a public AWS Elastic Container Registry repository). Likewise,
each AWS service controller has a separate Helm chart that installs (as a
Kubernetes `Deployment`) the AWS service controller, necessary custom resource
definitions (CRDs), Kubernetes RBAC manifests, and other supporting artifacts.

You may install a particular AWS service controller using the `helm install`
CLI command:

```sh
helm install [--namespace $KUBERNETES_NAMESPACE] ack/$SERVICE_ALIAS
```

for example, if you wanted to install the AWS S3 service controller into the
"ack-system" Kubernetes namespace, you would execute:


```sh
helm install --namespace ack-system ack/s3
```

### Static Kubernetes manifests

If you prefer not to use Helm, you may install a service controller using
static Kubernetes manifests.

Static Kubernetes manifests that install individual service controllers are
attached as artifacts to releases of AWS Controllers for Kubernetes. Select a
release from the [list of
releases](https://github.com/aws/aws-service-operator-k8s/releases) for AWS
Controllers for Kubernetes.

You will see a list of Assets for the release. One of those Assets will be
named `services/$SERVICE_ALIAS/all-resources.yaml`. For example, for the AWS S3
service controller, there will be an Asset named
`services/s3/all-resources.yaml` attached to the release. Click on the link to
download the YAML file. This YAML file may be fed to `kubectl apply -f`
directly to install the service controller, any CRDs that it manages, and all
necessary Kubernetes RBAC manifests.

For example:

```sh
kubectl apply -f https://github.com/aws/aws-service-operator-k8s/releases/download/v0.0.1/services/s3/all-resources.yaml
```

## Configure permissions

Because ACK bridges the Kubernetes and AWS APIs, before using ACK service
controllers, you will need to do some initial configuration around Kubernetes
and AWS Identity and Access Management (IAM) permissions.

### Configuring Kubernetes RBAC

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

### Configuring AWS IAM

Since ACK service controllers bridge the Kubernetes and AWS API worlds, in
addition to configuring Kubernetes RBAC permissions, you will need to ensure
that all AWS Identity and Access Management (IAM) roles and permissions have
been properly created.

TODO

#### Cross-account resource management

TODO

## Usage

### Prerequisites

Before using AWS Controllers for Kubernetes, first:

* [install](#Installation) one or more ACK service controllers.
* [configure permissions](#configure-permissions) of Kubernetes and AWS
  Identity and Access Management Roles.

### Creating an AWS resource via the Kubernetes API

TODO

### Viewing AWS resource information via the Kubernetes API

TODO

### Deleting an AWS resource via the Kubernetes API

TODO

### Modifying an AWS resource via the Kubernetes API

TODO

## Status

As of end of 2019 we are in the early stages of planning the redesign of ASO. Check the [issues list](https://github.com/aws/aws-service-operator-k8s/issues) for descriptions of work items. We invite any and all feedback and contributions, so please don't hesitate to submit an issue, a pull request or comment on an existing issue.

For discussions, please use the `#provider-aws` channel on the [Kubernetes Slack](https://kubernetes.slack.com) community or the [mailing list](https://groups.google.com/forum/#!forum/aws-service-operator-user/).

## Background

This repo contains the next generation AWS Service Operator for Kubernetes
(ASO).

An operator in Kubernetes is the combination of one or more [custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) and [controllers](https://kubernetes.io/docs/reference/glossary/?fundamental=true#term-controller) managing said custom resources.

The ASO will allow containerized applications and Kubernetes users to create, update, delete and retrieve the status of objects in AWS services such as S3 buckets, DynamoDB, RDS databases, SNS, etc. using the Kubernetes API, for example using 
Kubernetes manifests or `kubectl` plugins.

Read about the [motivation for and background on](docs/background.md) the ASO.

## License

This project is licensed under the Apache-2.0 License.
