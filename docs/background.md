# Background

The AWS Service Operator (ASO) project was introduced by [Chris Hein in 10/2018](https://aws.amazon.com/blogs/opensource/aws-service-operator-kubernetes-available/).
We reviewed the feedback from the wider community and stakeholders and [decided in 08/2019](https://github.com/aws/containers-roadmap/issues/456) to turn ASO into a first-tier OSS project with concrete commitments from the service team side, based on the following tenets:

- It is a community-driven project, based on a governance model defining roles and responsibilities.
- It is optimized for production usage with a full test coverage including performance and scalability test suites.
- It strives to be the only codebase exposing AWS services via a Kubernetes operator. 

Since then, we worked on [design issues](https://github.com/aws/aws-service-operator-k8s/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3Adesign) and gathering feedback around which services to prioritize.


## Custom controllers and operators in AWS

- [Sagemaker operator](https://github.com/aws/amazon-sagemaker-operator-for-k8s), allowing to use Sagemaker from Kubernetes 
- [App Mesh controller](https://github.com/aws/aws-app-mesh-controller-for-k8s), managing App Mesh resources from Kubernetes
- [EKS Pod Identity Webhook](https://github.com/aws/amazon-eks-pod-identity-webhook), providing IAM roles for service accounts functionality

## Related projects

- [Crossplane](https://crossplane.io/docs/v0.5/services/aws-services-guide.html)
- [aws-s3-provisioner](https://github.com/yard-turkey/aws-s3-provisioner)