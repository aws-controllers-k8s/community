# Releases

We use the following different release statuses:

* [PROPOSED](#proposed)
* [PLANNED](#planned)
* [BUILD](#build)
* [DEVELOPER PREVIEW](#developer-preview)
* [BETA](#beta)
* [GENERALLY AVAILABLE](#generally-available) (GA)

## PROPOSED

The `PROPOSEDD` status indicates that someone has expressed interest in
supporting said service in ACK.

There will be a Github Issue for tracking the build of the ACK service
controller for the service.

The GitHub Issue **WILL NOT** be associated with a GitHub Milestone.

## PLANNED

The `PLANNED` status indicates that **the AWS service is on our radar** for
building into ACK.

There will be a Github Issue for tracking the build of the ACK service
controller for the service.

The GitHub Issue **WILL BE** associated with a GitHub Milestone indicating the
target date for the `DEVELOPER PREVIEW` release of that service's controller.

## BUILD

The `BUILD` status indicates that the ACK service controller for the AWS
service is **actively being built** in preparation for `DEVELOPER PREVIEW`
release of that ACK service controller.

It is in the `BUILD` status that we will identify those **AWS service API
resources** that will be supported in `DEVELOPER PREVIEW` and which will be
supported by the `GENERALLY AVAILABLE` release of the controller.

!!! note "what do we mean by 'AWS service API resources'?
    An *AWS service API resource* is a top-level object that can be created by
    a particular AWS service API. For example, an SNS Topic or an S3 Bucket.
    Some service APIs have multiple top-level resources; SNS, for instance, has
    Topic, PlatformApplication and PlatformEndpoint top-level resources that
    may be created.

## DEVELOPER PREVIEW

The `DEVELOPER PREVIEW` status indicates that the **source code and some
documentation** for the ACK service controller for the AWS service has been
check into the ACK source repository along with **minimual end-to-end test
cases** that run against a local Kubernetes-in-Docker (KinD) cluster.

Notably, an ACK service controller in `DEVELOPER PREVIEW` **DOES NOT include
Helm charts or published binary Docker images** for easy installation of the
controller.

The following are release criteria for `DEVELOPER PREVIEW`:

* Source code for the controller checked into ACK source repository
* Mininal (smoke) tests for at least one service API resource
* Documentation on to test the controller using KinD

The Custom Resource Definition (CRD) `APIVersion` for resources managed by ACK
service controllers in `DEVELOPER PREVIEW` will carry an `alpha` designation --
e.g. `v1alpha3`. Developers can expect major changes to the structure and
format of the CRDs during the `DEVELOPER PREVIEW` release status.

## BETA

The `BETA` status indicates that the ACK service controller for the AWS service
**has Helm charts and published binary Docker images** that users can use to
easily install and configure the controller.

In addition, ACK service controllers in `BETA` status have **more extensive
end-to-end tests** included in the source repository and the successful running
of these tests **gate any changes to the service controller source code**.

The following are release criteria for `BETA`:

* End-to-end tests of all CRUD operations for at least one service API resource
* Documentation on how to install and configure the controller using Helm
* All necessary artifacts such as container images and Helm charts are
  available via a public container registry

The Custom Resource Definition (CRD) `APIVersion` for resources managed by ACK
service controllers in `BETA` will carry an `beta` designation -- e.g.
`v1beta1`. Developers can expect minor changes to the structure and format of
the CRDs during the `BETA` release status, however **any change to the CRD
format during the `BETA` release status will result in an incremented
`APIVersion` for the CRD**. For example, consider a CRD with a
`GroupVersionKind` (GVK) of `Bucket.s3.services.k8s.aws/v1beta2`. If the format
of that CRD changed, we guarantee that a new `v1beta3` package and
corresponding `Bucket.s3.services.k8s.aws/v1beta3` GVK would be released,
allowing developers to cleanly migrate between API versions of their custom
resources.

## GENERALLY AVAILABLE

An ACK service controller reaches the `GENERALLY AVAILABLE` (GA) release status
once a controller in the `BETA` release status has satisfied the following
criteria:

* End-to-end test coverage of all CRUD operations for all service API
  resources for the specific service
* Test coverage for "negative" or "fuzz" testing to ensure validating webhooks
  properly guard the resource creation and mutation
* Long-running "soak" testing to ensure controller reliability and stability
* Documentation is included that demonstrates usage of the custom resources
  managed by the controller

The Custom Resource Definition (CRD) `APIVersion` for resources managed by ACK
service controllers in `GENERALLY AVAILABLE` will have a major version with no
`alpha` or `beta` designation -- e.g.  `v1`. Developers can expect no changes
to the structure and format of the CRDs once the controller is in the
`GENERALLY AVAILABLE` release status.

!!! note "How long from developer preview to GA?"
    We aim to get an ACK service controller into the `GENERALLY AVAILABLE`
    release status within 3 months of when the controller is released as
    `DEVELOPER PREVIEW`. Some controllers will naturally take longer due to
    inconsistencies, corner cases or complexity of the underlying AWS service
    API.
