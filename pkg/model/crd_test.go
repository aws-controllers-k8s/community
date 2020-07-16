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

	"github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/aws/aws-controllers-k8s/pkg/testutil"
)

func attrCamelNames(fields map[string]*model.CRDField) []string {
	res := []string{}
	for _, attr := range fields {
		res = append(res, attr.Names.Camel)
	}
	sort.Strings(res)
	return res
}

func getCRDByName(name string, crds []*model.CRD) *model.CRD {
	for _, c := range crds {
		if c.Names.Original == name {
			return c
		}
	}
	return nil
}

func TestSNSTopic(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperForService(t, "sns")

	crds, err := sh.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Topic", crds)
	require.NotNil(crd)

	assert.Equal("Topic", crd.Names.Camel)
	assert.Equal("topic", crd.Names.CamelLower)
	assert.Equal("topic", crd.Names.Snake)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		"Attributes",
		"Name",
		"Tags",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	// None of the fields in the Topic resource's CreateTopicInput shape are
	// returned in the CreateTopicOutput shape, so none of them return any Go
	// code for setting a Status struct field to a corresponding Create Output
	// Shape member
	nameField := specFields["Name"]
	nameFieldGoCode := nameField.GoCodeSetFieldFromOutput(model.OpTypeCreate)
	assert.Equal("", nameFieldGoCode)

	expStatusFieldCamel := []string{
		// "TopicARN" is the only field in the output shape for CreateTopic,
		// but it is removed because it is the primary resource object's ARN
		// field and the SDKMapper has identified it as the source for the
		// standard Status.ACKResourceMetadata.ARN field
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))

	// The SNS Topic API is a little weird. There are Create and Delete
	// operations ("CreateTopic", "DeleteTopic") but there is no ReadOne
	// operation (there is a "GetTopicAttributes" call though) or Update
	// operation (there is a "SetTopicAttributes" call though). And there is a
	// ReadMany operation (ListTopics)
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.ReadMany)

	assert.Nil(crd.Ops.ReadOne)
	assert.Nil(crd.Ops.Update)
}

func TestEC2LaunchTemplate(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperForService(t, "ec2")

	crds, err := sh.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("LaunchTemplate", crds)
	require.NotNil(crd)

	assert.Equal("LaunchTemplate", crd.Names.Camel)
	assert.Equal("launchTemplate", crd.Names.CamelLower)
	assert.Equal("launch_template", crd.Names.Snake)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		// TODO(jaypipes): DryRun and ClientToken are examples of two fields in
		// the resource input shape that need to be stripped out of the CRD. We
		// need to instruct the code generator that these types of fields are
		// not germane to the resource itself...
		"ClientToken",
		"DryRun",
		"LaunchTemplateData",
		"LaunchTemplateName",
		// TODO(jaypipes): Here's an example of where we need to instruct the
		// code generator to rename the "TagSpecifications" field to simply
		// "Tags" and place it into the common Spec.Tags field.
		"TagSpecifications",
		"VersionDescription",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	// LaunchTemplateName is in the LaunchTemplate resource's CreateTopicInput shape and also
	// returned in the CreateLaunchTemplateResult shape, so it should have
	// Go code to set the Input Shape member from the Spec field but not set a
	// Status field from the Create Output Shape member
	nameField := specFields["LaunchTemplateName"]
	nameFieldGoCodeInputShape := nameField.GoCodeSetInputFromField(model.OpTypeCreate)
	assert.Equal("res.LaunchTemplateName = r.ko.Spec.LaunchTemplateName", nameFieldGoCodeInputShape)

	expStatusFieldCamel := []string{
		"CreateTime",
		"CreatedBy",
		"DefaultVersionNumber",
		"LatestVersionNumber",
		// TODO(jaypipes): Handle "Id" Fields like "LaunchTemplateId" in the
		// same way as we handle ARN-ified modern service APIs and use the
		// SDKMapper to instruct the code generator that this field represents
		// the primary resource object's identifier field.
		"LaunchTemplateID",
		// LaunchTemplateName excluded because it matches input shape.,
		// TODO(jaypipes): Tags field should be excluded because it is the same
		// as the input shape's "TagSpecifications" field...
		"Tags",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))

	// Check that we properly determined how to find the CreatedBy attribute
	// within the CreateLaunchTemplateResult shape, which has a single field called
	// "LaunchTemplate" that contains the CreatedBy field.
	createdByField := statusFields["CreatedBy"]
	createdByFieldOutputCode := createdByField.GoCodeSetFieldFromOutput(model.OpTypeCreate)
	assert.Equal("ko.Status.CreatedBy = resp.LaunchTemplate.CreatedBy", createdByFieldOutputCode)

	// The EC2 LaunchTemplate API has a "normal" set of CUD operations:
	//
	// * CreateLaunchTemplate
	// * ModifyLaunchTemplate
	// * DeleteLaunchTemplate
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.Update)

	// However, oddly, there is no ReadOne operation. There is only a
	// ReadMany/List operation "DescribeLaunchTemplates" :(
	//
	// TODO(jaypipes): Develop strategy for informing the code generator via
	// the SDKMapper that certain APIs don't have ReadOne but only ReadMany
	// APIs...
	assert.Nil(crd.Ops.ReadOne)
	assert.NotNil(crd.Ops.ReadMany)
}

func TestECRRepository(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperForService(t, "ecr")

	crds, err := sh.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Repository", crds)
	require.NotNil(crd)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	// The ECR API uses a REST-like API that uses "wrapper" single-member
	// objects in the JSON response for the create/describe calls. In other
	// words, the returned result from the CreateRepository API looks like
	// this:
	//
	// {
	//   "repository": {
	//     .. bunch of fields for the repository ..
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

	// ImageScanningConfiguration is in the Repository resource's
	// CreateRepositoryInput shape and also returned in the
	// CreateRepositoryOutput shape, so it should produce Go code to set the
	// appropriate input shape member.
	iscField := specFields["ImageScanningConfiguration"]
	iscFieldGoCodeInputShape := iscField.GoCodeSetInputFromField(model.OpTypeCreate)
	assert.Equal("res.ImageScanningConfiguration = r.ko.Spec.ImageScanningConfiguration", iscFieldGoCodeInputShape)

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

	// Check that we properly determined how to find the RegistryID attribute
	// within the CreateRepositoryOutput shape, which has a single field called
	// "Repository" that contains the RegistryId field.
	regIDField := statusFields["RegistryId"]
	regIDFieldOutputCode := regIDField.GoCodeSetFieldFromOutput(model.OpTypeCreate)
	assert.Equal("ko.Status.RegistryID = resp.Repository.RegistryId", regIDFieldOutputCode)

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
}

func TestCodeDeployDeployment(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperForService(t, "codedeploy")

	crds, err := sh.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Deployment", crds)
	require.NotNil(crd)

	assert.Equal("Deployment", crd.Names.Camel)
	assert.Equal("deployment", crd.Names.CamelLower)
	assert.Equal("deployment", crd.Names.Snake)

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

	// None of the fields in the Topic resource's CreateTopicInput shape are
	// returned in the CreateTopicOutput shape, so none of them return any Go
	// code for setting a Status struct field to a corresponding Create Output
	// Shape member
	nameField := specFields["ApplicationName"]
	nameFieldGoCodeCreate := nameField.GoCodeSetFieldFromOutput(model.OpTypeCreate)
	assert.Equal("", nameFieldGoCodeCreate)

	// However, all of the fields in the Deployment resource's
	// CreateDeploymentInput shape are returned in the GetDeploymentOutput
	// shape, and there is a DeploymentInfo wrapper struct for the output
	// shape, so the readOne accessor contains the wrapper struct's name.
	nameFieldGoCodeReadOne := nameField.GoCodeSetFieldFromOutput(model.OpTypeGet)
	assert.Equal("ko.Spec.ApplicationName = resp.DeploymentInfo.ApplicationName", nameFieldGoCodeReadOne)

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
