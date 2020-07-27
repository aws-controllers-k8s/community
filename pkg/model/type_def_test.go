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

package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/aws/aws-controllers-k8s/pkg/testutil"
)

func getTypeDefByName(name string, tdefs []*model.TypeDef) *model.TypeDef {
	for _, td := range tdefs {
		if td.Names.Original == name {
			return td
		}
	}
	return nil
}

func TestAPIGatewayV2_GetTypeDefs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperForService(t, "apigatewayv2")

	tdefs, timports, err := sh.GetTypeDefs()
	require.Nil(err)

	// APIGatewayV2 shapes have time.Time ("timestamp") types and so
	// GetTypeDefs() should return the special-cased with apimachinery/metav1
	// import, aliased as "metav1"
	expImports := map[string]string{"k8s.io/apimachinery/pkg/apis/meta/v1": "metav1"}
	assert.Equal(expImports, timports)

	// There is an "Api" Shape that is a struct that is an element of the
	// GetApis Operation. Its name conflicts with the CRD called API and thus
	// we need to check that its cleaned name is set to API_SDK (the _SDK
	// suffix is appended to the type name to avoid the conflict with
	// CRD-specific structs.
	tdef := getTypeDefByName("Api", tdefs)
	require.NotNil(tdef)

	assert.Equal("API_SDK", tdef.Names.Camel)
}
