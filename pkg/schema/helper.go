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
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/aws/aws-service-operator-k8s/pkg/model"
)

type Helper struct {
	api          *openapi3.Swagger
	serviceAlias string
	crds         []*model.CRD
	// A map of operation type and resource name to openapi3.Operation
	opMap *OperationMap
}

func (h *Helper) GetServiceAlias() string {
	if h.serviceAlias != "" {
		return h.serviceAlias
	}
	if h.api == nil || h.api.Info == nil {
		return "Unknown service alias"
	}
	apiAlias, found := h.api.Info.Extensions["x-aws-api-alias"]
	apiAliasStr := []byte("unknown")
	if found {
		apiAliasStr, _ = apiAlias.(json.RawMessage).MarshalJSON()
	}
	apiAliasStr = bytes.Replace(apiAliasStr, []byte("\""), []byte(""), -1)
	h.serviceAlias = string(apiAliasStr)
	return h.serviceAlias
}

func (h *Helper) GetAPIGroup() string {
	serviceAlias := h.GetServiceAlias()
	return fmt.Sprintf("%s.services.k8s.aws", serviceAlias)
}

func (h *Helper) GetSchema(schemaName string) *openapi3.Schema {
	if h.api == nil {
		return nil
	}
	schemaRef := h.api.Components.Schemas[schemaName]
	if schemaRef == nil {
		return nil
	}
	return schemaRef.Value
}

func NewHelper(api *openapi3.Swagger) *Helper {
	return &Helper{api, "", nil, nil}
}
