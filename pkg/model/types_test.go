package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-controllers-k8s/pkg/model"
)

func TestReplacePkgName(t *testing.T) {
	assert := assert.New(t)
	testCases := []struct {
		subject         string
		pkgName         string
		replacePkgAlias string
		keepPointer     bool
		want            string
	}{
		{ // most frequent case
			"*ecr.Repository",
			"ecr",
			"svcsdk",
			true,
			"*svcsdk.Repository",
		},
		{ // don't keep pointer
			"*ecr.Repository",
			"ecr",
			"svcsdk",
			false,
			"svcsdk.Repository",
		},
		{ // non sdk type
			"*time.Time",
			"ecr",
			"svcsdk",
			true,
			"*time.Time",
		},
		{ // map type
			"map[string]*ecr.Repository",
			"ecr",
			"svcsdk",
			true,
			"map[string]*svcsdk.Repository",
		},
		{ // nested map type
			"map[string]map[string]uint8",
			"ec2",
			"svcsdk",
			true,
			"map[string]map[string]uint8",
		},
		{ // slice type
			"[]ecr.Repository",
			"ecr",
			"svcsdk",
			true,
			"[]svcsdk.Repository",
		},
		{ // nested slices type
			"[][]*codedeploy.EC2TagFilter",
			"codedeploy",
			"svcsdk",
			true,
			"[][]*svcsdk.EC2TagFilter",
		},
	}

	for _, tc := range testCases {
		result := model.ReplacePkgName(
			tc.subject,
			tc.pkgName,
			tc.replacePkgAlias,
			tc.keepPointer,
		)
		assert.Equal(tc.want, result)
	}
}
