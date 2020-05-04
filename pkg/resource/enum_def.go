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

package resource

import (
	"bytes"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/aws/aws-service-operator-k8s/pkg/names"
)

type EnumValue struct {
	Original string
	Clean    string
}

type EnumDef struct {
	Names  names.Names
	GoType string
	Values []EnumValue
}

func EnumDefsFromAPI(
	api *openapi3.Swagger,
) ([]*EnumDef, error) {
	edefs := []*EnumDef{}

	for schemaName, schemaRef := range api.Components.Schemas {
		schema := getSchemaFromSchemaRef(api, schemaRef)
		if len(schema.Enum) == 0 {
			continue
		}

		goType := "unknown"
		switch schema.Type {
		case "string":
			goType = "string"
		case "integer":
			if schema.Format == "int32" {
				goType = "int32"
			} else {
				goType = "int64"
			}
		default:
			return nil, fmt.Errorf("cannot determine go type from enum schema type %s", schema.Type)
		}
		vals := make([]EnumValue, len(schema.Enum))
		for x, item := range schema.Enum {
			strVal, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("cannot convert %v to string", item)
			}
			vals[x] = newEnumVal(strVal)
		}
		edefs = append(edefs, &EnumDef{
			Names:  names.New(schemaName),
			GoType: goType,
			Values: vals,
		})
	}
	return edefs, nil
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
