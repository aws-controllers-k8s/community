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

func TestECRRepository(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "ecr")

	crds, err := g.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Repository", crds)
	require.NotNil(crd)

	// The ECR Repository API has just the C and R of the normal CRUD
	// operations:
	//
	// * CreateRepository
	// * DeleteRepository
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)

	// There is no DescribeRepository operation. There is a List operation for
	// Repositories, though: DescribeRepositories
	assert.Nil(crd.Ops.ReadOne)
	assert.NotNil(crd.Ops.ReadMany)

	// There is no update operation (you need to call various SetXXX operations
	// on the Repository's components
	assert.Nil(crd.Ops.Update)

	// The DescribeRepositories operation has the following definition:
	//
	//    "DescribeRepositories":{
	//      "name":"DescribeRepositories",
	//      "http":{
	//        "method":"POST",
	//        "requestUri":"/"
	//      },
	//      "input":{"shape":"DescribeRepositoriesRequest"},
	//      "output":{"shape":"DescribeRepositoriesResponse"},
	//      "errors":[
	//        {"shape":"ServerException"},
	//        {"shape":"InvalidParameterException"},
	//        {"shape":"RepositoryNotFoundException"}
	//      ]
	//    },
	//
	// NOTE: This is UNUSUAL for a List operation to return a 404 Not Found.
	// Typically a return of zero results for a List operation results in a 200
	// OK.
	//
	// Where the RepositoryNotFoundException shape looks like this:
	//
	//    "RepositoryNotFoundException":{
	//      "type":"structure",
	//      "members":{
	//        "message":{"shape":"ExceptionMessage"}
	//      },
	//      "exception":true
	//    },
	//
	// Which does not indicate that the error is a 404 :( So, the logic in the
	// CRD.ExceptionCode(404) method needs to get its override from the
	// generate.yaml configuration file.
	assert.Equal("RepositoryNotFoundException", crd.ExceptionCode(404))

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	// The ECR API uses a REST-like API that uses "wrapper" single-member
	// objects in the JSON response for the create/describe calls. In other
	// words, the returned result from the CreateRepository API looks like
	// this:
	//
	// {
	//   "repository": {
	//	 .. bunch of fields for the repository ..
	//   }
	// }
	//
	// This test is verifying that we're properly "unwrapping" the structs and
	// putting the repository object's fields into the Spec and Status for the
	// Repository CRD.
	expSpecFieldCamel := []string{
		"ImageScanningConfiguration",
		"ImageTagMutability",
		"RepositoryName",
		"Tags",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	expStatusFieldCamel := []string{
		"CreatedAt",
		// "ImageScanningConfiguration" removed because it is contained in the
		// input shape and therefore exists in the Spec
		// "ImageTagMutability" removed because it is contained in the input
		// shape and therefore exists in the Spec
		"RegistryID",
		// "RepositoryARN" removed because it refers to the primary object's
		// ARN and the SDKMapper identified it for mapping to the standard
		// Status.ACKResourceMetadata.ARN field
		// "RepositoryName" removed because it is contained in the input shape
		// and therefore exists in the Spec
		"RepositoryURI",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))
}
