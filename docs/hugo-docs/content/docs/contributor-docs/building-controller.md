---
title: "Building a Controller"
description: "How to build or regenerate an ACK service controller"
lead: "How to build or regenerate an ACK service controller"
draft: false
menu: 
  docs:
    parent: "contributor"
weight: 50
toc: true
---

## Prerequisites

You should have forked the `github.com/aws-controllers-k8s/code-generator`
repository and `git clone`'d it locally when [setting up](../setup) your
development environment,

With the prerequisites out of the way, let's move on to the first step:
building the code generator.

## Build code generator

Building an ACK service controller (or regenerating an existing one from a
newer API model file) requires the `ack-generate` binary, which is the main
code generator CLI tool.

To build the latest `ack-generate` binary, execute the following command from
the root directory of the `github.com/aws-controllers-k8s/code-generator`
source repository:

```
make build-ack-generate
```

!!! note "One-off build"
    You only have to do this once, overall. In other words: unless we change
    something upstream in terms of the code generation process, this is
    a one-off operation. Internally, the Makefile executes an `go build` here.

Don't worry if you forget this step, the script in the next step will complain
with a message along the line of `ERROR: Unable to find an ack-generate binary`
and will give you another opportunity to rectify the situation.

## Build an ACK service controller

Now that we have the basic code generation step done we will create the
respective ACK service controller and its supporting artifacts.

So first you have to select a service that you want to build and test.
You do that by setting the `SERVICE` environment variable. Let's say we want
to test the S3 service (creating an S3 bucket), so we would execute the
following:

```
export SERVICE=s3
```

Now we are in a position to generate the ACK service controller for the S3 API.

```
make build-controller SERVICE=$SERVICE
```

By default, running `make build-controller` will output the generated code to
ACK service controller for S3's source code repository (the
`$GOPATH/src/github.com/aws-controllers-k8s/s3-controller` directory). You can
override this behaviour with the `SERVICE_CONTROLLER_SOURCE_PATH` environment
variable.

!!! bug "Handle `controller-gen: command not found`"
    If you run into the `controller-gen: command not found` message when
    executing `make build-controller` then you want to check if the
    `controller-gen` binary is available in `$GOPATH/bin`, also ensure that `$GOPATH/bin` is part of your `$PATH`, see also
    [`#234`](https://github.com/aws/aws-controllers-k8s/issues/234).
    You can also install the required version of `controller-gen` using the
    `scripts/install-controller-gen.sh` helper script.

In addition to the ACK service controller code, above generates the
custom resource definition (CRD) manifests as well as the necessary RBAC
settings using the [`/scripts/build-controller.sh`][bc-script].

[bc-script]: https://github.com/aws-controllers-k8s/code-generator/blob/main/scripts/build-controller.sh

## Next Steps

Now that we have the generation part completed, we want to see if the
generated artifacts indeed are able to create an S3 bucket for us.

Learn about how to [run e2e tests for an ACK controller](../testing).
