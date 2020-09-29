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

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"

	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	ackcompare "github.com/aws/aws-controllers-k8s/pkg/compare"
)

// AWSResourceManager is responsible for providing a consistent way to perform
// CRUD+L operations in a backend AWS service API for Kubernetes custom
// resources (CR) corresponding to those AWS service API resources.
//
// Use an AWSResourceManagerFactory to create an AWSResourceManager for a
// particular APIResource and AWS account.
type AWSResourceManager interface {
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
	// Update attempts to mutate the supplied desired AWSResource in the backend AWS
	// service API, returning an AWSResource representing the newly-mutated
	// resource.
	// Note for specialized logic implementers can check to see how the latest
	// observed resource differs from the supplied desired state. The
	// higher-level reonciler determines whether or not the desired differs
	// from the latest observed and decides whether to call the resource
	// manager's Update method
	Update(context.Context, /* desired */ AWSResource, /* latest */ AWSResource, *ackcompare.Reporter) (AWSResource, error)

	// Delete attempts to destroy the supplied AWSResource in the backend AWS
	// service API.
	Delete(context.Context, AWSResource) error
	// ARNFromName returns an AWS Resource Name from a given string name. This
	// is useful for constructing ARNs for APIs that require ARNs in their
	// GetAttributes operations but all we have (for new CRs at least) is a
	// name for the resource
	ARNFromName(string) string
}

// AWSResourceManagerFactory returns an AWSResourceManager that can be used to
// manage AWS resources for a particular AWS account
type AWSResourceManagerFactory interface {
	// ResourceDescriptor returns an AWSResourceDescriptor that can be used by
	// the upstream controller-runtime to introspect the CRs that the resource
	// manager will manage as well as produce Kubernetes runtime object
	// prototypes
	ResourceDescriptor() AWSResourceDescriptor
	// ManagerFor returns an AWSResourceManager that manages AWS resources on
	// behalf of a particular AWS account and in a specific AWS region
	ManagerFor(
		AWSResourceReconciler,
		*session.Session,
		ackv1alpha1.AWSAccountID,
		ackv1alpha1.AWSRegion,
	) (AWSResourceManager, error)
}
