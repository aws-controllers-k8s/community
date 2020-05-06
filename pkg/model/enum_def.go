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

package model

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-service-operator-k8s/pkg/names"
)

type EnumValue struct {
	Original string
	Clean    string
}

// EnumDef is the definition of an enumeration type for a field present in
// either a CRD or a TypeDef
type EnumDef struct {
	Names  names.Names
	GoType string
	Values []EnumValue
}

func NewEnumDef(names names.Names, goType string, values []interface{}) (*EnumDef, error) {
	enumVals := make([]EnumValue, len(values))
	for x, item := range values {
		strVal, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("cannot convert %v to string", item)
		}
		enumVals[x] = newEnumVal(strVal)
	}
	return &EnumDef{names, goType, enumVals}, nil
}

func newEnumVal(orig string) EnumValue {
	// Convert values like "m5.xlarge" into "m5_xlarge"
	cleaner := func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}
	clean := bytes.Map(cleaner, []byte(orig))

	return EnumValue{
		Original: orig,
		Clean:    string(clean),
	}
}
