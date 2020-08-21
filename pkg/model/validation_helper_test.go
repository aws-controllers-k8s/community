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
	"testing"

	"github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/stretchr/testify/assert"
)

type MockErrInvalidParams struct {
	error   string
	message string
}

func (mockErrInvalidParams MockErrInvalidParams) Error() string {
	return mockErrInvalidParams.error
}

func (mockErrInvalidParams MockErrInvalidParams) Message() string {
	return mockErrInvalidParams.message
}

func TestIsValidationErrorIgnorable(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name                string
		ignorableFieldNames []string
		errInvalidParams    MockErrInvalidParams
		expected            bool
	}{
		{
			name:                "No status fields are required",
			ignorableFieldNames: []string{},
			errInvalidParams: MockErrInvalidParams{
				error:   "",
				message: "",
			},
			expected: false,
		},
		{
			name:                "invalidParamError has unexpected content",
			ignorableFieldNames: []string{"fieldA"},
			errInvalidParams: MockErrInvalidParams{
				error:   "some invalid content",
				message: "some invalid content",
			},
			expected: false,
		},
		{
			name:                "invalidParamError has more errors than required status fields",
			ignorableFieldNames: []string{"GetApiInput.ApiId"},
			errInvalidParams: MockErrInvalidParams{
				error:   "- missing required field, GetApiInput.ApiId.\n- missing required field, GetApiInput.Dummy",
				message: "2 validation error(s) found.",
			},
			expected: false,
		},
		{
			name:                "invalidParamError less more errors than required status fields",
			ignorableFieldNames: []string{"GetApiInput.ApiId", "GetApiInput.Dummy"},
			errInvalidParams: MockErrInvalidParams{
				error:   "- missing required field, GetApiInput.ApiId.",
				message: "1 validation error(s) found.",
			},
			expected: false,
		},
		{
			name:                "invalidParamsError does not contain all required status fields",
			ignorableFieldNames: []string{"GetApiInput.ApiId"},
			errInvalidParams: MockErrInvalidParams{
				error:   "- missing required field, GetApiInput.Dummy.",
				message: "1 validation error(s) found.",
			},
			expected: false,
		},
		{
			name:                "invalidParamsError contains all required status fields. Test Case Insensitivity",
			ignorableFieldNames: []string{"GetApiInput.APIID"},
			errInvalidParams: MockErrInvalidParams{
				error:   "- missing required field, GetApiInput.apiid.",
				message: "1 validation error(s) found.",
			},
			expected: true,
		},
		{
			name:                "invalidParamsError contains all required status fields. Test multiple parameters.",
			ignorableFieldNames: []string{"GetApiInput.ApiId", "GetApiInput.Dummy"},
			errInvalidParams: MockErrInvalidParams{
				error:   "- missing required field, GetApiInput.ApiId.\n- missing required field, GetApiInput.Dummy",
				message: "2 validation error(s) found.",
			},
			expected: true,
		},
		{
			name:                "invalidParamsError contains all required status fields. Happy Case",
			ignorableFieldNames: []string{"GetApiInput.ApiId"},
			errInvalidParams: MockErrInvalidParams{
				error:   "- missing required field, GetApiInput.ApiId.",
				message: "1 validation error(s) found.",
			},
			expected: true,
		},
	}

	validationHelper := model.ValidationHelper{}

	for _, test := range tests {
		assert.Equal(test.expected, validationHelper.IsValidationErrorIgnorable(test.ignorableFieldNames, test.errInvalidParams))
	}
}
