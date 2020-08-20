[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/aws/aws-controllers-k8s/issues)
![status is preview](https://img.shields.io/badge/status-preview-brightgreen?style=flat)
![GitHub issues](https://img.shields.io/github/issues-raw/aws/aws-controllers-k8s?style=flat)
![GitHub](https://img.shields.io/github/license/aws/aws-controllers-k8s?style=flat)


![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/aws/aws-controllers-k8s)
[![Go Report Card](https://goreportcard.com/badge/github.com/aws/aws-controllers-k8s)](https://goreportcard.com/report/github.com/aws/aws-controllers-k8s)
![GitHub watchers](https://img.shields.io/github/watchers/aws/aws-controllers-k8s?style=social)
![GitHub stars](https://img.shields.io/github/stars/aws/aws-controllers-k8s?style=social)
![GitHub forks](https://img.shields.io/github/forks/aws/aws-controllers-k8s?style=social)



# AWS Controllers for Kubernetes (ACK)
**AWS Controllers for Kubernetes (ACK)** lets you define and use AWS service resources directly from Kubernetes. With ACK, you can take advantage of AWS managed services for your Kubernetes applications without needing to define resources outside of the cluster or run services that provide supporting capabilities like databases or message queues within the cluster.

This is a new open source project built with ❤️ by AWS and available as a **Developer Preview**. We encourage you to [try it out](https://aws.github.io/aws-controllers-k8s/dev-docs/testing/), provide feedback and contribute to development.

> **IMPORTANT** Because this project is in developer preview, you may see breaking changes throughout. We encourage you to experiment with ACK but DO NOT adopt it for production use just yet*

* [Overview](#overview)
* [Getting Started](#getting-started)
* [Help & Feedback](#help--feedback)
* [Contributing](#contributing)
* [License](#license)

## Overview

Kubernetes applications often require a number of supporting resources like databases, message queues, and object stores. AWS provides a set of managed services that you can use to provide these resources for your apps, but provisioning and integrating them with Kubernetes was complex and time consuming. ACK lets you define and consume AWS services and resources directly from a Kubernetes cluster. It gives you a unified way to manage your application and its dependencies.

ACK is a collection of Kubernetes [custom resource definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) (CRDs) and custom controllers working together to extend the Kubernetes API and manage AWS resources on your behalf.

## Getting Started

Until we've graduated ACK [service controllers](https://aws.github.io/aws-controllers-k8s/services/) we ask you to [test-drive](https://aws.github.io/aws-controllers-k8s/dev-docs/testing/) them.

## Help & Feedback
For help, please consider the following venues (in order):

* [ACK project documentation](https://aws.github.io/aws-controllers-k8s/)
* [Search open issues](https://github.com/aws/aws-controllers-k8s/issues)
* [File an issue](https://github.com/aws/aws-controllers-k8s/issues/new/choose)
* Join our [mailing list](https://groups.google.com/forum/#!forum/aws-service-operator-user/).
* Chat with us on the `#provider-aws` channel in the [Kubernetes Slack](https://kubernetes.slack.com/) community.

## Contributing
We welcome community contributions and pull requests. See our [contribution guide](/CONTRIBUTING.md) for more information on how to report issues, set up a development environment, and submit code.

Check the [issues list](https://github.com/aws/aws-controllers-k8s/issues) for descriptions of work items. We invite any and all feedback and contributions, so please don't hesitate to submit an issue, a pull request or comment on an existing issue.

ACK adheres to the [Amazon Open Source Code of Conduct](https://aws.github.io/code-of-conduct). You can also learn more about our [Governance](/GOVERNANCE.md) structure.

## License
This project is licensed under the Apache-2.0 License.
