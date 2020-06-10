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
	k8srt "k8s.io/apimachinery/pkg/runtime"

	acktypes "github.com/aws/aws-service-operator-k8s/pkg/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	svcapitypes "github.com/aws/aws-service-operator-k8s/services/bookstore/apis/v1alpha1"
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
func (d *bookResourceDescriptor) EmptyObject() k8srt.Object {
	return &svcapitypes.Book{}
}

// ResourceFromObject returns an AWSResource that has been initialized with the
// supplied runtime.Object
func (d *bookResourceDescriptor) ResourceFromObject(
	obj k8srt.Object,
) acktypes.AWSResource {
	return &bookResource{
		ko: obj.(*svcapitypes.Book),
	}
}

// Equal returns true if the two supplied AWSResources have the same content.
// The underlying types of the two supplied AWSResources should be the same. In
// other words, the Equal() method should be called with the same concrete
// implementing AWSResource type
func (d *bookResourceDescriptor) Equal(
	a acktypes.AWSResource,
	b acktypes.AWSResource,
) bool {
	ac := a.(*bookResource)
	bc := b.(*bookResource)
	opts := cmpopts.EquateEmpty()
	return cmp.Equal(ac.sdko, bc.sdko, opts)
}

// Diff returns a string representing the difference between two supplied
// AWSResources/ The underlying types of the two supplied AWSResources should
// be the same. In other words, the Diff() method should be called with the
// same concrete implementing AWSResource type
func (d *bookResourceDescriptor) Diff(
	a acktypes.AWSResource,
	b acktypes.AWSResource,
) string {
	ac := a.(*bookResource)
	bc := b.(*bookResource)
	opts := cmpopts.EquateEmpty()
	return cmp.Diff(ac.sdko, bc.sdko, opts)
}

// UpdateCRStatus accepts an AWSResource object and changes the Status
// sub-object of the AWSResource's Kubernetes custom resource (CR) and
// returns whether any changes were made
func (d *bookResourceDescriptor) UpdateCRStatus(
	res acktypes.AWSResource,
) (bool, error) {
	updated := false
	return updated, nil
}
