# Background

In 10/2018 Chris Hein [introduced](https://aws.amazon.com/blogs/opensource/aws-service-operator-kubernetes-available/) the AWS Service Operator (ASO) project. We reviewed the feedback from the community and stakeholders and in 08/2019 [decided](https://github.com/aws/containers-roadmap/issues/456) to relaunch ASO as a first-tier open source project with concrete commitments from the container service team. In this process, we renamed the project to AWS Controllers for Kubernetes (ACK).

The tenets for the relaunch were:

1. ACK is a community-driven project, based on a governance model defining roles and responsibilities.
2. ACK is optimized for production usage with full test coverage including performance and scalability test suites.
3. ACK strives to be the only codebase exposing AWS services via a Kubernetes operator. 

Since then, we worked on [design issues](https://github.com/aws/aws-controllers-k8s/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3Adesign) and gathering feedback around which services to prioritize.


## Existing custom controllers

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