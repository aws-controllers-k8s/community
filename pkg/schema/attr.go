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
) ([]*model.Attr, error) {
	attrs := []*model.Attr{}
	if schemaRef == nil {
		return nil, fmt.Errorf("failed to find schema properties from request schema ref: schemaRef was nil")
	}
	schema := h.getSchemaFromSchemaRef(schemaRef)
	if schema == nil {
		return nil, fmt.Errorf("failed to find schema properties from request schema ref: no schema for ref %s", schemaRef.Ref)
	}
	if schema.Type != "object" {
		return nil, fmt.Errorf("failed to find schema properties from request schema ref: expected to find schema of type 'object' but found %s", schema.Type)
	}

	for propName, propSchemaRef := range schema.Properties {
		propNames := names.New(propName)
		goType, err := h.GetGoTypeFromSchemaRef(propNames, propSchemaRef)
		if err != nil {
			return nil, err
		}
		propSchema := h.getSchemaFromSchemaRef(propSchemaRef)
		attrs = append(attrs, model.NewAttr(propNames, goType, propSchema))
	}
	return attrs, nil
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
) ([]*model.Attr, error) {
	attrs := []*model.Attr{}
	if schemaRef == nil {
		return nil, fmt.Errorf("failed to find schema properties from response schema ref: schemaRef was nil")
	}
	var schema *openapi3.Schema
	schema = h.getSchemaFromSchemaRef(schemaRef)
	if schema == nil {
		return nil, fmt.Errorf("failed to find schema properties from response schema ref: no schema for ref %s", schemaRef.Ref)
	}
	if schema.Type != "object" {
		return nil, fmt.Errorf("failed to find schema properties from response schema ref: expected to find schema of type 'object' but found %s", schema.Type)
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
		propNames := names.New(propName)
		goType, err := h.GetGoTypeFromSchemaRef(propNames, propSchemaRef)
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, model.NewAttr(propNames, goType, propSchema))
	}
	return attrs, nil
}

// getAttrsFromOp returns two slices of Attr representing the input fields for
// the operation request and the output fields for the operation response
func (h *Helper) getAttrsFromOp(
	op *openapi3.Operation,
	crdName string,
) ([]*model.Attr, []*model.Attr, error) {
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
				return inAttrs, outAttrs, nil
			}
			if mediaType.Schema != nil {
				reqAttrs, err := h.getAttrsFromRequestSchemaRef(mediaType.Schema)
				if err != nil {
					return nil, nil, err
				}
				inAttrs = append(inAttrs, reqAttrs...)
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
					return inAttrs, outAttrs, nil
				}
				if mediaType.Schema != nil {
					respAttrs, err := h.getAttrsFromResponseSchemaRef(mediaType.Schema, crdName)
					if err != nil {
						return nil, nil, err
					}
					outAttrs = append(outAttrs, respAttrs...)
				}
			}
		}
	}
	return inAttrs, outAttrs, nil
}
