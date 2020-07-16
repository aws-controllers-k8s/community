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

package pet

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"

	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	ackerr "github.com/aws/aws-controllers-k8s/pkg/errors"
	ackrt "github.com/aws/aws-controllers-k8s/pkg/runtime"
	acktypes "github.com/aws/aws-controllers-k8s/pkg/types"

	// svcapitypes "github.com/aws/aws-sdk-go/service/apis/{{ .AWSServiceVersion}}

	// svcsdkapi "github.com/aws/aws-sdk-go/service/{{ .AWSServiceAlias }}/{{ .AWSServiceAlias }}iface"
	svcsdkapi "github.com/aws/aws-controllers-k8s/services/petstore/sdk/service/petstore/petstoreiface"
	// svcsdk "github.com/aws/aws-sdk-go/service/{{ .AWSServiceAlias }}"
	svcsdk "github.com/aws/aws-controllers-k8s/services/petstore/sdk/service/petstore"
)

// resourceManager is responsible for providing a consistent way to perform
// CRUD operations in a backend AWS service API for Pet custom resources.
type resourceManager struct {
	// awsAccountID is the AWS account identifier that contains the resources
	// managed by this resource manager
	awsAccountID ackv1alpha1.AWSAccountID
	// sess is the AWS SDK Session object used to communicate with the backend
	// AWS service API
	sess *session.Session
	// sdk is a pointer to the AWS service API interface exposed by the
	// aws-sdk-go/services/{alias}/{alias}iface package.
	sdkapi svcsdkapi.PetstoreAPI
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
	sdko, err := rm.findSDKResource(ctx, r)
	if err != nil {
		return nil, err
	}
	return &resource{
		ko:   r.ko,
		sdko: sdko,
	}, nil
}

// findSDKResource returns SDK-specific information about a supplied resource
func (rm *resourceManager) findSDKResource(
	ctx context.Context,
	r *resource,
) (*svcsdk.PetData, error) {
	input, err := rm.newDescribeRequestPayload(r)
	if err != nil {
		return nil, err
	}
	resp, err := rm.sdkapi.DescribePetWithContext(ctx, input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFoundException" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}
	return resp.Pet, nil
}

// newDescribeRequestPayload returns SDK-specific struct for the HTTP request
// payload of the Describe API call for the resource
func (rm *resourceManager) newDescribeRequestPayload(
	r *resource,
) (*svcsdk.DescribePetInput, error) {
	return &svcsdk.DescribePetInput{
		PetName:  r.ko.Spec.Name,
		PetOwner: r.ko.Spec.Owner,
	}, nil
}

// Create attempts to create the supplied AWSResource in the backend AWS
// service API, returning an AWSResource representing the newly-created
// resource
func (rm *resourceManager) Create(
	ctx context.Context,
	res acktypes.AWSResource,
) (acktypes.AWSResource, error) {
	r := rm.concreteResource(res)
	input, err := rm.newCreateRequestPayload(r)
	if err != nil {
		return nil, err
	}
	resp, err := rm.sdkapi.CreatePetWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return &resource{
		ko:   r.ko,
		sdko: resp.Pet,
	}, nil
}

// newCreateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newCreateRequestPayload(
	r *resource,
) (*svcsdk.CreatePetInput, error) {
	return &svcsdk.CreatePetInput{
		PetName: r.ko.Spec.Name,
	}, nil
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
	if r.sdko == nil {
		// Should never happen... if it does, it's buggy code.
		panic("resource manager's Update() method received resource with nil SDK object")
	}
	input, err := rm.newUpdateRequestPayload(r)
	if err != nil {
		return nil, err
	}
	resp, err := rm.sdkapi.UpdatePetWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return &resource{
		ko:   r.ko,
		sdko: resp.Pet,
	}, nil
}

// newUpdateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Update API call for the resource
func (rm *resourceManager) newUpdateRequestPayload(
	r *resource,
) (*svcsdk.UpdatePetInput, error) {
	return &svcsdk.UpdatePetInput{
		PetName: r.ko.Spec.Name,
	}, nil
}

// Delete attempts to destroy the supplied AWSResource in the backend AWS
// service API.
func (rm *resourceManager) Delete(
	ctx context.Context,
	res acktypes.AWSResource,
) error {
	r := rm.concreteResource(res)
	if r.sdko == nil {
		// Should never happen... if it does, it's buggy code.
		panic("resource manager's Update() method received resource with nil SDK object")
	}
	input, err := rm.newDeleteRequestPayload(r)
	if err != nil {
		return err
	}
	_, err = rm.sdkapi.DeletePetWithContext(ctx, input)
	return err
}

// newDeleteRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Delete API call for the resource
func (rm *resourceManager) newDeleteRequestPayload(
	r *resource,
) (*svcsdk.DeletePetInput, error) {
	return &svcsdk.DeletePetInput{
		PetName: r.sdko.PetName,
	}, nil
}

func newResourceManager(
	id ackv1alpha1.AWSAccountID,
) (*resourceManager, error) {
	sess, err := ackrt.NewSession()
	if err != nil {
		return nil, err
	}
	return &resourceManager{
		awsAccountID: id,
		sess:         sess,
	}, nil
}
