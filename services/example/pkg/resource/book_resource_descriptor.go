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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	acktypes "github.com/aws/aws-service-operator-k8s/pkg/types"

	svcapitypes "github.com/aws/aws-service-operator-k8s/services/example/apis/v1alpha1"
)

var (
	bookResourceGK = metav1.GroupKind{
		Group: "bookstore.services.k8s.aws",
		Kind:  "Book",
	}
)

// bookResourceDescriptor implements the
// `aws-service-operator-k8s/pkg/types.AWSResourceDescriptor` interface
type bookResourceDescriptor struct {
}

// GroupKind returns a Kubernetes metav1.GroupKind struct that describes the
// API Group and Kind of CRs described by the descriptor
func (d *bookResourceDescriptor) GroupKind() *metav1.GroupKind {
	return &bookResourceGK
}

// EmptyObject returns an empty object prototype that may be used in
// apimachinery and k8s client operations
func (d *bookResourceDescriptor) EmptyObject() runtime.Object {
	return &svcapitypes.Book{}
}

// ResourceFromObject returns an AWSResource that has been initialized with the
// supplied runtime.Object
func (d *bookResourceDescriptor) ResourceFromObject(
	obj runtime.Object,
) acktypes.AWSResource {
	return &bookResource{
		ko: obj.(*svcapitypes.Book),
	}
}
