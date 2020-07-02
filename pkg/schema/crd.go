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

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/aws/aws-service-operator-k8s/pkg/model"
	"github.com/aws/aws-service-operator-k8s/pkg/names"
)

func (h *Helper) GetCRDs() ([]*model.CRD, error) {
	if h.crds != nil {
		return h.crds, nil
	}
	crds := []*model.CRD{}

	opMap := h.GetOperationMap()

	createOps := (*opMap)[OpTypeCreate]
	readOneOps := (*opMap)[OpTypeGet]
	updateOps := (*opMap)[OpTypeUpdate]
	deleteOps := (*opMap)[OpTypeDelete]

	for resName, createOp := range createOps {
		names := names.New(resName)
		sdkObjType := h.sdkObjectTypeFromOp(createOp)
		crdOps := model.CRDOps{
			Create:  createOps[resName],
			ReadOne: readOneOps[resName],
			Update:  updateOps[resName],
			Delete:  deleteOps[resName],
		}
		crd := model.NewCRD(names, sdkObjType, crdOps)
		inAttrs, outAttrs, err := h.getAttrsFromOp(createOp, resName)
		if err != nil {
			return nil, err
		}
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

func (h *Helper) GetOperationMap() *OperationMap {
	if h.opMap != nil {
		return h.opMap
	}
	api := h.api
	// create an index of Operations by operation types and resource name
	opMap := OperationMap{}
	for _, pathItem := range api.Paths {
		for _, op := range pathItem.Operations() {
			opID := op.OperationID
			opType, resName := GetOpTypeAndResourceNameFromOpID(opID)
			if _, found := opMap[opType]; !found {
				opMap[opType] = map[string]*openapi3.Operation{}
			}
			opMap[opType][resName] = op
		}
	}
	h.opMap = &opMap
	return &opMap
}
