# ACK service controllers

This directory contains the individual ACK service controllers, contained in
subdirectories named for the alias of the AWS service (e.g. `s3` or
`elasticache`).

The majority of the code in these subdirectories is **generated** using the
`ack-generate` CLI tool. Therefore, before making changes to any code or
configuration file in a particular service controller directory, please check
with the ACK contributors either on Kubernetes Slack community (#provider-aws
channel) or by submitting a Github Issue with your thoughts on what you'd like
to change about a particular service controller.

## Supported services

See the [documentation](https://aws.github.io/aws-controllers-k8s/services) for a list of supported services.

### Adding a service controller

See our [contributors guide](../CONTRIBUTING.md) for information on adding a new service controller.
