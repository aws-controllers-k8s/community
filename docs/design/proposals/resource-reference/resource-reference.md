## Introduction

This proposal discusses adding the support for referencing an ACK custom resource
in another custom resource. The reconciler will wait for the referenced custom
resource to be created, and in a stable state, before proceeding to create the
resource which references it.

For example, currently if ACK customer needs to create an APIGatewayv2 `Integration`,
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
* Add support for references in custom fields that are not present in
* Add support for referencing resource from different api-groups
* Add support for not deleting a resource if it is referenced in another resource
* Add support for finite retries and terminal conditions when resolving references

## Preferred Approach

### High Level Overview

* Introduce new construct in generator.yaml which will indicate to code-generator
if a new spec field with type `*AWSResourceReferenceWrapper` or
`[]*AWSResourceReferenceWrapper` should be added to the custom resource definition.

* This new spec field will work as additional input for an existing spec field.
Addition of the extra spec field keeps the backward compatibility with existing
custom resource definitions.

* Validation will be present that only one of existing spec field OR new reference
field is present in custom resource, not both.

* If the existing spec field was marked as required, validation will be added that
at least one of existing spec field or new reference field is present.

* If `*AWSResourceReferenceWrapper` type field is present in the desired state,
first ACK controller will query for the referenced resource. If the referenced resource
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
    References *ReferencesConfig `json:"references,omitempty"`
    ...
}

// ReferencesConfig contains the instructions for how to add the referenced resource
// configuration for a field.
type ReferencesConfig struct {
	// Resource mentions the K8s resource which is read to resolve the
	// reference
	Resource string `json:"resource"`
	// Path refers to the the path of field which should be copied
	// to resolve the reference
	Path string `json:"path"`
}
```

`APIGatewayv2 generator.yaml` Example:
```yaml
resources:
  Integration:
    fields:
      ApiId:
        references:
          resource: API
          path: Status.APIID
```

#### New AWSResourceReference Types

```go
// AWSResourceReferenceWrapper provides a wrapper around *AWSResourceReference
// type to provide more user friendly syntax for references using 'from' field
// Ex:
// APIIDRef:
//   from:
//     name: my-api
type AWSResourceReferenceWrapper struct {
	From *AWSResourceReference `json:"from,omitempty"`
}

// AWSResourceReference provides ways to either provide an AWSResource identifier
// (Id/ARN/Name) by string value or refer to another k8s resource for finding
// the identifier
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
  apiIDRef:
    from:
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
			crd.AddSpecField(memberNames, memberShapeRef)

+			fConfig := m.cfg.ResourceFields(crdName)[fieldName]
+			if fConfig != nil && fConfig.Reference != nil {
+				referenceFieldNames := names.New(fieldName + "Ref")
+				rf := NewReferenceField(crd, referenceFieldNames, memberShapeRef)
+				crd.SpecFields[referenceFieldNames.Original] = rf
+				crd.Fields[referenceFieldNames.Camel] = rf
+			}
		}

		// Now any additional Spec fields that are required from other API
		// operations.
		for targetFieldName, fieldConfig := range m.cfg.ResourceFields(crdName) {
			if fieldConfig.IsReadOnly {
				// It's a Status field...
				continue
			}

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
	shapeRef *awssdkmodel.ShapeRef,
) *Field {
	gt := "*ackv1alpha1.AWSResourceReferenceWrapper"
	gtp := "*github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1.AWSResourceReferenceWrapper"
	gte := ""
	if shapeRef.Shape.Type == "list" {
		gt = "[]" + gt"
		gtp = "[]" + gtp
		gte = "*ackv1alpha1.AWSResourceReferenceWrapper"
	}
	return &Field{
		CRD:               crd,
		Names:             fieldNames,
		Path:              fieldNames.Original,
		ShapeRef:          nil,
		GoType:            gt,
		GoTypeElem:        gte,
		GoTypeWithPkgName: gtp,
		FieldConfig:       nil,
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
+       // It returns an AWSResource with resolved references & 'ACK.ReferenceResolved'
+       // condition and an error if the passed AWSResource's reference fields
+       // cannot be resolved
+       ResolveReferences(context.Context, client.Reader, AWSResource) (AWSResource, error)
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
+    	  res, _ = rm.ResolveReferences(ctx, r.apiReader, res)
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
+	resolvedRefDesired, err := rm.ResolveReferences(ctx, r.apiReader, desired)
+	rlog.Exit("rm.ResolveReferences", err)
+	if err != nil {
+		return resolvedRefDesired, err
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
      ApiId:
        references:
          resource: API
          path: Status.APIID
  VpcLink:
    fields:
      SecurityGroupIds:
        references:
          resource: API
          path: Status.APIID
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
) (acktypes.AWSResource, error) {
  ko := rm.concreteResource(res).ko.DeepCopy()
  referencePresent := false
  if ko.Spec.APIIDRef != nil && ko.Spec.APIID != nil {
    return ackcondition.WithReferenceResolvedCondition(&resource{ko}, fmt.Errorf("'APIID' field should not be present when using reference field 'APIIDRef'"))
  }
  if ko.Spec.APIIDRef == nil && ko.Spec.APIID == nil {
    return &resource{ko}, fmt.Errorf("At least one of 'APIID' or 'APIIDRef' field should be present")
  }
  // Checking Referenced Field APIIDRef
  if ko.Spec.APIIDRef != nil && ko.Spec.APIIDRef.From != nil {
    referencePresent = true
    arr := ko.Spec.APIIDRef.From
    if arr == nil || arr.Name == nil || *arr.Name == "" {
      return ackcondition.WithReferenceResolvedCondition(&resource{ko}, fmt.Errorf("provided resource reference is nil or empty"))
    }
    namespacedName := types.NamespacedName{Namespace: res.MetaObject().GetNamespace(), Name: *arr.Name}
    obj := acksvcv1alpha1.API{}
    err := apiReader.Get(ctx, namespacedName, &obj)
    if err != nil {
      return ackcondition.WithReferenceResolvedCondition(&resource{ko}, err)
    }
    var refResourceSynced bool
    for _, cond := range obj.Status.Conditions {
      if cond.Type == ackv1alpha1.ConditionTypeResourceSynced && cond.Status == corev1.ConditionTrue {
        refResourceSynced = true
        break
      }
    }
    if !refResourceSynced {
      return ackcondition.WithReferenceResolvedCondition(&resource{ko}, fmt.Errorf("referenced 'API' resource "+*arr.Name+" does not have 'ACK.ResourceSynced' condition status 'True'"))
    }
    if obj.Status.APIID == nil {
      return ackcondition.WithReferenceResolvedCondition(&resource{ko}, fmt.Errorf("'Status.APIID' is not yet present for referenced 'API' resource "+*arr.Name))
    }
    ko.Spec.APIID = obj.Status.APIID
  }
  if referencePresent {
    return ackcondition.WithReferenceResolvedCondition(&resource{ko}, nil)
  }
  return &resource{ko}, nil
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
) (acktypes.AWSResource, error) {
  ko := rm.concreteResource(res).ko.DeepCopy()
  referencePresent := false
  if ko.Spec.SecurityGroupIDsRef != nil && ko.Spec.SecurityGroupIDs != nil {
    return ackcondition.WithReferenceResolvedCondition(&resource{ko}, fmt.Errorf("'SecurityGroupIDs' field should not be present when using reference field 'SecurityGroupIDsRef'"))
  }
  // Checking Referenced Field SecurityGroupIDsRef
  if ko.Spec.SecurityGroupIDsRef != nil && len(ko.Spec.SecurityGroupIDsRef) > 0 {
    referencePresent = true
    resolvedReferences := []*string{}
    for _, arrw := range ko.Spec.SecurityGroupIDsRef {
      arr := arrw.From
      if arr == nil || arr.Name == nil || *arr.Name == "" {
        return ackcondition.WithReferenceResolvedCondition(&resource{ko}, fmt.Errorf("provided resource reference is nil or empty"))
      }
      namespacedName := types.NamespacedName{Namespace: res.MetaObject().GetNamespace(), Name: *arr.Name}
      obj := acksvcv1alpha1.API{}
      err := apiReader.Get(ctx, namespacedName, &obj)
      if err != nil {
        return ackcondition.WithReferenceResolvedCondition(&resource{ko}, err)
      }
      var refResourceSynced bool
      for _, cond := range obj.Status.Conditions {
        if cond.Type == ackv1alpha1.ConditionTypeResourceSynced && cond.Status == corev1.ConditionTrue {
          refResourceSynced = true
          break
        }
      }
      if !refResourceSynced {
        return ackcondition.WithReferenceResolvedCondition(&resource{ko}, fmt.Errorf("referenced 'API' resource "+*arr.Name+" does not have 'ACK.ResourceSynced' condition status 'True'"))
      }
      if obj.Status.APIID == nil {
        return ackcondition.WithReferenceResolvedCondition(&resource{ko}, fmt.Errorf("'Status.APIID' is not yet present for referenced 'API' resource "+*arr.Name))
      }
      resolvedReferences = append(resolvedReferences, obj.Status.APIID)
    }
    ko.Spec.SecurityGroupIDs = resolvedReferences
  }
  if referencePresent {
    return ackcondition.WithReferenceResolvedCondition(&resource{ko}, nil)
  }
  return &resource{ko}, nil
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
