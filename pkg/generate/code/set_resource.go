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
	"sort"
	"strings"

	awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"

	ackgenconfig "github.com/aws/aws-controllers-k8s/pkg/generate/config"
	"github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/aws/aws-controllers-k8s/pkg/names"
	"github.com/aws/aws-controllers-k8s/pkg/util"
)

// SetResource returns the Go code that sets a CRD's field value to the value
// of an output shape's member fields.  Status fields are always updated.
// Update of Spec fields depends on 'performSpecUpdate' parameter
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
func SetResource(
	cfg *ackgenconfig.Config,
	r *model.CRD,
	// The type of operation to look for the Output shape
	opType model.OpType,
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
	case model.OpTypeCreate:
		op = r.Ops.Create
	case model.OpTypeGet:
		op = r.Ops.ReadOne
	case model.OpTypeList:
		return setResourceReadMany(
			cfg, r,
			r.Ops.ReadMany, sourceVarName, targetVarName, indentLevel,
		)
	case model.OpTypeUpdate:
		op = r.Ops.Update
	case model.OpTypeDelete:
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
		var f *model.Field
		var found bool
		var targetMemberShapeRef *awssdkmodel.ShapeRef
		targetAdaptedVarName := targetVarName
		f, found = r.SpecFields[memberName]
		if found {
			targetAdaptedVarName += cfg.PrefixConfig.SpecField
			if !performSpecUpdate {
				continue
			}
		} else {
			f, found = r.StatusFields[memberName]
			if !found {
				// TODO(jaypipes): check generator config for exceptions?
				continue
			}
			targetAdaptedVarName += cfg.PrefixConfig.StatusField
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
				out += varEmptyConstructorK8sType(
					cfg, r,
					memberVarName,
					targetMemberShapeRef.Shape,
					indentLevel+1,
				)
				out += setResourceForContainer(
					cfg, r,
					f.Names.Camel,
					memberVarName,
					targetMemberShapeRef,
					sourceAdaptedVarName,
					sourceMemberShapeRef,
					indentLevel+1,
				)
				out += setResourceForScalar(
					cfg, r,
					f.Names.Camel,
					targetAdaptedVarName,
					memberVarName,
					sourceMemberShapeRef,
					indentLevel+1,
				)
			}
		default:
			out += setResourceForScalar(
				cfg, r,
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

// setResourceReadMany sets the supplied target variable from the results of a
// List operation. This is a special-case handling of those APIs where there is
// no ReadOne operation and instead the only way to grab information for a
// single object is to call the ReadMany/List operation with one of more
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
func setResourceReadMany(
	cfg *ackgenconfig.Config,
	r *model.CRD,
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
	matchFieldNames := r.ListOpMatchFieldNames()

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
		var f *model.Field
		var found bool
		var targetMemberShapeRef *awssdkmodel.ShapeRef
		targetAdaptedVarName := targetVarName
		f, found = r.SpecFields[memberName]
		if found {
			targetAdaptedVarName += cfg.PrefixConfig.SpecField
		} else {
			f, found = r.StatusFields[memberName]
			if !found {
				// TODO(jaypipes): check generator config for exceptions?
				continue
			}
			targetAdaptedVarName += cfg.PrefixConfig.StatusField
		}
		targetMemberShapeRef = f.ShapeRef
		out += fmt.Sprintf(
			"%s\tif %s != nil {\n", indent, sourceAdaptedVarName,
		)
		switch sourceMemberShape.Type {
		case "list", "structure", "map":
			{
				memberVarName := fmt.Sprintf("f%d", memberIndex)
				out += varEmptyConstructorK8sType(
					cfg, r,
					memberVarName,
					targetMemberShapeRef.Shape,
					indentLevel+2,
				)
				out += setResourceForContainer(
					cfg, r,
					f.Names.Camel,
					memberVarName,
					targetMemberShapeRef,
					sourceAdaptedVarName,
					sourceMemberShapeRef,
					indentLevel+2,
				)
				out += setResourceForScalar(
					cfg, r,
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
			out += setResourceForScalar(
				cfg, r,
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
	out += fmt.Sprintf("%s\t%s\n", indent, cfg.SetManyOutputNotFoundErrReturn)
	out += fmt.Sprintf("%s}\n", indent)
	return out
}

// ackResourceMetadataGuardConstructor returns Go code representing a nil-guard
// and constructor for an ACKResourceMetadata struct:
//
// if ko.Status.ACKResourceMetadata == nil {
//     ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
// }
func ackResourceMetadataGuardConstructor(
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

// SetResourceGetAttributes returns the Go code that sets the Status fields
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
func SetResourceGetAttributes(
	cfg *ackgenconfig.Config,
	r *model.CRD,
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
		msg := fmt.Sprintf(
			"called SetResourceGetAttributes for a resource '%s' that doesn't unpack attributes map",
			r.Ops.GetAttributes.Name,
		)
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
	adaptiveTargetVarName := targetVarName + cfg.PrefixConfig.StatusField

	// did we output an ACKResourceMetadata guard and constructor snippet?
	mdGuardOut := false
	fieldConfigs := cfg.ResourceFields(r.Names.Original)
	sortedAttrFieldNames := []string{}
	for fName, fConfig := range fieldConfigs {
		if fConfig.IsAttribute {
			sortedAttrFieldNames = append(sortedAttrFieldNames, fName)
		}
	}
	sort.Strings(sortedAttrFieldNames)
	for _, fieldName := range sortedAttrFieldNames {
		if r.IsPrimaryARNField(fieldName) {
			if !mdGuardOut {
				out += ackResourceMetadataGuardConstructor(
					adaptiveTargetVarName, indentLevel,
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
				adaptiveTargetVarName,
			)
			continue
		}

		fieldConfig := fieldConfigs[fieldName]
		if fieldConfig.IsOwnerAccountID && cfg.IncludeACKMetadata {
			if !mdGuardOut {
				out += ackResourceMetadataGuardConstructor(
					adaptiveTargetVarName, indentLevel,
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
				adaptiveTargetVarName,
			)
			continue
		}

		fieldNames := names.New(fieldName)
		if fieldConfig.IsReadOnly {
			out += fmt.Sprintf(
				"%s%s.%s = %s.Attributes[\"%s\"]\n",
				indent,
				adaptiveTargetVarName,
				fieldNames.Camel,
				sourceVarName,
				fieldName,
			)
		}
	}
	return out
}

// setResourceForContainer returns a string of Go code that sets the value of a
// target variable to that of a source variable. When the source variable type
// is a map, struct or slice type, then this function is called recursively on
// the elements or members of the source variable.
func setResourceForContainer(
	cfg *ackgenconfig.Config,
	r *model.CRD,
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
		return setResourceForStruct(
			cfg, r,
			targetFieldName,
			targetVarName,
			targetShapeRef,
			sourceVarName,
			sourceShapeRef,
			indentLevel,
		)
	case "list":
		return setResourceForSlice(
			cfg, r,
			targetFieldName,
			targetVarName,
			targetShapeRef,
			sourceVarName,
			sourceShapeRef,
			indentLevel,
		)
	case "map":
		return setResourceForMap(
			cfg, r,
			targetFieldName,
			targetVarName,
			targetShapeRef,
			sourceVarName,
			sourceShapeRef,
			indentLevel,
		)
	default:
		return setResourceForScalar(
			cfg, r,
			targetFieldName,
			targetVarName,
			sourceVarName,
			sourceShapeRef,
			indentLevel,
		)
	}
}

// setResourceForStruct returns a string of Go code that sets a target variable
// value to a source variable when the type of the source variable is a struct.
func setResourceForStruct(
	cfg *ackgenconfig.Config,
	r *model.CRD,
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
				out += varEmptyConstructorK8sType(
					cfg, r,
					memberVarName,
					targetMemberShapeRef.Shape,
					indentLevel+1,
				)
				out += setResourceForContainer(
					cfg, r,
					cleanNames.Camel,
					memberVarName,
					targetMemberShapeRef,
					sourceAdaptedVarName,
					memberShapeRef,
					indentLevel+1,
				)
				out += setResourceForScalar(
					cfg, r,
					cleanNames.Camel,
					targetVarName,
					memberVarName,
					memberShapeRef,
					indentLevel+1,
				)
			}
		default:
			out += setResourceForScalar(
				cfg, r,
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

// setResourceForSlice returns a string of Go code that sets a target variable
// value to a source variable when the type of the source variable is a slice.
func setResourceForSlice(
	cfg *ackgenconfig.Config,
	r *model.CRD,
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
	out += varEmptyConstructorK8sType(
		cfg, r,
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
	out += setResourceForContainer(
		cfg, r,
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

// setResourceForMap returns a string of Go code that sets a target variable
// value to a source variable when the type of the source variable is a map.
func setResourceForMap(
	cfg *ackgenconfig.Config,
	r *model.CRD,
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
	out += varEmptyConstructorK8sType(
		cfg, r,
		valVarName,
		targetShape.ValueRef.Shape,
		indentLevel+1,
	)
	//  f0val = *f0valiter
	containerFieldName := ""
	if sourceShape.ValueRef.Shape.Type == "structure" {
		containerFieldName = targetFieldName
	}
	out += setResourceForContainer(
		cfg, r,
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

// setResourceForScalar returns a string of Go code that sets a target variable
// value to a source variable when the type of the source variable is a scalar
// type (not a map, slice or struct).
func setResourceForScalar(
	cfg *ackgenconfig.Config,
	r *model.CRD,
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
