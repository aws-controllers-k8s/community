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
	"errors"
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

var (
	ErrAnonymousArrayElementObjectType = errors.New("found an array type with an unnamed element object definition")
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
				if err == ErrAnonymousArrayElementObjectType {
					// TODO(jaypipes) Figure this out...
					return "AnonymousArrayObjectType", nil
				}
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
	pluralize := pluralize.NewClient()
	guessObjectTypeName := arrayTypeName
	if strings.HasSuffix(arrayTypeName, "List") {
		guessObjectTypeName = strings.TrimSuffix(arrayTypeName, "List")
	} else if strings.HasSuffix(arrayTypeName, "Set") {
		guessObjectTypeName = strings.TrimSuffix(arrayTypeName, "Set")
	}
	if guessObjectTypeName != "" {
		singularName := guessObjectTypeName
		pluralName := guessObjectTypeName
		// First check to see if there is an object schema named the same as
		// the singular version of the guessed object type
		if pluralize.IsPlural(guessObjectTypeName) {
			singularName = pluralize.Singular(guessObjectTypeName)
		} else {
			pluralName = pluralize.Plural(guessObjectTypeName)
		}
		// Check to see if there is a schema matching the guessed object type
		// name and if so, return that schema name as the target object type
		_, found := h.api.Components.Schemas[singularName]
		if found {
			return "*" + singularName, nil
		}
		// The referred-to object type may be a "naturally pluralized" thing.
		// For instance, in the RDS API, the DBCluster Shape has a
		// DBClusterOptionGroupMemberships field that is a reference to a Shape
		// called DBClusterOptionGroupMemberships that is an array of objects
		// with some properties. There is no corresponding
		// "DBClusterOptionGroupMembership" shape that we can use to further
		// reduce the deduced referred-to object type
		_, found = h.api.Components.Schemas[pluralName]
		if found {
			return "*" + pluralName, nil
		}
		// Finally, if neither the singular OR plural names of the guessed
		// object type exist, then let's just try the original array type name
		// (without the stripped List/Set, etc) and fall back to that
		//
		// This is the case for the RDS API where there is something called an
		// OptionGroupOption that has a field called OptionGroupOptionVersions:
		//
		// OptionGroupOption:
		//   properties:
		//    ... lots of fields ...
		// 	  OptionGroupOptionVersions:
		// 	   $ref: '#/components/schemas/OptionGroupOptionVersionsList'
		//   type: object
		// OptionGroupOptionVersionsList:
		//   items:
		// 	  properties:
		// 	   IsDefault:
		// 		$ref: '#/components/schemas/Boolean'
		// 	   Version:
		// 		$ref: '#/components/schemas/String'
		// 	  type: object
		//   type: array
		//
		// There is no schema for either "OptionGroupOptionVersion" or
		// "OptionGroupOptionVersions". The only schema is the
		// "OptionGroupOptionVersionsList" shown above. :(
		//
		// We need to signal to that a struct type to represent the array's
		// elements is needed, since there is no named struct...
		arrayTypeSchemaRef, found := h.api.Components.Schemas[arrayTypeName]
		if found {
			arrayTypeSchema := h.getSchemaFromSchemaRef(arrayTypeSchemaRef)
			if arrayTypeSchema != nil && arrayTypeSchema.Type == "array" {
				return "", ErrAnonymousArrayElementObjectType
			}
		}
		// TODO(jaypipes): Look through all schemas and try to match on known
		// properties?
		return "", fmt.Errorf("failed to find %s when looking up arrayTypeName %s", guessObjectTypeName, arrayTypeName)
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
