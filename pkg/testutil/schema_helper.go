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
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	"github.com/aws/aws-controllers-k8s/pkg/model"
)

func NewGeneratorForService(t *testing.T, serviceAlias string) *generate.Generator {
	path := filepath.Clean("testdata")
	sdkHelper := model.NewSDKHelper(path)
	sdkAPI, err := sdkHelper.API(serviceAlias)
	if err != nil {
		t.Fatal(err)
	}
	generatorConfigPath := filepath.Join(path, "models", "apis", serviceAlias, "0000-00-00", "generator.yaml")
	if _, err := os.Stat(generatorConfigPath); os.IsNotExist(err) {
		generatorConfigPath = ""
	}
	g, err := generate.New(sdkAPI, "v1alpha1", generatorConfigPath, "")
	if err != nil {
		t.Fatal(err)
	}
	return g
}
