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

package model

import (
	"fmt"
	"sort"
	"strings"

	awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"

	"github.com/aws/aws-controllers-k8s/pkg/names"
	"github.com/aws/aws-controllers-k8s/pkg/util"
)

// GoCodeSetExceptionMessagePrefixCheck returns Go code that contains a
// condition to check if the message_prefix specified for a particular HTTP status code
// in generator config is a prefix for the exception message returned by AWS API.
// If message_prefix field was not specified for this HTTP code in generator config,
// we return an empty string
//
// Sample Output:
//
// && strings.HasPrefix(awsErr.Message(), "Could not find model")
func (r *CRD) GoCodeSetExceptionMessagePrefixCheck(httpStatusCode int) string {
	if r.cfg != nil {
		resGenConfig, found := r.cfg.Resources[r.Names.Original]
		if found && resGenConfig.Exceptions != nil {
			if excConfig, present := resGenConfig.Exceptions.Errors[httpStatusCode]; present &&
				resGenConfig.Exceptions.Errors[httpStatusCode].MessagePrefix != nil {
				return fmt.Sprintf("&& strings.HasPrefix(awsErr.Message(), \"%s\") ",
					*excConfig.MessagePrefix)
			}
		}
	}
	return ""
}

// GoCodeRequiredFieldsMissingFromShape returns Go code that contains a
// condition checking that the required fields in the supplied Shape have a
// non-nil value in the corresponding CR's Spec or Status substruct.
//
// Sample Output:
//
// return r.ko.Spec.APIID == nil || r.ko.Status.RouteID != nil
func (r *CRD) GoCodeRequiredFieldsMissingFromShape(
	opType OpType,
	koVarName string,
	indentLevel int,
) string {
	var op *awssdkmodel.Operation
	switch opType {
	case OpTypeGet:
		op = r.Ops.ReadOne
	case OpTypeGetAttributes:
		op = r.Ops.GetAttributes
	case OpTypeSetAttributes:
		op = r.Ops.SetAttributes
	default:
		return ""
	}

	shape := op.InputRef.Shape
	return r.goCodeRequiredFieldsMissingFromShape(
		koVarName,
		indentLevel,
		shape,
	)
}

func (r *CRD) goCodeRequiredFieldsMissingFromShape(
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

// GoCodeSetInput returns the Go code that sets an input shape's member fields
// from a CRD's fields.
//
// Assume a CRD called Repository that looks like this pseudo-schema:
//
// .Status
//   .Authors ([]*string)
//   .ImageData
//     .Location (*string)
//     .Tag (*string)
//   .Name (*string)
//
// And assume an SDK Shape CreateRepositoryInput that looks like this
// pseudo-schema:
//
// .Repository
//   .Authors ([]*string)
//   .ImageData
//     .Location (*string)
//     .Tag (*string)
//   .Name
//
// This function is called from a template that generates the Go code that
// represents linkage between the Kubernetes objects (CRs) and the aws-sdk-go
// (SDK) objects. If we call this function with the following parameters:
//
//  opType:			OpTypeCreate
//  sourceVarName:	ko
//  targetVarName:	res
//  indentLevel:	1
//
// Then this function should output something like this:
//
//   field1 := []*string{}
//   for _, elem0 := range r.ko.Spec.Authors {
//       elem0 := &string{*elem0}
//       field0 = append(field0, elem0)
//   }
//   res.Authors = field1
//   field1 := &svcsdk.ImageData{}
//   field1.SetLocation(*r.ko.Spec.ImageData.Location)
//   field1.SetTag(*r.ko.Spec.ImageData.Tag)
//   res.ImageData = field1
//	 res.SetName(*r.ko.Spec.Name)
//
// Note that for scalar fields, we use the SetXXX methods that are on all
// aws-sdk-go SDK structs
func (r *CRD) GoCodeSetInput(
	// The type of operation to look for the Input shape
	opType OpType,
	// String representing the name of the variable that we will grab the Input
	// shape from. This will likely be "r.ko" since in the templates that call
	// this method, the "source variable" is the CRD struct which is used to
	// populate the target variable, which is the Input shape
	sourceVarName string,
	// String representing the name of the variable that we will be **setting**
	// with values we get from the Output shape. This will likely be
	// "res" since that is the name of the "target variable" that the
	// templates that call this method use for the Input shape.
	targetVarName string,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	var op *awssdkmodel.Operation
	switch opType {
	case OpTypeCreate:
		op = r.Ops.Create
	case OpTypeGet:
		op = r.Ops.ReadOne
	case OpTypeList:
		op = r.Ops.ReadMany
	case OpTypeUpdate:
		op = r.Ops.Update
	case OpTypeDelete:
		op = r.Ops.Delete
	default:
		return ""
	}
	if op == nil {
		return ""
	}
	inputShape := op.InputRef.Shape
	if inputShape == nil {
		return ""
	}

	out := "\n"
	indent := strings.Repeat("\t", indentLevel)

	// Some input shapes for APIs that use GetAttributes API calls don't have
	// an Attributes member (example: all the Delete shapes...)
	_, foundAttrs := inputShape.MemberRefs["Attributes"]
	if r.UnpacksAttributesMap() && foundAttrs {
		// For APIs that use a pattern of a parameter called "Attributes" that
		// is of type `map[string]*string` to represent real, schema'd fields,
		// we need to set the input shape's "Attributes" member field to the
		// re-constructed, packed set of fields.
		//
		// Therefore, we output here something like this (example from SNS
		// Topic's Attributes map):
		//
		// attrMap := map[string]*string{}
		// attrMap["DeliveryPolicy"] = r.ko.Spec.DeliveryPolicy
		// attrMap["DisplayName"} = r.ko.Spec.DisplayName
		// attrMap["KmsMasterKeyId"] = r.ko.Spec.KMSMasterKeyID
		// attrMap["Policy"] = r.ko.Spec.Policy
		// res.SetAttributes(attrMap)
		attrMapConfig := r.cfg.Resources[r.Names.Original].UnpackAttributesMapConfig
		out += fmt.Sprintf("%sattrMap := map[string]*string{}\n", indent)
		sortedAttrFieldNames := []string{}
		for fieldName := range attrMapConfig.Fields {
			sortedAttrFieldNames = append(sortedAttrFieldNames, fieldName)
		}
		sort.Strings(sortedAttrFieldNames)
		for _, fieldName := range sortedAttrFieldNames {
			fieldConfig := attrMapConfig.Fields[fieldName]
			fieldNames := names.New(fieldName)
			if !fieldConfig.IsReadOnly {
				sourceAdaptedVarName := sourceVarName + ".Spec." + fieldNames.Camel
				out += fmt.Sprintf(
					"%sif %s != nil {\n",
					indent, sourceAdaptedVarName,
				)
				out += fmt.Sprintf(
					"%s\tattrMap[\"%s\"] = %s\n",
					indent, fieldName, sourceAdaptedVarName,
				)
				out += fmt.Sprintf(
					"%s}\n", indent,
				)
			}
		}
		out += fmt.Sprintf("%s%s.SetAttributes(attrMap)\n", indent, targetVarName)
	}

	opConfig, override := r.cfg.OverrideValues(op.Name)

	for memberIndex, memberName := range inputShape.MemberNames() {
		if r.UnpacksAttributesMap() && memberName == "Attributes" {
			continue
		}

		if override {
			value, ok := opConfig[memberName]
			memberShapeRef, _ := inputShape.MemberRefs[memberName]
			memberShape := memberShapeRef.Shape

			if ok {
				switch memberShape.Type {
				case "boolean", "integer":
				case "string":
					value = "\"" + value + "\""
				default:
					panic("Member type not handled")
				}

				out += fmt.Sprintf("%s%s.Set%s(%s)\n", indent, targetVarName, memberName, value)
				continue
			}
		}

		if r.IsPrimaryARNField(memberName) {
			// if ko.Status.ACKResourceMetadata != nil && ko.Status.ACKResourceMetadata.ARN != nil {
			//     res.SetTopicArn(string(*ko.Status.ACKResourceMetadata.ARN))
			// } else {
			//     res.SetTopicArn(rm.ARNFromName(*ko.Spec.Name))
			// }
			out += fmt.Sprintf(
				"%sif %s.Status.ACKResourceMetadata != nil && %s.Status.ACKResourceMetadata.ARN != nil {\n",
				indent, sourceVarName, sourceVarName,
			)
			out += fmt.Sprintf(
				"%s\t%s.Set%s(string(*%s.Status.ACKResourceMetadata.ARN))\n",
				indent, targetVarName, memberName, sourceVarName,
			)
			out += fmt.Sprintf(
				"%s} else {\n", indent,
			)
			nameField := r.NameField()
			out += fmt.Sprintf(
				"%s\t%s.Set%s(rm.ARNFromName(*%s.Spec.%s))\n",
				indent, targetVarName, memberName, sourceVarName, nameField,
			)
			out += fmt.Sprintf(
				"%s}\n", indent,
			)
			continue
		}
		renamedName, _ := r.InputFieldRename(op.Name, memberName)
		// Determine whether the input shape's field is in the Spec or the
		// Status struct and set the source variable appropriately.
		var f *Field
		var found bool
		sourceAdaptedVarName := sourceVarName
		f, found = r.SpecFields[renamedName]
		if found {
			sourceAdaptedVarName += r.cfg.PrefixConfig.SpecField
		} else {
			f, found = r.StatusFields[memberName]
			if !found {
				// TODO(jaypipes): check generator config for exceptions?
				continue
			}
			sourceAdaptedVarName += r.cfg.PrefixConfig.StatusField
		}
		sourceAdaptedVarName += "." + f.Names.Camel

		memberShapeRef, _ := inputShape.MemberRefs[memberName]
		memberShape := memberShapeRef.Shape

		// we construct variables containing temporary storage for sub-elements
		// and sub-fields that are structs. Names of fields are "f" appended by
		// the 0-based index of the field within the set of the target struct's
		// set of fields. Nested structs simply append another "f" and the
		// field index to the variable name.
		//
		// This means you can tell what field a temporary fields variable
		// represents by the name.
		//
		// For example, the field variable name "f0f5f2", it contains the third
		// field of the sixth field of the first field of the input shape being
		// constructed.
		//
		// If we have two levels of nested struct fields, we will end
		// up with a targetVarName of "field0f0f0" and the generated code
		// might look something like this:
		//
		// res := &sdkapi.CreateBookInput{}
		// f0 := &sdkapi.BookData{}
		// if ko.Spec.Author != nil {
		//     f0f0 := &sdkapi.Author{}
		//     if ko.Spec.Author.Address != nil {
		//         f0f0f0 := &sdkapi.Address{}
		//         f0f0f0.SetStreet(*ko.Spec.Author.Address.Street)
		//         f0f0f0.SetCity(*ko.Spec.Author.Address.City)
		//         f0f0f0.SetState(*ko.Spec.Author.Address.State)
		//         f0f0.Address = f0f0f0
		//     }
		//     if ko.Spec.Author.Name != nil {
		//         f0f0.SetName(*r.ko.Author.Name)
		//         f0.Author = f0f0
		//     }
		//     res.Book = f0
		// }
		//
		// It's ugly but at least consistent and mostly readable...
		//
		// For populating list fields, we need an iterator and a temporary
		// element variable. We name these "{fieldName}iter" and
		// "{fieldName}elem" respectively. For nested levels, the names will be
		// progressively longer.
		//
		// For list fields, we want to end up with something like this:
		//
		// res := &sdkapi.CreateCustomAvailabilityZoneInput{}
		// if ko.Spec.VPNGroupsMemberships != nil {
		//     f0 := []*sdkapi.VpnGroupMembership{}
		//     for _, f0iter := ko.Spec.VPNGroupMemberships {
		//         f0elem := &sdkapi.VpnGroupMembership{}
		//         f0elem.SetVpnId(f0elem.VPNID)
		//         f0 := append(f0, f0elem)
		//     }
		//     res.VpnMemberships = f0
		// }
		out += fmt.Sprintf(
			"%sif %s != nil {\n", indent, sourceAdaptedVarName,
		)

		switch memberShape.Type {
		case "list", "structure", "map":
			{
				memberVarName := fmt.Sprintf("f%d", memberIndex)
				out += r.goCodeVarEmptyConstructorSDKType(
					memberVarName,
					memberShape,
					indentLevel+1,
				)
				out += r.goCodeSetInputForContainer(
					memberName,
					memberVarName,
					sourceAdaptedVarName,
					memberShapeRef,
					indentLevel+1,
				)
				out += r.goCodeSetInputForScalar(
					memberName,
					targetVarName,
					inputShape.Type,
					memberVarName,
					memberShapeRef,
					indentLevel+1,
				)
			}
		default:
			out += r.goCodeSetInputForScalar(
				memberName,
				targetVarName,
				inputShape.Type,
				sourceAdaptedVarName,
				memberShapeRef,
				indentLevel+1,
			)
		}
		out += fmt.Sprintf(
			"%s}\n", indent,
		)
	}
	return out
}

// GoCodeGetAttributesSetInput returns the Go code that sets the Input shape for a
// resource's GetAttributes operation.
//
// As an example, for the GetTopicAttributes SNS API call, the returned code
// looks like this:
//
// res.SetTopicArn(string(*r.ko.Status.ACKResourceMetadata.ARN))
//
// For the SQS API's GetQueueAttributes call, the returned code looks like this:
//
// res.SetQueueUrl(*r.ko.Status.QueueURL)
//
// You will note the difference due to the special handling of the ARN fields.
func (r *CRD) GoCodeGetAttributesSetInput(
	// String representing the name of the variable that we will grab the
	// Input shape from. This will likely be "r.ko.Spec" since in the templates
	// that call this method, the "source variable" is the CRD struct's Spec
	// field which is used to populate the target variable, which is the Input
	// shape
	sourceVarName string,
	// String representing the name of the variable that we will be **setting**
	// with values we get from the Output shape. This will likely be
	// "res" since that is the name of the "target variable" that the
	// templates that call this method use for the Input shape.
	targetVarName string,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	op := r.Ops.GetAttributes
	if op == nil {
		return ""
	}
	inputShape := op.InputRef.Shape
	if inputShape == nil {
		return ""
	}
	if !r.UnpacksAttributesMap() {
		// This is a bug in the code generation if this occurs...
		msg := fmt.Sprintf("called GoCodeGetAttributesSetInput for a resource '%s' that doesn't unpack attributes map", r.Names.Original)
		panic(msg)
	}

	out := "\n"
	indent := strings.Repeat("\t", indentLevel)

	inputFieldOverrides := map[string][]string{}
	attrCfg := r.cfg.Resources[r.Names.Original].UnpackAttributesMapConfig
	if attrCfg.GetAttributesInput != nil {
		for memberName, override := range attrCfg.GetAttributesInput.Overrides {
			inputFieldOverrides[memberName] = override.Values
		}
	}

	for _, memberName := range inputShape.MemberNames() {
		if r.IsPrimaryARNField(memberName) {
			// if ko.Status.ACKResourceMetadata != nil && ko.Status.ACKResourceMetadata.ARN != nil {
			//     res.SetTopicArn(string(*ko.Status.ACKResourceMetadata.ARN))
			// } else {
			//     res.SetTopicArn(rm.ARNFromName(*ko.Spec.Name))
			// }
			out += fmt.Sprintf(
				"%sif %s.Status.ACKResourceMetadata != nil && %s.Status.ACKResourceMetadata.ARN != nil {\n",
				indent, sourceVarName, sourceVarName,
			)
			out += fmt.Sprintf(
				"%s\t%s.Set%s(string(*%s.Status.ACKResourceMetadata.ARN))\n",
				indent, targetVarName, memberName, sourceVarName,
			)
			out += fmt.Sprintf(
				"%s} else {\n", indent,
			)
			nameField := r.NameField()
			out += fmt.Sprintf(
				"%s\t%s.Set%s(rm.ARNFromName(*%s.Spec.%s))\n",
				indent, targetVarName, memberName, sourceVarName, nameField,
			)
			out += fmt.Sprintf(
				"%s}\n", indent,
			)
			continue
		}

		// Some APIs to retrieve the attributes for a resource require passing
		// specific fields and field values. For example, in order to get all
		// of an SQS Queue's attributes, the SQS GetQueueAttributes API call's
		// Input shape's AttributeNames member needs to be set to
		// []string{"All"}...
		//
		// Go code output in this section will look something like this:
		//
		// {
		//     tmpVals := []*string{}
		//     tmpVal0 := "All"
		//     tmpVals = append(tmpVals, &tmpVal0)
		//     res.SetAttributeNames(tmpVals)
		// }
		if overrideValues, ok := inputFieldOverrides[memberName]; ok {
			memberShapeRef := inputShape.MemberRefs[memberName]
			out += fmt.Sprintf("%s{\n", indent)
			// We need to output a set of temporary strings that we will take a
			// reference to when constructing the values of the []*string or
			// *string members.
			if memberShapeRef.Shape.Type == "list" {
				out += fmt.Sprintf("%s\ttmpVals := []*string{}\n", indent)
				for x, overrideValue := range overrideValues {
					out += fmt.Sprintf("%s\ttmpVal%d := \"%s\"\n", indent, x, overrideValue)
					out += fmt.Sprintf("%s\ttmpVals = append(tmpVals, &tmpVal%d)\n", indent, x)
				}
				out += fmt.Sprintf("%s\t%s.Set%s(tmpVals)\n", indent, targetVarName, memberName)
			} else {
				out += fmt.Sprintf("%s\ttmpVal := \"%s\"\n", indent, overrideValues[0])
				out += fmt.Sprintf("%s\t%s.Set%s(&tmpVal)\n", indent, targetVarName, memberName)
			}
			out += fmt.Sprintf("%s}\n", indent)
			continue
		}

		cleanMemberNames := names.New(memberName)
		cleanMemberName := cleanMemberNames.Camel

		sourceVarPath := sourceVarName
		field, found := r.SpecFields[memberName]
		if found {
			sourceVarPath = sourceVarName + ".Spec." + cleanMemberName
		} else {
			field, found = r.StatusFields[memberName]
			if !found {
				// If it isn't in our spec/status fields, just ignore it
				continue
			}
			sourceVarPath = sourceVarPath + ".Status." + cleanMemberName
		}
		out += fmt.Sprintf(
			"%sif %s != nil {\n",
			indent, sourceVarPath,
		)
		out += r.goCodeSetInputForScalar(
			memberName,
			targetVarName,
			inputShape.Type,
			sourceVarPath,
			field.ShapeRef,
			indentLevel+1,
		)
		out += fmt.Sprintf(
			"%s}\n", indent,
		)
	}
	return out
}

// GoCodeSetAttributesSetInput returns the Go code that sets the Input shape for a
// resource's SetAttributes operation.
//
// Unfortunately, the AWS SetAttributes API operations (even within the *same*
// API) are inconsistent regarding whether the SetAttributes sets a batch of
// attributes or a single attribute. We need to construct the method
// differently depending on this behaviour. For example, the SNS
// SetTopicAttributes API call actually only allows the caller to set a single
// attribute, which needs to be specified in an AttributeName and
// AttributeValue field in the Input shape. On the other hand, the SNS
// SetPlatformApplicationAttributes API call's Input shape has an Attributes
// field which is a map[string]string containing all the attribute key/value
// pairs to replace. Your guess is as good as mine as to why these APIs are
// different.
//
// The returned code looks something like this:
//
// attrMap := map[string]*string{}
// if r.ko.Spec.DeliveryPolicy != nil {
//     attrMap["DeliveryPolicy"] = r.ko.Spec.DeliveryPolicy
// }
// if r.ko.Spec.DisplayName != nil {
//     attrMap["DisplayName"} = r.ko.Spec.DisplayName
// }
// if r.ko.Spec.KMSMasterKeyID != nil {
//     attrMap["KmsMasterKeyId"] = r.ko.Spec.KMSMasterKeyID
// }
// if r.ko.Spec.Policy != nil {
//     attrMap["Policy"] = r.ko.Spec.Policy
// }
// res.SetAttributes(attrMap)
func (r *CRD) GoCodeSetAttributesSetInput(
	// String representing the name of the variable that we will grab the Input
	// shape from. This will likely be "r.ko" since in the templates that call
	// this method, the "source variable" is the CRD struct which is used to
	// populate the target variable, which is the Input shape
	sourceVarName string,
	// String representing the name of the variable that we will be **setting**
	// with values we get from the Output shape. This will likely be
	// "res" since that is the name of the "target variable" that the
	// templates that call this method use for the Input shape.
	targetVarName string,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	op := r.Ops.SetAttributes
	if op == nil {
		return ""
	}
	inputShape := op.InputRef.Shape
	if inputShape == nil {
		return ""
	}
	if !r.UnpacksAttributesMap() {
		// This is a bug in the code generation if this occurs...
		msg := fmt.Sprintf("called GoCodeSetAttributesSetInput for a resource '%s' that doesn't unpack attributes map", r.Names.Original)
		panic(msg)
	}

	if r.SetAttributesSingleAttribute() {
		// TODO(jaypipes): For now, because these APIs require *multiple* calls
		// to the backend, one for each attribute being set, we'll go ahead and
		// rely on the CustomOperation functionality to write code for these...
		return ""
	}

	out := "\n"
	indent := strings.Repeat("\t", indentLevel)

	for _, memberName := range inputShape.MemberNames() {
		if r.IsPrimaryARNField(memberName) {
			// if ko.Status.ACKResourceMetadata != nil && ko.Status.ACKResourceMetadata.ARN != nil {
			//     res.SetTopicArn(string(*ko.Status.ACKResourceMetadata.ARN))
			// } else {
			//     res.SetTopicArn(rm.ARNFromName(*ko.Spec.Name))
			// }
			out += fmt.Sprintf(
				"%sif %s.Status.ACKResourceMetadata != nil && %s.Status.ACKResourceMetadata.ARN != nil {\n",
				indent, sourceVarName, sourceVarName,
			)
			out += fmt.Sprintf(
				"%s\t%s.Set%s(string(*%s.Status.ACKResourceMetadata.ARN))\n",
				indent, targetVarName, memberName, sourceVarName,
			)
			out += fmt.Sprintf(
				"%s} else {\n", indent,
			)
			nameField := r.NameField()
			out += fmt.Sprintf(
				"%s\t%s.Set%s(rm.ARNFromName(*%s.Spec.%s))\n",
				indent, targetVarName, memberName, sourceVarName, nameField,
			)
			out += fmt.Sprintf(
				"%s}\n", indent,
			)
			continue
		}
		if memberName == "Attributes" {
			// For APIs that use a pattern of a parameter called "Attributes" that
			// is of type `map[string]*string` to represent real, schema'd fields,
			// we need to set the input shape's "Attributes" member field to the
			// re-constructed, packed set of fields.
			//
			// Therefore, we output here something like this (example from SNS
			// Topic's Attributes map):
			//
			// attrMap := map[string]*string{}
			// if r.ko.Spec.DeliveryPolicy != nil {
			//     attrMap["DeliveryPolicy"] = r.ko.Spec.DeliveryPolicy
			// }
			// if r.ko.Spec.DisplayName != nil {
			//     attrMap["DisplayName"} = r.ko.Spec.DisplayName
			// }
			// if r.ko.Spec.KMSMasterKeyID != nil {
			//     attrMap["KmsMasterKeyId"] = r.ko.Spec.KMSMasterKeyID
			// }
			// if r.ko.Spec.Policy != nil {
			//     attrMap["Policy"] = r.ko.Spec.Policy
			// }
			// res.SetAttributes(attrMap)
			attrMapConfig := r.cfg.Resources[r.Names.Original].UnpackAttributesMapConfig
			out += fmt.Sprintf("%sattrMap := map[string]*string{}\n", indent)
			sortedAttrFieldNames := []string{}
			for fieldName := range attrMapConfig.Fields {
				sortedAttrFieldNames = append(sortedAttrFieldNames, fieldName)
			}
			sort.Strings(sortedAttrFieldNames)
			for _, fieldName := range sortedAttrFieldNames {
				fieldConfig := attrMapConfig.Fields[fieldName]
				fieldNames := names.New(fieldName)
				if !fieldConfig.IsReadOnly {
					sourceAdaptedVarName := sourceVarName + ".Spec." + fieldNames.Camel
					out += fmt.Sprintf(
						"%sif %s != nil {\n",
						indent, sourceAdaptedVarName,
					)
					out += fmt.Sprintf(
						"%s\tattrMap[\"%s\"] = %s\n",
						indent, fieldName, sourceAdaptedVarName,
					)
					out += fmt.Sprintf(
						"%s}\n", indent,
					)
				}
			}
			out += fmt.Sprintf("%s%s.SetAttributes(attrMap)\n", indent, targetVarName)
			continue
		}

		// Handle setting any other Input shape fields that are not the ARN
		// field or the Attributes unpacked map. The field value may come from
		// either the Spec or the Status fields.
		cleanMemberNames := names.New(memberName)
		cleanMemberName := cleanMemberNames.Camel

		sourceVarPath := sourceVarName
		field, found := r.SpecFields[memberName]
		if found {
			sourceVarPath = sourceVarName + ".Spec." + cleanMemberName
		} else {
			field, found = r.StatusFields[memberName]
			if !found {
				// If it isn't in our spec/status fields, just ignore it
				continue
			}
			sourceVarPath = sourceVarPath + ".Status." + cleanMemberName
		}
		out += fmt.Sprintf(
			"%sif %s != nil {\n",
			indent, sourceVarPath,
		)
		out += r.goCodeSetInputForScalar(
			memberName,
			targetVarName,
			inputShape.Type,
			sourceVarPath,
			field.ShapeRef,
			indentLevel+1,
		)
		out += fmt.Sprintf(
			"%s}\n", indent,
		)
	}
	return out
}

// goCodeSetInputForContainer returns a string of Go code that sets the value
// of a target variable to that of a source variable. When the source variable
// type is a map, struct or slice type, then this function is called
// recursively on the elements or members of the source variable.
func (r *CRD) goCodeSetInputForContainer(
	// The name of the SDK Input shape member we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	// ShapeRef of the target struct field
	targetShapeRef *awssdkmodel.ShapeRef,
	indentLevel int,
) string {
	switch targetShapeRef.Shape.Type {
	case "structure":
		return r.goCodeSetInputForStruct(
			targetFieldName,
			targetVarName,
			targetShapeRef,
			sourceVarName,
			indentLevel,
		)
	case "list":
		return r.goCodeSetInputForSlice(
			targetFieldName,
			targetVarName,
			targetShapeRef,
			sourceVarName,
			indentLevel,
		)
	case "map":
		return r.goCodeSetInputForMap(
			targetFieldName,
			targetVarName,
			targetShapeRef,
			sourceVarName,
			indentLevel,
		)
	default:
		return r.goCodeSetInputForScalar(
			targetFieldName,
			targetVarName,
			targetShapeRef.Shape.Type,
			sourceVarName,
			targetShapeRef,
			indentLevel,
		)
	}
}

// goCodeSetInputForStruct returns a string of Go code that sets a target
// variable value to a source variable when the type of the source variable is
// a struct.
func (r *CRD) goCodeSetInputForStruct(
	// The name of the CR field we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// Shape Ref of the target struct field
	targetShapeRef *awssdkmodel.ShapeRef,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	targetShape := targetShapeRef.Shape

	for memberIndex, memberName := range targetShape.MemberNames() {
		memberShapeRef := targetShape.MemberRefs[memberName]
		memberShape := memberShapeRef.Shape
		cleanMemberNames := names.New(memberName)
		cleanMemberName := cleanMemberNames.Camel
		memberVarName := fmt.Sprintf("%sf%d", targetVarName, memberIndex)
		sourceAdaptedVarName := sourceVarName + "." + cleanMemberName
		out += fmt.Sprintf(
			"%sif %s != nil {\n", indent, sourceAdaptedVarName,
		)
		switch memberShape.Type {
		case "list", "structure", "map":
			{
				out += r.goCodeVarEmptyConstructorSDKType(
					memberVarName,
					memberShape,
					indentLevel+1,
				)
				out += r.goCodeSetInputForContainer(
					memberName,
					memberVarName,
					sourceAdaptedVarName,
					memberShapeRef,
					indentLevel+1,
				)
				out += r.goCodeSetInputForScalar(
					memberName,
					targetVarName,
					targetShape.Type,
					memberVarName,
					memberShapeRef,
					indentLevel+1,
				)
			}
		default:
			out += r.goCodeSetInputForScalar(
				memberName,
				targetVarName,
				targetShape.Type,
				sourceAdaptedVarName,
				memberShapeRef,
				indentLevel+1,
			)
		}
		out += fmt.Sprintf(
			"%s}\n", indent,
		)
	}
	return out
}

// goCodeSetInputForSlice returns a string of Go code that sets a target
// variable value to a source variable when the type of the source variable is
// a struct.
func (r *CRD) goCodeSetInputForSlice(
	// The name of the CR field we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// Shape Ref of the target struct field
	targetShapeRef *awssdkmodel.ShapeRef,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	targetShape := targetShapeRef.Shape

	iterVarName := fmt.Sprintf("%siter", targetVarName)
	elemVarName := fmt.Sprintf("%selem", targetVarName)
	// for _, f0iter := range r.ko.Spec.Tags {
	out += fmt.Sprintf("%sfor _, %s := range %s {\n", indent, iterVarName, sourceVarName)
	//		f0elem := string{}
	out += r.goCodeVarEmptyConstructorSDKType(
		elemVarName,
		targetShape.MemberRef.Shape,
		indentLevel+1,
	)
	//  f0elem = *f0iter
	//
	// or
	//
	//  f0elem.SetMyField(*f0iter)
	containerFieldName := ""
	if targetShape.MemberRef.Shape.Type == "structure" {
		containerFieldName = targetFieldName
	}
	out += r.goCodeSetInputForContainer(
		containerFieldName,
		elemVarName,
		iterVarName,
		&targetShape.MemberRef,
		indentLevel+1,
	)
	addressOfVar := ""
	switch targetShape.MemberRef.Shape.Type {
	case "structure", "list", "map":
		break
	default:
		addressOfVar = "&"
	}
	//  f0 = append(f0, elem0)
	out += fmt.Sprintf("%s\t%s = append(%s, %s%s)\n", indent, targetVarName, targetVarName, addressOfVar, elemVarName)
	out += fmt.Sprintf("%s}\n", indent)
	return out
}

// goCodeSetInputForMap returns a string of Go code that sets a target
// variable value to a source variable when the type of the source variable is
// a struct.
func (r *CRD) goCodeSetInputForMap(
	// The name of the CR field we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// Shape Ref of the target struct field
	targetShapeRef *awssdkmodel.ShapeRef,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	targetShape := targetShapeRef.Shape

	valIterVarName := fmt.Sprintf("%svaliter", targetVarName)
	keyVarName := fmt.Sprintf("%skey", targetVarName)
	valVarName := fmt.Sprintf("%sval", targetVarName)
	// for f0key, f0valiter := range r.ko.Spec.Tags {
	out += fmt.Sprintf("%sfor %s, %s := range %s {\n", indent, keyVarName, valIterVarName, sourceVarName)
	//		f0elem := string{}
	out += r.goCodeVarEmptyConstructorSDKType(
		valVarName,
		targetShape.ValueRef.Shape,
		indentLevel+1,
	)
	//  f0val = *f0valiter
	//
	// or
	//
	//  f0val.SetMyField(*f0valiter)
	containerFieldName := ""
	if targetShape.ValueRef.Shape.Type == "structure" {
		containerFieldName = targetFieldName
	}
	out += r.goCodeSetInputForContainer(
		containerFieldName,
		valVarName,
		valIterVarName,
		&targetShape.ValueRef,
		indentLevel+1,
	)
	addressOfVar := ""
	switch targetShape.ValueRef.Shape.Type {
	case "structure", "list", "map":
		break
	default:
		addressOfVar = "&"
	}
	// f0[f0key] = f0val
	out += fmt.Sprintf("%s\t%s[%s] = %s%s\n", indent, targetVarName, keyVarName, addressOfVar, valVarName)
	out += fmt.Sprintf("%s}\n", indent)
	return out
}

func (r *CRD) goCodeVarEmptyConstructorSDKType(
	varName string,
	// The shape we want to construct a new thing for
	shape *awssdkmodel.Shape,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	goType := shape.GoTypeWithPkgName()
	keepPointer := (shape.Type == "list" || shape.Type == "map")
	goType = replacePkgName(goType, r.sdkAPI.API.PackageName(), "svcsdk", keepPointer)
	switch shape.Type {
	case "structure":
		// f0 := &svcsdk.BookData{}
		out += fmt.Sprintf("%s%s := &%s{}\n", indent, varName, goType)
	case "list", "map":
		// f0 := []*string{}
		out += fmt.Sprintf("%s%s := %s{}\n", indent, varName, goType)
	default:
		// var f0 string
		out += fmt.Sprintf("%svar %s %s\n", indent, varName, goType)
	}
	return out
}

func (r *CRD) goCodeVarEmptyConstructorK8sType(
	varName string,
	// The shape we want to construct a new thing for
	shape *awssdkmodel.Shape,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	goType := shape.GoTypeWithPkgName()
	keepPointer := (shape.Type == "list" || shape.Type == "map")
	goType = replacePkgName(goType, r.sdkAPI.API.PackageName(), "svcapitypes", keepPointer)
	goTypeNoPkg := goType
	goPkg := ""
	hadPkg := false
	if strings.Contains(goType, ".") {
		parts := strings.Split(goType, ".")
		goTypeNoPkg = parts[1]
		goPkg = parts[0]
		hadPkg = true
	}
	renames := r.sdkAPI.GetTypeRenames(r.cfg)
	altTypeName, renamed := renames[goTypeNoPkg]
	if renamed {
		goTypeNoPkg = altTypeName
	} else if hadPkg {
		cleanNames := names.New(goTypeNoPkg)
		goTypeNoPkg = cleanNames.Camel
	}
	goType = goTypeNoPkg
	if hadPkg {
		goType = goPkg + "." + goType
	}

	switch shape.Type {
	case "structure":
		// f0 := &svcapitypes.BookData{}
		out += fmt.Sprintf("%s%s := &%s{}\n", indent, varName, goType)
	case "list", "map":
		// f0 := []*string{}
		out += fmt.Sprintf("%s%s := %s{}\n", indent, varName, goType)
	default:
		// var f0 string
		out += fmt.Sprintf("%svar %s %s\n", indent, varName, goType)
	}
	return out
}

// goCodeSetInputForScalar returns the Go code that sets the value of a target
// variable or field to a scalar value. For target variables that are structs,
// we output the aws-sdk-go's common SetXXX() method. For everything else, we
// output normal assignment operations.
func (r *CRD) goCodeSetInputForScalar(
	// The name of the Input SDK Shape member we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// The type of shape of the target variable
	targetVarType string,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	shapeRef *awssdkmodel.ShapeRef,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	setTo := sourceVarName
	shape := shapeRef.Shape
	if shape.Type == "timestamp" {
		setTo += ".Time"
	} else if shapeRef.UseIndirection() {
		setTo = "*" + setTo
	}
	if targetVarType == "structure" {
		out += fmt.Sprintf("%s%s.Set%s(%s)\n", indent, targetVarName, targetFieldName, setTo)
	} else {
		targetVarPath := targetVarName
		if targetFieldName != "" {
			targetVarPath += "." + targetFieldName
		}
		out += fmt.Sprintf("%s%s = %s\n", indent, targetVarPath, setTo)
	}
	return out
}

// GoCodeSetOutput returns the Go code that sets a CRD's field value to
// the value of an output shape's member fields.
// Status fields are always updated. Update of Spec fields depends on
// 'performSpecUpdate' parameter
//
// Assume a CRD called Repository that looks like this pseudo-schema:
//
// .Status
//   .Authors ([]*string)
//   .ImageData
//     .Location (*string)
//     .Tag (*string)
//   .Name (*string)
//
// And assume an SDK Shape CreateRepositoryOutput that looks like this
// pseudo-schema:
//
// .Repository
//   .Authors ([]*string)
//   .ImageData
//     .Location (*string)
//     .Tag (*string)
//   .Name
//
// This function is called from a template that generates the Go code that
// represents linkage between the Kubernetes objects (CRs) and the aws-sdk-go
// (SDK) objects. If we call this function with the following parameters:
//
//  opType:			OpTypeCreate
//  sourceVarName:	resp
//  targetVarName:	ko.Status
//  indentLevel:	1
//
// Then this function should output something like this:
//
//   field0 := []*string{}
//   for _, iter0 := range resp.Authors {
//       var elem0 string
//       elem0 = *iter
//       field0 = append(field0, &elem0)
//   }
//   ko.Status.Authors = field0
//   field1 := &svcapitypes.ImageData{}
//   field1.Location = resp.ImageData.Location
//   field1.Tag = resp.ImageData.Tag
//   ko.Status.ImageData = field1
//   ko.Status.Name = resp.Name
func (r *CRD) GoCodeSetOutput(
	// The type of operation to look for the Output shape
	opType OpType,
	// String representing the name of the variable that we will grab the
	// Output shape from. This will likely be "resp" since in the templates
	// that call this method, the "source variable" is the response struct
	// returned by the aws-sdk-go's SDK API call corresponding to the Operation
	sourceVarName string,
	// String representing the name of the variable that we will be **setting**
	// with values we get from the Output shape. This will likely be
	// "ko.Status" since that is the name of the "target variable" that the
	// templates that call this method use.
	targetVarName string,
	// Number of levels of indentation to use
	indentLevel int,
	// boolean to indicate whether Spec fields should be updated from opTypeOutput
	performSpecUpdate bool,
) string {
	var op *awssdkmodel.Operation
	switch opType {
	case OpTypeCreate:
		op = r.Ops.Create
	case OpTypeGet:
		op = r.Ops.ReadOne
	case OpTypeList:
		return r.goCodeSetOutputReadMany(
			r.Ops.ReadMany, sourceVarName, targetVarName, indentLevel,
		)
	case OpTypeUpdate:
		op = r.Ops.Update
	case OpTypeDelete:
		op = r.Ops.Delete
	default:
		return ""
	}
	if op == nil {
		return ""
	}
	outputShape := op.OutputRef.Shape
	if outputShape == nil {
		return ""
	}

	// We might be in a "wrapper" shape. Unwrap it to find the real object
	// representation for the CRD's createOp. If there is a single member
	// shape and that member shape is a structure, unwrap it.
	if outputShape.UsedAsOutput && len(outputShape.MemberRefs) == 1 {
		for memberName, memberRef := range outputShape.MemberRefs {
			if memberRef.Shape.Type == "structure" {
				sourceVarName += "." + memberName
				outputShape = memberRef.Shape
			}
		}
	}
	out := "\n"
	indent := strings.Repeat("\t", indentLevel)

	// Recursively descend down through the set of fields on the Output shape,
	// creating temporary variables, populating those temporary variables'
	// fields with further-nested fields as needed
	for memberIndex, memberName := range outputShape.MemberNames() {
		//TODO: (vijat@) should these field be renamed before looking them up in spec?
		sourceAdaptedVarName := sourceVarName + "." + memberName

		// Handle the special case of ARN for primary resource identifier
		if r.IsPrimaryARNField(memberName) {
			// if ko.Status.ACKResourceMetadata == nil {
			//     ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
			// }
			out += fmt.Sprintf(
				"%sif %s.Status.ACKResourceMetadata == nil {\n",
				indent,
				targetVarName,
			)
			out += fmt.Sprintf(
				"%s\t%s.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}\n",
				indent,
				targetVarName,
			)
			out += fmt.Sprintf("%s}\n", indent)

			// if resp.BookArn != nil {
			//     ko.Status.ACKResourceMetadata.ARN = resp.BookArn
			// }
			out += fmt.Sprintf(
				"%sif %s != nil {\n",
				indent,
				sourceAdaptedVarName,
			)
			out += fmt.Sprintf(
				"%s\tarn := ackv1alpha1.AWSResourceName(*%s)\n",
				indent,
				sourceAdaptedVarName,
			)
			out += fmt.Sprintf(
				"%s\t%s.Status.ACKResourceMetadata.ARN = &arn\n",
				indent,
				targetVarName,
			)
			out += fmt.Sprintf("%s}\n", indent)
			continue
		}

		sourceMemberShapeRef := outputShape.MemberRefs[memberName]
		if sourceMemberShapeRef.Shape == nil {
			// Technically this should not happen, so let's bail here if it
			// does...
			msg := fmt.Sprintf(
				"expected .Shape to not be nil for ShapeRef of memberName %s",
				memberName,
			)
			panic(msg)
		}

		sourceMemberShape := sourceMemberShapeRef.Shape

		// Determine whether the input shape's field is in the Spec or the
		// Status struct and set the source variable appropriately.
		var f *Field
		var found bool
		var targetMemberShapeRef *awssdkmodel.ShapeRef
		targetAdaptedVarName := targetVarName
		f, found = r.SpecFields[memberName]
		if found {
			targetAdaptedVarName += r.cfg.PrefixConfig.SpecField
			if !performSpecUpdate {
				continue
			}
		} else {
			f, found = r.StatusFields[memberName]
			if !found {
				// TODO(jaypipes): check generator config for exceptions?
				continue
			}
			targetAdaptedVarName += r.cfg.PrefixConfig.StatusField
		}
		targetMemberShapeRef = f.ShapeRef
		// fieldVarName is the name of the variable that is used for temporary
		// storage of complex member field values
		//
		// For struct fields, we want to output code sort of like this:
		//
		//   field0 := &svapitypes.ImageData{}
		//   if resp.ImageData.Location != nil {
		//	     field0.Location = resp.ImageData.Location
		//   }
		//   if resp.ImageData.Tag != nil {
		//       field0.Tag = resp.ImageData.Tag
		//   }
		//   r.ko.Status.ImageData = field0
		//   if resp.Name != nil {
		//	     r.ko.Status.Name = resp.Name
		//   }
		//
		// For list fields, we want to end up with something like this:
		//
		// field0 := []*svcapitypes.VpnGroupMembership{}
		// for _, iter0 := resp.CustomAvailabilityZone.VpnGroupMemberships {
		//     elem0 := &svcapitypes.VPNGroupMembership{}
		//     if iter0.VPNID != nil {
		//         elem0.VPNID = iter0.VPNID
		//     }
		//     field0 := append(field0, elem0)
		// }
		// ko.Status.VpnMemberships = field0

		out += fmt.Sprintf(
			"%sif %s != nil {\n", indent, sourceAdaptedVarName,
		)

		switch sourceMemberShape.Type {
		case "list", "structure", "map":
			{
				memberVarName := fmt.Sprintf("f%d", memberIndex)
				out += r.goCodeVarEmptyConstructorK8sType(
					memberVarName,
					targetMemberShapeRef.Shape,
					indentLevel+1,
				)
				out += r.goCodeSetOutputForContainer(
					f.Names.Camel,
					memberVarName,
					targetMemberShapeRef,
					sourceAdaptedVarName,
					sourceMemberShapeRef,
					indentLevel+1,
				)
				out += r.goCodeSetOutputForScalar(
					f.Names.Camel,
					targetAdaptedVarName,
					memberVarName,
					sourceMemberShapeRef,
					indentLevel+1,
				)
			}
		default:
			out += r.goCodeSetOutputForScalar(
				f.Names.Camel,
				targetAdaptedVarName,
				sourceAdaptedVarName,
				sourceMemberShapeRef,
				indentLevel+1,
			)
		}
		out += fmt.Sprintf(
			"%s}\n", indent,
		)
	}
	return out
}

// goCodeSetOutputReadMany sets the supplied target variable from the results
// of a List operation. This is a special-case handling of those APIs where
// there is no ReadOne operation and instead the only way to grab information
// for a single object is to call the ReadMany/List operation with one of more
// filtering fields and then look for one element in the returned array of
// results and unpack that into the target variable.
//
// As an example, for the DescribeCacheClusters Elasticache API call, the
// returned code looks like this:
//
// Note: "resp" is the source variable and represents the
//       DescribeCacheClustersOutput shape/struct in the aws-sdk-go API for
//       Elasticache
// Note: "ko" is the target variable and represents the thing we'll be
//		 setting fields on
//
//  if len(resp.CacheClusters) == 0 {
//      return nil, ackerr.NotFound
//  }
//  found := false
//  for _, elem := range resp.CacheClusters {
//      if elem.ARN != nil {
//          if ko.Status.ACKResourceMetadata == nil {
//              ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
//          }
//          tmpARN := ackv1alpha1.AWSResourceName(*elemARN)
//          ko.Status.ACKResourceMetadata.ARN = &tmpARN
//      }
//      if elem.AtRestEncryptionEnabled != nil {
//          ko.Status.AtRestEncryptionEnabled = elem.AtRestEncryptionEnabled
//      }
//      ...
//      if elem.CacheClusterId != nil {
//          if ko.Spec.CacheClusterID != nil {
//              if *elem.CacheClusterId != *ko.Spec.CacheClusterID {
//                  continue
//              }
//          }
//          r.ko.Spec.CacheClusterID = elem.CacheClusterId
//      }
//      found = true
//  }
//  if !found {
//      return nil, ackerr.NotFound
//  }
func (r *CRD) goCodeSetOutputReadMany(
	// The ReadMany operation descriptor
	op *awssdkmodel.Operation,
	// String representing the name of the variable that we will grab the
	// Output shape from. This will likely be "resp" since in the templates
	// that call this method, the "source variable" is the response struct
	// returned by the aws-sdk-go's SDK API call corresponding to the Operation
	sourceVarName string,
	// String representing the name of the variable that we will be **setting**
	// with values we get from the Output shape. This will likely be
	// "ko" since that is the name of the "target variable" that the
	// templates that call this method use.
	targetVarName string,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	outputShape := op.OutputRef.Shape
	if outputShape == nil {
		return ""
	}

	out := "\n"
	indent := strings.Repeat("\t", indentLevel)

	listShapeName := ""
	var sourceElemShape *awssdkmodel.Shape

	// Find the element in the output shape that contains the list of
	// resources. This heuristic is simplistic (just look for the field with a
	// list type) but seems to be followed consistently by the aws-sdk-go for
	// List operations.
	for memberName, memberShapeRef := range outputShape.MemberRefs {
		if memberShapeRef.Shape.Type == "list" {
			listShapeName = memberName
			sourceElemShape = memberShapeRef.Shape.MemberRef.Shape
			break
		}
	}

	if listShapeName == "" {
		panic("List output shape had no field of type 'list'")
	}

	// Set of field names in the element shape that, if the generator config
	// instructs us to, we will write Go code to filter results of the List
	// operation by checking for matching values in these fields.
	matchFieldNames := r.listOpMatchFieldNames()

	// found := false
	out += fmt.Sprintf("%sfound := false\n", indent)
	// for _, elem := range resp.CacheClusters {
	out += fmt.Sprintf(
		"%sfor _, elem := range %s.%s {\n",
		indent, sourceVarName, listShapeName,
	)
	for memberIndex, memberName := range sourceElemShape.MemberNames() {
		sourceMemberShapeRef := sourceElemShape.MemberRefs[memberName]
		sourceMemberShape := sourceMemberShapeRef.Shape
		sourceAdaptedVarName := "elem." + memberName
		if r.IsPrimaryARNField(memberName) {
			out += fmt.Sprintf(
				"%s\tif %s != nil {\n", indent, sourceAdaptedVarName,
			)
			//     if ko.Status.ACKResourceMetadata == nil {
			//  	   ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
			//     }
			out += fmt.Sprintf(
				"%s\t\tif %s.Status.ACKResourceMetadata == nil {\n",
				indent, targetVarName,
			)
			out += fmt.Sprintf(
				"%s\t\t\t%s.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}\n",
				indent, targetVarName,
			)
			out += fmt.Sprintf(
				"\t\t%s}\n", indent,
			)
			//          tmpARN := ackv1alpha1.AWSResourceName(*elemARN)
			//  		ko.Status.ACKResourceMetadata.ARN = &tmpARN
			out += fmt.Sprintf(
				"%s\t\ttmpARN := ackv1alpha1.AWSResourceName(*%s)\n",
				indent,
				sourceAdaptedVarName,
			)
			out += fmt.Sprintf(
				"%s\t\t%s.Status.ACKResourceMetadata.ARN = &tmpARN\n",
				indent,
				targetVarName,
			)
			out += fmt.Sprintf(
				"\t%s}\n", indent,
			)
			continue
		}
		// Determine whether the input shape's field is in the Spec or the
		// Status struct and set the source variable appropriately.
		var f *Field
		var found bool
		var targetMemberShapeRef *awssdkmodel.ShapeRef
		targetAdaptedVarName := targetVarName
		f, found = r.SpecFields[memberName]
		if found {
			targetAdaptedVarName += r.cfg.PrefixConfig.SpecField
		} else {
			f, found = r.StatusFields[memberName]
			if !found {
				// TODO(jaypipes): check generator config for exceptions?
				continue
			}
			targetAdaptedVarName += r.cfg.PrefixConfig.StatusField
		}
		targetMemberShapeRef = f.ShapeRef
		out += fmt.Sprintf(
			"%s\tif %s != nil {\n", indent, sourceAdaptedVarName,
		)
		switch sourceMemberShape.Type {
		case "list", "structure", "map":
			{
				memberVarName := fmt.Sprintf("f%d", memberIndex)
				out += r.goCodeVarEmptyConstructorK8sType(
					memberVarName,
					targetMemberShapeRef.Shape,
					indentLevel+2,
				)
				out += r.goCodeSetOutputForContainer(
					f.Names.Camel,
					memberVarName,
					targetMemberShapeRef,
					sourceAdaptedVarName,
					sourceMemberShapeRef,
					indentLevel+2,
				)
				out += r.goCodeSetOutputForScalar(
					f.Names.Camel,
					targetAdaptedVarName,
					memberVarName,
					sourceMemberShapeRef,
					indentLevel+2,
				)
			}
		default:
			//          if ko.Spec.CacheClusterID != nil {
			//              if *elem.CacheClusterId != *ko.Spec.CacheClusterID {
			//                  continue
			//              }
			//          }
			if util.InStrings(memberName, matchFieldNames) {
				out += fmt.Sprintf(
					"%s\t\tif %s.%s != nil {\n",
					indent,
					targetAdaptedVarName,
					f.Names.Camel,
				)
				out += fmt.Sprintf(
					"%s\t\t\tif *%s != *%s.%s {\n",
					indent,
					sourceAdaptedVarName,
					targetAdaptedVarName,
					f.Names.Camel,
				)
				out += fmt.Sprintf(
					"%s\t\t\t\tcontinue\n", indent,
				)
				out += fmt.Sprintf(
					"%s\t\t\t}\n", indent,
				)
				out += fmt.Sprintf(
					"%s\t\t}\n", indent,
				)
			}
			//          r.ko.Spec.CacheClusterID = elem.CacheClusterId
			out += r.goCodeSetOutputForScalar(
				f.Names.Camel,
				targetAdaptedVarName,
				sourceAdaptedVarName,
				sourceMemberShapeRef,
				indentLevel+2,
			)
		}
		out += fmt.Sprintf(
			"%s\t}\n", indent,
		)
	}
	// When we don't have custom matching/filtering logic for the list
	// operation, we just take the first element in the returned slice
	// of objects. When we DO have match fields, the generated Go code
	// above will output a `continue` when the required fields don't
	// match. Thus, we will break here only when getting a record where
	// all match fields have matched.
	out += fmt.Sprintf(
		"%s\tfound = true\n", indent,
	)
	out += fmt.Sprintf(
		"%s\tbreak\n", indent,
	)
	out += fmt.Sprintf("%s}\n", indent)
	//  if !found {
	//      return nil, ackerr.NotFound
	//  }
	out += fmt.Sprintf("%sif !found {\n", indent)
	out += fmt.Sprintf("%s\t%s\n", indent, r.cfg.SetManyOutputNotFoundErrReturn)
	out += fmt.Sprintf("%s}\n", indent)
	return out
}

// goCodeACKResourceMetadataGuardConstructor returns Go code representing a
// nil-guard and constructor for an ACKResourceMetadata struct:
//
// if ko.Status.ACKResourceMetadata == nil {
//     ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
// }
func goCodeACKResourceMetadataGuardConstructor(
	// String representing the name of the variable that we will be **setting**
	// with values we get from the Output shape. This will likely be
	// "ko.Status" since that is the name of the "target variable" that the
	// templates that call this method use.
	targetVarName string,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	indent := strings.Repeat("\t", indentLevel)
	out := fmt.Sprintf(
		"%sif %s.ACKResourceMetadata == nil {\n",
		indent,
		targetVarName,
	)
	out += fmt.Sprintf(
		"%s\t%s.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}\n",
		indent,
		targetVarName,
	)
	out += fmt.Sprintf("%s}\n", indent)
	return out
}

// GoCodeGetAttributesSetOutput returns the Go code that sets the Status fields
// from the Output shape returned from a resource's GetAttributes operation.
//
// As an example, for the GetTopicAttributes SNS API call, the returned code
// looks like this:
//
// if ko.Status.ACKResourceMetadata == nil {
//     ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
// }
// ko.Status.EffectiveDeliveryPolicy = resp.Attributes["EffectiveDeliveryPolicy"]
// ko.Status.ACKResourceMetadata.OwnerAccountID = ackv1alpha1.AWSAccountID(resp.Attributes["Owner"])
// ko.Status.ACKResourceMetadata.ARN = ackv1alpha1.AWSResourceName(resp.Attributes["TopicArn"])
func (r *CRD) GoCodeGetAttributesSetOutput(
	// String representing the name of the variable that we will grab the
	// Output shape from. This will likely be "resp" since in the templates
	// that call this method, the "source variable" is the response struct
	// returned by the aws-sdk-go's SDK API call corresponding to the Operation
	sourceVarName string,
	// String representing the name of the variable that we will be **setting**
	// with values we get from the Output shape. This will likely be
	// "ko.Status" since that is the name of the "target variable" that the
	// templates that call this method use.
	targetVarName string,
	// Number of levels of indentation to use
	indentLevel int,
) string {
	if !r.UnpacksAttributesMap() {
		// This is a bug in the code generation if this occurs...
		msg := fmt.Sprintf("called GoCodeGetAttributesSetOutput for a resource '%s' that doesn't unpack attributes map", r.Ops.GetAttributes.Name)
		panic(msg)
	}
	op := r.Ops.GetAttributes
	if op == nil {
		return ""
	}
	inputShape := op.InputRef.Shape
	if inputShape == nil {
		return ""
	}

	out := "\n"
	indent := strings.Repeat("\t", indentLevel)

	// did we output an ACKResourceMetadata guard and constructor snippet?
	mdGuardOut := false
	attrMapConfig := r.cfg.Resources[r.Names.Original].UnpackAttributesMapConfig
	sortedAttrFieldNames := []string{}
	for fieldName := range attrMapConfig.Fields {
		sortedAttrFieldNames = append(sortedAttrFieldNames, fieldName)
	}
	sort.Strings(sortedAttrFieldNames)
	for _, fieldName := range sortedAttrFieldNames {
		if r.IsPrimaryARNField(fieldName) {
			if !mdGuardOut {
				out += goCodeACKResourceMetadataGuardConstructor(
					targetVarName, indentLevel,
				)
				mdGuardOut = true
			}
			out += fmt.Sprintf(
				"%stmpARN := ackv1alpha1.AWSResourceName(*%s.Attributes[\"%s\"])\n",
				indent,
				sourceVarName,
				fieldName,
			)
			out += fmt.Sprintf(
				"%s%s.ACKResourceMetadata.ARN = &tmpARN\n",
				indent,
				targetVarName,
			)
			continue
		}

		fieldConfig := attrMapConfig.Fields[fieldName]
		if fieldConfig.IsOwnerAccountID {
			if !mdGuardOut {
				out += goCodeACKResourceMetadataGuardConstructor(
					targetVarName, indentLevel,
				)
				mdGuardOut = true
			}
			out += fmt.Sprintf(
				"%stmpOwnerID := ackv1alpha1.AWSAccountID(*%s.Attributes[\"%s\"])\n",
				indent,
				sourceVarName,
				fieldName,
			)
			out += fmt.Sprintf(
				"%s%s.ACKResourceMetadata.OwnerAccountID = &tmpOwnerID\n",
				indent,
				targetVarName,
			)
			continue
		}

		fieldNames := names.New(fieldName)
		if fieldConfig.IsReadOnly {
			out += fmt.Sprintf(
				"%s%s.%s = %s.Attributes[\"%s\"]\n",
				indent,
				targetVarName,
				fieldNames.Camel,
				sourceVarName,
				fieldName,
			)
		}
	}
	return out
}

// goCodeSetOutputForContainer returns a string of Go code that sets the value
// of a target variable to that of a source variable. When the source variable
// type is a map, struct or slice type, then this function is called
// recursively on the elements or members of the source variable.
func (r *CRD) goCodeSetOutputForContainer(
	// The name of the CR field we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// Shape Ref of the target struct field
	targetShapeRef *awssdkmodel.ShapeRef,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	// ShapeRef of the source struct field
	sourceShapeRef *awssdkmodel.ShapeRef,
	indentLevel int,
) string {
	switch sourceShapeRef.Shape.Type {
	case "structure":
		return r.goCodeSetOutputForStruct(
			targetFieldName,
			targetVarName,
			targetShapeRef,
			sourceVarName,
			sourceShapeRef,
			indentLevel,
		)
	case "list":
		return r.goCodeSetOutputForSlice(
			targetFieldName,
			targetVarName,
			targetShapeRef,
			sourceVarName,
			sourceShapeRef,
			indentLevel,
		)
	case "map":
		return r.goCodeSetOutputForMap(
			targetFieldName,
			targetVarName,
			targetShapeRef,
			sourceVarName,
			sourceShapeRef,
			indentLevel,
		)
	default:
		return r.goCodeSetOutputForScalar(
			targetFieldName,
			targetVarName,
			sourceVarName,
			sourceShapeRef,
			indentLevel,
		)
	}
}

// goCodeSetOutputForStruct returns a string of Go code that sets a target
// variable value to a source variable when the type of the source variable is
// a struct.
func (r *CRD) goCodeSetOutputForStruct(
	// The name of the CR field we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// Shape Ref of the target struct field
	targetShapeRef *awssdkmodel.ShapeRef,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	// ShapeRef of the source struct field
	sourceShapeRef *awssdkmodel.ShapeRef,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	sourceShape := sourceShapeRef.Shape
	targetShape := targetShapeRef.Shape

	for memberIndex, memberName := range sourceShape.MemberNames() {
		targetMemberShapeRef := targetShape.MemberRefs[memberName]
		if targetMemberShapeRef == nil {
			continue
		}
		memberVarName := fmt.Sprintf("%sf%d", targetVarName, memberIndex)
		memberShapeRef := sourceShape.MemberRefs[memberName]
		memberShape := memberShapeRef.Shape
		cleanNames := names.New(memberName)
		sourceAdaptedVarName := sourceVarName + "." + memberName
		out += fmt.Sprintf(
			"%sif %s != nil {\n", indent, sourceAdaptedVarName,
		)
		switch memberShape.Type {
		case "list", "structure", "map":
			{
				out += r.goCodeVarEmptyConstructorK8sType(
					memberVarName,
					targetMemberShapeRef.Shape,
					indentLevel+1,
				)
				out += r.goCodeSetOutputForContainer(
					cleanNames.Camel,
					memberVarName,
					targetMemberShapeRef,
					sourceAdaptedVarName,
					memberShapeRef,
					indentLevel+1,
				)
				out += r.goCodeSetOutputForScalar(
					cleanNames.Camel,
					targetVarName,
					memberVarName,
					memberShapeRef,
					indentLevel+1,
				)
			}
		default:
			out += r.goCodeSetOutputForScalar(
				cleanNames.Camel,
				targetVarName,
				sourceAdaptedVarName,
				memberShapeRef,
				indentLevel+1,
			)
		}
		out += fmt.Sprintf(
			"%s}\n", indent,
		)
	}
	return out
}

// goCodeSetOutputForSlice returns a string of Go code that sets a target
// variable value to a source variable when the type of the source variable is
// a slice.
func (r *CRD) goCodeSetOutputForSlice(
	// The name of the CR field we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// Shape Ref of the target slice field
	targetShapeRef *awssdkmodel.ShapeRef,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	// ShapeRef of the source slice field
	sourceShapeRef *awssdkmodel.ShapeRef,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	sourceShape := sourceShapeRef.Shape
	targetShape := targetShapeRef.Shape
	iterVarName := fmt.Sprintf("%siter", targetVarName)
	elemVarName := fmt.Sprintf("%selem", targetVarName)
	// for _, f0iter0 := range resp.TagSpecifications {
	out += fmt.Sprintf("%sfor _, %s := range %s {\n", indent, iterVarName, sourceVarName)
	//		var f0elem0 string
	out += r.goCodeVarEmptyConstructorK8sType(
		elemVarName,
		targetShape.MemberRef.Shape,
		indentLevel+1,
	)
	//  f0elem0 = *f0iter0
	//
	// or
	//
	//  f0elem0.SetMyField(*f0iter0)
	containerFieldName := ""
	if sourceShape.MemberRef.Shape.Type == "structure" {
		containerFieldName = targetFieldName
	}
	out += r.goCodeSetOutputForContainer(
		containerFieldName,
		elemVarName,
		&targetShape.MemberRef,
		iterVarName,
		&sourceShape.MemberRef,
		indentLevel+1,
	)
	addressOfVar := ""
	switch sourceShape.MemberRef.Shape.Type {
	case "structure", "list", "map":
		break
	default:
		addressOfVar = "&"
	}
	//  f0 = append(f0, elem0)
	out += fmt.Sprintf("%s\t%s = append(%s, %s%s)\n", indent, targetVarName, targetVarName, addressOfVar, elemVarName)
	out += fmt.Sprintf("%s}\n", indent)
	return out
}

// goCodeSetOutputForMap returns a string of Go code that sets a target
// variable value to a source variable when the type of the source variable is
// a map.
func (r *CRD) goCodeSetOutputForMap(
	// The name of the CR field we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// Shape Ref of the target map field
	targetShapeRef *awssdkmodel.ShapeRef,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	// ShapeRef of the source map field
	sourceShapeRef *awssdkmodel.ShapeRef,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	sourceShape := sourceShapeRef.Shape
	targetShape := targetShapeRef.Shape

	valIterVarName := fmt.Sprintf("%svaliter", targetVarName)
	keyVarName := fmt.Sprintf("%skey", targetVarName)
	valVarName := fmt.Sprintf("%sval", targetVarName)
	// for f0key, f0valiter := range resp.Tags {
	out += fmt.Sprintf("%sfor %s, %s := range %s {\n", indent, keyVarName, valIterVarName, sourceVarName)
	//		f0elem := string{}
	out += r.goCodeVarEmptyConstructorK8sType(
		valVarName,
		targetShape.ValueRef.Shape,
		indentLevel+1,
	)
	//  f0val = *f0valiter
	containerFieldName := ""
	if sourceShape.ValueRef.Shape.Type == "structure" {
		containerFieldName = targetFieldName
	}
	out += r.goCodeSetOutputForContainer(
		containerFieldName,
		valVarName,
		&targetShape.ValueRef,
		valIterVarName,
		&sourceShape.ValueRef,
		indentLevel+1,
	)
	addressOfVar := ""
	switch sourceShape.ValueRef.Shape.Type {
	case "structure", "list", "map":
		break
	default:
		addressOfVar = "&"
	}
	// f0[f0key] = f0val
	out += fmt.Sprintf("%s\t%s[%s] = %s%s\n", indent, targetVarName, keyVarName, addressOfVar, valVarName)
	out += fmt.Sprintf("%s}\n", indent)
	return out
}

// goCodeSetOutputForScalar returns a string of Go code that sets a target
// variable value to a source variable when the type of the source variable is
// a scalar type (not a map, slice or struct).
func (r *CRD) goCodeSetOutputForScalar(
	// The name of the Input SDK Shape member we're outputting for
	targetFieldName string,
	// The variable name that we want to set a value to
	targetVarName string,
	// The struct or struct field that we access our source value from
	sourceVarName string,
	shapeRef *awssdkmodel.ShapeRef,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	setTo := sourceVarName
	shape := shapeRef.Shape
	if shape.Type == "timestamp" {
		setTo = "&metav1.Time{*" + sourceVarName + "}"
	}
	targetVarPath := targetVarName
	if targetFieldName != "" {
		targetVarPath += "." + targetFieldName
	} else {
		setTo = "*" + setTo
	}
	out += fmt.Sprintf("%s%s = %s\n", indent, targetVarPath, setTo)
	return out
}
