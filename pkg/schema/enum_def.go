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

func (h *Helper) GetEnumDefs() ([]*model.EnumDef, error) {
	edefs := []*model.EnumDef{}

	for shapeName, shape := range h.sdkAPI.Shapes {
		if !shape.IsEnum() {
			continue
		}

		edef, err := model.NewEnumDef(names.New(shapeName), shape.Enum)
		if err != nil {
			return nil, err
		}
		edefs = append(edefs, edef)
	}
	sort.Slice(edefs, func(i, j int) bool {
		return edefs[i].Names.Camel < edefs[j].Names.Camel
	})
	return edefs, nil
}
