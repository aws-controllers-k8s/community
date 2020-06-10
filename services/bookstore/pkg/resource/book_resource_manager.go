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

package resource

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"

	ackerr "github.com/aws/aws-service-operator-k8s/pkg/errors"
	ackrt "github.com/aws/aws-service-operator-k8s/pkg/runtime"
	acktypes "github.com/aws/aws-service-operator-k8s/pkg/types"

	// svcapitypes "github.com/aws/aws-sdk-go/service/apis/{{ .AWSServiceVersion}}
	svcapitypes "github.com/aws/aws-service-operator-k8s/services/bookstore/apis/v1alpha1"
	// svcsdkapi "github.com/aws/aws-sdk-go/service/{{ .AWSServiceAlias }}/{{ .AWSServiceAlias }}iface"
	svcsdkapi "github.com/aws/aws-service-operator-k8s/services/bookstore/sdk/service/bookstore/bookstoreiface"
	// svcsdk "github.com/aws/aws-sdk-go/service/{{ .AWSServiceAlias }}"
	svcsdk "github.com/aws/aws-service-operator-k8s/services/bookstore/sdk/service/bookstore"
)

// bookResourceManager is responsible for providing a consistent way to perform
// CRUD operations in a backend AWS service API for Book custom resources.
type bookResourceManager struct {
	// awsAccountID is the AWS account identifier that contains the resources
	// managed by this resource manager
	awsAccountID acktypes.AWSAccountID
	// sess is the AWS SDK Session object used to communicate with the backend
	// AWS service API
	sess *session.Session
	// sdk is a pointer to the AWS service API interface exposed by the
	// aws-sdk-go/services/{alias}/{alias}iface package.
	sdkapi svcsdkapi.BookstoreAPI
}

// concreteResource returns a pointer to a bookResource from the supplied
// generic AWSResource interface
func (rm *bookResourceManager) concreteResource(
	res acktypes.AWSResource,
) *bookResource {
	// cast the generic interface into a pointer type specific to the concrete
	// implementing resource type managed by this resource manager
	return res.(*bookResource)
}

// Exists returns true if the supplied AWSResource exists in the backend AWS
// service API.
func (rm *bookResourceManager) Exists(
	ctx context.Context,
	res acktypes.AWSResource,
) bool {
	return false
}

// ReadOne returns the currently-observed state of the supplied AWSResource in
// the backend AWS service API.
func (rm *bookResourceManager) ReadOne(
	ctx context.Context,
	res acktypes.AWSResource,
) (acktypes.AWSResource, error) {
	r := rm.concreteResource(res)
	sdko, err := rm.findSDKBook(ctx, r)
	if err != nil {
		return nil, err
	}
	return &bookResource{
		ko:   r.ko,
		sdko: sdko,
	}, nil
}

// findSDKBook returns SDK-specific information about a supplied bookResource
func (rm *bookResourceManager) findSDKBook(
	ctx context.Context,
	r *bookResource,
) (*svcsdk.BookData, error) {
	input := svcsdk.DescribeBookInput{
		BookName:  r.ko.Spec.Name,
		BookOwner: r.ko.Spec.Owner,
	}
	resp, err := rm.sdkapi.DescribeBookWithContext(ctx, &input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFoundException" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}
	return resp.Book, nil
}

// Create attempts to create the supplied AWSResource in the backend AWS
// service API, returning an AWSResource representing the newly-created
// resource
func (rm *bookResourceManager) Create(
	ctx context.Context,
	res acktypes.AWSResource,
) (acktypes.AWSResource, error) {
	r := rm.concreteResource(res)
	input := svcsdk.CreateBookInput{
		BookName: r.ko.Spec.Name,
	}
	resp, err := rm.sdkapi.CreateBookWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}
	return &bookResource{
		ko:   r.ko,
		sdko: resp.Book,
	}, nil
}

// Update attempts to mutate the supplied AWSResource in the backend AWS
// service API, returning an AWSResource representing the newly-mutated
// resource. Note that implementers should NOT check to see if the latest
// observed resource differs from the supplied desired state. The higher-level
// reonciler determines whether or not the desired differs from the latest
// observed and decides whether to call the resource manager's Update method
func (rm *bookResourceManager) Update(
	ctx context.Context,
	res acktypes.AWSResource,
) (acktypes.AWSResource, error) {
	r := rm.concreteResource(res)
	if r.sdko == nil {
		// Should never happen... if it does, it's buggy code.
		panic("resource manager's Update() method received resource with nil SDK object")
	}
	desired, err := rm.sdkoFromKO(r.ko)
	if err != nil {
		return nil, err
	}
	input := svcsdk.UpdateBookInput{
		BookName: desired.BookName,
	}
	resp, err := rm.sdkapi.UpdateBookWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}
	return &bookResource{
		ko:   r.ko,
		sdko: resp.Book,
	}, nil
}

// sdkoFromKO constructs a BookData object from a Book CR
func (rm *bookResourceManager) sdkoFromKO(
	ko *svcapitypes.Book,
) (*svcsdk.BookData, error) {
	// TODO(jaypipes): isolate conversion/translation logic here. I'm not a
	// huge fan of the sigs.k8s.io/apimachinery/pkg/conversion package and
	// would prefer long-term to use something a bit more readable and less
	// verbose, and since we have type information for both the k8s side and
	// the SDK side, it should be possible to make non-generic conversion
	// functions.
	sdko := svcsdk.BookData{
		BookName: ko.Spec.Name,
	}
	return &sdko, nil
}

// Delete attempts to destroy the supplied AWSResource in the backend AWS
// service API.
func (rm *bookResourceManager) Delete(
	ctx context.Context,
	res acktypes.AWSResource,
) error {
	r := rm.concreteResource(res)
	if r.sdko == nil {
		// Should never happen... if it does, it's buggy code.
		panic("resource manager's Update() method received resource with nil SDK object")
	}
	input := svcsdk.DeleteBookInput{
		BookName: r.sdko.BookName,
	}
	_, err := rm.sdkapi.DeleteBookWithContext(ctx, &input)
	return err
}

func newBookResourceManager(
	id acktypes.AWSAccountID,
) (*bookResourceManager, error) {
	sess, err := ackrt.NewSession()
	if err != nil {
		return nil, err
	}
	return &bookResourceManager{
		awsAccountID: id,
		sess:         sess,
	}, nil
}
