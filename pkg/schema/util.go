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
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	compSchemasRef = "#/components/schemas/"
)

func inStrings(subject string, collection []string) bool {
	for _, item := range collection {
		if subject == item {
			return true
		}
	}
	return false
}

func isException(schema *openapi3.Schema) bool {
	_, found := schema.Extensions["x-aws-api-exception"]
	return found
}

func isSuccessResponseCode(rc string) bool {
	val, err := strconv.Atoi(rc)
	if err == nil {
		return val >= 200 && val < 300
	}
	return false
}

// getGoTypeFromSchema returns a string of the Go type given an openapi3.Schema
func (h *Helper) getGoTypeFromSchema(
	schemaName string,
	schema *openapi3.Schema,
) string {
	switch schema.Type {
	case "boolean":
		return "bool"
	case "string":
		return "string"
	case "number":
		if schema.Format != "" {
			switch schema.Format {
			case "int64":
				return "int64"
			case "float64":
				return "float64"
			}
		}
		return "int64"
	case "integer":
		return "int64"
	case "array":
		itemsSchema := h.getSchemaFromSchemaRef(schema.Items)
		itemType := h.getGoTypeFromSchema(schemaName, itemsSchema)
		return "[]" + itemType
	case "object":
		if schema.AdditionalPropertiesAllowed != nil && *schema.AdditionalPropertiesAllowed {
			return "map[string]string"
		}
		return "*" + schemaName
	}
	return "!!! UNKNOWN !!!"
}

// getSchemaFromSchemaRef returns an openapi3.Schema given a SchemaRef
func (h *Helper) getSchemaFromSchemaRef(
	schemaRef *openapi3.SchemaRef,
) *openapi3.Schema {
	if schemaRef.Ref != "" {
		if strings.HasPrefix(schemaRef.Ref, compSchemasRef) {
			refSchemaID := strings.TrimPrefix(schemaRef.Ref, compSchemasRef)
			schema, found := h.api.Components.Schemas[refSchemaID]
			if found {
				return schema.Value

			}
		}
	} else {
		return schemaRef.Value
	}
	return nil
}
