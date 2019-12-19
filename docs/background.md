# Background

The AWS Service Operator (ASO) project was introduced by [Chris Hein in 10/2018](https://aws.amazon.com/blogs/opensource/aws-service-operator-kubernetes-available/).
We reviewed the feedback from the wider community and stakeholders and [decided in 08/2019](https://github.com/aws/containers-roadmap/issues/456) to turn ASO into a first-tier OSS project with concrete commitments from the service team side, based on the following tenets:

1. It is a community-driven project, based on a governance model defining roles and responsibilities.
2. It is optimized for production usage with full test coverage including performance and scalability test suites.
3. It strives to be the only codebase exposing AWS services via a Kubernetes operator. 

Since then, we worked on [design issues](https://github.com/aws/aws-service-operator-k8s/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3Adesign) and gathering feedback around which services to prioritize.


## Custom controllers and operators in AWS

AWS service teams use custom controllers, webhooks, and operators for different use cases and based on different approaches. Examples include:

- [Sagemaker operator](https://github.com/aws/amazon-sagemaker-operator-for-k8s), allowing to use Sagemaker from Kubernetes 
- [App Mesh controller](https://github.com/aws/aws-app-mesh-controller-for-k8s), managing App Mesh resources from Kubernetes
- [EKS Pod Identity Webhook](https://github.com/aws/amazon-eks-pod-identity-webhook), providing IAM roles for service accounts functionality

While the autonomy in the different teams and project allows for rapid iterations and innovations, there are some drawbacks associated with it:

- The UX differs and that can lead to frustration when adopting an offering.
- A consistent quality bar across the different offerings is hard to establish and to verify.
- It's wasteful to re-invent the plumbing and necessary infrastructure (testing, etc.).

Above is the motivation for our 3rd tenet: we want to make sure that there is a common framework, implementing good practices as put forward, for example, in the [Operator Developer Guide](https://operators.gitbook.io/operator-developer-guide-for-red-hat-partners/) or in the [Programming Kubernetes](https://programming-kubernetes.info/) book.

## Related projects

Outside of AWS, there are projects that share similar goals we have with the ASO, for example:

- [Crossplane](https://crossplane.io/docs/v0.5/services/aws-services-guide.html)
- [aws-s3-provisioner](https://github.com/yard-turkey/aws-s3-provisioner)