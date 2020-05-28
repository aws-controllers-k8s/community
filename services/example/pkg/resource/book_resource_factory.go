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
	"k8s.io/apimachinery/pkg/runtime"

	acktypes "github.com/aws/aws-service-operator-k8s/pkg/types"

	svcapitypes "github.com/aws/aws-service-operator-k8s/services/example/apis/v1alpha1"
)

// bookResourceFactory implements the
// `aws-service-operator-k8s/pkg/types.AWSResourceFactory` interface
type bookResourceFactory struct {
}

// EmptyObject returns an empty object prototype that may be used in
// apimachinery and k8s client operations
func (r *bookResourceFactory) EmptyObject() runtime.Object {
	return &svcapitypes.Book{}
}

// ResourceFromObject returns an AWSResource that has been initialized with the
// supplied runtime.Object
func (r *bookResourceFactory) ResourceFromObject(
	obj runtime.Object,
) acktypes.AWSResource {
	return &bookResource{
		ko: obj.(*svcapitypes.Book),
	}
}
