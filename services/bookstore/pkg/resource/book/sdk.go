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

package book

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/awserr"

	ackerr "github.com/aws/aws-controllers-k8s/pkg/errors"

	// svcsdk "github.com/aws/aws-sdk-go/service/{{ .AWSServiceAlias }}"
	svcsdk "github.com/aws/aws-controllers-k8s/services/bookstore/sdk/service/bookstore"
)

// sdkFind returns SDK-specific information about a supplied resource
func (rm *resourceManager) sdkFind(
	ctx context.Context,
	r *resource,
) (*resource, error) {
	input, err := rm.newDescribeRequestPayload(r)
	if err != nil {
		return nil, err
	}
	resp, err := rm.sdkapi.DescribeBookWithContext(ctx, input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFoundException" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()
	ko.Spec.Name = resp.Book.BookName
	ko.Spec.Title = resp.Book.Title
	ko.Spec.Author = resp.Book.Author
	ko.Status.CreateTime = resp.CreateTime
	return &resource{ko}, nil
}

// newDescribeRequestPayload returns SDK-specific struct for the HTTP request
// payload of the Describe API call for the resource
func (rm *resourceManager) newDescribeRequestPayload(
	r *resource,
) (*svcsdk.DescribeBookInput, error) {
	return &svcsdk.DescribeBookInput{
		BookName: r.ko.Spec.Name,
	}, nil
}

// sdkCreate creates the supplied resource in the backend AWS service API and
// returns a new resource with any fields in the Status field filled in
func (rm *resourceManager) sdkCreate(
	ctx context.Context,
	r *resource,
) (*resource, error) {
	input, err := rm.newCreateRequestPayload(r)
	if err != nil {
		return nil, err
	}
	resp, err := rm.sdkapi.CreateBookWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()
	ko.Spec.Name = resp.Book.BookName
	ko.Spec.Title = resp.Book.Title
	ko.Spec.Author = resp.Book.Author
	ko.Status.CreateTime = resp.CreateTime
	return &resource{ko}, nil
}

// newCreateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newCreateRequestPayload(
	r *resource,
) (*svcsdk.CreateBookInput, error) {
	return &svcsdk.CreateBookInput{
		BookName: r.ko.Spec.Name,
		Title:    r.ko.Spec.Title,
		Author:   r.ko.Spec.Author,
	}, nil
}

// sdkUpdate patches the supplied resource in the backend AWS service API and
// returns a new resource with updated fields.
func (rm *resourceManager) sdkUpdate(
	ctx context.Context,
	r *resource,
) (*resource, error) {
	input, err := rm.newUpdateRequestPayload(r)
	if err != nil {
		return nil, err
	}
	resp, err := rm.sdkapi.UpdateBookWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()
	ko.Spec.Name = resp.Book.BookName
	ko.Spec.Title = resp.Book.Title
	ko.Spec.Author = resp.Book.Author
	return &resource{ko}, nil
}

// newUpdateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Update API call for the resource
func (rm *resourceManager) newUpdateRequestPayload(
	r *resource,
) (*svcsdk.UpdateBookInput, error) {
	return &svcsdk.UpdateBookInput{
		BookName: r.ko.Spec.Name,
		Title:    r.ko.Spec.Title,
		Author:   r.ko.Spec.Author,
	}, nil
}

// sdkDelete deletes the supplied resource in the backend AWS service API
func (rm *resourceManager) sdkDelete(
	ctx context.Context,
	r *resource,
) error {
	input, err := rm.newDeleteRequestPayload(r)
	if err != nil {
		return err
	}
	_, err = rm.sdkapi.DeleteBookWithContext(ctx, input)
	return err
}

// newDeleteRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Delete API call for the resource
func (rm *resourceManager) newDeleteRequestPayload(
	r *resource,
) (*svcsdk.DeleteBookInput, error) {
	return &svcsdk.DeleteBookInput{
		BookName: r.ko.Spec.Name,
	}, nil
}
