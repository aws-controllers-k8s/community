---
title: "Release"
description: "The release process for ACK service controller"
lead: "The release process for ACK service controller"
draft: false
menu: 
  docs:
    parent: "contributor"
weight: 70
toc: true
---

Remember that there is no single ACK binary. Rather, when we build a release
for ACK, that release is for one or more individual ACK service controllers
binaries, each of which are installed separately.

This documentation covers the steps involved for officially publishing
a ACK service controller's release artifacts.

Once ACK service controller changes are tested by the service team and they wish to
release latest artifacts, service team only needs to create a new release for service-controller
github repository with a semver tag (Ex: v0.0.1). 
Steps below show how to create a new release with semver tag.

!!! note "Semver"
    For more details on semantic versioning(semver), please read [releases.md](https://aws-controllers-k8s.github.io/community/releases/) 

Once the git repository is tagged with semver, a postsubmit prowjob builds binary 
docker image for ACK service controller and publish to public ecr repository `public.ecr.aws/aws-controllers-k8s/controller`.
Same prowjob also publishes the Helm charts for the ACK service controller to
public ecr repository `public.ecr.aws/aws-controllers-k8s/chart`.

## What is a release exactly?

A "release" is the combination of a Git tag containing a SemVer version tag
against this source repository and the collection of *artifacts* that allow the
individual ACK service controllers included in that Git commit to be easily
installed via Helm.

The Git tag points at a specific Git commit referencing the exact source code
that comprises the ACK service controllers in that "release".

The release artifacts include the following for one or more service
controllers:

* Docker image
* Helm chart

The Docker image is built and pushed with an image tag that indicates the
release version for the controller along with the AWS service. For example,
assume a release semver tag of `v0.1.0` that includes service controllers for
S3 and SNS. There would be two Docker images built for this release, one each
containing the ACK service controllers for S3 and SNS. The Docker images would
have the following image tags: `s3-v0.1.0` and `sns-v0.1.0`. Note
that the full image name would be
`public.ecr.aws/aws-controllers-k8s/controller:s3-v0.1.0`

The Helm chart artifact can be used to install the ACK service controller as a
Kubernetes Deployment; the Deployment's Pod image will refer to the exact
Docker image tag matching the release tag.

## Release steps
<// mkdocs does not support numbered lists with code blocks. So use 1) instead of 1. in numbered list. >
1) First check out a git branch for your release:
```bash
export RELEASE_VERSION=v0.0.1
git checkout -b release-$RELEASE_VERSION
```

2) Build the release artifacts for the controllers you wish to include in the
   release

   Run `make build-controller` for each service from code-generator repository.
    For instance, to build release artifacts for the SNS and S3 controllers I
    would do:

```bash
for SERVICE in sns s3; do
    export SERVICE;
    echo "building ACK controller for $SERVICE, Version: $RELEASE_VERSION"
    make build-controller;
done
```

3) You can review the release artifacts that were built for each service by looking in the `services/$SERVICE/helm`
directory:

`tree services/$SERVICE/helm`

or by doing:

`git diff`

!!! note
    When you run `make build-controller` for a service, it will overwrite any
    Helm chart files that had previously been generated in the `services/$SERVICE/helm`
    directory with files that refer to the Docker image with an image tag
    referring to the release you've just built artifacts for.
   
4) Commit your code and create a pull request:
```bash
git commit -a -m "release artifacts for release $RELEASE_VERSION"
```

5) Get your pull request reviewed and merged.

6) Upon merging the pull request
```bash
git tag -a $RELEASE_VERSION $( git rev-parse HEAD )
git push upstream main --tags
```

!!! note "TODO"
    A Github Action should execute the above step which will end up associating a Git tag (and therefore a Github
    release) with the SHA1 commit ID of the source code for the controllers and the release artifacts you built for
    that release version.

7) `git tag` operation from last step triggers a postsubmit prowjob which builds binary docker image and then publishes
both docker image and Helm chart to public ECR repository.
Service team can see the release prowjobs, their status and logs at https://prow.ack.aws.dev/

!!! note "Stable Helm Chart"
    * This same postsubmit prowjob also publishes the stable Helm charts, whenever there is a code push on `stable` git 
    branch.
    * To learn more about how to push changes to stable branch please read [releases.md](https://aws-controllers-k8s.github.io/community/releases/)
    * When this prowjob is triggered from `stable` branch, it does not build a docker image and only publishes the helm 
    artifacts with stable tag. Ex: `elasticache-v1-stable`
