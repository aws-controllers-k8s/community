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
	k8srt "k8s.io/apimachinery/pkg/runtime"
)

// AWSResource represents a custom resource object in the Kubernetes API that
// corresponds to a resource in an AWS service API.
type AWSResource interface {
	// AccountID returns the AWS account identifier in which the backend AWS
	// resource resides
	AccountID() AWSAccountID
	// IsBeingDeleted returns true if the Kubernetes resource has a non-zero
	// deletion timestemp
	IsBeingDeleted() bool
	// CR returns the Kubernetes custom resource (CR) representation
	// of the AWSResource
	CR() k8srt.Object
}
