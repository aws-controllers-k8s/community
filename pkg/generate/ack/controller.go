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

package ack

import (
	"path/filepath"
	"strings"
	ttpl "text/template"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	"github.com/aws/aws-controllers-k8s/pkg/generate/templateset"
	ackmodel "github.com/aws/aws-controllers-k8s/pkg/model"
)

var (
	controllerConfigTemplatePaths = []string{
		"config/controller/deployment.yaml.tpl",
		"config/controller/kustomization.yaml.tpl",
		"config/default/kustomization.yaml.tpl",
		"config/rbac/cluster-role-binding.yaml.tpl",
		"config/rbac/role-reader.yaml.tpl",
		"config/rbac/role-writer.yaml.tpl",
		"config/rbac/kustomization.yaml.tpl",
		"config/crd/kustomization.yaml.tpl",
	}
	controllerIncludePaths = []string{
		"boilerplate.go.tpl",
		"pkg/resource/sdk_find_read_one.go.tpl",
		"pkg/resource/sdk_find_get_attributes.go.tpl",
		"pkg/resource/sdk_find_read_many.go.tpl",
		"pkg/resource/sdk_find_not_implemented.go.tpl",
		"pkg/resource/sdk_update.go.tpl",
		"pkg/resource/sdk_update_custom.go.tpl",
		"pkg/resource/sdk_update_set_attributes.go.tpl",
		"pkg/resource/sdk_update_not_implemented.go.tpl",
	}
	controllerCopyPaths = []string{}
	controllerFuncMap   = ttpl.FuncMap{
		"ToLower": strings.ToLower,
		"ResourceExceptionCode": func(r *ackmodel.CRD, httpStatusCode int) string {
			return r.ExceptionCode(httpStatusCode)
		},
		"GoCodeSetReadOneOutput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int, performSpecUpdate bool) string {
			return r.GoCodeSetOutput(ackmodel.OpTypeGet, sourceVarName, targetVarName, indentLevel, performSpecUpdate)
		},
		"GoCodeSetReadOneInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(ackmodel.OpTypeGet, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetReadManyOutput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int, performSpecUpdate bool) string {
			return r.GoCodeSetOutput(ackmodel.OpTypeList, sourceVarName, targetVarName, indentLevel, performSpecUpdate)
		},
		"GoCodeSetReadManyInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(ackmodel.OpTypeList, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeGetAttributesSetInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeGetAttributesSetInput(sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetAttributesSetInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetAttributesSetInput(sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeGetAttributesSetOutput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeGetAttributesSetOutput(sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetCreateOutput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int, performSpecUpdate bool) string {
			return r.GoCodeSetOutput(ackmodel.OpTypeCreate, sourceVarName, targetVarName, indentLevel, performSpecUpdate)
		},
		"GoCodeSetCreateInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(ackmodel.OpTypeCreate, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetUpdateOutput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int, performSpecUpdate bool) string {
			return r.GoCodeSetOutput(ackmodel.OpTypeUpdate, sourceVarName, targetVarName, indentLevel, performSpecUpdate)
		},
		"GoCodeSetUpdateInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(ackmodel.OpTypeUpdate, sourceVarName, targetVarName, indentLevel)
		},
		"GoCodeSetDeleteInput": func(r *ackmodel.CRD, sourceVarName string, targetVarName string, indentLevel int) string {
			return r.GoCodeSetInput(ackmodel.OpTypeDelete, sourceVarName, targetVarName, indentLevel)
		},
		"Empty": func(subject string) bool {
			return strings.TrimSpace(subject) == ""
		},
		"GoCodeRequiredFieldsMissingFromReadOneInput": func(r *ackmodel.CRD, koVarName string, indentLevel int) string {
			return r.GoCodeRequiredFieldsMissingFromShape(ackmodel.OpTypeGet, koVarName, indentLevel)
		},
		"GoCodeRequiredFieldsMissingFromGetAttributesInput": func(r *ackmodel.CRD, koVarName string, indentLevel int) string {
			return r.GoCodeRequiredFieldsMissingFromShape(ackmodel.OpTypeGetAttributes, koVarName, indentLevel)
		},
		"GoCodeRequiredFieldsMissingFromSetAttributesInput": func(r *ackmodel.CRD, koVarName string, indentLevel int) string {
			return r.GoCodeRequiredFieldsMissingFromShape(ackmodel.OpTypeSetAttributes, koVarName, indentLevel)
		},
	}
)

// Controller returns a pointer to a TemplateSet containing all the templates
// for generating ACK service controller implementations
func Controller(
	g *generate.Generator,
	templateBasePath string,
) (*templateset.TemplateSet, error) {
	crds, err := g.GetCRDs()
	if err != nil {
		return nil, err
	}

	ts := templateset.New(
		templateBasePath,
		controllerIncludePaths,
		controllerCopyPaths,
		controllerFuncMap,
	)

	metaVars := g.MetaVars()

	// First add all the CRD pkg/resource templates
	targets := []string{
		"descriptor.go.tpl",
		"identifiers.go.tpl",
		"manager.go.tpl",
		"manager_factory.go.tpl",
		"resource.go.tpl",
		"sdk.go.tpl",
	}
	for _, crd := range crds {
		for _, target := range targets {
			outPath := filepath.Join("pkg/resource", crd.Names.Snake, strings.TrimSuffix(target, ".tpl"))
			tplPath := filepath.Join("pkg/resource", target)
			crdVars := &templateCRDVars{
				metaVars,
				crd,
			}
			if err = ts.Add(outPath, tplPath, crdVars); err != nil {
				return nil, err
			}
		}
	}
	if err = ts.Add("pkg/resource/registry.go", "pkg/resource/registry.go.tpl", metaVars); err != nil {
		return nil, err
	}

	// Next add the template for the main.go file
	snakeCasedCRDNames := make([]string, 0)
	for _, crd := range crds {
		snakeCasedCRDNames = append(snakeCasedCRDNames, crd.Names.Snake)
	}
	cmdVars := &templateCmdVars{
		metaVars,
		snakeCasedCRDNames,
	}
	if err = ts.Add("cmd/controller/main.go", "cmd/controller/main.go.tpl", cmdVars); err != nil {
		return nil, err
	}

	// Finally, add the configuration YAML file templates
	for _, path := range controllerConfigTemplatePaths {
		outPath := strings.TrimSuffix(path, ".tpl")
		if err = ts.Add(outPath, path, metaVars); err != nil {
			return nil, err
		}
	}
	return ts, nil
}

// templateCmdVars contains template variables for the template that outputs Go
// code for a single top-level resource's API definition
type templateCmdVars struct {
	templateset.MetaVars
	SnakeCasedCRDNames []string
}
