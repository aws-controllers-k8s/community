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

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
)

type Attr struct {
	Name     string
	JSONName string
	GoType   string
	Schema   *openapi3.Schema
}

func newAttr(
	api *openapi3.Swagger,
	name string,
	schema *openapi3.Schema,
) *Attr {
	return &Attr{
		Name:     strcase.ToCamel(name),
		JSONName: name,
		GoType:   getGoTypeFromSchema(strcase.ToCamel(name), api, schema),
		Schema:   schema,
	}
}

// getAttrsFromRequestSchemaRef returns a slice of Attr representing the fields
// for a Schema related to an HTTP request for an Operation with
// application/json semantics.
func (r *Resource) getAttrsFromRequestSchemaRef(
	schemaRef *openapi3.SchemaRef,
) []*Attr {
	attrs := []*Attr{}
	if schemaRef == nil {
		return attrs
	}
	schema := getSchemaFromSchemaRef(r.api, schemaRef)
	if schema == nil {
		fmt.Printf("failed to find object schema for ref %s\n", schemaRef.Ref)
		return attrs
	}
	if schema.Type != "object" {
		fmt.Printf("expected to find object schema but found %s\n", schema.Type)
		return attrs
	}

	for propName, propSchemaRef := range schema.Properties {
		propSchema := getSchemaFromSchemaRef(r.api, propSchemaRef)
		attrs = append(attrs, newAttr(r.api, propName, propSchema))
	}
	return attrs
}

// getAttrsFromResponseSchemaRef returns a slice of Attr representing the fields
// for a Schema related to an HTTP response for an Operation with
// application/json semantics. If the HTTP response uses a strategy of
// "wrapping" the returned response object in a JSON object with a single
// attribute named the same as the created resource, we "flatten" the returned
// attributes to be the attributes of the wrapped JSON object schema.
func (r *Resource) getAttrsFromResponseSchemaRef(
	schemaRef *openapi3.SchemaRef,
) []*Attr {
	attrs := []*Attr{}
	if schemaRef == nil {
		return attrs
	}
	var schema *openapi3.Schema
	schema = getSchemaFromSchemaRef(r.api, schemaRef)
	if schema == nil {
		fmt.Printf("failed to find object schema for ref %s\n", schemaRef.Ref)
		return attrs
	}
	if schema.Type != "object" {
		fmt.Printf("expected to find object schema but found %s\n", schema.Type)
		return attrs
	}

	if len(schema.Properties) == 1 {
		for propName, propSchemaRef := range schema.Properties {
			if strings.ToLower(propName) == strings.ToLower(r.Kind) {
				// "flatten" by unwrapping the wrapped response schema
				schema = getSchemaFromSchemaRef(r.api, propSchemaRef)
				break
			}
		}
	}

	for propName, propSchemaRef := range schema.Properties {
		propSchema := getSchemaFromSchemaRef(r.api, propSchemaRef)
		attrs = append(attrs, newAttr(r.api, propName, propSchema))
	}
	return attrs
}

// getAttrsFromOp returns two slices of Attr representing the input fields for
// the operation request and the output fields for the operation response
func (r *Resource) getAttrsFromOp(
	op *openapi3.Operation,
) ([]*Attr, []*Attr) {
	inAttrs := []*Attr{}
	outAttrs := []*Attr{}
	if op.RequestBody != nil {
		// Look to see if the request body has a content element that refers to
		// a schema describing the object being created/patched
		if op.RequestBody.Ref != "" {
			fmt.Printf("found request body ref: %s\n", op.RequestBody.Ref)
		} else {
			rb := op.RequestBody.Value
			mediaType, found := rb.Content["application/json"]
			if !found {
				fmt.Printf("skipping non-JSON operation %s\n", op.OperationID)
				return inAttrs, outAttrs
			}
			if mediaType.Schema != nil {
				inAttrs = append(
					inAttrs,
					r.getAttrsFromRequestSchemaRef(mediaType.Schema)...,
				)
			}
		}
	}
	for rc, responseRef := range op.Responses {
		if isSuccessResponseCode(rc) {
			// Look to see if the response body has a content element that refers
			// to a schema describing the object that was created/patched
			if responseRef.Ref != "" {
				fmt.Printf("found response body ref: %s\n", responseRef.Ref)
			} else {
				resp := responseRef.Value
				mediaType, found := resp.Content["application/json"]
				if !found {
					fmt.Printf("skipping non-JSON operation %s\n", op.OperationID)
					return inAttrs, outAttrs
				}
				if mediaType.Schema != nil {
					outAttrs = append(
						outAttrs,
						r.getAttrsFromResponseSchemaRef(mediaType.Schema)...,
					)
				}
			}
		}
	}
	return inAttrs, outAttrs
}
