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
	"k8s.io/apimachinery/pkg/runtime"
)

// AWSResourceDescriptor provides metadata that describes the Kubernetes
// metadata associated with an AWSResource, the Kubernetes runtime.Object
// prototype for that AWSResource, and the relationships between the
// AWSResource and other AWSResources
type AWSResourceDescriptor interface {
	// EmptyObject returns an empty object prototype that may be used in
	// apimachinery and k8s client operations
	EmptyObject() runtime.Object
	// ResourceFromObject returns an AWSResource that has been initialized with
	// the supplied runtime.Object
	ResourceFromObject(runtime.Object) AWSResource
}
