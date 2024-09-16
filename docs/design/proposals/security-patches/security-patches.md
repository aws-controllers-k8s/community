# Security Patches and CVEs in ACK

## Background

The AWS Controllers for Kubernetes (ACK) project plays a role in bridging the
gap between Kubernetes clusters and various AWS services, offering users a way
to manage and integrate their AWS resources within Kubernetes environments.
However, like any software project, ACK is not immune to security vulnerabilities,
particularly those identified with CVE scanners such as [quay][quay] and 
[trivy][trivy] reports. These vulnerabilities may arise from outdated
dependencies, including libraries, base images, and compilers, leaving the
delivered controller for our users vulnerable to security risks.

## Problem statement

The core of our current issue lies in the fact that, until recently, ACK's
images and components relied on outdated fundemental elements. For example
a significant number of reported CVEs pointed out the usage of an outdated
`eks-distro-build-minimal` image and an older Go compiler (version `1.19`)
during the compilation process. This situation persisted because ACK
contributors had to manually update these dependencies. Given the rapid pace
at which these dependencies were updated, it was easy to forget about them.
This presented a substantial challenge to the project's security stance.

To address these vulnerabilities and improve the security of ACK controllers,
we had to take some actions: The first of these was upgrading the
`eks-distro-build-minimal` image and the Go compiler to version `1.21`
(the latest one as of september 2023), the second was re-releasing
new controller versions using those up-to-date depdendecies. While this
solved an important portionn of the security issues, it did not entirely
mitigate the risk, as it remained necessary to proactively keep an eye the
newest dependencies and ship the necessary patches.

## Purpose of this document

Right now, the project faces the challenge of automating the process of updating
the controllers "in-image-runtime" and dependencies to address security
vulnerabilities without disrupting the already established release chain. This
document explores the existing procedures for releasing new controllers, the
current approach to handling CVEs, and proposes solutions to enhance those
processes.

In a nutchel, the objective is to ensure that the ACK project can proactively
manage security patches, minimize CVE reports in its images, and keep its 
dependencies up-to-date. Achieving this goal requires a careful consideration
of the project's release workflows, security update procedures, and
the intersection of these two aspects. The solutions proposed here aim to strike
a balance between automation and control, facilitating the project's ability to
rapidly respond to security vulnerabilities while maintaining a stable and
reliable release pace.

## Scope

- Development and implementation of strategies for automating the detection
  and mitigation of security vulnerabilities.
- Integration of security patching processes into the existing ACK release
  workflows and procedures
- Updates and changes to the project's code and infrastructure necessary to
  automate security patching.
- Solutions that ensure the proactive management of security patches while
  maintaining the reliability of the project's release process. (needs rewording)

## Out of scope

- In-depth discussions or analysis of specific CVEs or individual security
  vulnerabilities.
- Completly reworking the release chain of ACK controllers containers and
  helm charts.
- Non-security-related updates or enhancements to the ACK project.

## Description of the current release chain

[//]: <> (move this section to test-infra repository)

## Current release process

In ACK, The release process is carefully structured and includes steps
such as detecting code changes, testing, approvals, git tagging, prow jobs,
and documentation updates. The process ensures that customers have access
to the latest controller functionality and improvements while maintaining
compatibility with the Kubernetes environments

Keep in mind that ACK releases involve two primary deliverable that are shipped
for users:
- controller container image.
- Helm chart release.

The release process follows a structured workflow that can be described
as followed:
- ACK maintainers and contributors make changes to a specific controller.
- a release pull request (PR) is created, which includes version updates
  in the Helm chart. (incrementing from the previous release, e.g., `0.0.6`)
- E2E and unit tests have to pass in the raised PR ^
- One of the maintainers reviews and approves the PR, typically by adding
  an `/lgtm`` comment.
- Prow (tide component) merge to the main branch, incorporating the changes.
- A Prow job detects the changes and  tags the repository with the next
  patch release version, such as `0.0.6`
- A GitHub action automatically creates a GitHub release, detailing the
  changes since the last release. (add example here)
- Multiple Prow jobs are triggered to release various artifacts:
  - A container image is tagged with the new release version, such as `0.0.6` (*1)
  - A Helm chart is tagged with the new release version (`0.0.6`) (out of scope)
  - Documentation updates are made in the community repository. (out of scope)
  - A pull request is raised for integration with Operator Lifecycle Manager
    (OLM) to ship artifacts to the operator hub platforms. (out of scope)

[//]: <> (NOTES(a-hilaly): insert diagrams)

(*1): The container image used in releases depends on a base image defined
in the project's code, linked to the [code-generator][code-generator]'s
Dockerfile and the environment variables and images in [test-infra][test-infra]
Prow jobs.

### CVE reports

When a CVE is reported for a specific image, ACK maintainers face a series of
tasks. The primary actions revolve around updating the project's components
to mitigate the identified security risk. This typically involves five main
steps:

- Base image version bumping for the main [Dockerfile][code-generator-dockerfile]
  used to build the controllers images. 
- Go compiler version bumping in the environment variables of the prowjobs
  responsible of releasing container images and helm charts.
- ACK runtime depdencies bumping, those generally made by dependabot by ACK
  maintainers can also raise PR to address similar issues ([example PR][runtime-deps-bump])
- re-release the [code-generator][code-generator] to open PR bumping the
  versions and depedencies for all the controllers ([example prowjob][code-generator-autogen])
- Maintainers monitor the tests for each the controller repository and merges
  and merges the PRs as soon as they pass. 

[//]: <> (NOTES(a-hilaly): insert diagram)
  
## Solutions

The goal of this section is to find a balance between enhancing security and
maintaining the efficiency and reliabiliity of the current release process.
The objective is to proactively address security vulnerabilities, particularly
CVEs, while minimizing disruptions to the already-functioning release chain.

### Solution 1: Periodic scanning and CVE detection (prefered)

To address these challenges while ensuring timely security updates, we propose
implementing a solution focused on detecting CVEs, newly released base images
and Go releases periodically. Here's how it would work:

- **Automated Detection**: implement automated scripts periodically scan for
  CVE reports related to ACK's dependencies, track updates to base images,
  and monitor Go releases. These scripts would run at regular intervals,
  ensuring that security vulnerabilities and dependency changes are quickly
  identified.

- **Pull Request Generation**: when a security vulnerability, new base image
 or Go release is detected, an automated process opens pull requests against
 the concerned repositories. 

- **Notification to Maintainers**: Parallely, notifications are sent to
  project maintainers, providing them with essential information about the
  detected vulnerabilities, dependency updates, and the a link to the generated
  PRs. Maintainers would have a clear overview of the necessary actions required.

- **Maintainer Actions**: maintainers review the generated PRs and take
  appropriate actions. This may involve temporarily reverting unfinished
  features, making necessary code adjustments, or completing unfinished
  features before merging the PRs.

#### Technical details:

[//]: <> (NOTES(a-hilaly): develop each section with technical details on the approach and technologies that will be leveraged)

### Solution 2: Tweaking the release process, by introducing branch based releases

[//]: <> (NOTES(a-hilaly): add more solutions if needed)

[//]: <> (NOTES(a-hilaly): add more solutions if needed)

[quay]: https://github.com/quay/clair
[trivy]: https://github.com/aquasecurity/trivy
[code-generator]: https://github.com/aws-controllers-k8s/code-generator
[code-generator-dockerfile]: https://github.com/aws-controllers-k8s/code-generator/blob/main/Dockerfile
[code-generator-autogen]: https://prow.ack.aws.dev/view/s3/ack-prow-logs/logs/auto-generate-controllers/1702457878050246656
[runtime]: https://github.com/aws-controllers-k8s/runtime
[runtime-deps-bump]: https://github.com/aws-controllers-k8s/runtime/pull/125
[test-infra]: https://github.com/aws-controllers-k8s/test-infra
[test-infra-release-prowjob]: https://github.com/aws-controllers-k8s/test-infra/blob/main/prow/jobs/jinja/postsubmits/controller_release.jinja2#L2-L33