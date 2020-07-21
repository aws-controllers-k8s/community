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
	ttpl "text/template"

	"github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/aws/aws-controllers-k8s/pkg/template"
)

type CRDSDKGoTemplateVars struct {
	APIVersion   string
	APIGroup     string
	ServiceAlias string
	CRD          *model.CRD
}

func NewCRDSDKGoTemplate(tplDir string) (*ttpl.Template, error) {
	tplPath := filepath.Join(tplDir, "pkg", "crd_sdk.go.tpl")
	tplContents, err := ioutil.ReadFile(tplPath)
	if err != nil {
		return nil, err
	}
	t := ttpl.New("crd_sdk")
	t = t.Funcs(ttpl.FuncMap{
		"GoCodeSetReadOneOutput": func(r *model.CRD, outVarName string, koVarName string, indentLevel int) string {
			return r.GoCodeSetOutput(model.OpTypeGet, outVarName, koVarName, indentLevel)
		},
		"GoCodeSetReadOneInput": func(r *model.CRD, inVarName string, koVarName string, indentLevel int) string {
			return r.GoCodeSetInput(model.OpTypeGet, inVarName, koVarName, indentLevel)
		},
		"GoCodeSetCreateOutput": func(r *model.CRD, outVarName string, koVarName string, indentLevel int) string {
			return r.GoCodeSetOutput(model.OpTypeCreate, outVarName, koVarName, indentLevel)
		},
		"GoCodeSetCreateInput": func(r *model.CRD, inVarName string, koVarName string, indentLevel int) string {
			return r.GoCodeSetInput(model.OpTypeCreate, inVarName, koVarName, indentLevel)
		},
		"GoCodeSetUpdateOutput": func(r *model.CRD, outVarName string, koVarName string, indentLevel int) string {
			return r.GoCodeSetOutput(model.OpTypeUpdate, outVarName, koVarName, indentLevel)
		},
		"GoCodeSetUpdateInput": func(r *model.CRD, inVarName string, koVarName string, indentLevel int) string {
			return r.GoCodeSetInput(model.OpTypeUpdate, inVarName, koVarName, indentLevel)
		},
		"GoCodeSetDeleteInput": func(r *model.CRD, inVarName string, koVarName string, indentLevel int) string {
			return r.GoCodeSetInput(model.OpTypeDelete, inVarName, koVarName, indentLevel)
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
