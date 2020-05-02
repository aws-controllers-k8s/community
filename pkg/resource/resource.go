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
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-service-operator-k8s/pkg/names"
	"github.com/gertd/go-pluralize"
	"github.com/getkin/kin-openapi/openapi3"
)

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

const (
	compSchemasRef = "#/components/schemas/"
)

type ResourceOps struct {
	CreateOp *openapi3.Operation
}

type Resource struct {
	api         *openapi3.Swagger
	Names       names.Names
	Kind        string
	Plural      string
	Ops         ResourceOps
	SpecAttrs   map[string]*Attr
	StatusAttrs map[string]*Attr
}

func isSuccessResponseCode(rc string) bool {
	val, err := strconv.Atoi(rc)
	if err == nil {
		return val >= 200 && val < 300
	}
	return false
}

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

// getGoTypeFromSchema returns a string of the Go type given an openapi3.Schema
func getGoTypeFromSchema(
	schemaName string,
	api *openapi3.Swagger,
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
		itemsSchema := getSchemaFromSchemaRef(api, schema.Items)
		itemType := getGoTypeFromSchema(schemaName, api, itemsSchema)
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
func getSchemaFromSchemaRef(
	api *openapi3.Swagger,
	schemaRef *openapi3.SchemaRef,
) *openapi3.Schema {
	if schemaRef.Ref != "" {
		if strings.HasPrefix(schemaRef.Ref, compSchemasRef) {
			refSchemaID := strings.TrimPrefix(schemaRef.Ref, compSchemasRef)
			schema, found := api.Components.Schemas[refSchemaID]
			if found {
				return schema.Value

			}
		}
	} else {
		return schemaRef.Value
	}
	return nil
}

func (r *Resource) loadAttrs() {
	inAttrs, outAttrs := r.getAttrsFromOp(r.Ops.CreateOp)
	inAttrMap := make(map[string]*Attr, len(inAttrs))
	for _, inAttr := range inAttrs {
		inAttrMap[inAttr.Names.Original] = inAttr
	}
	outAttrMap := make(map[string]*Attr, len(outAttrs))
	for _, outAttr := range outAttrs {
		outAttrMap[outAttr.Names.Original] = outAttr
	}
	r.SpecAttrs = inAttrMap
	r.StatusAttrs = outAttrMap
}

func ResourcesFromAPI(api *openapi3.Swagger) ([]*Resource, error) {
	pluralize := pluralize.NewClient()
	resources := []*Resource{}

	// create an index of Operations by operation types
	opTypeIndex := map[opType]map[string]*openapi3.Operation{}
	for _, pathItem := range api.Paths {
		if pathItem.Post != nil {
			op := pathItem.Post
			opID := op.OperationID
			opType := getOpTypeFromOpID(opID)
			if _, found := opTypeIndex[opType]; !found {
				opTypeIndex[opType] = map[string]*openapi3.Operation{}
			}
			opTypeIndex[opType][opID] = op
		}
		if pathItem.Put != nil {
			op := pathItem.Put
			opID := op.OperationID
			opType := getOpTypeFromOpID(opID)
			if _, found := opTypeIndex[opType]; !found {
				opTypeIndex[opType] = map[string]*openapi3.Operation{}
			}
			opTypeIndex[opType][opID] = op
		}
	}

	createOps := opTypeIndex[otCreate]
	for opID, createOp := range createOps {
		resourceName := strings.TrimPrefix(opID, "Create")
		singularName := pluralize.Singular(resourceName)
		pluralName := pluralize.Plural(resourceName)
		if pluralName == resourceName {
			// For now, ignore batch create operations since we're just trying
			// to determine top-level singular resources
			fmt.Printf("operation %s looks to be a batch create operation.\n", opID)
			continue
		}

		names := names.New(singularName)
		kind := names.GoExported
		plural := pluralize.Plural(kind)

		resource := &Resource{
			api:    api,
			Names:  names,
			Kind:   kind,
			Plural: plural,
			Ops:    ResourceOps{createOp},
		}
		resource.loadAttrs()
		resources = append(resources, resource)
	}
	return resources, nil
}
