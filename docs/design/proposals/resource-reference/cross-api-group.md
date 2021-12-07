## Introduction

This proposal builds on top of initial [resource-reference proposal](resource-reference.md).
This document proposes the approach to reference resource(s) from another ACK service controller
i.e. different api-group.

For Example, in the [initial proposal](resource-reference.md) referencing an `API`
resource inside `Integration` resource is possible because both resources are part
of `apigatewayv2-controller`(same apigroup `apigatewayv2.services.k8s.aws`) But
referencing `EC2's SecurityGroup and Subnet` resources was not possible inside
`apigatewayv2's VPCLink` resource.

This proposal aims to remove this blocker.

## Scope
* Referring a resource across namespace will **NOT** be supported.
* Only resources from other ACK service controller will be supported as
references. Referencing K8s native resources like `configMap` will **NOT** be supported
in this implementation.
* Only referencing fields of type `string`(main usecase being `identifiers`) will be
supported. Referencing structs from another resource will **NOT** be supported in this
implementation.
* `go.mod` changes will **NOT** be auto-generated, service teams will need to add another ACK
service controller's dependency in `go.mod` file manually for now.
* code-generator will generate new `CRDs` including the reference fields
* code-generator will update the `helm` and `kustomize` artifacts to include `Read and List
RBAC permissions` for resource referenced from another service controller
* code-generator will generate the source code for resolving the references for
resources from another service controller and set `Status.Condition` correctly
in case there are failure during reference resolution.
* It will be customer's responsibility to install controller/crd for the resource
being referenced. For example, even if apigatewayv2's `VPCLink` resource can reference
`SecurityGroup` resource from `ec2-controller`, `apigatewayv2-controller` installation will
**NOT** install `ec2-controller` or `SecurityGroup` CRD.
* Documentation changes will be made to help customers on steps to reference resources from
another service controller

## High Level Overview
### Declaring The References (AWS Service Team)
* For generating a corresponding reference field, service teams will add a `references`
entry inside `generator.yaml`.
If the `references` field refers to a resource from another ACK service
controller, service teams will provide name of that service in `references` field.
If no service name is provided, it defaults to AWS service name of the controller
where `generator.yaml` exists. See [how to add resource reference](#generatoryaml-example)

### Generating Service Controller (AWS Service Team)
* Apart from updating `generator.yaml` file as mentioned [here](#generatoryaml-example),
service teams will also need to update `go.mod` file to include ACK service controller
repository whose resource will be referenced.
For example: To able to resolve `SecurityGroup` resource from `ec2-controller`, `go.mod`
file in `apigatewayv2-controller` will need to add `github.com/aws-controllers-k8s/ec2-controller`
as dependency
* Apart from updating `go.mod` file, there are no changes in code-generation steps. Service teams can
execute `make build-controller` from `code-generator` repo to generate service controller source
code.

### Running Service Controller (ACK Customer)
* There will be no changes in installation steps if customers do not use the Reference
fields which refer to resource from another service controller.
* If customer wants to refer resource from another service controller, it will be
customer's responsibility to install that controller in same namespace.
* Another section will be added in ACK documentation on `how to refer resources across
service controllers` to guide the customers.

## Low Level Implementation

### Reference Config
* A new field `ServiceName` will be introduced to find the AWS service name of controller
where the resource will be referred from.
* This will be an optional field. When missing, the ServiceName will default to AWS service
name of controller where `generator.yaml` exists.
* This field is used to generate the `APIGroup` for reading the referred resource.
```go
type ReferencesConfig struct {
+ // ServiceName mentions the AWS service name where referenced 'Resource' exists.
+ // ServiceName is the package name for AWS service package in
+ // aws-sdk-go/models/apis/<package_name>
+ // When not specified, 'ServiceName' defaults to service name of controller
+ // which contains generator.yaml
+ ServiceName string `json:"service_name,omitempty"`
  // Resource mentions the K8s resource which is read to resolve the
  // reference
  Resource string `json:"resource"`
  // Path refers to the the path of field which should be copied
  // to resolve the reference
  Path string `json:"path"`
}
```

### Code Generator
* During template execution, if a `CRD` has references from another AWS service,
`code-generator` will import api-types from those service.
* `code-generator` will also add `get` and `list` rbac permissions for
`apigroup/resource` from another AWS service in `references.go` file. For
example, following permissions are added to `vpc_link/reference.go` since
`VPCLink` references `Subnet` and `SecurityGroup` resources from `ec2-controller`
```
// +kubebuilder:rbac:groups=ec2.services.k8s.aws,resources=securitygroups,verbs=get;list
// +kubebuilder:rbac:groups=ec2.services.k8s.aws,resources=securitygroups/status,verbs=get;list

// +kubebuilder:rbac:groups=ec2.services.k8s.aws,resources=subnets,verbs=get;list
// +kubebuilder:rbac:groups=ec2.services.k8s.aws,resources=subnets/status,verbs=get;list
```
* See sample code implementation for code-generator [here](#code-generator-changes)

### NOTE:
* There will be no changes needed in ACK runtime to support cross controller resource references.
* See the sample generated code for apigatewayv2-controller's `vpc_link/references.go` file
[here](#generated-code-for-apigatewayv2-controller)

### Controller Installation
* ACK Documentation will be updated to guide customers on how to install multiple controller to
use cross controller resource reference functionality
* If customers do not install CRDs/controller of the referenced resource, `Status.Condition` property
on the resource will indicate the error message guiding them towards mitigating the issue.


## Sample Code Changes

### generator.yaml example
```yaml
resources:
  Integration:
    fields:
      ApiId:
        references:
          resource: API
          path: Status.APIID
  VpcLink:
    fields:
+     SecurityGroupIds:
+       references:
+         resource: SecurityGroup
+         path: Status.ID
+         service_name: ec2
+     SubnetIds:
+       references:
+         resource: Subnet
+         path: Status.SubnetID
+         service_name: ec2
```

### code-generator changes
###### New Methods In `model/field.go`
```go
// ReferencedServiceName returns the serviceName for the referenced resource
// when the field has a corresponding reference field.
// If the field does not have corresponding reference field, empty string is
// returned
func (f *Field) ReferencedServiceName() (referencedServiceName string) {
	if f.FieldConfig != nil && f.FieldConfig.References != nil {
		if f.FieldConfig.References.ServiceName != "" {
			return f.FieldConfig.References.ServiceName
		} else {
			return f.CRD.sdkAPI.API.PackageName()
		}
	}
	return referencedServiceName
}

// ReferencedResourceNamePlural returns the plural of referenced resource
// when the field has a corresponding reference field.
// If the field does not have corresponding reference field, empty string is
// returned
func (f *Field) ReferencedResourceNamePlural() string {
	var referencedResourceName string
	pluralize := pluralize.NewClient()
	if f.FieldConfig != nil && f.FieldConfig.References != nil {
		referencedResourceName = f.FieldConfig.References.Resource
	}
	if referencedResourceName != "" {
		return pluralize.Plural(referencedResourceName)
	}
	return referencedResourceName
}
```

###### Changes In `templates/pkg/resource/references.go.tpl`
```go
import (
	"context"
{{ if .CRD.HasReferenceFields -}}
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
{{ end -}}
	"sigs.k8s.io/controller-runtime/pkg/client"

{{ if .CRD.HasReferenceFields -}}
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
{{ end -}}
	acktypes "github.com/aws-controllers-k8s/runtime/pkg/types"
+	svcapitypes "github.com/aws-controllers-k8s/{{ .ServicePackageName }}-controller/apis/{{ .APIVersion }}"
+{{ $servicePackageName := .ServicePackageName -}}
+{{ if .CRD.HasReferenceFields -}}
+{{ range $fieldName, $field := .CRD.Fields -}}
+{{ if and $field.HasReference (not (eq $field.ReferencedServiceName $servicePackageName)) -}}
+    {{ $field.ReferencedServiceName }}apitypes "github.com/aws-controllers-k8s/{{ $field.ReferencedServiceName }}-controller/apis/{{ .APIVersion }}"
+{{ end -}}
+{{ end -}}
+{{ end -}}
)

+{{ if .CRD.HasReferenceFields -}}
+{{ range $fieldName, $field := .CRD.Fields -}}
+{{ if and $field.HasReference (not (eq $field.ReferencedServiceName $servicePackageName)) -}}
+// +kubebuilder:rbac:groups={{ $field.ReferencedServiceName -}}.services.k8s.aws,resources={{ ToLower $field.ReferencedResourceNamePlural }},verbs=get;list
+// +kubebuilder:rbac:groups={{ $field.ReferencedServiceName -}}.services.k8s.aws,resources={{ ToLower $field.ReferencedResourceNamePlural }}/status,verbs=get;list
+
+{{ end -}}
+{{ end -}}
+{{ end -}}
```

### Generated Code for `apigatewayv2-controller`

###### go.mod
```
module github.com/aws-controllers-k8s/apigatewayv2-controller

go 1.14

require (
+	github.com/aws-controllers-k8s/ec2-controller v0.0.1
	github.com/aws-controllers-k8s/runtime v0.15.2
...
)

```

###### generator.yaml
```yaml
resources:
  ...
  Integration:
    fields:
      ApiId:
        references:
          resource: API
          path: Status.APIID
+ VpcLink:
+   fields:
+     SecurityGroupIds:
+       references:
+         resource: SecurityGroup
+         path: Status.ID
+         service_name: ec2
+     SubnetIds:
+       references:
+         resource: Subnet
+         path: Status.SubnetID
+         service_name: ec2
...

```

###### Additions In `helm/templates/cluster-role-controller.yaml`
```yaml
...
- apiGroups:
    - ec2.services.k8s.aws
  resources:
    - securitygroups
  verbs:
    - get
    - list
- apiGroups:
    - ec2.services.k8s.aws
  resources:
    - securitygroups/status
  verbs:
    - get
    - list
- apiGroups:
    - ec2.services.k8s.aws
  resources:
    - subnets
  verbs:
    - get
    - list
- apiGroups:
    - ec2.services.k8s.aws
  resources:
    - subnets/status
  verbs:
    - get
    - list
- apiGroups:
    - services.k8s.aws
  resources:
    - adoptedresources
...
```

###### pkg/resource/vpc_link/references.go
```go
// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Code generated by ack-generate. DO NOT EDIT.

package vpc_link

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/aws-controllers-k8s/apigatewayv2-controller/apis/v1alpha1"
	ec2apitypes "github.com/aws-controllers-k8s/ec2-controller/apis/v1alpha1"
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	acktypes "github.com/aws-controllers-k8s/runtime/pkg/types"
)

// +kubebuilder:rbac:groups=ec2.services.k8s.aws,resources=securitygroups,verbs=get;list
// +kubebuilder:rbac:groups=ec2.services.k8s.aws,resources=securitygroups/status,verbs=get;list

// +kubebuilder:rbac:groups=ec2.services.k8s.aws,resources=subnets,verbs=get;list
// +kubebuilder:rbac:groups=ec2.services.k8s.aws,resources=subnets/status,verbs=get;list

// ResolveReferences finds if there are any Reference field(s) present
// inside AWSResource passed in the parameter and attempts to resolve
// those reference field(s) into target field(s).
// It returns an AWSResource with resolved reference(s), and an error if the
// passed AWSResource's reference field(s) cannot be resolved.
// This method also adds/updates the ConditionTypeReferencesResolved for the
// AWSResource.
func (rm *resourceManager) ResolveReferences(
	ctx context.Context,
	apiReader client.Reader,
	res acktypes.AWSResource,
) (acktypes.AWSResource, error) {
	namespace := res.MetaObject().GetNamespace()
	ko := rm.concreteResource(res).ko.DeepCopy()
	err := validateReferenceFields(ko)
	if err == nil {
		err = resolveReferenceForSecurityGroupIDs(ctx, apiReader, namespace, ko)
	}
	if err == nil {
		err = resolveReferenceForSubnetIDs(ctx, apiReader, namespace, ko)
	}
	if hasNonNilReferences(ko) {
		return ackcondition.WithReferencesResolvedCondition(&resource{ko}, err)
	}
	return &resource{ko}, err
}

// validateReferenceFields validates the reference field and corresponding
// identifier field.
func validateReferenceFields(ko *svcapitypes.VPCLink) error {
	if ko.Spec.SecurityGroupIDsRef != nil && ko.Spec.SecurityGroupIDs != nil {
		return ackerr.ResourceReferenceAndIDNotSupportedFor("SecurityGroupIDs", "SecurityGroupIDsRef")
	}
	if ko.Spec.SubnetIDsRef != nil && ko.Spec.SubnetIDs != nil {
		return ackerr.ResourceReferenceAndIDNotSupportedFor("SubnetIDs", "SubnetIDsRef")
	}
	if ko.Spec.SubnetIDsRef == nil && ko.Spec.SubnetIDs == nil {
		return ackerr.ResourceReferenceOrIDRequiredFor("SubnetIDs", "SubnetIDsRef")
	}
	return nil
}

// hasNonNilReferences returns true if resource contains a reference to another
// resource
func hasNonNilReferences(ko *svcapitypes.VPCLink) bool {
	return false || ko.Spec.SecurityGroupIDsRef != nil || ko.Spec.SubnetIDsRef != nil
}

// resolveReferenceForSecurityGroupIDs reads the resource referenced
// from SecurityGroupIDsRef field and sets the SecurityGroupIDs
// from referenced resource
func resolveReferenceForSecurityGroupIDs(
	ctx context.Context,
	apiReader client.Reader,
	namespace string,
	ko *svcapitypes.VPCLink,
) error {
	if ko.Spec.SecurityGroupIDsRef != nil &&
		len(ko.Spec.SecurityGroupIDsRef) > 0 {
		resolvedReferences := []*string{}
		for _, arrw := range ko.Spec.SecurityGroupIDsRef {
			arr := arrw.From
			if arr == nil || arr.Name == nil || *arr.Name == "" {
				return fmt.Errorf("provided resource reference is nil or empty")
			}
			namespacedName := types.NamespacedName{
				Namespace: namespace,
				Name:      *arr.Name,
			}
			obj := ec2apitypes.SecurityGroup{}
			err := apiReader.Get(ctx, namespacedName, &obj)
			if err != nil {
				return err
			}
			var refResourceSynced, refResourceTerminal bool
			for _, cond := range obj.Status.Conditions {
				if cond.Type == ackv1alpha1.ConditionTypeResourceSynced &&
					cond.Status == corev1.ConditionTrue {
					refResourceSynced = true
				}
				if cond.Type == ackv1alpha1.ConditionTypeTerminal &&
					cond.Status == corev1.ConditionTrue {
					refResourceTerminal = true
				}
			}
			if refResourceTerminal {
				return ackerr.ResourceReferenceTerminalFor(
					"SecurityGroup",
					namespace, *arr.Name)
			}
			if !refResourceSynced {
				//TODO(vijtrip2) Uncomment below return statment once
				// ConditionTypeResourceSynced(True/False) is set for all resources
				//return ackerr.ResourceReferenceNotSyncedFor(
				//	"SecurityGroup",
				//	namespace, *arr.Name)
			}
			if obj.Status.ID == nil {
				return ackerr.ResourceReferenceMissingTargetFieldFor(
					"SecurityGroup",
					namespace, *arr.Name,
					"Status.ID")
			}
			resolvedReferences = append(resolvedReferences,
				obj.Status.ID)
		}
		ko.Spec.SecurityGroupIDs = resolvedReferences
	}
	return nil
}

// resolveReferenceForSubnetIDs reads the resource referenced
// from SubnetIDsRef field and sets the SubnetIDs
// from referenced resource
func resolveReferenceForSubnetIDs(
	ctx context.Context,
	apiReader client.Reader,
	namespace string,
	ko *svcapitypes.VPCLink,
) error {
	if ko.Spec.SubnetIDsRef != nil &&
		len(ko.Spec.SubnetIDsRef) > 0 {
		resolvedReferences := []*string{}
		for _, arrw := range ko.Spec.SubnetIDsRef {
			arr := arrw.From
			if arr == nil || arr.Name == nil || *arr.Name == "" {
				return fmt.Errorf("provided resource reference is nil or empty")
			}
			namespacedName := types.NamespacedName{
				Namespace: namespace,
				Name:      *arr.Name,
			}
			obj := ec2apitypes.Subnet{}
			err := apiReader.Get(ctx, namespacedName, &obj)
			if err != nil {
				return err
			}
			var refResourceSynced, refResourceTerminal bool
			for _, cond := range obj.Status.Conditions {
				if cond.Type == ackv1alpha1.ConditionTypeResourceSynced &&
					cond.Status == corev1.ConditionTrue {
					refResourceSynced = true
				}
				if cond.Type == ackv1alpha1.ConditionTypeTerminal &&
					cond.Status == corev1.ConditionTrue {
					refResourceTerminal = true
				}
			}
			if refResourceTerminal {
				return ackerr.ResourceReferenceTerminalFor(
					"Subnet",
					namespace, *arr.Name)
			}
			if !refResourceSynced {
				//TODO(vijtrip2) Uncomment below return statment once
				// ConditionTypeResourceSynced(True/False) is set for all resources
				//return ackerr.ResourceReferenceNotSyncedFor(
				//	"Subnet",
				//	namespace, *arr.Name)
			}
			if obj.Status.SubnetID == nil {
				return ackerr.ResourceReferenceMissingTargetFieldFor(
					"Subnet",
					namespace, *arr.Name,
					"Status.SubnetID")
			}
			resolvedReferences = append(resolvedReferences,
				obj.Status.SubnetID)
		}
		ko.Spec.SubnetIDs = resolvedReferences
	}
	return nil
}
```

