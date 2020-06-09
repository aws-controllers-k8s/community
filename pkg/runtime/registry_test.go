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

package runtime_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ackrt "github.com/aws/aws-service-operator-k8s/pkg/runtime"
	acktypes "github.com/aws/aws-service-operator-k8s/pkg/types"

	bookstoretypes "github.com/aws/aws-service-operator-k8s/pkg/test/fixture/bookstore/apis/v1alpha1"
)

type bookRM struct{}

func (rm *bookRM) Exists(ctx context.Context, r acktypes.AWSResource) bool {
	return false
}
func (rm *bookRM) ReadOne(ctx context.Context, r acktypes.AWSResource) (acktypes.AWSResource, error) {
	return nil, nil
}
func (rm *bookRM) Create(ctx context.Context, r acktypes.AWSResource) (acktypes.AWSResource, error) {
	return nil, nil
}
func (rm *bookRM) Update(ctx context.Context, r acktypes.AWSResource) (acktypes.AWSResource, error) {
	return nil, nil
}
func (rm *bookRM) Delete(ctx context.Context, r acktypes.AWSResource) error {
	return nil
}

type bookRes struct {
	ko *bookstoretypes.Book
}

func (r *bookRes) IsDeleted() bool {
	return !r.ko.DeletionTimestamp.IsZero()
}

func (r *bookRes) AccountID() acktypes.AWSAccountID {
	return "example-account-id"
}

type bookRD struct{}

func (d *bookRD) GroupKind() *metav1.GroupKind {
	return &metav1.GroupKind{
		Group: "bookstore.services.k8s.aws",
		Kind:  "Book",
	}

}

func (d *bookRD) EmptyObject() runtime.Object {
	return &bookstoretypes.Book{}
}
func (d *bookRD) ResourceFromObject(obj runtime.Object) acktypes.AWSResource {
	return &bookRes{ko: obj.(*bookstoretypes.Book)}
}

type bookRMF struct{}

func (f *bookRMF) ResourceDescriptor() acktypes.AWSResourceDescriptor {
	return &bookRD{}
}
func (f *bookRMF) ManagerFor(id acktypes.AWSAccountID) (acktypes.AWSResourceManager, error) {
	return &bookRM{}, nil
}

func TestRegistry(t *testing.T) {
	require := require.New(t)

	reg := ackrt.NewRegistry()
	rmf := &bookRMF{}

	rmfs := reg.GetResourceManagerFactories()
	require.Empty(rmfs)

	reg.RegisterResourceManagerFactory(rmf)
	rmfs = reg.GetResourceManagerFactories()
	require.NotEmpty(rmfs)
	require.Contains(rmfs, rmf)
}
