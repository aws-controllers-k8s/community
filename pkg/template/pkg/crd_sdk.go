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

package pkg

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	ttpl "text/template"

	"github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/aws/aws-controllers-k8s/pkg/template"
)

type CRDSDKGoTemplateVars struct {
	APIVersion              string
	APIGroup                string
	ServiceAlias            string
	SDKAPIInterfaceTypeName string
	CRD                     *model.CRD
}

func NewCRDSDKGoTemplate(tplDir string) (*ttpl.Template, error) {
	tplPath := filepath.Join(tplDir, "pkg", "crd_sdk.go.tpl")
	tplContents, err := ioutil.ReadFile(tplPath)
	if err != nil {
		return nil, err
	}
	t := ttpl.New("crd_sdk")
	t = t.Funcs(ttpl.FuncMap{
		"ResourceExceptionCode": func(r *model.CRD, httpStatusCode int) string {
			return r.ExceptionCode(httpStatusCode)
		},
		"GoCodeSetReadOneOutput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetOutput(model.OpTypeGet, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetReadOneInput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(model.OpTypeGet, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetReadManyOutput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetOutput(model.OpTypeList, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetReadManyInput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(model.OpTypeList, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeGetAttributesSetInput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeGetAttributesSetInput(sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeGetAttributesSetOutput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeGetAttributesSetOutput(sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetCreateOutput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetOutput(model.OpTypeCreate, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetCreateInput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(model.OpTypeCreate, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetUpdateOutput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetOutput(model.OpTypeUpdate, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetUpdateInput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(model.OpTypeUpdate, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetDeleteInput": func(r *model.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(model.OpTypeDelete, sourceVarName, targetVarName, indentLevel)
		},
		"Empty": func(subject string) bool {
			return strings.TrimSpace(subject) == ""
		},
		"GoCodeRequiredStatusFieldsForReadOneInput": func(r *model.CRD, indentLevel int) string {
			return r.RequiredStatusFieldsForReadOneInput(indentLevel)
		},
	})
	if t, err = t.Parse(string(tplContents)); err != nil {
		return nil, err
	}
	includes := []string{
		"boilerplate",
	}
	for _, include := range includes {
		if t, err = template.IncludeTemplate(t, tplDir, include); err != nil {
			return nil, err
		}
	}
	return t, nil
}
