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
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
)

// TODO(muvaf): Template file names are hard-coded but we are able to write output
// to any file we want. So, we have to reuse the existing template files for any
// file we'd like to be generated, though we're free to change content of any
// template file.

type ControllerGeneratorChain []func(*generate.Generator, string) error

func (a ControllerGeneratorChain) Generate(g *generate.Generator, controllerPath string) error {
	for _, f := range a {
		if err := f(g, controllerPath); err != nil {
			return err
		}
	}
	return nil
}

type ControllerGeneratorFn func(*generate.Generator, string) error

func (a ControllerGeneratorFn) Generate(g *generate.Generator, controllerPath string) error {
	return a(g, controllerPath)
}

func GenerateController(g *generate.Generator, controllerPath string) error {
	crds, err := g.GetCRDs()
	if err != nil {
		return err
	}
	for _, crd := range crds {
		// TODO(muvaf): "manager" is hard-coded in ACK.
		b, err := g.GenerateResourcePackageFile(crd.Names.Original, "manager")
		if err != nil {
			return err
		}
		dir := filepath.Join(controllerPath, crd.Names.Lower)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return errors.Wrap(err, "cannot create controller dir")
		}
		path := filepath.Join(dir, "zz_controller.go")
		if err := ioutil.WriteFile(path, b.Bytes(), 0666); err != nil {
			return err
		}
	}
	return nil
}

func GenerateConversions(g *generate.Generator, controllerPath string) error {
	crds, err := g.GetCRDs()
	if err != nil {
		return err
	}
	for _, crd := range crds {
		// TODO(muvaf): "sdk" is hard-coded in ACK.
		b, err := g.GenerateResourcePackageFile(crd.Names.Original, "sdk")
		if err != nil {
			return err
		}
		dir := filepath.Join(controllerPath, crd.Names.Lower)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return errors.Wrap(err, "cannot create controller dir")
		}
		path := filepath.Join(dir, "zz_conversions.go")
		if err := ioutil.WriteFile(path, b.Bytes(), 0666); err != nil {
			return err
		}
	}
	return nil
}

func GenerateHooksBoilerplate(g *generate.Generator, controllerPath string) error {
	crds, err := g.GetCRDs()
	if err != nil {
		return err
	}
	for _, crd := range crds {
		// TODO(muvaf): "resource" is hard-coded in ACK.
		b, err := g.GenerateResourcePackageFile(crd.Names.Original, "resource")
		if err != nil {
			return err
		}
		dir := filepath.Join(controllerPath, crd.Names.Lower)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return errors.Wrap(err, "cannot create controller dir")
		}
		path := filepath.Join(dir, fmt.Sprintf("hooks.go"))
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			// NOTE(muvaf): Hook files are generated once and can be edited by
			// the user later on.
			continue
		}
		if err := ioutil.WriteFile(path, b.Bytes(), 0666); err != nil {
			return err
		}
	}
	return nil
}
