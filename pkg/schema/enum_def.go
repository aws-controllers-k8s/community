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

import (
	"fmt"

	"github.com/aws/aws-service-operator-k8s/pkg/model"
	"github.com/aws/aws-service-operator-k8s/pkg/names"
)

func (h *Helper) GetEnumDefs() ([]*model.EnumDef, error) {
	api := h.api
	edefs := []*model.EnumDef{}

	for schemaName, schemaRef := range api.Components.Schemas {
		schema := h.getSchemaFromSchemaRef(schemaRef)
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
		edef, err := model.NewEnumDef(names.New(schemaName), goType, schema.Enum)
		if err != nil {
			return nil, err
		}
		edefs = append(edefs, edef)
	}
	return edefs, nil
}
