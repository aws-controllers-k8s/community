{{ template "boilerplate" }}

package {{ .CRD.Names.Snake }}

import (
	"context"
	"fmt"

	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	ackrt "github.com/aws/aws-controllers-k8s/pkg/runtime"
	acktypes "github.com/aws/aws-controllers-k8s/pkg/types"
	"github.com/aws/aws-sdk-go/aws/session"

	svcsdk "github.com/aws/aws-sdk-go/service/{{ .ServiceIDClean }}"
	svcsdkapi "github.com/aws/aws-sdk-go/service/{{ .ServiceIDClean }}/{{ .ServiceIDClean }}iface"
)

// +kubebuilder:rbac:groups={{ .APIGroup }},resources={{ ToLower .CRD.Plural }},verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups={{ .APIGroup }},resources={{ ToLower .CRD.Plural }}/status,verbs=get;update;patch

// resourceManager is responsible for providing a consistent way to perform
// CRUD operations in a backend AWS service API for Book custom resources.
type resourceManager struct {
	// rr is the AWSResourceReconciler which can be used for various utility
	// functions such as querying for Secret values given a SecretReference
	rr acktypes.AWSResourceReconciler
	// awsAccountID is the AWS account identifier that contains the resources
	// managed by this resource manager
	awsAccountID ackv1alpha1.AWSAccountID
	// The AWS Region that this resource manager targets
	awsRegion ackv1alpha1.AWSRegion
	// sess is the AWS SDK Session object used to communicate with the backend
	// AWS service API
	sess *session.Session
	// sdk is a pointer to the AWS service API interface exposed by the
	// aws-sdk-go/services/{alias}/{alias}iface package.
	sdkapi svcsdkapi.{{ .SDKAPIInterfaceTypeName }}API
}

// concreteResource returns a pointer to a resource from the supplied
// generic AWSResource interface
func (rm *resourceManager) concreteResource(
	res acktypes.AWSResource,
) *resource {
	// cast the generic interface into a pointer type specific to the concrete
	// implementing resource type managed by this resource manager
	return res.(*resource)
}

// ReadOne returns the currently-observed state of the supplied AWSResource in
// the backend AWS service API.
func (rm *resourceManager) ReadOne(
	ctx context.Context,
	res acktypes.AWSResource,
) (acktypes.AWSResource, error) {
	r := rm.concreteResource(res)
	if r.ko == nil {
		// Should never happen... if it does, it's buggy code.
		panic("resource manager's ReadOne() method received resource with nil CR object")
	}
	observed, err := rm.sdkFind(ctx, r)
	if err != nil {
		return nil, err
	}
	return observed, nil
}

// Create attempts to create the supplied AWSResource in the backend AWS
// service API, returning an AWSResource representing the newly-created
// resource
func (rm *resourceManager) Create(
	ctx context.Context,
	res acktypes.AWSResource,
) (acktypes.AWSResource, error) {
	r := rm.concreteResource(res)
	if r.ko == nil {
		// Should never happen... if it does, it's buggy code.
		panic("resource manager's Create() method received resource with nil CR object")
	}
	created, err := rm.sdkCreate(ctx, r)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// Update attempts to mutate the supplied AWSResource in the backend AWS
// service API, returning an AWSResource representing the newly-mutated
// resource. Note that implementers should NOT check to see if the latest
// observed resource differs from the supplied desired state. The higher-level
// reonciler determines whether or not the desired differs from the latest
// observed and decides whether to call the resource manager's Update method
func (rm *resourceManager) Update(
	ctx context.Context,
	res acktypes.AWSResource,
) (acktypes.AWSResource, error) {
	r := rm.concreteResource(res)
	if r.ko == nil {
		// Should never happen... if it does, it's buggy code.
		panic("resource manager's Update() method received resource with nil CR object")
	}
	updated, err := rm.sdkUpdate(ctx, r)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// Delete attempts to destroy the supplied AWSResource in the backend AWS
// service API.
func (rm *resourceManager) Delete(
	ctx context.Context,
	res acktypes.AWSResource,
) error {
	r := rm.concreteResource(res)
	if r.ko == nil {
		// Should never happen... if it does, it's buggy code.
		panic("resource manager's Update() method received resource with nil CR object")
	}
	return rm.sdkDelete(ctx, r)
}

// ARNFromName returns an AWS Resource Name from a given string name. This
// is useful for constructing ARNs for APIs that require ARNs in their
// GetAttributes operations but all we have (for new CRs at least) is a
// name for the resource
func (rm *resourceManager) ARNFromName(name string) string {
	return fmt.Sprintf(
		"arn:aws:{{ .ServiceIDClean }}:%s:%s:%s",
		rm.awsRegion,
		rm.awsAccountID,
		name,
	)
}

// newResourceManager returns a new struct implementing
// acktypes.AWSResourceManager
func newResourceManager(
	rr acktypes.AWSResourceReconciler,
	id ackv1alpha1.AWSAccountID,
	region ackv1alpha1.AWSRegion,
) (*resourceManager, error) {
	sess, err := ackrt.NewSession()
	if err != nil {
		return nil, err
	}
	return &resourceManager{
		rr: rr,
		awsAccountID: id,
		awsRegion: region,
		sess:		 sess,
		sdkapi:	   svcsdk.New(sess),
	}, nil
}
