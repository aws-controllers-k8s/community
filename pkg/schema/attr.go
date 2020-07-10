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

	awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"

	"github.com/aws/aws-service-operator-k8s/pkg/model"
	"github.com/aws/aws-service-operator-k8s/pkg/names"
)

// getAttrsFromInputShape returns a slice of Attr representing the fields
// for a Shape related to an HTTP request
func (h *Helper) getAttrsFromInputShape(
	shape *awssdkmodel.Shape,
) ([]*model.Attr, error) {
	attrs := []*model.Attr{}
	if shape == nil {
		// NOTE(jaypipes): aws-sdk-go guarantees that each Operation has a
		// top-level InputRef and OutputRef shape reference:
		//
		// https://github.com/aws/aws-sdk-go/blob/ae9d6422f7b6f945bc66ab87bc18cba57840a410/private/model/api/passes.go#L439-L454
		//
		// So, we panic() here because there should not happen.
		panic("failed to find member shape refs from input shape: shape was nil")
	}
	if shape.Type != "structure" {
		return nil, fmt.Errorf("failed to find member shape refs from input shape ref: expected to find Shape of type 'structure' but found %s", shape.Type)
	}

	for memberName, memberShapeRef := range shape.MemberRefs {
		memberNames := names.New(memberName)
		if memberShapeRef.Shape == nil {
			return nil, fmt.Errorf("failed to find member shape refs from input shape ref: no Shape for member ref %s", memberName)
		}
		goType := memberShapeRef.Shape.GoType()
		attrs = append(attrs, model.NewAttr(memberNames, goType, memberShapeRef.Shape))
	}
	return attrs, nil
}

// getAttrsFromOutputShape returns a slice of Attr representing the fields
// for a Shape related to an HTTP response
//
// If the HTTP response uses a strategy of "wrapping" the returned response
// object in a JSON object with a single attribute named the same as the
// created resource, we "flatten" the returned attributes to be the attributes
// of the wrapped JSON object schema.
func (h *Helper) getAttrsFromOutputShape(
	shape *awssdkmodel.Shape,
	crdName string,
) ([]*model.Attr, error) {
	attrs := []*model.Attr{}
	if shape == nil {
		// NOTE(jaypipes): aws-sdk-go guarantees that each Operation has a
		// top-level InputRef and OutputRef shape reference:
		//
		// https://github.com/aws/aws-sdk-go/blob/ae9d6422f7b6f945bc66ab87bc18cba57840a410/private/model/api/passes.go#L439-L454
		//
		// So, we panic() here because there should not happen.
		panic("failed to find member shape refs from output shape: shape was nil")
	}
	if shape.Type != "structure" {
		return nil, fmt.Errorf("failed to find member shape refs from output shape ref: expected to find Shape of type 'structure' but found %s", shape.Type)
	}

	for memberName, memberShapeRef := range shape.MemberRefs {
		memberNames := names.New(memberName)
		if memberShapeRef.Shape == nil {
			return nil, fmt.Errorf("failed to find member shape refs from output shape ref: no Shape for member ref %s", memberName)
		}
		goType := memberShapeRef.Shape.GoType()
		attrs = append(attrs, model.NewAttr(memberNames, goType, memberShapeRef.Shape))
	}
	return attrs, nil
}

// getAttrsFromOp returns two slices of Attr representing the input fields for
// the operation request and the output fields for the operation response
func (h *Helper) getAttrsFromOp(
	op *awssdkmodel.Operation,
	crdName string,
) ([]*model.Attr, []*model.Attr, error) {
	var err error
	inAttrs := []*model.Attr{}
	outAttrs := []*model.Attr{}
	inAttrs, err = h.getAttrsFromInputShape(op.InputRef.Shape)
	if err != nil {
		return nil, nil, err
	}
	outputShape := op.OutputRef.Shape
	if outputShape.UsedAsOutput && len(outputShape.MemberRefs) == 1 {
		// We might be in a "wrapper" shape. Unwrap it to find the real object
		// representation for the CRD's createOp. If there is a single member
		// shape and that member shape is a structure, unwrap it.
		for _, memberRef := range outputShape.MemberRefs {
			if memberRef.Shape.Type == "structure" {
				outputShape = memberRef.Shape
			}
		}
	}
	outAttrs, err = h.getAttrsFromOutputShape(outputShape, crdName)
	if err != nil {
		return nil, nil, err
	}
	return inAttrs, outAttrs, nil
}
