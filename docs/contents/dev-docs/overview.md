# Overview

This section of the docs is for ACK contributors.

## Code Organization

ACK is a collection of source repositories containing a common runtime and type
system, a code generator and individual service controllers that manage
resources in a specific AWS API.

* `github.com/aws-controllers-k8s/community` docs and common tests (this repo)
* `github.com/aws-controllers-k8s/runtime`: common ACK runtime and types
* `github.com/aws-controllers-k8s/code-generator`: the code generator and
  templates
* `github.com/aws-controllers-k8s/$SERVICE-controller`: individual ACK
  controllers for AWS services.

### `github.com/aws-controllers-k8s/community` (this repo)

The `github.com/aws-controllers-k8s/community` source code repository (this
repo) contains the common test scripts and documentation that gets published to
https://aws-controllers-k8s.github.io/community/.

### `github.com/aws-controllers-k8s/runtime`

The `github.com/aws-controllers-k8s/runtime` source code repository contains
the common ACK controller runtime (`/pkg/runtime`, `/pkg/types`) and core
public Kubernetes API types (`/apis/core`).

### `github.com/aws-controllers-k8s/code-generator`

The `github.com/aws-controllers-k8s/code-generator` source code repository
contains the `ack-generate` CLI tool (`/cmd/ack-generate`), the Go packages
that are used in API inference and code generation (`/pkg/generate`,
`/pkg/model`) and Bash scripts to build an ACK service controller
(`/scripts/build-controller.sh`).

### `github.com/aws-controllers-k8s/$SERVICE-controller`

Each AWS API that has had a Kubernetes controller built to manage resources in
that API has its own source code repository in the
`github.com/aws-controllers-k8s` Github Organization. The source repos will be
called `$SERVICE-controller`.

These service controller repositories contain Go code for the main controller
binary (`/cmd/controller/`), the public API types for the controllers
(`/apis`), the Go code for the resource managers used by the controller
(`/pkg/resource/*/`), static configuration manifests (`/config`) and Helm
charts for the controller installation (`/helm`).

## API Inference

Read about [how the code generator infers][api-inference] information about a
Kubernetes Custom Resource Definitions (CRDs) from an AWS API model file.

[api-inference]: https://aws-controllers-k8s.github.io/community/dev-docs/api-inference/

## Code Generation

The [code generation](../code-generation/) section gives you a bit of background
on how we go about automating the code generation for controllers and supporting
artifacts.

## Setting up a Development Environemnt

In the [setup](../setup/) section we walk you through setting up your local Git
environment with the repo and how advise you on how we handle contributions.

## Building an ACK Service Controller

After getting your development environment established, you will want to learn
[how to build an ACK service controller](../build-controller).

## Testing an ACK Service Controller

Last but not least, in the [testing](../testing/) section we show you how to
test ACK locally.
