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

// sdkObjectTypeFromOp returns the string name of the aws-sdk-go struct that is
// returned for a successful creation of the resource.  This is typically
// called either "{Resource}" or "{Resource}Data", where "{Resource}" is the
// name of the resource. For example, the AppMesh API calls this struct
// MeshData for the Mesh type. It is contained in the CreateMeshOutput
// payload/wrapper struct. The ECR API calls this struct Repository for the
// Repository type. It is contained in the CreateRepositoryResponse
// payload/wrapper struct.
func (h *Helper) sdkObjectTypeFromOp(op *openapi3.Operation) string {
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
				if len(schema.Properties) == 1 {
					// This is the payload/wrapper struct. Grab the name of the
					// single property and look up the referred-to schema, as
					// that is going to be the struct that represents the CRD's
					// primary aws-sdk-go struct.
					//
					// For example, from ECR API:
					//
					// CreateRepositoryResponse:
					//   properties:
					//	   repository:
					//	     $ref: '#/components/schemas/Repository'
					//
					// Right now, we're in the CreateRepositoryResponse schema
					// and we need to grab the #/components/schemes/Repository
					// object name (which is Repository)
					for _, propSchemaRef := range schema.Properties {
						return strings.TrimPrefix(propSchemaRef.Ref, compSchemasRef)
					}
				}
			}
		}
	}
	return ""
}
