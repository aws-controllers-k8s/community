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

package model_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-controllers-k8s/pkg/model"
)

var (
	lambda *model.SDKAPI
)

func lambdaSDKAPI(t *testing.T) *model.SDKAPI {
	if lambda != nil {
		return lambda
	}
	path := filepath.Clean("../generate/testdata")
	sdkHelper := model.NewSDKHelper(path)
	lambda, err := sdkHelper.API("lambda")
	if err != nil {
		t.Fatal(err)
	}
	return lambda
}

func TestGetInputShapeRef(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	description := "Description"
	s3Key := "S3Key"

	tests := []struct {
		opID            string
		path            string
		expShapeRefName *string
		expFound        bool
	}{
		{
			// non-nested path search
			"CreateFunction",
			"Description",
			&description,
			true,
		},
		{
			// nested path search
			"CreateFunction",
			"Code.S3Key",
			&s3Key,
			true,
		},
		{
			// no such op
			"CreateNonexisting",
			"Foo",
			nil,
			false,
		},
		{
			// no such member
			"CreateFunction",
			"Foo",
			nil,
			false,
		},
		{
			// no such member path
			"CreateFunction",
			"Code.Foo",
			nil,
			false,
		},
	}
	api := lambdaSDKAPI(t)
	for _, test := range tests {
		got, found := api.GetInputShapeRef(test.opID, test.path)
		require.Equal(test.expFound, found, test.path)
		if test.expShapeRefName == nil {
			assert.Nil(got)
		} else {
			assert.Equal(*test.expShapeRefName, got.ShapeName)
		}
	}
}

func TestGetOutputShapeRef(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tags := "Tags"
	stringshape := "String"

	tests := []struct {
		opID            string
		path            string
		expShapeRefName *string
		expFound        bool
	}{
		{
			// non-nested path search
			"GetFunction",
			"Tags",
			&tags,
			true,
		},
		{
			// nested path search
			"GetFunction",
			"Code.Location",
			// Note that unlike CreateFunctionRequest.Description above, which
			// is a Shape called `Description` that is itself a Go string type,
			// the name of the `GetFunctionResponse.Code.Location` Shape is
			// actually called `String`. Yep... this is why we can't have nice
			// things.
			&stringshape,
			true,
		},
		{
			// no such op
			"GetNonexisting",
			"Foo",
			nil,
			false,
		},
		{
			// no such member
			"GetFunction",
			"Foo",
			nil,
			false,
		},
		{
			// no such member path
			"GetFunction",
			"Code.Foo",
			nil,
			false,
		},
	}
	api := lambdaSDKAPI(t)
	for _, test := range tests {
		got, found := api.GetOutputShapeRef(test.opID, test.path)
		require.Equal(test.expFound, found, test.path)
		if test.expShapeRefName == nil {
			assert.Nil(got)
		} else {
			assert.Equal(*test.expShapeRefName, got.ShapeName)
		}
	}
}
