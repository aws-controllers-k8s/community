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
	"fmt"
	"github.com/aws/aws-controllers-k8s/pkg/generate"
	"github.com/aws/aws-controllers-k8s/pkg/generate/templateset"
	ackmodel "github.com/aws/aws-controllers-k8s/pkg/model"
)

// templateAPIVars contains template variables for templates that output Go
// code in the /services/$SERVICE/apis/$API_VERSION directory
type templateAPIVars struct {
	templateset.MetaVars
	EnumDefs []*ackmodel.EnumDef
	TypeDefs []*ackmodel.TypeDef
	Imports  map[string]string
}

// templateCRDVars contains template variables for the template that outputs Go
// code for a single top-level resource's API definition
type templateCRDVars struct {
	templateset.MetaVars
	CRD *ackmodel.CRD
}

// AddAPIFiles initializes templates for APIs and adds them to TemplateSet.
func AddAPIFiles(g *generate.Generator, ts *templateset.TemplateSet, meta templateset.MetaVars) error {
	enumDefs, err := g.GetEnumDefs()
	if err != nil {
		return err
	}
	typeDefs, typeImports, err := g.GetTypeDefs()
	if err != nil {
		return err
	}

	apiVars := &templateAPIVars{
		meta,
		enumDefs,
		typeDefs,
		typeImports,
	}
	for tmpl, target := range APIFilePairs {
		out := fmt.Sprintf(target, meta.ServiceIDClean, meta.APIVersion)
		if err = ts.Add(out, tmpl, apiVars); err != nil {
			return err
		}
	}
	return nil
}

// AddCRDFiles initializes templates for CRD-specific files and adds them to TemplateSet.
func AddCRDFiles(g *generate.Generator, ts *templateset.TemplateSet, meta templateset.MetaVars) error {
	crds, err := g.GetCRDs()
	if err != nil {
		return err
	}
	for _, crd := range crds {
		for tmpl, target := range CRDFilePairs {
			crdVars := &templateCRDVars{
				meta,
				crd,
			}
			out := fmt.Sprintf(target, meta.ServiceIDClean, meta.APIVersion, crd.Names.Snake)
			if err = ts.Add(out, tmpl, crdVars); err != nil {
				return err
			}
		}
	}
	return nil
}

// AddCRDFiles initializes templates for controller files and adds them to TemplateSet.
func AddControllerFiles(g *generate.Generator, ts *templateset.TemplateSet, meta templateset.MetaVars) error {
	crds, err := g.GetCRDs()
	if err != nil {
		return err
	}
	for _, crd := range crds {
		for tmpl, target := range ControllerFilePairs {
			out := fmt.Sprintf(target, meta.ServiceIDClean, crd.Names.Lower)
			crdVars := &templateCRDVars{
				meta,
				crd,
			}
			if err = ts.Add(out, tmpl, crdVars); err != nil {
				return err
			}
		}
	}
	return nil
}
