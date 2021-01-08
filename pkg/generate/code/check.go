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

package code

import (
	"fmt"
	"strings"

	awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"

	ackgenconfig "github.com/aws/aws-controllers-k8s/pkg/generate/config"
	"github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/aws/aws-controllers-k8s/pkg/names"
)

// CheckExceptionMessagePrefix returns Go code that contains a condition to
// check if the message_prefix specified for a particular HTTP status code in
// generator config is a prefix for the exception message returned by AWS API.
// If message_prefix field was not specified for this HTTP code in generator
// config, we return an empty string
//
// Sample Output:
//
// && strings.HasPrefix(awsErr.Message(), "Could not find model")
func CheckExceptionMessagePrefix(
	cfg *ackgenconfig.Config,
	r *model.CRD,
	httpStatusCode int,
) string {
	rConfig, ok := cfg.ResourceConfig(r.Names.Original)
	if ok && rConfig.Exceptions != nil {
		excConfig, ok := rConfig.Exceptions.Errors[httpStatusCode]
		if ok && excConfig.MessagePrefix != nil {
			return fmt.Sprintf("&& strings.HasPrefix(awsErr.Message(), \"%s\") ",
				*excConfig.MessagePrefix)
		}
	}
	return ""
}

// CheckRequiredFieldsMissingFromShape returns Go code that contains a
// condition checking that the required fields in the supplied Shape have a
// non-nil value in the corresponding CR's Spec or Status substruct.
//
// Sample Output:
//
// return r.ko.Spec.APIID == nil || r.ko.Status.RouteID == nil
func CheckRequiredFieldsMissingFromShape(
	r *model.CRD,
	opType model.OpType,
	koVarName string,
	indentLevel int,
) string {
	var op *awssdkmodel.Operation
	switch opType {
	case model.OpTypeGet:
		op = r.Ops.ReadOne
	case model.OpTypeGetAttributes:
		op = r.Ops.GetAttributes
	case model.OpTypeSetAttributes:
		op = r.Ops.SetAttributes
	default:
		return ""
	}

	shape := op.InputRef.Shape
	return checkRequiredFieldsMissingFromShape(
		r,
		koVarName,
		indentLevel,
		shape,
	)
}

func checkRequiredFieldsMissingFromShape(
	r *model.CRD,
	koVarName string,
	indentLevel int,
	shape *awssdkmodel.Shape,
) string {
	indent := strings.Repeat("\t", indentLevel)
	if shape == nil || len(shape.Required) == 0 {
		return fmt.Sprintf("%sreturn false", indent)
	}

	// Loop over the required member fields in the shape and identify whether
	// the field exists in either the Status or the Spec of the resource and
	// generate an if condition checking for all required fields having non-nil
	// corresponding resource Spec/Status values
	missing := []string{}
	for _, memberName := range shape.Required {
		if r.UnpacksAttributesMap() {
			// We set the Attributes field specially... depending on whether
			// the SetAttributes API call uses the batch or single attribute
			// flavor
			if r.SetAttributesSingleAttribute() {
				if memberName == "AttributeName" || memberName == "AttributeValue" {
					continue
				}
			} else {
				if memberName == "Attributes" {
					continue
				}
			}
		}
		if r.IsPrimaryARNField(memberName) {
			primaryARNCondition := fmt.Sprintf(
				"(%s.Status.ACKResourceMetadata == nil || %s.Status.ACKResourceMetadata.ARN == nil)",
				koVarName, koVarName,
			)
			missing = append(missing, primaryARNCondition)
			continue
		}
		cleanMemberNames := names.New(memberName)
		cleanMemberName := cleanMemberNames.Camel

		resVarPath := koVarName
		_, found := r.SpecFields[memberName]
		if found {
			resVarPath = resVarPath + ".Spec." + cleanMemberName
		} else {
			_, found = r.StatusFields[memberName]
			if !found {
				// If it isn't in our spec/status fields, we have a problem!
				msg := fmt.Sprintf(
					"GENERATION FAILURE! there's a required field %s in "+
						"Shape %s that isn't in either the CR's Spec or "+
						"Status structs!",
					memberName, shape.ShapeName,
				)
				panic(msg)
			}
			resVarPath = resVarPath + ".Status." + cleanMemberName
		}
		missing = append(missing, fmt.Sprintf("%s == nil", resVarPath))
	}
	// Use '||' because if any of the required fields are missing the object
	// is not created yet
	missingCondition := strings.Join(missing, " || ")
	return fmt.Sprintf("%sreturn %s\n", indent, missingCondition)
}
