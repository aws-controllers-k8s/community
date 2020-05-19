# `pkg/runtime`

This package contains a set of concrete structs and helper functions that
provide a common controller *implementation* for an AWS service.

The top-level container struct in the package is `ServiceController`. A single
instance of `ServiceController` is created (using `NewServiceController()`)
from the `cmd/controller/main.go` file that contains the ACK controller
entrypoint for a specific AWS service API.

`ServiceController` primarily serves as a way to glue the upstream
`sigs.k8s.io/controller-runtime` (here on called `ctrlrt` for short since that
alias we use in the ACK codebase to refer to that upstream repository)
machinery together with ACK types that handle communication with the AWS
service API.

The main `ctrlrt` types that `ServiceController` glues together
are the `ctrlrt.Manager` and `ctrlrt.Reconciler` types. The `ctrlrt.Manager`
type is used to bind a bunch of `sigs.k8s.io/client-go` and
`sigs.k8s.io/apimachinery` infrastructure together into a common network
server/listener structure. The `ctrlrt.Reconciler` type is an interface that
provides a single `Reconcile()` method whose job is to reconcile the state of a
single custom resource (CR) object.

The `ServiceController.BindControllerManager()` method accepts a
`ctrlrt.Manager` object and is responsible for creating a reconciler for each
kind of CR that the service controller will handle.

But how does the `ServiceController` know what kinds of CRs that it will
handle?

There is a `ServiceController.WithResourceManagerFactories()` method
that sets the `ServiceController`'s collection of objects that implement the
`types.AWSResourceManagerFactory` interface.

These resource manager factories *produce* objects that implement the
`types.AWSResourceManager` interface, which is basic CRUD+L operations for a
particular AWS resource against the backend AWS service API. The
`types.AWSResourceManagerFactory.For()` method returns a
`types.AWSResourceManager` object that has been created to handle a specific
AWS service API resources for one AWS account. In this way, the single service
controller can manage resources across multiple AWS accounts.

Resource manager factories are *registered* with a `Registry` object that is
package-scoped to the individual service controller's
`services/{service}/pkg/resource` package. See the example service's
[`pkg/resource/registry.go`](../services/example/pkg/resource/registry.go) file
for how this package-scoped registry works. Individual resource manager
factories are registered with this package-scoped `Registry` object using an
`init()` call within a file named `{resource}_manager_factory.go`. For example,
the `Book` resource in the example service's `pkg/resource` package has its
resource manager factory registered in the `init()` function in the
[`pkg/resource/book_resource_manager_factory.go`](../services/example/pkg/resource/book_resource_manager_factory.go)
file.
