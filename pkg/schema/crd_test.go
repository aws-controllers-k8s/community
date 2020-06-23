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

func TestGetCRDs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperFromFile(t, "topic-api.yaml")

	crds, err := sh.GetCRDs()
	require.Nil(err)

	assert.Equal(1, len(crds))

	topicCRD := crds[0]

	assert.Equal("Topic", topicCRD.Names.Camel)
	assert.Equal("topic", topicCRD.Names.CamelLower)
	assert.Equal("topic", topicCRD.Names.Snake)

	specAttrs := topicCRD.SpecAttrs
	statusAttrs := topicCRD.StatusAttrs

	assert.Equal(3, len(specAttrs))
	expSpecAttrCamel := []string{
		"Attributes",
		"Name",
		"Tags",
	}
	assert.Equal(expSpecAttrCamel, attrCamelNames(specAttrs))
	assert.Equal(1, len(statusAttrs))
	expStatusAttrCamel := []string{
		"TopicARN",
	}
	assert.Equal(expStatusAttrCamel, attrCamelNames(statusAttrs))
}
