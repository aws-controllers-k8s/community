## AWS Service Operator for Kubernetes

This repo contains the next generation AWS Service Operator for Kubernetes
(ASO).

An operator in Kubernetes is the combination of one or more [customer resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) and one or more [controllers](https://kubernetes.io/docs/reference/glossary/?fundamental=true#term-controller) managing said custom resources.

The ASO will allow containerized applications and Kubernetes users to create, update, delete and retrieve the status of AWS service such as S3 buckets, DynamoDB, RDS databases, SNS, etc. using the Kubernetes API, for example using 
Kubernetes manifests or `kubectl` plugins.

## Status

As of end of 2019 we are in the early stages of planning the redesign of ASO. Check the [issues list](https://github.com/aws/aws-service-operator-k8s/issues) for descriptions of work items. We invite any and all feedback and contributions, so please don't hesitate to submit an issue, a pull request or comment on an existing issue.

## Background

Read about the [motivation for and background on](docs/background.md) the ASO.

## License

This project is licensed under the Apache-2.0 License.
