## AWS Controllers for Kubernetes (ACK)

This repo contains a set of Kubernetes
[controllers](https://kubernetes.io/docs/reference/glossary/?fundamental=true#term-controller)
that manage resources in AWS service APIs.

ACK allows containerized applications and Kubernetes users to create, update,
delete and retrieve the status of objects in AWS services such as S3 buckets,
DynamoDB, RDS databases, SNS, etc. using the Kubernetes API, for example using
Kubernetes manifests or `kubectl` plugins.

[Documentation](https://aws.github.io/aws-controllers-k8s/), including installation and usage instructions, are
available online.

[TODO]: # (link to generated documentation)

Check the [issues list](https://github.com/aws/aws-controllers-k8s/issues) for
descriptions of work items. We invite any and all feedback and contributions,
so please don't hesitate to submit an issue, a pull request or comment on an
existing issue.

For discussions, please use the `#provider-aws` channel on the [Kubernetes
Slack](https://kubernetes.slack.com) community.

Read about the [motivation for and background on](docs/background.md) ACK.

## License

This project is licensed under the Apache-2.0 License.
