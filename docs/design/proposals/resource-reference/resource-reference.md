## Introduction

This proposal discusses adding the support for referencing an ACK custom resource
in another custom resource. The reconciler will wait for the referenced custom
resource to be created, and in a stable state, before proceeding to create the
resource which references it.

For example, currently if ACK customer needs to create an APIGatewayv2 integration,
they first need to
1. create an `API` custom resource
2. wait for `API` resource creation to finish
3. manually find the `APIID` from `Status` of `API` resource
4. create an `Integration` resource with the `APIID`

Having to manually wait for the resource creation prevents declarative configuration
techniques that rely on describing all desired resources at the same time using
one or more Kubernetes YAML manifests. But with approach of referenced resources,
ACK controller will wait for `API` resource to get created and then use the API
reference to find `APIID` and create Integration resource.

The deletion strategy for the referenced resource is out of scope for this proposal
right now. For Ex: If API resource is referenced in integration resource, the deletion
of API resource will not be blocked. We plan to handle deletion strategy in future.

## Scope

### First Revision
* Allow adding resource name(s) from same api-group as reference in the custom
resource definition
* Only the resources from same namespace will be allowed to use as reference
* Only the fields of type '*string' or '[]*string' can be referenced from
another resource. This satisfies the use case of referencing 'ID(s)/Name(s)/ARN(s)'
from other resources

### Future Improvements
* Add support for custom hooks while dereferencing resources
* Add support for `selector` for referencing the objects
* Add support for referencing resource from different api-groups
* Add support for not deleting a resource if it is referenced in another resource
* Add support for finite retries and terminal conditions when resolving references

## Preferred Approach

### High Level Overview

* Introduce new construct in generator.yaml which will indicate to code-generator
if a new spec field with type `AWSResourceReference` should be added to the custom
resource definition.

* This new spec field will work as additional input for an existing spec field.
Addition of the extra spec field keeps the backward compatibility with existing
custom resource definitions.

* Validation will be present that only one of existing spec field OR new reference
field is present in custom resource, not both.

* If the existing spec field was marked as required, validation will be added that
at least one of existing spec field or new reference field is present.

* If `AWSResourceReference` type field is present in the desired state, first ACK
controller will query for the referenced resource. If the referenced resource
has `ACK.ResourceSynced` condition status `True` & specified field is found
in the referenced resource, the reconciliation loop will progress. Otherwise
the request will get requeued for future processing.

* Watchers for the referenced resource type will not be implemented in the initial
implementation. The initial implementation will mostly be used for finding
identifier of referenced resource and any update in referenced resource should not
change it's identifier. As mentioned earlier, deletion will also be handled in
future.

* The code to determine when the referenced resource is synced will be generated,
and additional hooks will be added for customization. Ex: Do not just check presence
of a field but also check if the Status is available etc...

### Low Level Implementation

#### generator.yaml Changes

```go
type FieldConfig struct {
    ...
    Reference *ReferenceConfig `json:"references,omitempty"`
    ...
}

// ReferenceConfig contains the instructions for how to add the referenced resource
// configuration for a field.
type ReferenceConfig struct {
	// Kind mentions the K8s Kind of referenced resource
	Kind string `json:"kind"`
	// FieldPath refers to the the path of field which should be copied
	// from referenced resource into TargetFieldName
	FieldPath string `json:"field_path"`
	// TargetFieldName is spec field name which gets value from
	// referenced resource's FieldPath
	TargetFieldName string `json:"target_field"`
	// IsList mentions whether the new field should be AWSResourceReference or
	// List of AWSResourceReference
	IsList          bool   `json:"is_list"`
	// Required mentions whether either of AWSResourceReference or
	// TargetFieldName must be present in custom resource
	Required        bool   `json:"required"`
}
```

`APIGatewayv2 generator.yaml` Example:
```yaml
resources:
  Integration:
    fields:
      Api:
        references:
          kind: API
          field_path: Status.APIID
          target_field: APIID
          is_list: false
          required: true
      ApiId:
        is_required: false
```
NOTE: `ApiId` field is marked as not required because the new reference field
`Api` can also provide APIId.

#### New AWSResourceReference Type

```go
type AWSResourceReference struct {
	Name *string `json:"name,omitempty"`
}
```

and the new manifests will look like following:

a) Same as earlier (keeping backwards compatibility)

```yaml
apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
kind: Integration
metadata:
  name: my-integration
spec:
  apiID: my-api-id
  integrationType: HTTP_PROXY
  integrationURI: https://example.org
  integrationMethod: GET
  payloadFormatVersion: "1.0"
```

b) Using the API resource reference
```yaml
apiVersion: apigatewayv2.services.k8s.aws/v1alpha1
kind: Integration
metadata:
  name: my-integration
spec:
  api:
    name: my-api
  integrationType: HTTP_PROXY
  integrationURI: https://example.org
  integrationMethod: GET
  payloadFormatVersion: "1.0"
```

#### API Generation Changes

During CRD generation, we will inspect if any reference field needs to be added in
CRD spec.

```go
// model.go
// GetCRDs returns a slice of `CRD` structs that describe the
// top-level resources discovered by the code generator for an AWS service API
func (m *Model) GetCRDs() ([]*CRD, error) {
	if m.crds != nil {
		return m.crds, nil
	}
	crds := []*CRD{}

	opMap := m.SDKAPI.GetOperationMap(m.cfg)

	createOps := (*opMap)[OpTypeCreate]
	readOneOps := (*opMap)[OpTypeGet]
	readManyOps := (*opMap)[OpTypeList]
	updateOps := (*opMap)[OpTypeUpdate]
	deleteOps := (*opMap)[OpTypeDelete]
	getAttributesOps := (*opMap)[OpTypeGetAttributes]
	setAttributesOps := (*opMap)[OpTypeSetAttributes]

	for crdName, createOp := range createOps {
		if m.cfg.IsIgnoredResource(crdName) {
			continue
		}
		crdNames := names.New(crdName)
...
		// Now any additional Spec fields that are required from other API
		// operations.
		for targetFieldName, fieldConfig := range m.cfg.ResourceFields(crdName) {
			if fieldConfig.IsReadOnly {
				// It's a Status field...
				continue
			}

+			if fieldConfig.Reference != nil {
+				// Add the spec field here
+				referenceFieldNames := names.New(targetFieldName)
+				rf := NewReferenceField(crd, referenceFieldNames, fieldConfig)
+				crd.SpecFields[referenceFieldNames.Original] = rf
+				crd.Fields[referenceFieldNames.Camel] = rf
+			}

...
	m.crds = crds
	return crds, nil
}
```

```go
//field.go
func NewReferenceField(
	crd *CRD,
	fieldNames names.Names,
	cfg *ackgenconfig.FieldConfig,
) *Field {
	gt := "*ackv1alpha1.AWSResourceReference"
	gtp := "*ackv1alpha1.AWSResourceReference"
	if cfg.Reference.IsList {
		gt = "[]" + gt
		gtp = "[]" + gtp
	}
	return &Field{
		CRD:               crd,
		Names:             fieldNames,
		Path:              fieldNames.Original,
		ShapeRef:          nil,
		GoType:            gt,
		GoTypeElem:        "AWSResourceReference",
		GoTypeWithPkgName: gtp,
		FieldConfig:       cfg,
	}
}
```

#### Runtime Changes

```go
// aws_resource_manager.go

type AWSResourceManager interface {
        ...
+       // ResolveReferences finds if there are any AWSResourceReference fields
+       // present in the AWSResource passed in the parameters and attempts to resolve
+       // those reference fields into target field.
+       // It returns an AWSResource with resolved references, boolean representing
+       // whether any reference fields were present and an error if the passed
+       // AWSResource's reference fields cannot be resolved
+       ResolveReferences(context.Context, client.Reader, AWSResource) (AWSResource, bool, error)
}
```

```go
// reconciler.go

// reconcile either cleans up a deleted resource or ensures that the supplied
// AWSResource's backing API resource matches the supplied desired state.
//
// It returns a copy of the resource that represents the latest observed state.
func (r *resourceReconciler) reconcile(
ctx context.Context,
rm acktypes.AWSResourceManager,
res acktypes.AWSResource,
) (acktypes.AWSResource, error) {
    if res.IsBeingDeleted() {
+       // Resolve references before deleting the resource.
+       // Ignore any errors while resolving the references
+    	  res, _, _ = rm.ResolveReferences(ctx, r.apiReader, res)
        return r.deleteResource(ctx, rm, res)
    }
    return r.Sync(ctx, rm, res)
}

// Sync ensures that the supplied AWSResource's backing API resource
// matches the supplied desired state.
//
// It returns a copy of the resource that represents the latest observed state.
func (r *resourceReconciler) Sync(
	ctx context.Context,
	rm acktypes.AWSResourceManager,
	desired acktypes.AWSResource,
) (acktypes.AWSResource, error) {
	var err error
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("r.Sync")
	defer exit(err)

	var latest acktypes.AWSResource // the newly created or mutated resource

	isAdopted := IsAdopted(desired)
	rlog.WithValues("is_adopted", isAdopted)
+	rlog.Enter("rm.ResolveReferences")
+	resolvedRefDesired, containsResourceRef, err := rm.ResolveReferences(ctx, r.apiReader, desired)
+	rlog.Exit("rm.ResolveReferences", err)
+	if err != nil {
+		// Set the condition in copy of desired resources so that it can be
+		//patched back in etcd
+		desiredCopy := desired.DeepCopy()
+		if containsResourceRef {
+			condition.SetReferenceResolved(desiredCopy, corev1.ConditionFalse,
+				&condition.ReferenceNotResolvedMessage,
+				&condition.ReferenceNotResolvedReason)
+		}
+		return desiredCopy, err
+	}

+	// If the desired resource had references, set the ReferenceResolved
+	//condition to true
+	if containsResourceRef {
+		condition.SetReferenceResolved(resolvedRefDesired, corev1.ConditionTrue,
+			&condition.ReferenceResolvedMessage,
+			&condition.ReferenceResolvedReason)
+	} else {
+		condition.RemoveReferenceResolved(resolvedRefDesired)
+	}
+	desired = resolvedRefDesired

	rlog.Enter("rm.ReadOne")
	latest, err = rm.ReadOne(ctx, desired)
	rlog.Exit("rm.ReadOne", err)
	if err != nil {
		if err != ackerr.NotFound {
			return latest, err
		}
		if isAdopted {
			return nil, ackerr.AdoptedResourceNotFound
		}
		if latest, err = r.createResource(ctx, rm, desired); err != nil {
			return latest, err
		}
	} else {
		if latest, err = r.updateResource(ctx, rm, desired, latest); err != nil {
			return latest, err
		}
	}
	// Attempt to late initialize the resource. If there are no fields to
	// late initialize, this operation will be a no-op.
	if latest, err = r.lateInitializeResource(ctx, rm, latest); err != nil {
		return latest, err
	}
	return r.handleRequeues(ctx, latest)
}
```

#### Service Controller Generated Code
**generator.yaml**

NOTE: The VpcLink configuration below is just for generating sample code for
list for references. Using APIID as SecurityGroupId is not a legitimate example.
This is due to the restriction that in initial implementation we are only allowing
referencing the resource from same api-group.
```yaml
resources:
  Integration:
    fields:
      Api:
        references:
          kind: API
          field_path: Status.APIID
          target_field: APIID
          is_list: false
          required: true
      ApiId:
        is_required: false
  VpcLink:
    fields:
      SecurityGroups:
        references:
          kind: API
          field_path: Status.APIID
          target_field: SecurityGroupIDs
          is_list: true
          required: false
```

**integration/manager.go**
```go
package integration

import (
...

+	acksvcv1alpha1 "github.com/aws-controllers-k8s/apigatewayv2-controller/apis/v1alpha1"
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
...
	svcsdk "github.com/aws/aws-sdk-go/service/apigatewayv2"
	svcsdkapi "github.com/aws/aws-sdk-go/service/apigatewayv2/apigatewayv2iface"
+	"k8s.io/apimachinery/pkg/types"
+	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups=apigatewayv2.services.k8s.aws,resources=integrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apigatewayv2.services.k8s.aws,resources=integrations/status,verbs=get;update;patch

var lateInitializeFieldNames = []string{}

+var (
+	_ = types.NamespacedName{}
+	_ = acksvcv1alpha1.Integration{}
+)

...

// ResolveReferences finds if there are any AWSResourceReference fields present
// in the AWSResource passed in parameter and attempts to resolve those reference
// fields into target field.
// It returns a resolved AWSResource, boolean representing whether any reference
// fields were present and an error if the passed AWSResource's reference fields
// cannot be resolved
func (rm *resourceManager) ResolveReferences(
	ctx context.Context,
	apiReader client.Reader,
	res acktypes.AWSResource,
) (acktypes.AWSResource, bool, error) {
	ko := rm.concreteResource(res).ko.DeepCopy()
	referencePresent := false
	if ko.Spec.API != nil && ko.Spec.APIID != nil {
		return &resource{ko}, true, fmt.Errorf("'APIID' field should not be present when using reference field 'API'")
	}
	if ko.Spec.API == nil && ko.Spec.APIID == nil {
		return &resource{ko}, referencePresent, fmt.Errorf("At least one of 'APIID' or 'API' field should be present")
	}
	// Checking Referenced Field API
	if ko.Spec.API != nil {
		referencePresent = true
		arr := ko.Spec.API
		namespacedName := types.NamespacedName{Namespace: res.MetaObject().GetNamespace(), Name: *arr.Name}
		obj := acksvcv1alpha1.API{}
		err := apiReader.Get(ctx, namespacedName, &obj)
		if err != nil {
			return &resource{ko}, true, err
		}
		var refResourceSynced bool
		for _, cond := range obj.Status.Conditions {
			if cond.Type == ackv1alpha1.ConditionTypeResourceSynced && cond.Status == corev1.ConditionTrue {
				refResourceSynced = true
				break
			}
		}
		if !refResourceSynced {
			return &resource{ko}, true, fmt.Errorf("referenced 'API' resource " + *arr.Name + " does not have 'ACK.ResourceSynced' condition status 'True'")
		}
		if obj.Status.APIID == nil {
			return &resource{ko}, true, fmt.Errorf("'Status.APIID' is not yet present for referenced 'API' resource " + *arr.Name)
		}
		ko.Spec.APIID = obj.Status.APIID
	}
	return &resource{ko}, referencePresent, nil
}
```

**vpc_link/manager.go**
```go
package vpc_link

import (
...

+	acksvcv1alpha1 "github.com/aws-controllers-k8s/apigatewayv2-controller/apis/v1alpha1"
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
...
	svcsdk "github.com/aws/aws-sdk-go/service/apigatewayv2"
	svcsdkapi "github.com/aws/aws-sdk-go/service/apigatewayv2/apigatewayv2iface"
+	"k8s.io/apimachinery/pkg/types"
+	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups=apigatewayv2.services.k8s.aws,resources=integrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apigatewayv2.services.k8s.aws,resources=integrations/status,verbs=get;update;patch

var lateInitializeFieldNames = []string{}

+var (
+	_ = types.NamespacedName{}
+	_ = acksvcv1alpha1.Integration{}
+)

...

// ResolveReferences finds if there are any AWSResourceReference fields present
// in the AWSResource passed in parameter and attempts to resolve those reference
// fields into target field.
// It returns a resolved AWSResource, boolean representing whether any reference
// fields were present and an error if the passed AWSResource's reference fields
// cannot be resolved
func (rm *resourceManager) ResolveReferences(
	ctx context.Context,
	apiReader client.Reader,
	res acktypes.AWSResource,
) (acktypes.AWSResource, bool, error) {
	ko := rm.concreteResource(res).ko.DeepCopy()
	referencePresent := false
	if ko.Spec.SecurityGroups != nil && ko.Spec.SecurityGroupIDs != nil {
		return &resource{ko}, true, fmt.Errorf("'SecurityGroupIDs' field should not be present when using reference field 'SecurityGroups'")
	}
	// Checking Referenced Field SecurityGroups
	if ko.Spec.SecurityGroups != nil && len(ko.Spec.SecurityGroups) > 0 {
		referencePresent = true
		resolvedReferences := []*string{}
		for _, arr := range ko.Spec.SecurityGroups {
			namespacedName := types.NamespacedName{Namespace: res.MetaObject().GetNamespace(), Name: *arr.Name}
			obj := acksvcv1alpha1.API{}
			err := apiReader.Get(ctx, namespacedName, &obj)
			if err != nil {
				return &resource{ko}, true, err
			}
			var refResourceSynced bool
			for _, cond := range obj.Status.Conditions {
				if cond.Type == ackv1alpha1.ConditionTypeResourceSynced && cond.Status == corev1.ConditionTrue {
					refResourceSynced = true
					break
				}
			}
			if !refResourceSynced {
				return &resource{ko}, true, fmt.Errorf("referenced 'API' resource " + *arr.Name + " does not have 'ACK.ResourceSynced' condition status 'True'")
			}
			if obj.Status.APIID == nil {
				return &resource{ko}, true, fmt.Errorf("'Status.APIID' is not yet present for referenced 'API' resource " + *arr.Name)
			}
			resolvedReferences = append(resolvedReferences, obj.Status.APIID)
		}
		ko.Spec.SecurityGroupIDs = resolvedReferences
	}
	return &resource{ko}, referencePresent, nil
}
```

## Alternate Approach
* Change the type of member from `*string` to `*AWSResourceReference` and
  `[]*string` to `[]*AWSResourceReference`
* `AWSResourceReference` will either have the value for member field(string) or
  contains the reference for another resource.
* This approach was not chosen because it breaks the backward compatibility for
existing custom resource definitions by changing type of existing field.
* Even with the approach of custom `UnmarshallJson`, old manifests fail during
kubernetes crd validation.
