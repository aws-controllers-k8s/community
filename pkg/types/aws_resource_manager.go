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

package types

import "context"

// AWSResourceManager is responsible for providing a consistent way to perform
// CRUD+L operations in a backend AWS service API for Kubernetes custom
// resources (CR) corresponding to those AWS service API resources.
//
// Use an AWSResourceManagerFactory to create an AWSResourceManager for a
// particular APIResource and AWS account.
type AWSResourceManager interface {
	// Exists returns true if the supplied resource exists in the backend AWS
	// service API.
	Exists(context.Context, AWSResource) bool
	// ReadOne returns the currently-observed state of the supplied AWSResource
	// in the backend AWS service API.
	//
	// Implementers should return (nil, ackerrors.NotFound) when the backend
	// AWS service API cannot find the resource identified by the supplied
	// AWSResource's AWS identifier information.
	ReadOne(context.Context, AWSResource) (AWSResource, error)
	// Create attempts to create the supplied AWSResource in the backend AWS
	// service API, returning an AWSResource representing the newly-created
	// resource
	Create(context.Context, AWSResource) (AWSResource, error)
	// Update attempts to mutate the supplied AWSResource in the backend AWS
	// service API, returning an AWSResource representing the newly-mutated
	// resource. Note that implementers should NOT check to see if the latest
	// observed resource differs from the supplied desired state. The
	// higher-level reonciler determines whether or not the desired differs
	// from the latest observed and decides whether to call the resource
	// manager's Update method
	Update(context.Context, AWSResource) (AWSResource, error)
	// Delete attempts to destroy the supplied AWSResource in the backend AWS
	// service API.
	Delete(context.Context, AWSResource) error
}

// AWSResourceManagerFactory returns an AWSResourceManager that can be used to
// manage AWS resources for a particular AWS account
type AWSResourceManagerFactory interface {
	// GroupKind returns a string representation of the CRs handled by resource
	// managers returned by this factory
	GroupKind() string
	// ResourceDescriptor returns an AWSResourceDescriptor that can be used by
	// the upstream controller-runtime to introspect the CRs that the resource
	// manager will manage as well as produce Kubernetes runtime object
	// prototypes
	ResourceDescriptor() AWSResourceDescriptor
	// ManagerFor returns an AWSResourceManager that manages AWS resources on
	// behalf of a particular AWS account
	ManagerFor(AWSAccountID) (AWSResourceManager, error)
}
