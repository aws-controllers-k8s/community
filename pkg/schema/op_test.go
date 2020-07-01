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
	"testing"

	"github.com/aws/aws-service-operator-k8s/pkg/schema"
	"github.com/stretchr/testify/assert"
)

func TestGetOpTypeAndResourceNameFromOpID(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		opID       string
		expOpType  schema.OpType
		expResName string
	}{
		{
			"CreateTopic",
			schema.OpTypeCreate,
			"Topic",
		},
		{
			"CreateOrUpdateTopic",
			schema.OpTypeReplace,
			"Topic",
		},
		{
			"CreateBatchTopics",
			schema.OpTypeCreateBatch,
			"Topic",
		},
		{
			"CreateBatchTopic",
			schema.OpTypeCreateBatch,
			"Topic",
		},
		{
			"BatchCreateTopics",
			schema.OpTypeCreateBatch,
			"Topic",
		},
		{
			"BatchCreateTopic",
			schema.OpTypeCreateBatch,
			"Topic",
		},
		{
			"CreateTopics",
			schema.OpTypeCreateBatch,
			"Topic",
		},
		{
			"DescribeEC2Instances",
			schema.OpTypeList,
			"EC2Instance",
		},
		{
			"DescribeEC2Instance",
			schema.OpTypeGet,
			"EC2Instance",
		},
		{
			"UpdateTopic",
			schema.OpTypeUpdate,
			"Topic",
		},
		{
			"DeleteTopic",
			schema.OpTypeDelete,
			"Topic",
		},
		{
			"PauseEC2Instance",
			schema.OpTypeUnknown,
			"",
		},
	}
	for _, test := range tests {
		ot, resName := schema.GetOpTypeAndResourceNameFromOpID(test.opID)
		assert.Equal(test.expOpType, ot, test.opID)
		assert.Equal(test.expResName, resName, test.opID)
	}
}
