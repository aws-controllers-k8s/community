# AWS Controllers for Kubernetes (ACK)

The AWS Controllers for Kubernetes (ACK) will allow containerized applications and Kubernetes users to create, update, delete and retrieve the status of resources in AWS services such as S3 buckets, DynamoDB, RDS databases, SNS, etc. using the Kubernetes API, for example using Kubernetes manifests or kubectl plugins.

[ACK](https://github.com/aws/aws-controllers-k8s/) comprises a set of Kubernetes
custom [controllers](https://kubernetes.io/docs/reference/glossary/?fundamental=true#term-controller).
Each controller manages [custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
representing API resources of a single AWS service API. For example, the
service controller for AWS Simple Storage Service (S3) manages custom resources
that represent AWS S3 buckets, keys, etc.

Instead of logging into the AWS console or using the `aws` CLI tool to interact
with the AWS service API, Kubernetes users can install a controller for an AWS
service and then create, update, read and delete AWS resources using the Kubernetes
API.

This means they can use the Kubernetes API to fully describe both their
containerized applications, using Kubernetes resources like `Deployment` and
`Service`, as well as any AWS managed services upon which those applications
depend.

We are currently in the MVP phase, see also the [issue 22](https://github.com/aws/aws-controllers-k8s/issues/22) for details.

If you have feedback, questions, or suggestions please don't hesitate to submit an [issue](https://github.com/aws/aws-controllers-k8s/issues), a pull request or comment on an existing issue.
