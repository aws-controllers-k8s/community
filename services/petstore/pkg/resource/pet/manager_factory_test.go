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

package pet_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	acktypes "github.com/aws/aws-controllers-k8s/pkg/types"

	resource "github.com/aws/aws-controllers-k8s/services/petstore/pkg/resource"
	// The pet package's init() registers the pet resource manager factory
	// with the resource package's registry
	_ "github.com/aws/aws-controllers-k8s/services/petstore/pkg/resource/pet"
)

func TestManagerFor(t *testing.T) {
	require := require.New(t)

	rmfs := resource.GetManagerFactories()
	require.NotEmpty(rmfs)

	var petRMF acktypes.AWSResourceManagerFactory
	for _, rmf := range rmfs {
		if rmf.ResourceDescriptor().GroupKind().String() == "Pet.petstore.services.k8s.aws" {
			petRMF = rmf
		}
	}

	require.NotNil(petRMF)

	acctID := ackv1alpha1.AWSAccountID("aws-account-id")
	region := ackv1alpha1.AWSRegion("us-west-2")
	petRM, err := petRMF.ManagerFor(nil, acctID, region)
	require.Nil(err)
	require.NotNil(petRM)
	require.Implements((*acktypes.AWSResourceManager)(nil), petRM)

	// The resource manager factory should keep a cache of resource manager
	// objects, keyed by account ID.
	otherPetRM, err := petRMF.ManagerFor(nil, acctID, region)
	require.Nil(err)
	require.Exactly(petRM, otherPetRM)
}
