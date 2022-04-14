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

{{% hint type="info" title="Semver" %}}
For more details on semantic versioning(semver), please read our [release phase guide](../../community/releases/)
{{% /hint %}}

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
directory with files that refer to the Docker image with an image tag
referring to the release you've just built artifacts for.

{{% /hint %}}

4) Commit the generated release artifacts and create a pull request:
```bash
git commit -a -m "release artifacts for release $RELEASE_VERSION"
git push origin release-$RELEASE_VERSION
```

5) Get your pull request reviewed and merged. After merge, tag is automatically applied and pushed.

6) `git tag` operation (applied automatically in last step) triggers a postsubmit prowjob which builds binary docker image and then publishes
both docker image and Helm chart to public ECR repository.
Service team can see the release prowjobs, their status and logs at https://prow.ack.aws.dev/

## Stable Release
The postsubmit prowjob mentioned above also publishes the stable Helm charts,
whenever there is a code push on `stable` git branch. Follow the steps below
to cut a stable release for an ACK controller.

1) Checkout the ACK controller release which will be marked as stable.
Example below uses s3-controller v0.0.19 release.
```bash
cd $GOSRC/github.com/aws-controllers-k8s
export SERVICE=s3
export STABLE_RELEASE=<v0.0.19-do-not-copy> #Update this tag for the specific controller
cd $SERVICE-controller
git fetch --all --tags
git checkout -b stable-$STABLE_RELEASE $STABLE_RELEASE
```

2) Update the helm chart version to the stable version. To learn more about
nomenclature of stable branch and helm chart version please read our
[release phase guide](../../community/releases/).

For the above example, replace `version: v0.0.19` inside `helm/Chart.yaml`
with `version: v0-stable`. Without this update the postsubmit prowjob will
fail because validation error due to chart version mismatch.

3) Commit your changes from step2
```bash
git add helm/Chart.yaml
git commit -m "Updating the helm chart version for stable release"
```

4) Determine the remote which points to `aws-controllers-k8s/$SERVICE-controller`
and not your personal fork. Execute `git remote --verbose` command to find out the
remote name. Example: In the command below, `origin` points to the
`aws-controllers-k8s/s3-controller` repository.

```bash
git remote --verbose

origin  https://github.com/aws-controllers-k8s/s3-controller.git (fetch)
origin  https://github.com/aws-controllers-k8s/s3-controller.git (push)
vj      https://github.com/vijtrip2/s3-controller.git (fetch)
vj      https://github.com/vijtrip2/s3-controller.git (push)
```

5) Push the changes to the `stable` branch for remote pointing to
`aws-controllers-k8s/$SERVICE-controller`
```bash
git push -u origin stable-$STABLE_RELEASE:stable
```
The above command will create a new `stable` branch if it does not exist
and trigger the ACK postsubmit prowjob for stable release. This prowjob will
not build a docker image and only publishes the helm artifacts with stable tag.

If the git push command fails, use `--force` option to update the upstream
`stable` branch with your local changes.

