---
title: "Authentication and Credentials"
description: "Understanding how AWS credentials are determined for an ACK controller"
lead: "Understanding AWS credentials in ACK"
draft: false
menu:
  docs:
    parent: "getting-started"
weight: 65
toc: true
---

When an ACK service controller communicates with an AWS service API, the
controller uses a set of [*AWS Credentials*][aws-creds].

This document explains the process by which these credentials are determined by
the ACK service controller and how ACK users can configure the ACK service
controller to use a particular set of credentials.

[aws-creds]: https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html

## Background

Each ACK service controller uses the [`aws-sdk-go`][aws-sdk-go] library to call
the AWS service APIs.

When initiating communication with an AWS service API, the ACK controller
[creates][rt-new-sess] a new `aws-sdk-go` `Session` object. This `Session`
object is automatically configured during construction by code in the
`aws-sdk-go` library that [looks for credential information][look-creds] in the
following places, *in this specific order*:

1. If the `AWS_PROFILE` environment variable is set, [find][prof-find] that
   specified profile in the configured [credentials file][creds-file] and use
   that profile's credentials.

2. If the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables
   are both set, these [values are used][env-find] by `aws-sdk-go` to set the
   AWS credentials.

3. If the `AWS_WEB_IDENTITY_TOKEN_FILE` environment variable is set,
   `aws-sdk-go` will load the credentials from the JSON web token (JWT) present
   in the file pointed to by this environment variable. Note that this
   environment variable is set to the value
   `/var/run/secrets/eks.amazonaws.com/serviceaccount/token` by the IAM Roles
   for Service Accounts (IRSA) pod identity webhook and the contents of this
   file are automatically rotated by the webhook with temporary credentials.

4. If there is a credentials file present at the location specified in the
   `AWS_SHARED_CREDENTIALS_FILE` environment variable (or
   `$HOME/.aws/credentials` if empty), `aws-sdk-go` will load the "default"
   profile present in the credentials file.

[rt-new-sess]: https://github.com/aws-controllers-k8s/runtime/blob/7abfd4e9bf9c835b76e06603617cae50c39af42e/pkg/runtime/session.go#L58
[look-creds]: https://github.com/aws/aws-sdk-go/blob/2c3daca245ce07c2e12beb7ccbf6ce4cf7a97c5a/aws/session/credentials.go#L19
[prof-find]: https://github.com/aws/aws-sdk-go/blob/2c3daca245ce07c2e12beb7ccbf6ce4cf7a97c5a/aws/session/credentials.go#L85
[creds-file]: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html
[env-find]: https://github.com/aws/aws-sdk-go/blob/2c3daca245ce07c2e12beb7ccbf6ce4cf7a97c5a/aws/credentials/env_provider.go#L41-L69
[aws-sdk-go]: https://github.com/aws/aws-sdk-go/

## Configuring credentials

There are multiple ways in which you can configure an ACK service controller to
use a particular set of AWS credentials:

* Web identity token file (recommended)
* Shared credentials file
* Access key and secret access key environment variables (not recommended)

{{% hint type="info" title="Understand the AWS credentials file format" %}}
It is important to understand the [AWS credentials file format](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html),
especially if you choose not to use the web identity token file method of
credential configuration.
{{% /hint %}}

### Use a web identity token file (recommended)

Our recommended approach for configuring the AWS credentials that an ACK
service controller will use to communicate with AWS services is to use the IAM
Roles for Service Accounts (IRSA) functionality provided by the IAM Pod
Identity Webhook and OIDC connector.

Using IRSA means that you only need to configure the IAM Role that is
associated with the Kubernetes Service Account of the ACK service controller's
Kubernetes Pod. The Pod Identity Webhook will be responsible for automatically
injecting *and periodically rotating* the web identity token file into your ACK
service controller's Pod.

Learn [how to configure IRSA][conf-irsa].

[conf-irsa]: ../irsa/

{{% hint type="info" title="Understand the AWS credentials file format" %}}
IRSA is enabled and installed on EKS clusters by default, however must be
manually configured if you are using a non-EKS cluster. See the IRSA
[self-hosted documentation][self-hosted] for information about installing the
pod identity webhook in non-EKS clusters.
[self-hosted]: https://github.com/aws/amazon-eks-pod-identity-webhook/blob/master/SELF_HOSTED_SETUP.md
{{% /hint %}}

### Use a shared credentials file

If you are not using IAM Roles for Service Accounts (IRSA) or are running in an
environment where IRSA isn't feasible (such as running KinD clusters within
Kubernetes Pods using Docker-in-Docker), you can choose to instruct the ACK
service controller to use AWS credentials found in a
[shared credentials file][creds-file].

When using a shared credentials file, the ACK service controller will need read
access to a known credentials file location.

If you do *not* set the `AWS_SHARED_CREDENTIALS_FILE` environment variable, the
controller will look for a readable file at `$HOME/.aws/credentials`.

Practically, this means that the `Deployment` spec you use to deploy the ACK
service controller should have a [volume mount][k8s-mount] that mounts a
readonly file containing the credentials file.

Let's assume you have stored your local AWS credentials file content in a
Kubernetes `Secret` named `aws-creds`:

```bash
CREDS_CONTENT=$(cat ~/.aws/credentials)
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: aws-creds
type: Opaque
stringData:
  credentials-file: |
  $CREDS_CONTENT
EOF
```

You would want to mount a readonly volume into the `Deployment` for your ACK
service controller. Here's how you might do this for a sample ACK controller:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ack-s3-controller
spec:
  replicas: 1
  template:
    spec:
      containers:
      - command:
        - ./bin/controller
        image: controller:latest
        name: controller
        ports:
          - name: http
            containerPort: 8080
        resources:
          limits:
            cpu: 100m
            memory: 300Mi
          requests:
            cpu: 100m
            memory: 200Mi
        volumeMounts:
        - name: aws-creds
          mountPath: "/root/.aws/credentials"
          readOnly: true
      volumes:
      - name: aws-creds
        secret:
          secretName: aws-creds
```

You can instruct the service controller to use a specific profile within the
shared credentials file by setting the `AWS_PROFILE` environment variable for
the `Pod`:

```yaml
      env:
        - name: AWS_PROFILE
          value: my-profile
```

[k8s-mount]: https://kubernetes.io/docs/concepts/storage/volumes/
[cm-volume]: https://kubernetes.io/docs/concepts/storage/volumes/#configmap
[cm]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/

### Use access key and secret access key environment variables (not recommended)

Finally, you can choose to manually set the `AWS_ACCESS_KEY_ID`,
`AWS_SECRET_ACCESS_KEY` and optionally the `AWS_SESSION_TOKEN` environment
variables on the ACK service controller's `Pod`:

```bash
kubectl -n ack-system set env deployment/ack-s3-controller \
    AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
    AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
    AWS_SESSION_TOKEN="$AWS_SESSION_TOKEN"
```
