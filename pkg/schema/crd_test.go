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

func attrCamelNames(attrs map[string]*model.Attr) []string {
	res := []string{}
	for _, attr := range attrs {
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

	specAttrs := crd.SpecAttrs
	statusAttrs := crd.StatusAttrs

	expSpecAttrCamel := []string{
		"Attributes",
		"Name",
		"Tags",
	}
	assert.Equal(expSpecAttrCamel, attrCamelNames(specAttrs))
	expStatusAttrCamel := []string{
		// "TopicARN" is the only field in the output shape for CreateTopic,
		// but it is removed because it is the primary resource object's ARN
		// field and the SDKMapper has identified it as the source for the
		// standard Status.ACKResourceMetadata.ARN field
	}
	assert.Equal(expStatusAttrCamel, attrCamelNames(statusAttrs))

	// The SNS Topic API is a little weird. There are Create and Delete
	// operations ("CreateTopic", "DeleteTopic") but there is no ReadOne
	// operation (there is a "GetTopicAttributes" call though) or Update
	// operation (there is a "SetTopicAttributes" call though)
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)

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

	specAttrs := crd.SpecAttrs
	statusAttrs := crd.StatusAttrs

	expSpecAttrCamel := []string{
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
	assert.Equal(expSpecAttrCamel, attrCamelNames(specAttrs))
	expStatusAttrCamel := []string{
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
	assert.Equal(expStatusAttrCamel, attrCamelNames(statusAttrs))

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
}

func TestECRRepository(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperForService(t, "ecr")

	crds, err := sh.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Repository", crds)
	require.NotNil(crd)

	specAttrs := crd.SpecAttrs
	statusAttrs := crd.StatusAttrs

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
	expSpecAttrCamel := []string{
		"ImageScanningConfiguration",
		"ImageTagMutability",
		"RepositoryName",
		"Tags",
	}
	assert.Equal(expSpecAttrCamel, attrCamelNames(specAttrs))
	expStatusAttrCamel := []string{
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
	assert.Equal(expStatusAttrCamel, attrCamelNames(statusAttrs))
}
