## AWS Service Operator for Kubernetes

This repo contains the next generation AWS Service Operator for Kubernetes
(ASO).

A Kubernetes Operator is the combination of one or more [*customer resource
definitions*](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
(CRDs) and a
[*controller*](https://kubernetes.io/docs/reference/glossary/?fundamental=true#term-controller)
that reconciles the state of those CRDs.  ASO allows containerized applications
and Kubernetes users to create, update, delete and retrieve various AWS service
objects -- e.g. S3 buckets or RDS databases -- by submitting standard
Kubernetes manifests containing the ASO CRDs.

The [original AWS Service
Operator](https://github.com/awslabs/aws-service-operator) is no longer being
actively developed, and this repo will serve as the new home for an updated
Kubernetes Operator that exposes various AWS API objects as Kubernetes custom
resource definitions (CRDs).

We are in the early stages of planning the redesign of ASO. Check the [Issuesi
list](https://github.com/aws/aws-service-operator-k8s/issues) for descriptions
of work items. We invite any and all feedback and contributions, so please
don't hesitate to submit an issue, a pull request or comment on an existing
issue!

You can read about the redesign plans for ASO on the [AWS containers
roadmap](https://github.com/aws/containers-roadmap/issues/456).

## License

This project is licensed under the Apache-2.0 License.
