# AWS Controllers for Kubernetes (ACK)

**AWS Controllers for Kubernetes (ACK)** lets you define and use AWS service resources directly from Kubernetes. With ACK, you can take advantage of AWS managed services for your Kubernetes applications without needing to define resources outside of the cluster or run services that provide supporting capabilities like databases or message queues within the cluster.

This is a new open source project built with ❤️ by AWS and available as a **Developer Preview**. We encourage you to try it out, provide feedback and contribute to development.

**IMPORTANT *Because this project is a preview, there may be significant and breaking changes introduced in the future. We encourage you to try and experiment with this project. Please do not adopt it for production use.**

### Contents

* Overview
* Getting Started
* Supported Services
* Help & Feedback
* Contributing
* License

## Overview
Kubernetes applications often require a number of supporting resources like databases, message queues, and object stores to operate. AWS provides a set of managed services that you can use to provide these resources for your applications, but provisioning and integrating them with Kubernetes was complex and time consuming. ACK lets you define and consume many AWS services and resources directly within a Kubernetes cluster. ACK gives you a unified, operationally seamless way to manage your application and its dependencies.

ACK is a collection of [Kubernetes Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) (CRDs) and controllers which work together to extend the Kubernetes API and create AWS resources on your cluster’s behalf.

Read more on the [motivation and background](docs/background.md) for the this project.

## Getting Started

To get started, [choose and install](https://github.com/aws/aws-service-operator-k8s/blob/mvp/docs/contents/user-docs/install.md) the controllers for the AWS services you want to manage. Then, [configure permissions](https://github.com/aws/aws-service-operator-k8s/blob/mvp/docs/contents/user-docs/permissions.md) and [define your first AWS resource](https://github.com/aws/aws-service-operator-k8s/blob/mvp/docs/contents/user-docs/usage.md).

See a full walk through in our _documentation_.

## Help & Feedback
For help, please consider the following venues (in order):

* [ACK project documentation](https://aws.github.io/aws-controllers-k8s/)
* [AWS service documentation]
* [Search open issues]
* File a new issue]
* Mailing list: [ACK](https://groups.google.com/forum/#!forum/aws-service-operator-user/).
* Slack: Talk to us on the #provider-aws channel on the Kubernetes Slack (https://kubernetes.slack.com/) community.

## Contributing
ACK adheres to the [Amazon Open Source Code of Conduct](https://aws.github.io/code-of-conduct).

We welcome community contributions and pull requests. See our [contribution guide](/CONTRIBUTING.md) for more information on how to report issues, set up a development environment and submit code.

Check the [issues list](https://github.com/aws/aws-controllers-k8s/issues) for
descriptions of work items. We invite any and all feedback and contributions,
so please don't hesitate to submit an issue, a pull request or comment on an
existing issue.

You can also learn more about our [Governance](/GOVERNANCE.md) structure.

## License
This project is licensed under the Apache-2.0 License.
