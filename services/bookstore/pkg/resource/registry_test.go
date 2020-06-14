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

package resource_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	resource "github.com/aws/aws-service-operator-k8s/services/bookstore/pkg/resource"

	_ "github.com/aws/aws-service-operator-k8s/services/bookstore/pkg/resource/book"
)

func TestRegistry(t *testing.T) {
	require := require.New(t)

	rmfs := resource.GetManagerFactories()
	require.NotEmpty(rmfs)

	// There should be a resource manager factory for Book resources
	foundBookRMF := false
	for _, rmf := range rmfs {
		if rmf.ResourceDescriptor().GroupKind().String() == "Book.bookstore.services.k8s.aws" {
			foundBookRMF = true
		}
	}

	require.True(foundBookRMF)
}
