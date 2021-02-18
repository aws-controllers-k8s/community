# Release

Here we document the release process for ACK service controllers.

Remember that there is no single ACK binary. Rather, when we build a release
for ACK, that release is for one or more individual ACK service controllers
binaries, each of which are installed separately.

This documentation covers the steps involved in building a service controller's
release artifacts, including the Helm charts and binary Docker images for an
ACK service controller, publishing those artifacts and tagging the Git source
repository appropriately to create an official "release".

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

1. First check out a git branch for your release:
 
```bash
export RELEASE_VERSION=v0.0.1
git checkout -b release-$RELEASE_VERSION
 ```

2. Build the release artifacts for the controllers you wish to include in the
   release

   Run `scripts/build-controller-release.sh` for each service. For
   instance, to build release artifacts for the SNS and S3 controllers I would
   do:

```bash
for SERVICE in sns s3; do
    ./scripts/build-controller-release.sh $SERVICE $RELEASE_VERSION;
done
```

3. You can review the release artifacts that were built for each service by
   looking in the `services/$SERVICE/helm` directory:

    `tree services/$SERVICE/helm`

    or by doing:

    `git diff`

!!! note
    When you run `scripts/build-controller-release.sh` for a service, it will
    overwrite any Helm chart files that had previously been generated in the
    `services/$SERVICE/helm` directory with files that refer to the
    Docker image with an image tag referring to the release you've just built
    artifacts for.

4. Commit your code and create a pull request:

```bash
git commit -a -m "release artifacts for release $RELEASE_VERSION"
```

5. Get your pull request reviewed and merged.

6. Upon merging the pull request

```bash
git tag -a $RELEASE_VERSION $( git rev-parse HEAD )
git push upstream main --tags
```

!!! todo
    A Github Action should execute the above

which will end up associating a Git tag (and therefore a Github release) with
the SHA1 commit ID of the source code for the controllers and the release
artifacts you built for that release version.

7. Publish the controller images

First, ensure you are logged in to the ECR public repository:

```bash
aws --profile ecrpush ecr-public get-login-password --region us-east-1 | docker login -u AWS --password-stdin public.ecr.aws
Login Succeeded
```

!!! note
    Above, I have a set of AWS CLI credentials in a profile called "ecrpush"
    that I use for pushing to the ACK public ECR repository. You will need
    something similar.

Now publish all the controller images to the ECR public repository for
controller images:

```bash
export DOCKER_REPOSITORY=public.ecr.aws/aws-controllers-k8s/controller
for SERVICE in s3 sns; do
    export AWS_SERVICE_DOCKER_IMG=$DOCKER_REPOSITORY:$SERVICE-$RELEASE_VERSION
    ./scripts/publish-controller-image.sh $SERVICE
done
```

!!! todo

    The same Github Action should run the
    `scripts/publish-controller-images.sh` script to build the Docker images
    for the service controllers included in the release and push the images to
    the `public.ecr.aws/aws-controllers-k8s/controller` image repository.

8. Publish the Helm Charts

First, ensure you are logged in to the ECR public repository for Helm:

```bash
aws --profile ecrpush ecr-public get-login-password --region us-east-1 | HELM_EXPERIMENTAL_OCI=1 helm registry login -u AWS --password-stdin public.ecr.aws
Login succeeded
```

!!! note
    Above, I have a set of AWS CLI credentials in a profile called "ecrpush"
    that I use for pushing to the ACK public ECR repository. You will need
    something similar.

Now publish all the controller images to the ECR public repository for
controller image using the `scripts/helm-publish-chart.sh` script for each service:

```bash
for SERVICE in apigatewayv2 sns; do
    RELEASE_VERSION=v0.0.1 ./scripts/helm-publish-chart.sh $SERVICE;
done
```

```bash
Generating Helm chart package for apigatewayv2@v0.0.1 ... ref:     public.ecr.aws/aws-controllers-k8s/chart:apigatewayv2-v0.0.1
digest:  0e24159c9afb840677ba64e63c19a65a6de2dcc87e80df95b3daf0cdb5c54de6
size:    6.4 KiB
name:    ack-apigatewayv2-controller
version: v0.0.1
apigatewayv2-v0.0.1: saved
ok.
The push refers to repository [public.ecr.aws/aws-controllers-k8s/chart]
ref:     public.ecr.aws/aws-controllers-k8s/chart:apigatewayv2-v0.0.1
digest:  0e24159c9afb840677ba64e63c19a65a6de2dcc87e80df95b3daf0cdb5c54de6
size:    6.4 KiB
name:    ack-apigatewayv2-controller
version: v0.0.1
apigatewayv2-v0.0.1: pushed to remote (1 layer, 6.4 KiB total)
<snip>
Generating Helm chart package for sns@v0.0.1 ... ref:     public.ecr.aws/aws-controllers-k8s/chart:sns-v0.0.1
digest:  d5c1a79f85f8c320210c3418e7175da5398fba4e5644cd49f107c19db9e1e6d1
size:    4.0 KiB
name:    ack-sns-controller
version: v0.0.1
sns-v0.0.1: saved
ok.
The push refers to repository [public.ecr.aws/aws-controllers-k8s/chart]
ref:     public.ecr.aws/aws-controllers-k8s/chart:sns-v0.0.1
digest:  d5c1a79f85f8c320210c3418e7175da5398fba4e5644cd49f107c19db9e1e6d1
size:    4.0 KiB
name:    ack-sns-controller
version: v0.0.1
sns-v0.0.1: pushed to remote (1 layer, 4.0 KiB total)
```

All services that have had a Helm chart generated will have a
corresponding Helm chart pushed to the ECR public repository.
