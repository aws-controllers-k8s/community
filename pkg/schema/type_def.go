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
	"sort"

	"github.com/aws/aws-service-operator-k8s/pkg/model"
	"github.com/aws/aws-service-operator-k8s/pkg/names"
)

func (h *Helper) GetTypeDefs() ([]*model.TypeDef, error) {
	crds, err := h.GetCRDs()
	if err != nil {
		return nil, err
	}
	api := h.api
	tdefs := []*model.TypeDef{}

	payloads := h.getPayloads()

	crdNames := []string{}
	for _, crd := range crds {
		crdNames = append(crdNames, crd.Kind)
	}

	for schemaName, schemaRef := range api.Components.Schemas {
		if inStrings(schemaName, crdNames) {
			// CRDs are already top-level structs
			continue
		}
		if inStrings(schemaName, payloads) {
			// Payloads are not type defs
			continue
		}
		schema := h.getSchemaFromSchemaRef(schemaRef)
		if schema.Type != "object" {
			continue
		}
		if isException(schema) {
			// Neither are exceptions
			continue
		}
		attrs := map[string]*model.Attr{}
		for propName, propSchemaRef := range schema.Properties {
			propSchema := h.getSchemaFromSchemaRef(propSchemaRef)
			propNames := names.New(propName)
			goType := h.getGoTypeFromSchema(propNames.Camel, propSchema)
			attrs[propName] = model.NewAttr(propNames, goType, propSchema)
		}
		if len(attrs) == 0 {
			// Just ignore these...
			continue
		}
		tdefs = append(tdefs, &model.TypeDef{
			Names: names.New(schemaName),
			Attrs: attrs,
		})
	}
	sort.Slice(tdefs, func(i, j int) bool {
		return tdefs[i].Names.Camel < tdefs[j].Names.Camel
	})
	return tdefs, nil
}
