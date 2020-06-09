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

	"github.com/aws/aws-sdk-go/aws/session"

	ackrt "github.com/aws/aws-service-operator-k8s/pkg/runtime"
	"github.com/aws/aws-service-operator-k8s/pkg/types"
	acktypes "github.com/aws/aws-service-operator-k8s/pkg/types"

	// svcsdkapi "github.com/aws/aws-sdk-go/service/{{ .AWSServiceAlias }}/{{ .AWSServiceAlias }}iface"
	svcsdkapi "github.com/aws/aws-service-operator-k8s/services/example/sdk/service/bookstore/bookstoreiface"
	// svcsdk "github.com/aws/aws-sdk-go/service/{{ .AWSServiceAlias }}"
	svcsdk "github.com/aws/aws-service-operator-k8s/services/example/sdk/service/bookstore"
)

// bookResourceManager is responsible for providing a consistent way to perform
// CRUD operations in a backend AWS service API for Book custom resources.
type bookResourceManager struct {
	// awsAccountID is the AWS account identifier that contains the resources
	// managed by this resource manager
	awsAccountID types.AWSAccountID
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
	return nil, nil
}

// Update attempts to mutate the supplied AWSResource in the backend AWS
// service API, returning an AWSResource representing the newly-mutated
// resource
func (rm *bookResourceManager) Update(
	ctx context.Context,
	res acktypes.AWSResource,
) (acktypes.AWSResource, error) {
	return nil, nil
}

// Delete attempts to destroy the supplied AWSResource in the backend AWS
// service API.
func (rm *bookResourceManager) Delete(
	ctx context.Context,
	res acktypes.AWSResource,
) error {
	return nil
}

func newBookResourceManager(
	id types.AWSAccountID,
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
