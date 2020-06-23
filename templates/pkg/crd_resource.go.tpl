{{ template "boilerplate" }}

package {{ .CRD.Names.Snake }}

import (
	ackv1alpha1 "github.com/aws/aws-service-operator-k8s/apis/core/v1alpha1"
	acktypes "github.com/aws/aws-service-operator-k8s/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8srt "k8s.io/apimachinery/pkg/runtime"

	svcapitypes "github.com/aws/aws-service-operator-k8s/service/{{ .ServiceAlias }}/apis/{{ .APIVersion}}
	svcsdk "github.com/aws/aws-sdk-go/service/{{ .ServiceAlias }}"
)

// {{ .CRD.Names.CamelLower }}Resource implements the `aws-service-operator-k8s/pkg/types.AWSResource`
// interface
type {{ .CRD.Names.CamelLower }}Resource struct {
	// The Kubernetes-native CR representing the resource
	ko *svcapitypes.{{ .CRD.Names.Camel }}
	// The aws-sdk-go-native representation of the resource
	sdko *svcsdk.{{ .CRD.SDKObjectType }}
}

// Identifiers returns an AWSResourceIdentifiers object containing various
// identifying information, including the AWS account ID that owns the
// resource, the resource's AWS Resource Name (ARN)
func (r *{{ .CRD.Names.CamelLower }}Resource) Identifiers() acktypes.AWSResourceIdentifiers {
	return &{{ .CRD.Names.CamelLower }}ResourceIdentifiers{r.ko.Status.ACKResourceMetadata}
}

// IsBeingDeleted returns true if the Kubernetes resource has a non-zero
// deletion timestemp
func (r *{{ .CRD.Names.CamelLower }}Resource) IsBeingDeleted() bool {
	return !r.ko.DeletionTimestamp.IsZero()
}

// RuntimeObject returns the Kubernetes apimachinery/runtime representation of
// the AWSResource
func (r *{{ .CRD.Names.CamelLower }}Resource) RuntimeObject() k8srt.Object {
	return r.ko
}

// MetaObject returns the Kubernetes apimachinery/apis/meta/v1.Object
// representation of the AWSResource
func (r *{{ .CRD.Names.CamelLower }}Resource) MetaObject() metav1.Object {
	return r.ko
}

// RuntimeMetaObject returns an object that implements both the Kubernetes
// apimachinery/runtime.Object and the Kubernetes
// apimachinery/apis/meta/v1.Object interfaces
func (r *{{ .CRD.Names.CamelLower }}Resource) RuntimeMetaObject() acktypes.RuntimeMetaObject {
	return r.ko
}

// Conditions returns the ACK Conditions collection for the AWSResource
func (r *{{ .CRD.Names.CamelLower }}Resource) Conditions() []*ackv1alpha1.Condition {
	return r.ko.Status.Conditions
}
