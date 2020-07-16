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
	"testing"

	"github.com/stretchr/testify/require"

	ackrt "github.com/aws/aws-controllers-k8s/pkg/runtime"
	acktypes "github.com/aws/aws-controllers-k8s/pkg/types"

	"github.com/aws/aws-controllers-k8s/services/bookstore/pkg/resource"
	_ "github.com/aws/aws-controllers-k8s/services/bookstore/pkg/resource/book"
)

func newBookResource() acktypes.AWSResource {
	rmfs := resource.GetManagerFactories()
	var rd acktypes.AWSResourceDescriptor
	for _, rmf := range rmfs {
		if rmf.ResourceDescriptor().GroupKind().String() == "Book.bookstore.services.k8s.aws" {
			rd = rmf.ResourceDescriptor()
		}
	}
	if rd == nil {
		panic("expected to find Book resource manager")
	}
	return rd.ResourceFromRuntimeObject(rd.EmptyRuntimeObject())
}

func TestIsAdopted(t *testing.T) {
	require := require.New(t)

	res := newBookResource()
	require.False(ackrt.IsAdopted(res))
}

func TestIsSynced(t *testing.T) {
	require := require.New(t)

	res := newBookResource()
	require.False(ackrt.IsSynced(res))
}
