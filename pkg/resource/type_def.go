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
	"strings"

	"github.com/aws/aws-service-operator-k8s/pkg/names"
	"github.com/getkin/kin-openapi/openapi3"
)

type TypeDef struct {
	Names names.Names
	Attrs map[string]*Attr
}

func getPayloadsFromOp(api *openapi3.Swagger, op *openapi3.Operation) []string {
	res := []string{}
	if op.RequestBody != nil {
		// Look to see if the request body has a content element that refers to
		// a schema describing the object being created/patched
		if op.RequestBody.Ref == "" {
			rb := op.RequestBody.Value
			mediaType, found := rb.Content["application/json"]
			if found {
				schemaRef := mediaType.Schema
				if schemaRef != nil {
					schema := getSchemaFromSchemaRef(api, schemaRef)
					if schema != nil && schema.Type == "object" {
						res = append(res, strings.TrimPrefix(schemaRef.Ref, compSchemasRef))
					}
				}
			}
		}
	}
	for rc, responseRef := range op.Responses {
		if !isSuccessResponseCode(rc) {
			continue
		}
		// Look to see if the response body has a content element that refers
		// to a schema describing the object that was created/patched
		if responseRef.Ref != "" {
			continue
		}
		resp := responseRef.Value
		mediaType, found := resp.Content["application/json"]
		if !found {
			fmt.Printf("skipping non-JSON operation %s\n", op.OperationID)
			continue
		}
		schemaRef := mediaType.Schema
		if schemaRef != nil {
			schema := getSchemaFromSchemaRef(api, schemaRef)
			if schema != nil && schema.Type == "object" {
				res = append(res, strings.TrimPrefix(schemaRef.Ref, compSchemasRef))
			}
		}
	}
	return res
}

func TypeDefsFromAPI(
	api *openapi3.Swagger,
	resources []*Resource,
) ([]*TypeDef, error) {
	tdefs := []*TypeDef{}

	// Payload schemas are not struct defs, so let's go through all our
	// Operations and gather the names of all the payload structs
	payloads := []string{}
	for _, pathItem := range api.Paths {
		if pathItem.Delete != nil {
			payloads = append(payloads, getPayloadsFromOp(api, pathItem.Delete)...)
		}
		if pathItem.Get != nil {
			payloads = append(payloads, getPayloadsFromOp(api, pathItem.Get)...)
		}
		if pathItem.Post != nil {
			payloads = append(payloads, getPayloadsFromOp(api, pathItem.Post)...)
		}
		if pathItem.Put != nil {
			payloads = append(payloads, getPayloadsFromOp(api, pathItem.Put)...)
		}
		if pathItem.Patch != nil {
			payloads = append(payloads, getPayloadsFromOp(api, pathItem.Patch)...)
		}
	}

	resourceNames := []string{}
	for _, res := range resources {
		resourceNames = append(resourceNames, res.Kind)
	}

	for schemaName, schemaRef := range api.Components.Schemas {
		if inStrings(schemaName, resourceNames) {
			// Resources are already top-level structs
			continue
		}
		if inStrings(schemaName, payloads) {
			// Payloads are not type defs
			continue
		}
		schema := getSchemaFromSchemaRef(api, schemaRef)
		if schema.Type != "object" {
			continue
		}
		if isException(schema) {
			// Neither are exceptions
			continue
		}
		attrs := map[string]*Attr{}
		for propName, propSchemaRef := range schema.Properties {
			propSchema := getSchemaFromSchemaRef(api, propSchemaRef)
			attrs[propName] = newAttr(api, propName, propSchema)
		}
		if len(attrs) == 0 {
			// Just ignore these...
			continue
		}
		tdefs = append(tdefs, &TypeDef{
			Names: names.New(schemaName),
			Attrs: attrs,
		})
	}
	return tdefs, nil
}

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
