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

package testutil

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"

	"github.com/aws/aws-service-operator-k8s/pkg/schema"
)

func NewSchemaHelperFromFile(t *testing.T, yamlFile string) *schema.Helper {
	path := filepath.Join("testdata", yamlFile)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return NewSchemaHelperFromYAML(t, string(b))
}

func NewSchemaHelperFromYAML(t *testing.T, yamlContents string) *schema.Helper {
	jsonb, err := yaml.YAMLToJSON([]byte(yamlContents))
	if err != nil {
		t.Fatal(err)
	}
	api, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(jsonb)
	if err != nil {
		t.Fatal(err)
	}
	return schema.NewHelper(api, nil)
}
