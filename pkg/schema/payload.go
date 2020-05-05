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
)

func (h *Helper) getPayloads() []string {
	res := []string{}
	for _, pathItem := range h.api.Paths {
		if pathItem.Delete != nil {
			res = append(res, h.getPayloadsFromOp(pathItem.Delete)...)
		}
		if pathItem.Get != nil {
			res = append(res, h.getPayloadsFromOp(pathItem.Get)...)
		}
		if pathItem.Post != nil {
			res = append(res, h.getPayloadsFromOp(pathItem.Post)...)
		}
		if pathItem.Put != nil {
			res = append(res, h.getPayloadsFromOp(pathItem.Put)...)
		}
		if pathItem.Patch != nil {
			res = append(res, h.getPayloadsFromOp(pathItem.Patch)...)
		}
	}
	return res
}

func (h *Helper) getPayloadsFromOp(op *openapi3.Operation) []string {
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
					schema := h.getSchemaFromSchemaRef(schemaRef)
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
			schema := h.getSchemaFromSchemaRef(schemaRef)
			if schema != nil && schema.Type == "object" {
				res = append(res, strings.TrimPrefix(schemaRef.Ref, compSchemasRef))
			}
		}
	}
	return res
}
