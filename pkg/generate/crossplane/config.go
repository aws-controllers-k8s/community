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

package crossplane

import (
	"strings"
	"text/template"

	"github.com/aws/aws-controllers-k8s/pkg/generate/config"
	"github.com/aws/aws-controllers-k8s/pkg/model"
)

// DefaultConfig is the default config object for Initialize controllers.
var DefaultConfig = config.Config{
	PrefixConfig: config.PrefixConfig{
		SpecField:   ".Spec.ForProvider",
		StatusField: ".Status.AtProvider",
	},
	IncludeACKMetadata:             false,
	SetManyOutputNotFoundErrReturn: "return cr",
}

var (
	// APIFilePairs is the list of files that will be generated for every AWS API,
	// such as apigatewayv2 or dynamodb.
	APIFilePairs = map[string]string{
		"apis/doc.go.tpl":               "apis/%s/%s/zz_doc.go",
		"apis/enums.go.tpl":             "apis/%s/%s/zz_enums.go",
		"apis/groupversion_info.go.tpl": "apis/%s/%s/zz_groupversion_info.go",
		"apis/types.go.tpl":             "apis/%s/%s/zz_types.go",
	}
	// ControllerFilePairs is the list of files that will be generated once for
	// every controller.
	ControllerFilePairs = map[string]string{
		"pkg/controller.go.tpl":  "pkg/controller/%s/%s/zz_controller.go",
		"pkg/conversions.go.tpl": "pkg/controller/%s/%s/zz_conversions.go",
		// TODO(muvaf): Hooks file needs to be generated only once.
		"pkg/hooks.go.tpl": "pkg/controller/%s/%s/hooks.go",
	}
	// CRDFilePairs is the list of files that will be generated once for every CRD
	// in apis folder.
	CRDFilePairs = map[string]string{
		"apis/crd.go.tpl": "apis/%s/%s/zz_%s.go",
	}
	// IncludePaths are templates snippets that are imported by other templates.
	IncludePaths = []string{
		"boilerplate.go.tpl",
		"apis/enum_def.go.tpl",
		"apis/type_def.go.tpl",
		"pkg/sdk_find_read_one.go.tpl",
		"pkg/sdk_find_read_many.go.tpl",
		"pkg/sdk_find_get_attributes.go.tpl",
	}
	// CopyPaths is the list of files that will be copied as is.
	CopyPaths = []string{}
)

var TemplateFuncs = template.FuncMap{
	"ToLower": strings.ToLower,
	"ResourceExceptionCode": func(r *model.CRD, httpStatusCode int) string {
		return r.ExceptionCode(httpStatusCode)
	},
	"GoCodeSetExceptionMessagePrefixCheck": func(r *model.CRD, httpStatusCode int) string {
		return r.GoCodeSetExceptionMessagePrefixCheck(httpStatusCode)
	},
	"GoCodeSetReadOneOutput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int, performSpecUpdate bool) string {
		return r.GoCodeSetOutput(model.OpTypeGet, sourceVarName, targetVarName, indentLevel, performSpecUpdate)
	},
	"GoCodeSetReadOneInput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
		return r.GoCodeSetInput(model.OpTypeGet, sourceVarName, targetVarName, indentLevel)
	},
	"GoCodeSetReadManyOutput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int, performSpecUpdate bool) string {
		return r.GoCodeSetOutput(model.OpTypeList, sourceVarName, targetVarName, indentLevel, performSpecUpdate)
	},
	"GoCodeSetReadManyInput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
		return r.GoCodeSetInput(model.OpTypeList, sourceVarName, targetVarName, indentLevel)
	},
	"GoCodeSetCreateOutput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int, performSpecUpdate bool) string {
		return r.GoCodeSetOutput(model.OpTypeCreate, sourceVarName, targetVarName, indentLevel, performSpecUpdate)
	},
	"GoCodeSetCreateInput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
		return r.GoCodeSetInput(model.OpTypeCreate, sourceVarName, targetVarName, indentLevel)
	},
	"GoCodeSetUpdateInput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
		return r.GoCodeSetInput(model.OpTypeUpdate, sourceVarName, targetVarName, indentLevel)
	},
	"GoCodeSetDeleteInput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
		return r.GoCodeSetInput(model.OpTypeDelete, sourceVarName, targetVarName, indentLevel)
	},
	"GoCodeGetAttributesSetInput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
		return r.GoCodeGetAttributesSetInput(sourceVarName, targetVarName, indentLevel)
	},
	"GoCodeSetAttributesSetInput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
		return r.GoCodeSetAttributesSetInput(sourceVarName, targetVarName, indentLevel)
	},
	"GoCodeGetAttributesSetOutput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
		return r.GoCodeGetAttributesSetOutput(sourceVarName, targetVarName, indentLevel)
	},
	"Empty": func(subject string) bool {
		return strings.TrimSpace(subject) == ""
	},
}
