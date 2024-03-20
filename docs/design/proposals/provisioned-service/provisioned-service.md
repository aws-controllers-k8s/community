# Service Binding Specification for Kubernetes

## Problem Statement

Application developers require specific values to connect their
applications to various services. In order to facilitate this connection
problem, the [Service Binding Specification for
Kubernetes](https://servicebinding.io) defines Provisioned
Services as a standard way to expose connection information from backing
services. Supporting Provisioned Services within AWS Controllers for
Kubernetes makes application connectivity to ACK-backed services easy
and seamless.

## Existing Solutions

AWS Controllers for Kubernetes have support for a [FieldExport
API](https://github.com/aws-controllers-k8s/community/blob/main/docs/design/proposals/native-binding/native-binding.md).
This can be used to create a Secret that can be used for [Direct Secret
Reference as per the
specification](https://github.com/servicebinding/spec#direct-secret-reference).
This solution needs more effort from the application developers.

The FieldExport resources required for application developers are the
same, but every developer needs to repeat this same configuration for
every application they deploy. The proposed solution avoids this
repetition through Provisioned Services.

## Proposed Solution

The proposed solution is to make various ACK resources become
Provisioned Services. For example, make the DBInstance resource provided
by the RDS controller a Provisioned Service. Each controller would use
their main custom resource as a Provisioned Service resource. For each
Provisioned Service, identify the fields that need to be exported to
create the Secret resource.

The application developer experience is explained in this demo video:
[https://www.youtube.com/watch?v=AXXWv7N12JM](https://www.youtube.com/watch?v=AXXWv7N12JM)

(This is created as part of a [Proof of
Concept](https://github.com/aws-controllers-k8s/community/issues/1289))

The structure change for DBInstance:

```
// ServiceBindingSecretReference defines a mirror of corev1.LocalObjectReference
type ServiceBindingSecretReference struct {
	// Name of the referent secret.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name"`
}

// DBInstanceStatus defines the observed state of DBInstance
type DBInstanceStatus struct {
	// Binding exposes the Secret for Service Binding to conform the Provisioned Service
	// as per the Service Binding Specification for Kubernetes.
	// Ref. https://github.com/servicebinding/spec#provisioned-service
	Binding *ServiceBindingSecretReference `json:"binding,omitempty"`
```

### Benefits

When there are many microservices connecting to the same service, the
label selector feature helps to easily bind all of them with minimal
configuration. Service Binding Specification standard provides a uniform
experience for all the applications connectivity with services.

### Implementation

Since the API changes required for various custom resources are
standard, and they can be generated for all controllers. However, the
changes for each of these controllers can be implemented gradually.

### RBAC

Since a Secret resource needs to be created, updated, and deleted, the
ClusterRole requires "*create",* "*update", and "delete"* permissions.
Currently, this permission is not set in the existing controllers.

## RDS Controller Fields

The RDS controller supports multiple databases. Based on the database,
the value for type could be set. For example, postgresql for PostgreSQL
and mysql for MySQL.

Here is a table with required fields for PostgreSQL and MySQL:

|**Field**  | **Value**                |  **Remarks**
|-----------|--------------------------|-------------------------------
|type       |  postgresql/mysql        |   Based on Spec.Engine
|provider   |  aws                     |   
|host       |  Status.Endpoint.Address |   
|port       |  Status.Endpoint.Port    |   
|username   |  Spec.MasterUsername     |   
|password   |                          |   Based on Spec.MasterUserPassword
|database   |  Spec.Engine?            |   

Some of the fields will take more time to set the value. For example,
the address of an RDS service will only get set once the service has
become ready. The controller can set these fields dynamically.

## Conclusion

Provisioned Service support for ACK would make application connectivity
with services easy and seamless. When there are many microservices
connecting to the same service, the label selector feature helps to
easily bind all of them with minimal configuration. The application
developers receive a consistent experience across all services.
