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

package names_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-service-operator-k8s/pkg/names"
)

func TestNames(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		original         string
		expectExported   string
		expectUnexported string
	}{
		{"Identifier", "Identifier", "identifier"},
		{"Id", "ID", "id"},
		{"ID", "ID", "id"},
		{"KeyIdentifier", "KeyIdentifier", "keyIdentifier"},
		{"KeyId", "KeyID", "keyID"},
		{"KeyID", "KeyID", "keyID"},
		{"SSEKMSKeyID", "SSEKMSKeyID", "sseKMSKeyID"},
		{"DbiResourceId", "DBIResourceID", "dbiResourceID"},
		{"DbInstanceId", "DBInstanceID", "dbInstanceID"},
		{"DBInstanceId", "DBInstanceID", "dbInstanceID"},
		{"DBInstanceID", "DBInstanceID", "dbInstanceID"},
		{"DBInstanceIdentifier", "DBInstanceIdentifier", "dbInstanceIdentifier"},
	}
	for _, tc := range testCases {
		n := names.New(tc.original)
		msg := fmt.Sprintf("for original %s expected exported name of %s but got %s", tc.original, tc.expectExported, n.GoExported)
		assert.Equal(tc.expectExported, n.GoExported, msg)
		msg = fmt.Sprintf("for original %s expected unexported name of %s but got %s", tc.original, tc.expectUnexported, n.GoUnexported)
		assert.Equal(tc.expectUnexported, n.GoUnexported, msg)
	}
}
