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

package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	"github.com/aws/aws-controllers-k8s/pkg/model"
)

// GetCRDByName returns a CRD model with the supplied name
func GetCRDByName(
	t *testing.T,
	g *generate.Generator,
	name string,
) *model.CRD {
	require := require.New(t)

	crds, err := g.GetCRDs()
	require.Nil(err)

	for _, c := range crds {
		if c.Names.Original == name {
			return c
		}
	}
	return nil
}
