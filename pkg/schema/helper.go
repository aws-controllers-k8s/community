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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/aws/aws-service-operator-k8s/pkg/model"
)

type Helper struct {
	api  *openapi3.Swagger
	crds []*model.CRD
}

func (h *Helper) GetAPIGroup() string {
	apiAlias, found := h.api.Info.Extensions["x-aws-api-alias"]
	apiAliasStr := []byte("unknown")
	if found {
		apiAliasStr, _ = apiAlias.(json.RawMessage).MarshalJSON()
	}
	apiGroup := fmt.Sprintf("%s.services.k8s.aws", apiAliasStr)
	return strings.Replace(apiGroup, "\"", "", -1)
}

func NewHelper(api *openapi3.Swagger) *Helper {
	return &Helper{api, nil}
}
