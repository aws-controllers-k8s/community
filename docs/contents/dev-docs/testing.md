# Testing

Testing is tracked in the umbrella [Issue 6](https://github.com/aws/aws-controllers-k8s/issues/6).

The current plan is to modify the [e2e integration test setup](https://github.com/aws/amazon-vpc-cni-k8s/blob/bc04604397889430f0a3d5f6e4766b399c1d5fcc/scripts/run-integration-tests.sh) of the AWS VPC CNI plugin repository, which uses [aws-k8s-tester](https://github.com/aws/aws-k8s-tester) for the initial setup of the Kubernetes cluster.

For local development and/or testing we use [kind](https://kind.sigs.k8s.io/),
integration tests run against an EKS cluster.
