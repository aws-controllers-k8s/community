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

	"github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/aws/aws-controllers-k8s/pkg/testutil"
)

func TestCodeDeploy_Deployment(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "codedeploy")

	crds, err := g.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Deployment", crds)
	require.NotNil(crd)

	assert.Equal("Deployment", crd.Names.Camel)
	assert.Equal("deployment", crd.Names.CamelLower)
	assert.Equal("deployment", crd.Names.Snake)

	// The GetDeployment operation has the following definition:
	//
	//    "GetDeployment":{
	//      "name":"GetDeployment",
	//      "http":{
	//        "method":"POST",
	//        "requestUri":"/"
	//      },
	//      "input":{"shape":"GetDeploymentInput"},
	//      "output":{"shape":"GetDeploymentOutput"},
	//      "errors":[
	//        {"shape":"DeploymentIdRequiredException"},
	//        {"shape":"InvalidDeploymentIdException"},
	//        {"shape":"DeploymentDoesNotExistException"}
	//      ]
	//    },
	//
	// Where the DeploymentDoesNotExistException shape looks like this:
	//
	//    "DeploymentDoesNotExistException":{
	//      "type":"structure",
	//      "members":{
	//      },
	//      "exception":true
	//    },
	//
	// Which does not indicate that the error is a 404 :( So, the logic in the
	// CRD.ExceptionCode(404) method needs to get its override from the
	// generate.yaml configuration file.
	assert.Equal("DeploymentDoesNotExistException", crd.ExceptionCode(404))

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		"ApplicationName",
		"AutoRollbackConfiguration",
		"DeploymentConfigName",
		"DeploymentGroupName",
		"Description",
		"FileExistsBehavior",
		"IgnoreApplicationStopFailures",
		"Revision",
		"TargetInstances",
		"UpdateOutdatedInstancesOnly",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	// However, all of the fields in the Deployment resource's
	// CreateDeploymentInput shape are returned in the GetDeploymentOutput
	// shape, and there is a DeploymentInfo wrapper struct for the output
	// shape, so the readOne accessor contains the wrapper struct's name.
	expCreateOutput := `
	if resp.DeploymentId != nil {
		ko.Status.DeploymentID = resp.DeploymentId
	}
`
	assert.Equal(expCreateOutput, crd.GoCodeSetOutput(model.OpTypeCreate, "resp", "ko", 1))

	expStatusFieldCamel := []string{
		// All of the fields in the Deployment resource's CreateDeploymentInput
		// shape are returned in the CreateDeploymentOutput shape so there are
		// not Status fields
		//
		// There is a DeploymentID field in addition to the Spec fields, though.
		"DeploymentID",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))

	// The CodeDeploy Deployment API actually CR+L operations:
	//
	// * CreateDeployment
	// * GetDeployment
	// * ListDeployments
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.ReadOne)
	assert.NotNil(crd.Ops.ReadMany)

	// But sadly, has no Update or Delete operation :(
	assert.Nil(crd.Ops.Update)
	assert.Nil(crd.Ops.Delete)
}
