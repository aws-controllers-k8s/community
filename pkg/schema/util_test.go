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

package schema_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-service-operator-k8s/pkg/names"
	"github.com/aws/aws-service-operator-k8s/pkg/testutil"
)

func TestGoTypeFromSchemaRef(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tests := []struct {
		name       string
		schemaName string
		propNames  names.Names
		expGoType  string
		expErr     error
	}{
		{
			"string type",
			"Attribute",
			names.New("value"),
			"string",
			nil,
		},
		{
			"boolean type",
			"ImageScanningConfiguration",
			names.New("scanOnPush"),
			"bool",
			nil,
		},
		{
			"int64 type",
			"ImageDetail",
			names.New("imageSizeInBytes"),
			"int64",
			nil,
		},
		{
			"object type",
			"ImageDetail",
			names.New("imageScanStatus"),
			"*ImageScanStatus",
			nil,
		},
		{
			"simple array type",
			"ImageDetail",
			names.New("imageTags"),
			"[]string",
			nil,
		},
		{
			"object array type",
			"ImageScanFindings",
			names.New("findings"),
			"[]*ImageScanFinding",
			nil,
		},
		{
			"object array type for only-pluralized type name",
			"DBCluster",
			names.New("DBClusterOptionGroupMemberships"),
			"[]*DBClusterOptionGroupMemberships",
			nil,
		},
		{
			"object array type for List-suffixed wrapper struct name with pluralized referred-to object type",
			"OptionGroup",
			names.New("Options"),
			"[]*Option",
			nil,
		},
		{
			"object array type for List-suffixed wrapper struct name where there is no referred-to object type, pluralized or singular",
			"OptionGroupOption",
			names.New("OptionGroupOptionVersions"),
			"AnonymousArrayObjectType",
			nil,
		},
		{
			"map[string]string type",
			"AnyTag",
			names.New("tags"),
			"map[string]string",
			nil,
		},
		{
			"unknown type",
			"UnknownType",
			names.New("field"),
			"",
			fmt.Errorf("failed to determine Go type. schema.Type was %s", "unknown"),
		},
	}
	for _, test := range tests {
		sh := testutil.NewSchemaHelperFromFile(t, "complex-types.yaml")

		schema := sh.GetSchema(test.schemaName)
		require.NotNil(schema, test.name)

		propSchemaRef := schema.Properties[test.propNames.Original]
		require.NotNil(propSchemaRef, test.name)

		gt, err := sh.GetGoTypeFromSchemaRef(test.propNames, propSchemaRef)
		if test.expErr != nil {
			assert.Error(err)
			assert.Equal(test.expErr, err, test.name)
		} else {
			assert.Nil(err, "expected err to be nil but got %s", err)
		}
		assert.Equal(test.expGoType, gt, test.name)
	}
}
