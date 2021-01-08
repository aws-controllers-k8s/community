// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	 http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package generate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-controllers-k8s/pkg/testutil"
)

func TestSQS_Queue(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "sqs")

	crds, err := g.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Queue", crds)
	require.NotNil(crd)

	assert.Equal("Queue", crd.Names.Camel)
	assert.Equal("queue", crd.Names.CamelLower)
	assert.Equal("queue", crd.Names.Snake)

	// The SQS Queue API has CD+L operations:
	//
	// * CreateQueue
	// * DeleteQueue
	// * ListQueues
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.ReadMany)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.GetAttributes)
	assert.NotNil(crd.Ops.SetAttributes)

	// But sadly, has no Update or ReadOne operation :(
	// There is, however, GetQueueUrl and GetQueueAttributes calls...
	assert.Nil(crd.Ops.ReadOne)
	assert.Nil(crd.Ops.Update)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		"ContentBasedDeduplication",
		"DelaySeconds",
		"FifoQueue",
		"KMSDataKeyReusePeriodSeconds",
		"KMSMasterKeyID",
		"MaximumMessageSize",
		"MessageRetentionPeriod",
		"Policy",
		"QueueName",
		"ReceiveMessageWaitTimeSeconds",
		"RedrivePolicy",
		"Tags",
		"VisibilityTimeout",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	expStatusFieldCamel := []string{
		// There are a set of Attribute map keys that are readonly
		// fields...
		"CreatedTimestamp",
		"LastModifiedTimestamp",
		// There is only a QueueURL field returned from CreateQueueResult shape
		"QueueURL",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))
}
