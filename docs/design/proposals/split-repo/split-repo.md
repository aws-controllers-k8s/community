# Split Monorepo

This proposal seeks to split the existing monolithic source code repository
into a set of smaller source code repositories containing individual ACK
service controllers. This will allow the common ACK runtime code to be
versioned separately from ACK service controllers and allow individual ACK
service controllers to implement their own release branching mechanics.

## The Problem

ACK is currently contained in a single monolithic source repository
(github.com/aws/aws-controllers-k8s). Within this monorepo are a number of
distinct pieces:

* ACK core API types (`/apis/core`)
* ACK common controller configuration and runtime (`/pkg/runtime`, `/pkg/config`, `/pkg/types`)
* ACK code generator (`/cmd/ack-generate`, `/pkg/generate`, `/pkg/model`, `/templates`)
* ACK common build and test scripts (`/scripts`, `/test`)
* Individual AWS service-specific controllers (`/services/$SERVICE`)
* Individual AWS service-specific tests (`/test/e2e/$SERVICE`)

When we generate service controller code in ACK, we build the `ack-generate`
CLI from the code in this monorepo and output the generated code into the
`/services/$SERVICE` directory also within this monorepo.

When we build releases in ACK, the releases are Git tags on the monorepo --
e.g. `v0.0.2` -- and then individual ACK service controller binaries (Docker
images) are built and pushed to the [ECR public ACK controller image][ecr]
registry with an image tag of `$SERVICE-$RELEASE_VERSION` -- e.g. `s3-v0.0.2`

[ecr]: https://gallery.ecr.aws/aws-controllers-k8s/controller

### Problem #1: Release series inflexibility

This monorepo tightly couples the ACK common runtime code, code
generator/templates and the individual service controller code, making the
release process inflexible and cumbersome. The Elasticache team cannot have
separate release series -- say, a `stable` and a `preview` release series --
because releases are Git tags, and Git tags are specific to one source
repository. Numeric releases such as `v0.2.0` refer to a specific
content-addressable SHA1 Git commit, and with all service controllers in the
same source repository, this means that numeric Git tags cannot be applied to
two different service controllers to represent an incremental change in one
specific service controller's codebase.

### Problem #2: Common runtime import inflexibility

Furthermore, due to the code generator and ACK common runtime code being in the
same repo as the service controllers that `import` this code, it's not easy to
depend on a specific version of that common runtime from within the service
controller code itself. Individual service controller package directories would
need their own `go.mod` files and managing this would be a pain and wouldn't
solve Problem #1 anyway.

## Solution

We propose splitting the monorepo into a set of source code repositories that
can be separately Git-tagged and have their releases managed according to the
preferences of the team that owns that component or individual ACK service
controller.

Specifically, we propose the following changes:

1) All code for generating code will move into a new
   `github.com/aws-controllers-k8s/code-generator` source repository.

   This includes all content in the existing `/cmd/ack-generate`, `/templates`,
   `/pkg/generate`, `/pkg/testutil`,`/pkg/util` and `/pkg/model` directories of
   the existing `github.com/aws/aws-controllers-k8s` source repository.

   The code generator repository will have a single Git branch `main` and all
   releases will be SemVer -- e.g. `v0.0.3`.

   The `ack-generate` CLI tool will get a new `--aws-sdk-version` flag that
   will cause the code generator to consider the API model in a specific SemVer
   release of the `github.com/aws/aws-sdk-go` dependent repository.

2) The ACK common runtime and core ACK types will be split out into a
   `github.com/aws-controllers-k8s/runtime` source repository.

   The ACK common runtime repository will have a single Git branch `main` and
   all releases will be SemVer -- e.g. `v0.0.3`.

   All service controllers will import -- via a `go.mod` dependency line -- a
   specific SemVer version of the common ACK runtime.

3) Each ACK service controller will move into a new
   `github.com/aws-controllers-k8s/$SERVICE-controller` source repository.

   There will be a single Git branch `main` for these individual service
   controller source repositories. Teams that own individual service
   controllers may create an additional `stable` Git branch if they would like
   to have releaseable configuration artifacts that point to configuration
   values or controller image tags that are "baked" longer in a separate Git
   branch.

4) The primary `github.com/aws/aws-controllers-k8s` repository will be
   **renamed** to `github.com/aws-controllers-k8s/community` in order to move
   Github Issues, Projects, activity/stars, etc.

5) The `/test` directory will remain in the newly-renamed
   `github.com/aws-controllers-k8s/community` source repository and be adapted to
   build controller images and configuration manifests for test purposes from the
   new separated controller source repositories instead of a `/services/$SERVICE`
   directory in the `github.com/aws-controllers-k8s/community` repository.

6) The common build and image/chart publication scripts in the newly-renamed
   `github.com/aws-controllers-k8s/community` will remain in that repo but will be
   adapted as needed to allow pointing at different source repositories instead
   of different subdirectories in the main monorepo. We will continue to
   publish Docker images and Helm charts to the same single ECR registry and
   continue to use the same basic image tagging scheme:
   `$SERVICE-$RELEASE_VERSION`.

   We may also publish Helm chart tags that point to the latest Git tag in
   either the `main` or `stable` Git branch for the service controller,
   allowing users to install a Helm chart, e.g. `chart/s3-stable` that pulls an
   image for the S3 controller matching the last Git tag on the `stable` branch
   of the `github.com/aws-controllers-k8s/s3-controller` source repository.
