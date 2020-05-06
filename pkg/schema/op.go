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

package schema

import "strings"

type opType int

const (
	otUnknown opType = iota
	otCreate
	otCreateBatch
	otDelete
	otReplace
	otPatch
	otUpdateAttr
	otAddChild
	otAddChildren
	otRemoveChild
	otRemoveChildren
	otGet
	otList
)

// Guess the type of operation from the OperationID...
func getOpTypeFromOpID(opID string) opType {
	if strings.HasPrefix(opID, "CreateOrUpdate") {
		return otReplace
	} else if strings.HasPrefix(opID, "Create") {
		return otCreate
	} else if strings.HasPrefix(opID, "Delete") {
		return otDelete
	} else if strings.HasPrefix(opID, "Describe") {

	}
	return otUnknown
}
