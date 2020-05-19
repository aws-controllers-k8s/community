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

import "k8s.io/apimachinery/pkg/runtime"

// AWSResourceManager is responsible for providing a consistent way to perform
// CRUD+L operations in a backend AWS service API for Kubernetes custom
// resources (CR) corresponding to those AWS service API resources.
//
// Use an AWSResourceManagerFactory to create an AWSResourceManager for a
// particular APIResource and AWS account.
type AWSResourceManager interface {
	// Exists returns true if the supplied resource exists in the backend AWS
	// service API.
	Exists(AWSResource) bool
	// ReadOne returns the currently-observed state of the supplied Resource
	// in the backend AWS service API.
	ReadOne(AWSResource) (AWSResource, error)
	// Create attempts to create the supplied Resource in the backend AWS
	// service API.
	Create(AWSResource) error
	// Delete attempts to destroy the supplied Resource in the backend AWS
	// service API.
	Delete(AWSResource) error
}

// AWSResourceManagerFactory returns an AWSResourceManager that can be used to
// manage AWS resources for a particular AWS account
type AWSResourceManagerFactory interface {
	// GroupKind returns a string representation of the CRs handled by resource
	// managers returned by this factory
	GroupKind() string
	// ObjectPrototype returns a pointer to a runtime.Object that can be used
	// by the upstream controller-runtime to introspect the CRs that the
	// resource manager will manage
	ObjectPrototype() runtime.Object
	// For returns an AWSResourceManager that manages AWS resources on behalf
	// of a particular AWS account
	For(AWSAccountID) (AWSResourceManager, error)
}
