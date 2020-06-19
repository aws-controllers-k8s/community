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
	"sort"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/aws/aws-service-operator-k8s/pkg/model"
	"github.com/aws/aws-service-operator-k8s/pkg/names"
)

func (h *Helper) GetCRDs() ([]*model.CRD, error) {
	if h.crds != nil {
		return h.crds, nil
	}
	api := h.api
	pluralize := pluralize.NewClient()
	crds := []*model.CRD{}

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
		crdName := strings.TrimPrefix(opID, "Create")
		singularName := pluralize.Singular(crdName)
		pluralName := pluralize.Plural(crdName)
		if pluralName == crdName {
			// For now, ignore batch create operations since we're just trying
			// to determine top-level singular crds
			fmt.Printf("operation %s looks to be a batch create operation.\n", opID)
			continue
		}

		names := names.New(singularName)
		crd := model.NewCRD(names, createOp)
		inAttrs, outAttrs := h.getAttrsFromOp(createOp, crdName)
		inAttrMap := make(map[string]*model.Attr, len(inAttrs))
		for _, inAttr := range inAttrs {
			inAttrMap[inAttr.Names.Original] = inAttr
		}
		outAttrMap := make(map[string]*model.Attr, len(outAttrs))
		for _, outAttr := range outAttrs {
			outAttrMap[outAttr.Names.Original] = outAttr
		}
		crd.SpecAttrs = inAttrMap
		crd.StatusAttrs = outAttrMap
		crds = append(crds, crd)
	}
	sort.Slice(crds, func(i, j int) bool {
		return crds[i].Names.Camel < crds[j].Names.Camel
	})
	h.crds = crds
	return crds, nil
}
