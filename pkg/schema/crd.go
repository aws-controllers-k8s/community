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
	"strings"

	awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"
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
		crdOps := model.CRDOps{
			Create:  createOps[resName],
			ReadOne: readOneOps[resName],
			Update:  updateOps[resName],
			Delete:  deleteOps[resName],
		}
		crd := model.NewCRD(names, crdOps)
		sdkMapper := model.NewSDKMapper(crd)
		crd.SDKMapper = sdkMapper
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

		// At this point, we have two maps of attributes, one for the
		// attributes passed to the Create request's input shape, the other
		// returned from the API response's output shape.
		//
		// We need to determine fields in the output shape that:
		//
		// * Are not the same as the fields in the input shape
		// * Are not ARN fields for the primary object (since that is always in
		//   `Status.ACKResourceMetadata.ARN`)
		//
		// For example, assume the following input attributes for "Book"
		// resource:
		//
		// "name", "title", "author"
		//
		// And the following output attributes:
		//
		// "bookARN", "name", "title", "author", "createdOn"
		//
		// We want to reduce the output attributes to just "createdOn", since
		// all the other attributes are either the primary object's ARN or are
		// in the input attributes (and would be in the CRD's Spec field)
		for outAttrName := range outAttrMap {
			_, found := inAttrMap[outAttrName]
			if found {
				delete(outAttrMap, outAttrName)
			}
			if strings.EqualFold(outAttrName, "arn") ||
				strings.EqualFold(outAttrName, resName+"arn") {
				sdkMapper.SetPrimaryResourceARNField(createOp, outAttrName)
				delete(outAttrMap, outAttrName)
			}
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

// GetOperationMap returns a map, keyed by the operation type and operation
// ID/name, of aws-sdk-go private/model/api.Operation struct pointers
func (h *Helper) GetOperationMap() *OperationMap {
	if h.opMap != nil {
		return h.opMap
	}
	// create an index of Operations by operation types and resource name
	opMap := OperationMap{}
	for opID, op := range h.sdkAPI.Operations {
		opType, resName := GetOpTypeAndResourceNameFromOpID(opID)
		if _, found := opMap[opType]; !found {
			opMap[opType] = map[string]*awssdkmodel.Operation{}
		}
		opMap[opType][resName] = op
	}
	h.opMap = &opMap
	return &opMap
}
