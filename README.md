[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/aws-controllers-k8s/community/issues)
![GitHub issues](https://img.shields.io/github/issues-raw/aws-controllers-k8s/community?style=flat)
![GitHub](https://img.shields.io/github/license/aws-controllers-k8s/community?style=flat)
![GitHub watchers](https://img.shields.io/github/watchers/aws-controllers-k8s/community?style=social)
![GitHub stars](https://img.shields.io/github/stars/aws-controllers-k8s/community?style=social)
![GitHub forks](https://img.shields.io/github/forks/aws-controllers-k8s/community?style=social)

# AWS Controllers for Kubernetes (ACK)

**AWS Controllers for Kubernetes (ACK)** lets you define and use AWS service
resources directly from Kubernetes. With ACK, you can take advantage of AWS
managed services for your Kubernetes applications without needing to define
resources outside of the cluster or run services that provide supporting
capabilities like databases or message queues within the cluster.

ACK is an open source project built with ❤️  by AWS. The project is composed of
many source code repositories containing a [common runtime][runtime-repo], a
[code generator][codegen-repo], [common testing tools][test-infra-repo] and
Kubernetes custom controllers for individual AWS service APIs.

[runtime-repo]: https://github.com/aws-controllers-k8s/runtime
[codegen-repo]: https://github.com/aws-controllers-k8s/code-generator
[test-infra-repo]: https://github.com/aws-controllers-k8s/test-infra

> **IMPORTANT** Please be sure to read our documentation about
> [release versioning and maintenance phases][releases] and note that ACK
> service controllers in the `Preview` maintenance phase are not recommended
> for production use. Use of ACK controllers in `Preview` maintenance phase is
> subject to the terms and conditions contained in the
> [AWS Service Terms][aws-service-terms], particularly the Beta Service
> Participation Service Terms, and apply to any service controllers in a
> `Preview` maintenance phase.

[releases]: https://aws-controllers-k8s.github.io/community/docs/community/releases/
[aws-service-terms]: https://aws.amazon.com/service-terms

* [Overview](#overview)
* [Getting Started](#getting-started)
* [Help & Feedback](#help--feedback)
* [Contributing](#contributing)
* [License](#license)

## Overview

Kubernetes applications often require a number of supporting resources like
databases, message queues, and object stores. AWS provides a set of managed
services that you can use to provide these resources for your apps, but
provisioning and integrating them with Kubernetes was complex and time
consuming. ACK lets you define and consume AWS services and resources directly
from a Kubernetes cluster. It gives you a unified way to manage your
application and its dependencies.

ACK is a collection of Kubernetes [custom resource definitions][crd] (CRDs) and
custom controllers working together to extend the Kubernetes API and manage AWS
resources on your behalf.

[crd]: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/

## Getting Started

Please see the list of ACK [service controllers][services] currently in one of
our [project stages][proj-stages].

[proj-stages]: https://aws-controllers-k8s.github.io/community/docs/community/releases/#project-stages

You can [install][install] any of the controllers in the `RELEASED` project stage using
Helm (recommended) or manually using the raw Kubernetes manifests contained in
the individual ACK service controller's source repository.

[services]: https://aws-controllers-k8s.github.io/community/docs/community/services/
[install]: https://aws-controllers-k8s.github.io/community/docs/user-docs/install/

Once installed, Kubernetes users may apply a custom resource (CR) corresponding
to one of the resources exposed by the ACK service controller for the service.

To view the list of custom resources and each CR's schema, visit our
[reference documentation][ref-docs].

[ref-docs]: https://aws-controllers-k8s.github.io/community/reference/

## Help & Feedback

For help, please consider the following venues (in order):

* [ACK project documentation](https://aws-controllers-k8s.github.io/community/)
* [Search open issues](https://github.com/aws-controllers-k8s/community/issues)
* [File an issue](https://github.com/aws-controllers-k8s/community/issues/new/choose)
* Chat with us on the `#aws-controllers-k8s` channel in the [Kubernetes Slack](https://kubernetes.slack.com/) community.

## Contributing

We welcome community contributions and pull requests.

See our [contribution guide](/CONTRIBUTING.md) for more information on how to
report issues, set up a development environment, and submit code.

We adhere to the [Amazon Open Source Code of Conduct][coc].

You can also learn more about our [Governance](/GOVERNANCE.md) structure.

[coc]: https://aws.github.io/code-of-conduct

## Community Meeting

ACK Community meeting is held every week.
Everyone is welcome to participate.

#### Details 
* **Agenda/Notes**: [link][meeting-notes]
  * Notes from each meeting are captured here.
* **When:** every Thursday at 9:00 AM [PST][pst-timezone]
* **Where:** [Zoom meeting][zoom-meeting-link]

[zoom-meeting-link]: https://zoom.us/j/95069096871?pwd=OXc3eWk1NVluUlozcVg3b1VtdGl5Zz09
[meeting-notes]: https://docs.google.com/document/d/1G9Nl-vBXuOBRoOt-9N-fQMpY05V8fCP8vPg94iTZ9gA
[utc-timezone]: https://dateful.com/convert/pst-pdt-pacific-time?t=9am

## License

This project is [licensed](/LICENSE) under the Apache-2.0 License.