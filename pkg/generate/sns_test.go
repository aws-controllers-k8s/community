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

func TestSNS_Topic(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "sns")

	crds, err := g.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Topic", crds)
	require.NotNil(crd)

	assert.Equal("Topic", crd.Names.Camel)
	assert.Equal("topic", crd.Names.CamelLower)
	assert.Equal("topic", crd.Names.Snake)

	// The GetTopicAttributes operation has the following definition:
	//
	//    "GetTopicAttributes":{
	//      "name":"GetTopicAttributes",
	//      "http":{
	//        "method":"POST",
	//        "requestUri":"/"
	//      },
	//      "input":{"shape":"GetTopicAttributesInput"},
	//      "output":{
	//        "shape":"GetTopicAttributesResponse",
	//        "resultWrapper":"GetTopicAttributesResult"
	//      },
	//      "errors":[
	//        {"shape":"InvalidParameterException"},
	//        {"shape":"InternalErrorException"},
	//        {"shape":"NotFoundException"},
	//        {"shape":"AuthorizationErrorException"},
	//        {"shape":"InvalidSecurityException"}
	//      ]
	//    },
	//
	// Where the NotFoundException shape looks like this:
	//
	//    "NotFoundException":{
	//      "type":"structure",
	//      "members":{
	//        "message":{"shape":"string"}
	//      },
	//      "error":{
	//        "code":"NotFound",
	//        "httpStatusCode":404,
	//        "senderFault":true
	//      },
	//      "exception":true
	//    },
	//
	// So, we expect that the normal logic in CRD.ExceptionCode(404)
	// identifies the above as the 404 Not Found error with a code of
	// "NotFound"
	assert.Equal("NotFound", crd.ExceptionCode(404))

	// The SNS Topic API is a little weird. There are Create and Delete
	// operations ("CreateTopic", "DeleteTopic") but there is no ReadOne
	// operation (there is a "GetTopicAttributes" call though) or Update
	// operation (there is a "SetTopicAttributes" call though). And there is a
	// ReadMany operation (ListTopics)
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.ReadMany)
	assert.NotNil(crd.Ops.GetAttributes)
	assert.NotNil(crd.Ops.SetAttributes)

	assert.Nil(crd.Ops.ReadOne)
	assert.Nil(crd.Ops.Update)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	// The SNS Topic uses an "Attributes" map[string]*string to masquerade
	// real fields. DeliveryPolicy, Policy, KMSMasterKeyID and DisplayName are
	// all examples of these masqueraded fields...
	expSpecFieldCamel := []string{
		"DeliveryPolicy",
		"DisplayName",
		"KMSMasterKeyID",
		"Name",
		"Policy",
		"Tags",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	// The SNS Topic uses an "Attributes" map[string]*string to masquerade
	// real fields. EffectiveDeliveryPolicy and Owner are
	// examples of these masqueraded fields that are ReadOnly and thus belong
	// in the Status struct
	expStatusFieldCamel := []string{
		// "TopicARN" is in the output shape for CreateTopic, but it is removed
		// because it is the primary resource object's ARN field and the
		// SDKMapper has identified it as the source for the standard
		// Status.ACKResourceMetadata.ARN field
		"EffectiveDeliveryPolicy",
		"Owner",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))
}
