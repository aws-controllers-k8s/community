# Testing

**NOTE**: Testing is tracked in the umbrella [Issue 6](https://github.com/aws/aws-controllers-k8s/issues/6).

For local development and/or testing we use [kind](https://kind.sigs.k8s.io/).

To build and test an ACK controller against a KinD cluster, execute the
following from the root directory of your checked-out source repository:

```
make build-ack-generate
# Replace with the service you want to build and test an ACK controller for...
SERVICE_TO_BUILD="ecr"
./scripts/build-controller.sh $SERVICE_TO_BUILD
./scripts/kind-build-test.sh -s $SERVICE_TO_BUILD
```

The above does the following:

* Builds the latest `ack-generate` binary
* Generates the ACK service controller for the AWS ECR API and output the
  generated code to the `services/$SERVICE_TO_BUILD` directory
* Generates the custom resource definition (CRD) manifests for resources
  managed by that ACK service controller
* Generates the Helm chart that can be used to install those CRD manifests and
  a Deployment manifest that runs the ACK service controller in a Pod on a
  Kubernetes cluster (still TODO)
* Provisions a KinD Kubernetes cluster
* Builds a Docker image containing the ACK service controller
* Loads the Docker image for the ACK service controller into the KinD cluster
* Installs the ACK service controller and related Kubernetes manifests into the
  KinD cluster using `helm install` (still TODO)
* Runs a series of Bash test scripts that call `kubectl` and the `aws` CLI
  tools to verify that custom resources (CRs) of the type managed by the ACK
  service controller are created, updated and deleted appropriately (still
  TODO)
