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
github repository with a semver tag (Ex: `0.0.1`).
Steps below show how to create a new release with semver tag.

{{% hint type="info" title="Semver" %}}
For more details on semantic versioning(semver), please read our [release phase guide](../../community/releases/)
{{% /hint %}}

Once the git repository is tagged with semver, a postsubmit prowjob builds 
container image for ACK service controller and publish to public ecr repository `public.ecr.aws/aws-controllers-k8s/controller`.
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

* Container image
* Helm chart

The container image is built and pushed with an image tag that indicates the
release version for the controller along with the AWS service. For example,
assume a release semver tag of `0.1.0` that includes service controllers for
S3 and SNS. There would be two container images built for this release, one each
containing the ACK service controllers for S3 and SNS. The container images would
have the following image tags: `s3-0.1.0` and `sns-0.1.0`. Note
that the full image name would be
`public.ecr.aws/aws-controllers-k8s/s3-controller:0.1.0`

The Helm chart artifact can be used to install the ACK service controller as a
Kubernetes Deployment; the Deployment's Pod image will refer to the exact
container image tag matching the release tag.

## Release steps

0) Rebase $SERVICE-controller repo with latest code:
```bash
cd $GOSRC/github.com/aws-controllers-k8s
export SERVICE=s3
cd $SERVICE-controller
git fetch --all --tags
# Optionally fetch and rebase the latest code generator
cd ../code-generator
git checkout main && git fetch --all --tags && git rebase upstream/main
```

1) Navigate to $SERVICE-controller repo and check out a git branch for your release:
```bash
cd ../$SERVICE-controller
export RELEASE_VERSION=v0.0.1
git checkout -b release-$RELEASE_VERSION
git branch --set-upstream-to=origin/main release-$RELEASE_VERSION
```

2) Navigate to code-generator repo and build the release artifacts for the $SERVICE-controller:
```bash
cd ../code-generator
make build-controller
```


3) Navigate to $SERVICE-controller repo to review the release artifacts that were built for each service by looking in the `helm`
directory:
```bash
cd ../$SERVICE-controller
git diff
```

{{% hint %}}
When you run `make build-controller` for a service, it will overwrite any
Helm chart files that had previously been generated in the `$SERVICE-controller/helm`
directory with files that refer to the container image with an image tag
referring to the release you've just built artifacts for.

{{% /hint %}}

4) Commit the generated release artifacts and create a pull request:
```bash
git commit -a -m "release artifacts for release $RELEASE_VERSION"
git push origin release-$RELEASE_VERSION
```

5) Get your pull request reviewed and merged. After merge, tag is automatically applied and pushed.

6) `git tag` operation (applied automatically in last step) triggers a postsubmit prowjob which builds container image and then publishes
both container image and Helm chart to public ECR repository.
Service team can see the release prowjobs, their status and logs at https://prow.ack.aws.dev/
