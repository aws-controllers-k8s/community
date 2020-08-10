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

package book_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	acktypes "github.com/aws/aws-controllers-k8s/pkg/types"

	resource "github.com/aws/aws-controllers-k8s/services/bookstore/pkg/resource"
	// The book package's init() registers the book resource manager factory
	// with the resource package's registry
	_ "github.com/aws/aws-controllers-k8s/services/bookstore/pkg/resource/book"
)

func TestManagerFor(t *testing.T) {
	require := require.New(t)

	rmfs := resource.GetManagerFactories()
	require.NotEmpty(rmfs)

	var bookRMF acktypes.AWSResourceManagerFactory
	for _, rmf := range rmfs {
		if rmf.ResourceDescriptor().GroupKind().String() == "Book.bookstore.services.k8s.aws" {
			bookRMF = rmf
		}
	}

	require.NotNil(bookRMF)

	acctID := ackv1alpha1.AWSAccountID("aws-account-id")
	bookRM, err := bookRMF.ManagerFor(nil, acctID)
	require.Nil(err)
	require.NotNil(bookRM)
	require.Implements((*acktypes.AWSResourceManager)(nil), bookRM)

	// The resource manager factory should keep a cache of resource manager
	// objects, keyed by account ID.
	otherBookRM, err := bookRMF.ManagerFor(nil, acctID)
	require.Nil(err)
	require.Exactly(bookRM, otherBookRM)
}
