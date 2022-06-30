---
title: "Testing"
description: "How to test an ACK service controller"
lead: ""
draft: false
menu: 
  docs:
    parent: "contributor"
weight: 60
toc: true
---

In the following, we will take you through the steps to run end-to-end (e2e)
tests for the ACK service controller for S3. You may use these steps to run e2e
tests for other ACK service controllers.

If you run into any problems when testing a service controller, please
[raise an issue](https://github.com/aws-controllers-k8s/community/issues/new/choose)
with the details so we can reproduce your issue.

## Prerequisites

For local development and testing we use "Kubernetes in Docker" (`kind`), 
which in turn requires Docker.

{{% hint type="warning" title="Footprint" %}}
When you run the `scripts/start.sh` script the first time,
the step that builds the container image for the target ACK service controller
can take up to 10 or more minutes. This is because the container image contains
a lot of dependencies. Once you successfully built the target image this base
image layer is cached locally, and the build takes a much shorter amount of
time. We are aware of this and aim to reduce both in the fullness of time.
{{% /hint %}}

In summary, in order to test ACK you will need to have the following tools
installed and configured:

1. [Golang 1.17+](https://golang.org/doc/install)
1. `make`
1. [Docker](https://docs.docker.com/get-docker/)
1. [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
1. [kubectl](https://kubernetes.io/docs/tasks/tools/)
1. [Helm](https://helm.sh/docs/intro/install/)
1. [kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/binaries/)
1. [jq](https://github.com/stedolan/jq/wiki/Installation)
1. [yq](https://mikefarah.gitbook.io/yq/#install)

To build and test an ACK controller with `kind`, **execute the commands as
described in the following from the root directory of the
`github.com/aws-controllers-k8s/test-infra` repository**.

You should have forked this repository and `git clone`'d it locally when
[setting up your development environment](../setup/).

{{% hint type="info" title="Recommended RAM" %}}
Given that our test setup creates the container images and then launches
a test cluster, we recommend that you have at least 4GB of RAM available
for the tests.
{{% /hint %}}

With the prerequisites out of the way, let's move on to running e2e tests for a
service controller.

## Run tests

Time to run the end-to-end test.

### Test configuration file setup

The e2e tests should be configured through a `test_config.yaml` file that lives
in the root of your `test-infra` directory. We have provided a
`test_config.example.yaml` file which contains the description for each
configuration option and its default value. Copy this configuration file and
customize it for your own needs:

```bash
cp test_config.example.yaml test_config.yaml
```

Take some time to look over each of the available options in the configuration
file and make changes to suit your preferences.

#### IAM Setup

In order for the ACK service controller to manage the S3 bucket, it needs an
identity. In other words, it needs an IAM role that represents the ACK service
controller towards the S3 service.

First, define the name of the IAM role that will have the permission to manage
S3 buckets on your behalf:

```bash
export ACK_TEST_IAM_ROLE=Admin-k8s
```

Now we need to verify the IAM principal (likely an IAM user) that is going to
assume the IAM role `ACK_TEST_IAM_ROLE`. So to get its ARN, execute:

```bash
export ACK_TEST_PRINCIPAL_ARN=$(aws sts get-caller-identity --query 'Arn' --output text)
```

You can verify if that worked using `echo $ACK_TEST_PRINCIPAL_ARN` and that should
print something along the lines of `arn:aws:iam::1234567890121:user/ausername`.

Next up, create the IAM role, adding the necessary trust relationship to the
role, using the following commands:

```bash
cat > trust-policy.json << EOF
{
	"Version": "2012-10-17",
	"Statement": {
		"Effect": "Allow",
		"Principal": {
			"AWS": "$ACK_TEST_PRINCIPAL_ARN"
		},
		"Action": "sts:AssumeRole"
	}
}
EOF
```

Using above trust policy, we can now create the IAM role:

```bash
aws iam create-role \
    --role-name $ACK_TEST_IAM_ROLE \
    --assume-role-policy-document file://trust-policy.json
```

Now we're in the position to give the IAM role `ACK_TEST_IAM_ROLE` the
permission to handle S3 buckets for us, using:

```bash
aws iam attach-role-policy \
    --role-name $ACK_TEST_IAM_ROLE \
    --policy-arn "arn:aws:iam::aws:policy/AmazonS3FullAccess"
```

{{% hint title="IAM policies for other services" %}}
If you are running tests on a service other than S3, you will need to find the
recommended policy ARN for the given service. The ARN is stored in
[`config/iam/recommended-policy-arn`][recc-arn] in each controller repository.

Some services don't have a single policy ARN to represent all of the permissions
required to run their controller. Instead you can find an [inline
policy][inline-policy] in the
[`config/iam/recommended-inline-policy`][recc-inline] in each applicable
controller repository. This can be applied to the role using [`aws iam
put-role-policy`][put-role-policy].
{{% /hint %}}

{{% hint title="Access delegation in IAM" %}}
If you're not that familiar with IAM access delegation, we recommend you
to peruse the [IAM documentation][iam-docs]
{{% /hint %}}

Next, in order for our test to generate [temporary credentials][temp-creds]
we need to tell it to use the IAM role we created in the previous step.
To generate the IAM role ARN and update your configuration file, do:

```bash
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query 'Account' --output text) && \
ACK_ROLE_ARN=arn:aws:iam::${AWS_ACCOUNT_ID}:role/${ACK_TEST_IAM_ROLE} \
yq -i '.aws.assumed_role_arn = env(ACK_ROLE_ARN)' test_config.yaml
```

{{% hint type="info" %}}
The tests uses the `generate_temp_creds` function from the
`scripts/lib/aws.sh` script, executing effectively 
`aws sts assume-role --role-session-arn $ACK_ROLE_ARN --role-session-name $TEMP_ROLE `
which fetches temporarily `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`,
and an `AWS_SESSION_TOKEN` used in turn to authentication the ACK
controller. The duration of the session token is 900 seconds (15 minutes).
{{% /hint %}}

Phew that was a lot to set up, but good news: you're almost there.

[iam-docs]: https://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_cross-account-with-roles.html
[inline-policy]: https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_managed-vs-inline.html#inline-policies
[put-role-policy]: https://docs.aws.amazon.com/cli/latest/reference/iam/put-role-policy.html
[recc-arn]: https://github.com/aws-controllers-k8s/s3-controller/tree/main/config/iam
[recc-inline]: https://github.com/aws-controllers-k8s/eks-controller/blob/main/config/iam/recommended-inline-policy
[temp-creds]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp.html

### Run end-to-end test

Before you proceed, make sure that you've done the configuration file setup in
the previous step.

Now we're finally in the position to execute the end-to-end test:

```bash
make kind-test SERVICE=$SERVICE
```

This provisions a Kubernetes cluster using `kind`, builds a container image with
the ACK service controller, and loads the container image into the `kind`
cluster.

It then installs the ACK service controller and related Kubernetes manifests
into the `kind` cluster using `kustomize build | kubectl apply -f -`.

First, it will attempt to install the Helm chart for the controller to ensure
the default values are safe and that the controller stands up properly.

Then, the above script builds a testing container, containing a Python
environment and the testing libraries we use, and runs the e2e tests for the
controller within that environment. These tests create, update and delete each
of the ACK resources and ensure their properties are properly mirrored in the
AWS service.

#### Repeat for other services

We have end-to-end tests for all services listed in the `DEVELOPER-PREVIEW`,
`BETA` and `GA` release statuses in our [service listing](../../community/services)
document. Simply replace your `SERVICE` environment variable with the name of a
supported service and re-run the IAM and test steps outlined above.

### Unit testing

We use [mockery](https://github.com/vektra/mockery) for unit testing.
You can install it by following the guideline on mockery's GitHub or simply
by running our handy script at `./scripts/install-mockery.sh` for general
Linux environments.

## Clean up

To clean up a `kind` cluster, including the container images and configuration 
files created by the script specifically for said test cluster, execute:

```bash
kind delete cluster --name $CLUSTER_NAME
```

If you want to delete all `kind` cluster running on your machine, use: 
```bash
kind delete clusters --all
```

With this the testing is completed. Thanks for your time and we appreciate your
feedback.
