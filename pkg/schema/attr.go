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
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/aws/aws-service-operator-k8s/pkg/model"
	"github.com/aws/aws-service-operator-k8s/pkg/names"
)

// getAttrsFromRequestSchemaRef returns a slice of Attr representing the fields
// for a Schema related to an HTTP request for an Operation with
// application/json semantics.
func (h *Helper) getAttrsFromRequestSchemaRef(
	schemaRef *openapi3.SchemaRef,
) []*model.Attr {
	attrs := []*model.Attr{}
	if schemaRef == nil {
		return attrs
	}
	schema := h.getSchemaFromSchemaRef(schemaRef)
	if schema == nil {
		fmt.Printf("failed to find object schema for ref %s\n", schemaRef.Ref)
		return attrs
	}
	if schema.Type != "object" {
		fmt.Printf("expected to find object schema but found %s\n", schema.Type)
		return attrs
	}

	for propName, propSchemaRef := range schema.Properties {
		propSchema := h.getSchemaFromSchemaRef(propSchemaRef)
		names := names.New(propName)
		goType := h.getGoTypeFromSchema(names.GoExported, schema)
		attrs = append(attrs, model.NewAttr(names, goType, propSchema))
	}
	return attrs
}

// getAttrsFromResponseSchemaRef returns a slice of Attr representing the fields
// for a Schema related to an HTTP response for an Operation with
// application/json semantics. If the HTTP response uses a strategy of
// "wrapping" the returned response object in a JSON object with a single
// attribute named the same as the created resource, we "flatten" the returned
// attributes to be the attributes of the wrapped JSON object schema.
func (h *Helper) getAttrsFromResponseSchemaRef(
	schemaRef *openapi3.SchemaRef,
	crdName string,
) []*model.Attr {
	attrs := []*model.Attr{}
	if schemaRef == nil {
		return attrs
	}
	var schema *openapi3.Schema
	schema = h.getSchemaFromSchemaRef(schemaRef)
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
			if strings.ToLower(propName) == strings.ToLower(crdName) {
				// "flatten" by unwrapping the wrapped response schema
				schema = h.getSchemaFromSchemaRef(propSchemaRef)
				break
			}
		}
	}

	for propName, propSchemaRef := range schema.Properties {
		propSchema := h.getSchemaFromSchemaRef(propSchemaRef)
		names := names.New(propName)
		goType := h.getGoTypeFromSchema(names.GoExported, schema)
		attrs = append(attrs, model.NewAttr(names, goType, propSchema))
	}
	return attrs
}

// getAttrsFromOp returns two slices of Attr representing the input fields for
// the operation request and the output fields for the operation response
func (h *Helper) getAttrsFromOp(
	op *openapi3.Operation,
	crdName string,
) ([]*model.Attr, []*model.Attr) {
	inAttrs := []*model.Attr{}
	outAttrs := []*model.Attr{}
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
					h.getAttrsFromRequestSchemaRef(mediaType.Schema)...,
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
						h.getAttrsFromResponseSchemaRef(mediaType.Schema, crdName)...,
					)
				}
			}
		}
	}
	return inAttrs, outAttrs
}
