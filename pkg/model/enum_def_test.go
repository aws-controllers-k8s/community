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

func getEnumDefByName(name string, enumDefs []*model.EnumDef) *model.EnumDef {
	for _, e := range enumDefs {
		if e.Names.Original == name {
			return e
		}
	}
	return nil
}

func TestEnumDefs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tests := []struct {
		name              string
		service           string
		expNameCamel      string
		expNameCamelLower string
		expValuesOrig     []string
		expValuesClean    []string
	}{
		{
			"original same as clean value",
			"ecr",
			"ScanStatus",
			"scanStatus",
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
			"ec2",
			"InstanceLifecycle",
			"instanceLifecycle",
			[]string{
				"on-demand",
				"spot",
			},
			[]string{
				"on_demand",
				"spot",
			},
		},
	}
	for _, test := range tests {
		sh := testutil.NewSchemaHelperForService(t, test.service)

		edefs, err := sh.GetEnumDefs()
		require.Nil(err)

		edef := getEnumDefByName(test.expNameCamel, edefs)
		require.NotNil(edef)

		assert.Equal(test.expNameCamelLower, edef.Names.CamelLower)

		assert.Equal(len(test.expValuesOrig), len(edef.Values))
		assert.Equal(test.expValuesOrig, sortedOriginalValues(edef.Values))
		assert.Equal(test.expValuesClean, sortedCleanValues(edef.Values))
	}
}
