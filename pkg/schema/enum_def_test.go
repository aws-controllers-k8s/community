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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-service-operator-k8s/pkg/model"
	"github.com/aws/aws-service-operator-k8s/pkg/testutil"
)

func sortedOriginalValues(vals []model.EnumValue) []string {
	res := []string{}
	for _, val := range vals {
		res = append(res, val.Original)
	}
	sort.Strings(res)
	return res
}

func sortedCleanValues(vals []model.EnumValue) []string {
	res := []string{}
	for _, val := range vals {
		res = append(res, val.Clean)
	}
	sort.Strings(res)
	return res
}

func TestEnumDefs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tests := []struct {
		name              string
		yaml              string
		expNameExported   string
		expNameUnexported string
		expGoType         string
		expValuesOrig     []string
		expValuesClean    []string
	}{
		{
			"original same as clean value",
			`
components:
  schemas:
    ScanStatus:
      enum:
      - IN_PROGRESS
      - COMPLETE
      - FAILED
      type: string
`,
			"ScanStatus",
			"scanStatus",
			"string",
			[]string{
				"COMPLETE",
				"FAILED",
				"IN_PROGRESS",
			},
			[]string{
				"COMPLETE",
				"FAILED",
				"IN_PROGRESS",
			},
		},
		{
			"value strings need cleaning for Go output",
			`
components:
  schemas:
    InstanceLifecycle:
      enum:
      - spot
      - on-demand
      type: string
`,
			"InstanceLifecycle",
			"instanceLifecycle",
			"string",
			[]string{
				"on-demand",
				"spot",
			},
			[]string{
				"on_demand",
				"spot",
			},
		},
		{
			"int32 enum",
			`
components:
  schemas:
    PowersOfTwo:
      enum:
      - 1
      - 2
      - 4
      type: integer
      format: int32
`,
			"PowersOfTwo",
			"powersOfTwo",
			"int32",
			[]string{
				"1",
				"2",
				"4",
			},
			[]string{
				"1",
				"2",
				"4",
			},
		},
	}
	for _, test := range tests {
		sh := testutil.NewSchemaHelperFromYAML(t, test.yaml)

		edefs, err := sh.GetEnumDefs()
		require.Nil(err)

		assert.Equal(1, len(edefs))

		edef := edefs[0]
		assert.Equal(test.expNameExported, edef.Names.GoExported)
		assert.Equal(test.expNameUnexported, edef.Names.GoUnexported)

		assert.Equal(len(test.expValuesOrig), len(edef.Values))
		assert.Equal(test.expValuesOrig, sortedOriginalValues(edef.Values))
		assert.Equal(test.expValuesClean, sortedCleanValues(edef.Values))
	}
}
