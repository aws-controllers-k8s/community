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
	"io/ioutil"
	"path/filepath"
	"sort"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-service-operator-k8s/pkg/model"
	"github.com/aws/aws-service-operator-k8s/pkg/schema"
)

func loadSwagger(t *testing.T, yamlFile string) *openapi3.Swagger {
	path := filepath.Join("testdata", yamlFile)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var jsonb []byte

	if jsonb, err = yaml.YAMLToJSON(b); err != nil {
		t.Fatal(err)
	}

	api, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(jsonb)
	if err != nil {
		t.Fatal(err)
	}
	return api
}

func attrExportedNames(attrs map[string]*model.Attr) []string {
	res := []string{}
	for _, attr := range attrs {
		res = append(res, attr.Names.GoExported)
	}
	sort.Strings(res)
	return res
}

func TestGetCRDs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	api := loadSwagger(t, "topic-api.yaml")
	sh := schema.NewHelper(api)

	crds, err := sh.GetCRDs()
	require.Nil(err)

	assert.Equal(1, len(crds))

	topicCRD := crds[0]

	assert.Equal("Topic", topicCRD.Names.GoExported)
	assert.Equal("topic", topicCRD.Names.GoUnexported)

	specAttrs := topicCRD.SpecAttrs
	statusAttrs := topicCRD.StatusAttrs

	assert.Equal(3, len(specAttrs))
	expSpecAttrExported := []string{
		"Attributes",
		"Name",
		"Tags",
	}
	assert.Equal(expSpecAttrExported, attrExportedNames(specAttrs))
	assert.Equal(1, len(statusAttrs))
	expStatusAttrExported := []string{
		"TopicARN",
	}
	assert.Equal(expStatusAttrExported, attrExportedNames(statusAttrs))
}
