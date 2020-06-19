{{ template "boilerplate" }}

package {{ .CRD.Names.Snake }}

import (
	acktypes "github.com/aws/aws-service-operator-k8s/pkg/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sapirt "k8s.io/apimachinery/pkg/runtime"
	k8sctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	svcapitypes "github.com/aws/aws-service-operator-k8s/services/{{ .ServiceAlias }}/apis/{{ .APIVersion }}"
)

const (
	{{ .CRD.Names.CamelLower }}FinalizerString = "finalizers.{{ .APIGroup }}/{{ .CRD.Kind }}"
)

var (
	{{ .CRD.Names.CamelLower }}ResourceGK = metav1.GroupKind{
		Group: "{{ .APIGroup }}",
		Kind:  "{{ .CRD.Kind }}",
	}
)

// {{ .CRD.Names.CamelLower }}ResourceDescriptor implements the
// `aws-service-operator-k8s/pkg/types.AWSResourceDescriptor` interface
type {{ .CRD.Names.CamelLower }}ResourceDescriptor struct {
}

// GroupKind returns a Kubernetes metav1.GroupKind struct that describes the
// API Group and Kind of CRs described by the descriptor
func (d *{{ .CRD.Names.CamelLower }}ResourceDescriptor) GroupKind() *metav1.GroupKind {
	return &{{ .CRD.Names.CamelLower }}ResourceGK
}

// EmptyRuntimeObject returns an empty object prototype that may be used in
// apimachinery and k8s client operations
func (d *{{ .CRD.Names.CamelLower }}ResourceDescriptor) EmptyRuntimeObject() k8sapirt.Object {
	return &svcapitypes.{{ .CRD.Kind }}{}
}

// ResourceFromRuntimeObject returns an AWSResource that has been initialized
// with the supplied runtime.Object
func (d *{{ .CRD.Names.CamelLower }}ResourceDescriptor) ResourceFromRuntimeObject(
	obj k8sapirt.Object,
) acktypes.AWSResource {
	return &{{ .CRD.Names.CamelLower }}Resource{
		ko: obj.(*svcapitypes.{{ .CRD.Kind }}),
	}
}

// Equal returns true if the two supplied AWSResources have the same content.
// The underlying types of the two supplied AWSResources should be the same. In
// other words, the Equal() method should be called with the same concrete
// implementing AWSResource type
func (d *{{ .CRD.Names.CamelLower }}ResourceDescriptor) Equal(
	a acktypes.AWSResource,
	b acktypes.AWSResource,
) bool {
	ac := a.(*{{ .CRD.Names.CamelLower }}Resource)
	bc := b.(*{{ .CRD.Names.CamelLower }}Resource)
	opts := cmpopts.EquateEmpty()
	return cmp.Equal(ac.sdko, bc.sdko, opts)
}

// Diff returns a string representing the difference between two supplied
// AWSResources/ The underlying types of the two supplied AWSResources should
// be the same. In other words, the Diff() method should be called with the
// same concrete implementing AWSResource type
func (d *{{ .CRD.Names.CamelLower }}ResourceDescriptor) Diff(
	a acktypes.AWSResource,
	b acktypes.AWSResource,
) string {
	ac := a.(*{{ .CRD.Names.CamelLower }}Resource)
	bc := b.(*{{ .CRD.Names.CamelLower }}Resource)
	opts := cmpopts.EquateEmpty()
	return cmp.Diff(ac.sdko, bc.sdko, opts)
}

// UpdateCRStatus accepts an AWSResource object and changes the Status
// sub-object of the AWSResource's Kubernetes custom resource (CR) and
// returns whether any changes were made
func (d *{{ .CRD.Names.CamelLower }}ResourceDescriptor) UpdateCRStatus(
	res acktypes.AWSResource,
) (bool, error) {
	updated := false
	return updated, nil
}

// IsManaged returns true if the supplied AWSResource is under the management
// of an ACK service controller. What this means in practice is that the
// underlying custom resource (CR) in the AWSResource has had a
// resource-specific finalizer associated with it.
func (d *{{ .CRD.Names.CamelLower }}ResourceDescriptor) IsManaged(
	res acktypes.AWSResource,
) bool {
	obj := res.RuntimeMetaObject()
	if obj == nil {
		// Should not happen. If it does, there is a bug in the code
		panic("nil RuntimeMetaObject in AWSResource")
	}
	// Remove use of custom code once
	// https://github.com/kubernetes-sigs/controller-runtime/issues/994 is
	// fixed. This should be able to be:
	//
	// return k8sctrlutil.ContainsFinalizer(obj, {{ .CRD.Names.CamelLower }}FinalizerString)
	return containsFinalizer(obj, {{ .CRD.Names.CamelLower }}FinalizerString)
}

// Remove once https://github.com/kubernetes-sigs/controller-runtime/issues/994
// is fixed.
func containsFinalizer(obj acktypes.RuntimeMetaObject, finalizer string) bool {
	f := obj.GetFinalizers()
	for _, e := range f {
		if e == finalizer {
			return true
		}
	}
	return false
}

// MarkManaged places the supplied resource under the management of ACK.  What
// this typically means is that the resource manager will decorate the
// underlying custom resource (CR) with a finalizer that indicates ACK is
// managing the resource and the underlying CR may not be deleted until ACK is
// finished cleaning up any backend AWS service resources associated with the
// CR.
func (d *{{ .CRD.Names.CamelLower }}ResourceDescriptor) MarkManaged(
	res acktypes.AWSResource,
) {
	obj := res.RuntimeMetaObject()
	if obj == nil {
		// Should not happen. If it does, there is a bug in the code
		panic("nil RuntimeMetaObject in AWSResource")
	}
	k8sctrlutil.AddFinalizer(obj, {{ .CRD.Names.CamelLower }}FinalizerString)
}

// MarkUnmanaged removes the supplied resource from management by ACK.  What
// this typically means is that the resource manager will remove a finalizer
// underlying custom resource (CR) that indicates ACK is managing the resource.
// This will allow the Kubernetes API server to delete the underlying CR.
func (d *{{ .CRD.Names.CamelLower }}ResourceDescriptor) MarkUnmanaged(
	res acktypes.AWSResource,
) {
	obj := res.RuntimeMetaObject()
	if obj == nil {
		// Should not happen. If it does, there is a bug in the code
		panic("nil RuntimeMetaObject in AWSResource")
	}
	k8sctrlutil.RemoveFinalizer(obj, {{ .CRD.Names.CamelLower }}FinalizerString)
}

