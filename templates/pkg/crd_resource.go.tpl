{{ template "boilerplate" }}

package {{ .CRD.Names.Snake }}

import (
	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	acktypes "github.com/aws/aws-controllers-k8s/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8srt "k8s.io/apimachinery/pkg/runtime"

	svcapitypes "github.com/aws/aws-controllers-k8s/services/{{ .ServiceIDClean }}/apis/{{ .APIVersion}}"
)

// resource implements the `aws-service-operator-k8s/pkg/types.AWSResource`
// interface
type resource struct {
	// The Kubernetes-native CR representing the resource
	ko *svcapitypes.{{ .CRD.Names.Camel }}
}

// Identifiers returns an AWSResourceIdentifiers object containing various
// identifying information, including the AWS account ID that owns the
// resource, the resource's AWS Resource Name (ARN)
func (r *resource) Identifiers() acktypes.AWSResourceIdentifiers {
	return &resourceIdentifiers{r.ko.Status.ACKResourceMetadata}
}

// IsBeingDeleted returns true if the Kubernetes resource has a non-zero
// deletion timestemp
func (r *resource) IsBeingDeleted() bool {
	return !r.ko.DeletionTimestamp.IsZero()
}

// RuntimeObject returns the Kubernetes apimachinery/runtime representation of
// the AWSResource
func (r *resource) RuntimeObject() k8srt.Object {
	return r.ko
}

// MetaObject returns the Kubernetes apimachinery/apis/meta/v1.Object
// representation of the AWSResource
func (r *resource) MetaObject() metav1.Object {
	return r.ko
}

// RuntimeMetaObject returns an object that implements both the Kubernetes
// apimachinery/runtime.Object and the Kubernetes
// apimachinery/apis/meta/v1.Object interfaces
func (r *resource) RuntimeMetaObject() acktypes.RuntimeMetaObject {
	return r.ko
}

// Conditions returns the ACK Conditions collection for the AWSResource
func (r *resource) Conditions() []*ackv1alpha1.Condition {
	return r.ko.Status.Conditions
}
