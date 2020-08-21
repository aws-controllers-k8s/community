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
