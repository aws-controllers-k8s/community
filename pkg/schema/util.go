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

	"github.com/aws/aws-service-operator-k8s/pkg/names"
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

// GetGoTypeFromSchemaRef returns a string of the Go type given an
// openapi3.SchemaRef
func (h *Helper) GetGoTypeFromSchemaRef(
	propNames names.Names, // This is the name of the field/attribute
	schemaRef *openapi3.SchemaRef,
) (string, error) {
	schema := h.getSchemaFromSchemaRef(schemaRef)
	switch schema.Type {
	case "boolean":
		return "bool", nil
	case "string":
		if schema.Format == "byte" {
			return "[]byte", nil
		}
		return "string", nil
	case "number":
		if schema.Format == "float64" {
			return "float64", nil
		}
		return "int64", nil
	case "integer":
		return "int64", nil
	case "array":
		itemsSchemaRefName := schema.Items.Ref
		if itemsSchemaRefName == "" {
			if schemaRef.Ref != "" {
				// Here, we deal with a situation like we find in the ECR API
				// with ImageScanFinding. There is an ImageScanFindings schema
				// of type "object" which has a "findings" property that is a
				// reference to an ImageScanFindingsList schema that is of type
				// "array" of "object" with a set of properties that is
				// identical to the ImageScanFinding schema:
				//
				//     ImageScanFinding:
				//       properties:
				//         attributes:
				//           $ref: '#/components/schemas/AttributeList'
				//         description:
				//           $ref: '#/components/schemas/FindingDescription'
				//         name:
				//           $ref: '#/components/schemas/FindingName'
				//         severity:
				//           $ref: '#/components/schemas/FindingSeverity'
				//         uri:
				//           $ref: '#/components/schemas/Url'
				//       type: object
				//     ImageScanFindingList:
				//       items:
				//         properties:
				//           attributes:
				//             $ref: '#/components/schemas/AttributeList'
				//           description:
				//             $ref: '#/components/schemas/FindingDescription'
				//           name:
				//             $ref: '#/components/schemas/FindingName'
				//           severity:
				//             $ref: '#/components/schemas/FindingSeverity'
				//           uri:
				//             $ref: '#/components/schemas/Url'
				//         type: object
				//       type: array
				//     ImageScanFindings:
				//       properties:
				//         findingSeverityCounts:
				//           $ref: '#/components/schemas/FindingSeverityCounts'
				//         findings:
				//           $ref: '#/components/schemas/ImageScanFindingList'
				//         imageScanCompletedAt:
				//           $ref: '#/components/schemas/ScanTimestamp'
				//         vulnerabilitySourceUpdatedAt:
				//           $ref: '#/components/schemas/VulnerabilitySourceUpdateTimestamp'
				//       type: object
				//
				// In order to come up with a Go type for the
				// `ImageScanFindings.Findings` struct field, we need to search
				// for a schema called ImageScanFindingList, not a schema
				// called "findings".
				itemsSchemaRefName = strings.TrimPrefix(schemaRef.Ref, compSchemasRef)
			} else {
				itemsSchemaRefName = propNames.Original
			}
		} else {
			itemsSchemaRefName = strings.TrimPrefix(itemsSchemaRefName, compSchemasRef)
		}
		itemsSchema := h.getSchemaFromSchemaRef(schema.Items)
		var itemType string
		var err error
		if itemsSchema.Type == "object" {
			itemType, err = h.deduceArrayOfObjectsType(itemsSchemaRefName, itemsSchema)
			if err != nil {
				return "", err
			}
		} else {
			itemType, err = h.GetGoTypeFromSchemaRef(propNames, schema.Items)
			if err != nil {
				return "", err
			}
		}
		return "[]" + itemType, nil
	case "object":
		if schema.AdditionalPropertiesAllowed != nil && *schema.AdditionalPropertiesAllowed {
			return "map[string]string", nil
		}
		return "*" + propNames.Camel, nil
	}
	return "", fmt.Errorf("failed to determine Go type. schema.Type was %s", schema.Type)
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
) (string, error) {
	guessObjectTypeName := arrayTypeName
	if strings.HasSuffix(arrayTypeName, "List") {
		guessObjectTypeName = strings.TrimSuffix(arrayTypeName, "List")
	} else if strings.HasSuffix(arrayTypeName, "Set") {
		guessObjectTypeName = strings.TrimSuffix(arrayTypeName, "Set")
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
			return "*" + guessObjectTypeName, nil
		}
		return "", fmt.Errorf("failed to find %s when looking up arrayTypeName %s", guessObjectTypeName, arrayTypeName)
		// TODO(jaypipes): Look through all schemas and try to match on known
		// properties?
	}
	return "", fmt.Errorf("failed to guess an object type name when looking up arrayTypeName %s", arrayTypeName)
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
