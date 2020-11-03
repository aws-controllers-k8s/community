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

	"github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestGetOpTypeAndResourceNameFromOpID(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		opID       string
		expOpType  model.OpType
		expResName string
	}{
		{
			"CreateTopic",
			model.OpTypeCreate,
			"Topic",
		},
		{
			"CreateOrUpdateTopic",
			model.OpTypeReplace,
			"Topic",
		},
		{
			"CreateBatchTopics",
			model.OpTypeCreateBatch,
			"Topic",
		},
		{
			"CreateBatchTopic",
			model.OpTypeCreateBatch,
			"Topic",
		},
		{
			"BatchCreateTopics",
			model.OpTypeCreateBatch,
			"Topic",
		},
		{
			"BatchCreateTopic",
			model.OpTypeCreateBatch,
			"Topic",
		},
		{
			"CreateTopics",
			model.OpTypeCreateBatch,
			"Topic",
		},
		{
			"DescribeEC2Instances",
			model.OpTypeList,
			"EC2Instance",
		},
		{
			"DescribeEC2Instance",
			model.OpTypeGet,
			"EC2Instance",
		},
		{
			"UpdateTopic",
			model.OpTypeUpdate,
			"Topic",
		},
		{
			"DeleteTopic",
			model.OpTypeDelete,
			"Topic",
		},
		{
			"DescribeInstances",
			model.OpTypeList,
			"Instance",
		},
		{
			"ListDeploymentGroups",
			model.OpTypeList,
			"DeploymentGroup",
		},
		{
			"GetDeployment",
			model.OpTypeGet,
			"Deployment",
		},
		{
			"PauseEC2Instance",
			model.OpTypeUnknown,
			"PauseEC2Instance",
		},
	}
	for _, test := range tests {
		ot, resName := model.GetOpTypeAndResourceNameFromOpID(test.opID)
		assert.Equal(test.expOpType, ot, test.opID)
		assert.Equal(test.expResName, resName, test.opID)
	}
}
