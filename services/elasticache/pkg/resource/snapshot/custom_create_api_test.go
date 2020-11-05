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

package snapshot

import (
	"context"
	svcapitypes "github.com/aws/aws-controllers-k8s/services/elasticache/apis/v1alpha1"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Helper methods to setup tests
// provideResourceManager returns pointer to resourceManager
func provideResourceManager() *resourceManager {
	return &resourceManager{
		rr:           nil,
		awsAccountID: "",
		awsRegion:    "",
		sess:         nil,
		sdkapi:       nil,
	}
}

// provideResource returns pointer to resource
func provideResource() *resource {
	return &resource{
		ko: &svcapitypes.Snapshot{},
	}
}

func Test_CustomCreateSnapshot_NotCopySnapshot(t *testing.T) {
	assert := assert.New(t)
	// Setup
	rm := provideResourceManager()

	desired := provideResource()

	var ctx context.Context

	res, err := rm.CustomCreateSnapshot(ctx, desired)
	assert.Nil(res)
	assert.Nil(err)
}

func Test_CustomCreateSnapshot_InvalidParam(t *testing.T) {
	assert := assert.New(t)
	// Setup
	rm := provideResourceManager()
	desired := provideResource()
	sourceSnapshotName := "test-rg-backup"
	rgId := "rgId"
	desired.ko.Spec = svcapitypes.SnapshotSpec{SourceSnapshotName: &sourceSnapshotName,
		ReplicationGroupID: &rgId}
	var ctx context.Context

	res, err := rm.CustomCreateSnapshot(ctx, desired)
	assert.Nil(res)
	assert.NotNil(err)
	assert.Equal(err.Error(), "InvalidParameterCombination: Cannot specify CacheClusteId or ReplicationGroupId while SourceSnapshotName is specified")
}
