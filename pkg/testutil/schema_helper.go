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
	"path/filepath"
	"testing"

	"github.com/aws/aws-controllers-k8s/pkg/model"
)

func NewSchemaHelperForService(t *testing.T, serviceAlias string) *model.Helper {
	path := filepath.Clean("testdata")
	sdkHelper := model.NewSDKHelper(path)
	sdkAPI, err := sdkHelper.API(serviceAlias)
	if err != nil {
		t.Fatal(err)
	}
	return model.NewHelper(sdkAPI)
}
