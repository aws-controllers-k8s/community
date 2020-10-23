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
	"io/ioutil"
	"path/filepath"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
)

type APIFileGeneratorChain []func(*generate.Generator, string) error

func (a APIFileGeneratorChain) Generate(g *generate.Generator, apiPath string) error {
	for _, f := range a {
		if err := f(g, apiPath); err != nil {
			return err
		}
	}
	return nil
}

type APIFileGeneratorFn func(*generate.Generator, string) error

func (a APIFileGeneratorFn) Generate(g *generate.Generator, apiPath string) error {
	return a(g, apiPath)
}

func GenerateCRDFiles(g *generate.Generator, apiPath string) error {
	crds, err := g.GetCRDs()
	if err != nil {
		return errors.Wrap(err, "cannot generate CRDs")
	}
	for _, crd := range crds {
		// TODO(muvaf): ACK hard-codes the suffix for the API group as "services.k8s.aws"
		content, err := g.GenerateCRDFile(crd.Names.Original)
		if err != nil {
			return errors.Wrap(err, "cannot generate crd file")
		}
		path := filepath.Join(apiPath, fmt.Sprintf("zz_%s.go", strcase.ToSnake(crd.Kind)))
		if err := ioutil.WriteFile(path, content.Bytes(), 0666); err != nil {
			return errors.Wrap(err, "cannot write crd file")
		}
	}
	return nil
}

func GenerateTypesFile(g *generate.Generator, apiPath string) error {
	typeDefs, _, err := g.GetTypeDefs()
	if err != nil {
		return errors.Wrap(err, "cannot generate type definitions")
	}
	if len(typeDefs) == 0 {
		return nil
	}
	content, err := g.GenerateAPIFile("types")
	if err != nil {
		return errors.Wrap(err, "cannot generate types file")
	}
	path := filepath.Join(apiPath, "zz_types.go")
	return errors.Wrap(ioutil.WriteFile(path, content.Bytes(), 0666), "cannot write types file")
}

func GenerateEnumsFile(g *generate.Generator, apiPath string) error {
	enumDefs, err := g.GetEnumDefs()
	if err != nil {
		return errors.Wrap(err, "cannot generate enum definitions")
	}
	if len(enumDefs) == 0 {
		return nil
	}
	content, err := g.GenerateAPIFile("enums")
	if err != nil {
		return errors.Wrap(err, "cannot generate enums file")
	}
	path := filepath.Join(apiPath, "zz_enums.go")
	return errors.Wrap(ioutil.WriteFile(path, content.Bytes(), 0666), "cannot write enums file")
}

func GenerateGroupVersionInfoFile(g *generate.Generator, apiPath string) error {
	gvi, err := g.GenerateAPIFile("groupversion_info")
	if err != nil {
		return errors.Wrap(err, "cannot generate groupversion_info file")
	}
	path := filepath.Join(apiPath, "zz_groupversion_info.go")
	return errors.Wrap(ioutil.WriteFile(path, gvi.Bytes(), 0666), "cannot write groupversion_info file")
}

func GenerateDocFile(g *generate.Generator, apiPath string) error {
	gvi, err := g.GenerateAPIFile("doc")
	if err != nil {
		return errors.Wrap(err, "cannot generate doc file")
	}
	path := filepath.Join(apiPath, "zz_doc.go")
	return errors.Wrap(ioutil.WriteFile(path, gvi.Bytes(), 0666), "cannot write doc file")
}
