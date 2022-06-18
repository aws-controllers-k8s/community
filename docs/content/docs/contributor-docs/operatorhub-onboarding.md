---
title: "OperatorHub Onboarding"
description: "How controllers end up in OperatorHub"
lead: ""
draft: false
menu:
  docs:
    parent: "contributor"
weight: 20
toc: true
---

There are two ways a user can install an operator, one is via an OLM enabled cluster using OperatorHub.io and the other
is via the embedded OperatorHub within an OpenShift cluster. In order to onboard a new controller and have it appear in both
places, the below steps should be followed. After these steps are completed, the build/release process will then raise pull
requests against the proper repos.

## Add an OLM Config File to Controller Repository

The OLM config file is used during the build/release process of a controller to assist in the generation of the
[ClusterServiceVersion](https://olm.operatorframework.io/docs/concepts/crds/clusterserviceversion/) ("CSV") in the controller's bundle.
The file should live at `./olm/olmconfig.yaml` in the project structure of a controller. It should also contain a sample for each `CustomResource`
managed by the controller. Please see the S3 controller's `olmconfig.yaml` found
[here](https://github.com/aws-controllers-k8s/s3-controller/blob/main/olm/olmconfig.yaml) for proper formatting.

## Validate the Generated CSV

After the `olmconfig.yaml` has been generated it's a good practice to validate that the CSV for the controller looks
as expected, this ensures proper AWS branding for the controller. This does not need to be done after ever change to the
controller, but if changes to the `olmconfig.yaml` are done, or a new CR is added to the controller, it's best to validate that
the new changes in the CSV appears as expected.

1. Build the controller locally using the `code-generator` project
   1. Install Operator SDK in the `code-generator` `/bin` directory using the below script
      1. `$ ./scripts/install-operator-sdk.sh`
   2. Target the appropriate controller
      1. `$ export SERVICE=s3`
   3. Build the controller and generate the bundle
      1. `$ ACK_GENERATE_OLM=true make build-controller SERVICE=$SERVICE`
2. Validate that the CSV was generated
   1. Unless overridden, the CSV will be at `$GOPATH/src/github.com/aws-controllers-k8s/s3-controller/olm/bundle/manifests/ack-s3-controller.clusterserviceversion.yaml`
3. The CSV can be [previewed](https://operatorhub.io/preview) by copying and pasting the CSV

   ![OperatorHub.io Preview](../images/operatorhub-preview.png)

## Raise Pull Requests to Community Operators Repositories

Both repositories rely on the same folder structure for each operator, which is laid out below. For the initial onboarding
all that needs to be worried about is adding `./operators/ack-new-controller` and the `ci.yaml` file. Since the ACK project
releases all the operators using semantic versioning, each ACK operator CI file will be identical, so an existing ACK operator's
CI file can be copied and used in the Pull Request.

```shell
.
└── ack-new-controller
    ├── 0.0.1
    ├── 0.0.2
    └── ci.yaml
```

1. Raise a Pull Request for OperatorHub.io [here](https://github.com/k8s-operatorhub/community-operators)
  - Below is a quote from this repository's Readme file explaining its usage
    > This repo is the canonical source for Kubernetes Operators that appear on [OperatorHub.io](https://operatorhub.io).
    The solutions merged on this repository are distributed via the [OLM][olm] index catalog [quay.io/operatorhubio/catalog][quay.io].
    Users can install [OLM][olm] in any Kubernetes or vendor such as Openshift to consume this content by adding a new CatalogSource for the index image
    `quay.io/operatorhubio/catalog`. [(more info)][catalog]

2. Raise a Pull Request for embedded OperatorHub in OpenShift [here](https://github.com/redhat-openshift-ecosystem/community-operators-prod)
  - Below is a quote from this repository's Readme file explaining its usage
    > This repo is the canonical source for Kubernetes Operators that appear on [OpenShift Container Platform](https://openshift.com)
    and [OKD](https://www.okd.io/).

The build for these Pull Requests will fail since no bundle has been provided, but that is okay, the maintainers will still review and
merge the pull request. After these Pull Requests have been merged, the new controller is now onboarded and ready for a release.

[olm]: https://github.com/operator-framework/operator-lifecycle-manager
[quay.io]: https://quay.io/repository/operatorhubio/catalog?tag=latest&tab=tags
[catalog]: https://k8s-operatorhub.github.io/community-operators/testing-operators/#1-create-the-catalogsource
