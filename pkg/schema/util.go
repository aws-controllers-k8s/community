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
	"strconv"
	"strings"

	"github.com/gertd/go-pluralize"
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
		if schema.Format == "byte" {
			return "[]byte"
		}
		return "string"
	case "number":
		if schema.Format == "float64" {
			return "float64"
		}
		return "int64"
	case "integer":
		return "int64"
	case "array":
		itemsSchema := h.getSchemaFromSchemaRef(schema.Items)
		var itemType string
		if itemsSchema.Type == "object" {
			itemType = h.deduceArrayOfObjectsType(schemaName, itemsSchema)
			if itemType == "" {
				itemType = "!!! UNKNOWN ARRAY OBJECT TYPE !!!"
			}
		} else {
			itemType = h.getGoTypeFromSchema(schemaName, itemsSchema)
		}
		return "[]" + itemType
	case "object":
		if schema.AdditionalPropertiesAllowed != nil && *schema.AdditionalPropertiesAllowed {
			return "map[string]string"
		}
		return "*" + schemaName
	}
	return "!!! UNKNOWN !!!"
}

// Some shapes in AWS service APIs are arrays of unnamed objects. A good
// example of this is the SNS API's TagList shape, transformed into the
// following JSONSchema:
//
//    TagList:
//      items:
//        properties:
//          Key:
//            $ref: '#/components/schemas/TagKey'
//          Value:
//            $ref: '#/components/schemas/TagValue'
//        required:
//        - Key
//        - Value
//        type: object
//      type: array
//    Tag:
//      properties:
//        Key:
//          $ref: '#/components/schemas/TagKey'
//        Value:
//          $ref: '#/components/schemas/TagValue'
//      required:
//      - Key
//      - Value
//      type: object
//
// We need to convert a reference to the TagList schema (i.e.
// #/components/schemas/TagList) into the following Go type definition:
//
//   []*Tag
//
// Because we don't know that the TagList items are actually Tag objects, we
// need to try and deduce this information. We first examine the name of the
// array type and see if there is an object type with the same name without the
// "List" or "s" or "Set" suffixes. We then check to see if the schema of the
// found object type is identical to the schema of the items in the array type.
func (h *Helper) deduceArrayOfObjectsType(
	arrayTypeName string,
	objectTypeSchema *openapi3.Schema,
) string {
	guessObjectTypeName := ""
	if strings.HasSuffix(arrayTypeName, "List") {
		guessObjectTypeName = strings.TrimSuffix(arrayTypeName, "List")
	} else if strings.HasSuffix(arrayTypeName, "Set") {

	} else {
		// Fall back to determining if the array type name is a plural
		pluralize := pluralize.NewClient()
		if pluralize.IsPlural(arrayTypeName) {
			guessObjectTypeName = pluralize.Singular(arrayTypeName)
		}
	}
	if guessObjectTypeName != "" {
		// Check to see if there is a schema matching the guessed object type
		// name and if so, return that schema name as the target object type
		_, found := h.api.Components.Schemas[guessObjectTypeName]
		if found {
			return "*" + guessObjectTypeName
		}
		fmt.Printf("Failed to find %s when looking up arrayTypeName %s\n", guessObjectTypeName, arrayTypeName)
		// TODO(jaypipes): Look through all schemas and try to match on known
		// properties?
	}
	return ""
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
