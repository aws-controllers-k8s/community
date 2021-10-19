---
title: "Code Organization"
description: "How the source code for ACK is organized"
lead: ""
draft: false
menu:
  docs:
    parent: "contributor"
weight: 20
toc: true
---

ACK is a collection of source repositories containing a common runtime and type
system, a code generator and individual service controllers that manage
resources in a specific AWS API.

* [`github.com/aws-controllers-k8s/community`][comm-repo]: docs, issues and
  project management (this repo)
* [`github.com/aws-controllers-k8s/runtime`][rt-repo]: common ACK runtime and types
* [`github.com/aws-controllers-k8s/code-generator`][codegen-repo]: the code generator and
  templates
* [`github.com/aws-controllers-k8s/test-infra`][testinfra-repo]: common test code and infrastructure
* `github.com/aws-controllers-k8s/$SERVICE-controller`: individual ACK
  controllers for AWS services.

## `github.com/aws-controllers-k8s/community` (this repo)

The [`github.com/aws-controllers-k8s/community`][comm-repo] source code
repository (this repo) contains the documentation that gets published to
https://aws-controllers-k8s.github.io/community/.

{{% hint type="info" title="Bug reports and feature requests" %}}
**NOTE**: All [bug reports and feature requests][issues] for all ACK source repositories
are contained in this repository.
{{% /hint %}}

## `github.com/aws-controllers-k8s/runtime`

The [`github.com/aws-controllers-k8s/runtime`][rt-repo] source code repository contains
the common ACK controller runtime (`/pkg/runtime`, `/pkg/types`) and core
public Kubernetes API types (`/apis/core`).

## `github.com/aws-controllers-k8s/code-generator`

The [`github.com/aws-controllers-k8s/code-generator`][codegen-repo] source code repository
contains the `ack-generate` CLI tool (`/cmd/ack-generate`), the Go packages
that are used in API inference and code generation (`/pkg/generate`,
`/pkg/model`) and Bash scripts to build an ACK service controller
(`/scripts/build-controller.sh`).

## `github.com/aws-controllers-k8s/test-infra`

The [`github.com/aws-controllers-k8s/test-infra`][testinfra-repo] source code repository
contains the `acktest` Python package for common ACK e2e test code, the CDK to
deploy our Prow CI/CD system and the scripts for running tests locally.

## `github.com/aws-controllers-k8s/$SERVICE-controller`

Each AWS API that has had a Kubernetes controller built to manage resources in
that API has its own source code repository in the
`github.com/aws-controllers-k8s` Github Organization. The source repos will be
called `$SERVICE-controller`, for example the ACK service controller for S3 is
located at [`github.com/aws-controllers-k8s/s3-controller`][s3-repo].

These service controller repositories contain Go code for the main controller
binary (`/cmd/controller/`), the public API types for the controllers
(`/apis`), the Go code for the resource managers used by the controller
(`/pkg/resource/*/`), static configuration manifests (`/config`), Helm
charts for the controller installation (`/helm`) along with a set of end-to-end
tests for the resources exposed by that controller (`/test/e2e`).

[issues]: https://github.com/aws-controllers-k8s/community/issues
[comm-repo]: https://github.com/aws-controllers-k8s/community/
[rt-repo]: https://github.com/aws-controllers-k8s/runtime/
[codegen-repo]: https://github.com/aws-controllers-k8s/code-generator/
[testinfra-repo]: https://github.com/aws-controllers-k8s/test-infra/
[s3-repo]: https://github.com/aws-controllers-k8s/s3-controller/
