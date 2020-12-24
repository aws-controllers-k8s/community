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

func TestLambda_Function(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "lambda")

	crds, err := g.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Function", crds)
	require.NotNil(crd)

	assert.Equal("Function", crd.Names.Camel)
	assert.Equal("function", crd.Names.CamelLower)
	assert.Equal("function", crd.Names.Snake)

	// The Lambda Function API has Create, Delete, ReadOne and ReadMany
	// operations, however has no single Update operation. Instead, there are
	// multiple Update operations, depending on the attributes of the function
	// being changed...
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.ReadOne)
	assert.NotNil(crd.Ops.ReadMany)

	assert.Nil(crd.Ops.GetAttributes)
	assert.Nil(crd.Ops.SetAttributes)
	assert.Nil(crd.Ops.Update)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		"Code",
		"DeadLetterConfig",
		"Description",
		"Environment",
		"FileSystemConfigs",
		"FunctionName",
		"Handler",
		"KMSKeyARN",
		"Layers",
		"MemorySize",
		"Publish",
		"Role",
		"Runtime",
		"Tags",
		"Timeout",
		"TracingConfig",
		"VPCConfig",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	expStatusFieldCamel := []string{
		// Added from generator.yaml
		"CodeLocation",
		"CodeRepositoryType",
		"CodeSHA256",
		"CodeSize",
		// "FunctionArn", <-- ACKMetadata.ARN
		"LastModified",
		"LastUpdateStatus",
		"LastUpdateStatusReason",
		"LastUpdateStatusReasonCode",
		"MasterARN",
		"RevisionID",
		"State",
		"StateReason",
		"StateReasonCode",
		"Version",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))
}
