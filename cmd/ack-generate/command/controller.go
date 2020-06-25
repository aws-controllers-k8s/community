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

package command

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	k8sversion "k8s.io/apimachinery/pkg/version"

	"github.com/aws/aws-service-operator-k8s/pkg/model"
	"github.com/aws/aws-service-operator-k8s/pkg/schema"
	dockertemplate "github.com/aws/aws-service-operator-k8s/pkg/template"
	cmdtemplate "github.com/aws/aws-service-operator-k8s/pkg/template/cmd"
	pkgtemplate "github.com/aws/aws-service-operator-k8s/pkg/template/pkg"
)

var (
	optControllerOutputPath string
	cmdControllerPath       string
	pkgResourcePath         string
	latestAPIVersion        string
)

var controllerCmd = &cobra.Command{
	Use:   "controller <service>",
	Short: "Generates Go files containing service controller implementation for a given service",
	RunE:  generateController,
}

func init() {
	controllerCmd.PersistentFlags().StringVarP(
		&optControllerOutputPath, "output", "o", "", "path to root directory to create generated files. Defaults to "+optServicesDir+"/$service",
	)
	rootCmd.AddCommand(controllerCmd)
}

// generateController generates the Go files for a service controller
func generateController(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("please specify the service alias for the AWS service API to generate")
	}
	svcAlias := args[0]
	if optControllerOutputPath == "" {
		optControllerOutputPath = filepath.Join(optServicesDir, svcAlias)
	}

	if !optDryRun {
		cmdControllerPath = filepath.Join(optControllerOutputPath, "cmd", "controller")
		if _, err := ensureDir(cmdControllerPath); err != nil {
			return err
		}
		pkgResourcePath = filepath.Join(optControllerOutputPath, "pkg", "resource")
		if _, err := ensureDir(pkgResourcePath); err != nil {
			return err
		}
	}

	sh, err := getSchemaHelper()
	if err != nil {
		return err
	}
	latestAPIVersion, err = getLatestAPIVersion()
	if err != nil {
		return err
	}

	if err = writeControllerMainGo(sh); err != nil {
		return err
	}
	if err = writeResourcePackage(sh); err != nil {
		return err
	}
	if err = writeDockerfile(sh); err != nil {
		return err
	}
	return nil
}

func writeControllerMainGo(sh *schema.Helper) error {
	var b bytes.Buffer
	vars := &cmdtemplate.ControllerMainTemplateVars{
		APIVersion:   latestAPIVersion,
		ServiceAlias: sh.GetServiceAlias(),
	}
	tpl, err := cmdtemplate.NewControllerMainTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= cmd/controller/main.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(cmdControllerPath, "main.go")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeResourcePackage(sh *schema.Helper) error {
	crds, err := sh.GetCRDs()
	if err != nil {
		return err
	}
	for _, crd := range crds {
		pkgCRDResourcePath := filepath.Join(pkgResourcePath, crd.Names.Snake)
		if !optDryRun {
			if _, err := ensureDir(pkgCRDResourcePath); err != nil {
				return err
			}
		}
		if err = writeCRDResourceGo(sh, crd); err != nil {
			return err
		}
		if err = writeCRDIdentifiersGo(sh, crd); err != nil {
			return err
		}
		if err = writeCRDDescriptorGo(sh, crd); err != nil {
			return err
		}
		if err = writeCRDManagerFactoryGo(sh, crd); err != nil {
			return err
		}
	}
	return writeResourcePackageRegistryGo(sh)
}

func writeResourcePackageRegistryGo(sh *schema.Helper) error {
	var b bytes.Buffer
	vars := &pkgtemplate.ResourceRegistryGoTemplateVars{
		APIVersion:   latestAPIVersion,
		ServiceAlias: sh.GetServiceAlias(),
	}
	tpl, err := pkgtemplate.NewResourceRegistryGoTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= pkg/resource/registry.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(pkgResourcePath, "registry.go")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeCRDResourceGo(sh *schema.Helper, crd *model.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDResourceGoTemplateVars{
		APIVersion:   latestAPIVersion,
		ServiceAlias: sh.GetServiceAlias(),
		CRD:          crd,
	}
	tpl, err := pkgtemplate.NewCRDResourceGoTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= pkg/resource/" + crd.Names.Snake + "/resource.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(pkgResourcePath, crd.Names.Snake, "resource.go")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeCRDIdentifiersGo(sh *schema.Helper, crd *model.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDIdentifiersGoTemplateVars{
		APIVersion:   latestAPIVersion,
		ServiceAlias: sh.GetServiceAlias(),
		CRD:          crd,
	}
	tpl, err := pkgtemplate.NewCRDIdentifiersGoTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= pkg/resource/" + crd.Names.Snake + "/identifiers.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(pkgResourcePath, crd.Names.Snake, "identifiers.go")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeCRDDescriptorGo(sh *schema.Helper, crd *model.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDDescriptorGoTemplateVars{
		APIVersion:   latestAPIVersion,
		APIGroup:     sh.GetAPIGroup(),
		ServiceAlias: sh.GetServiceAlias(),
		CRD:          crd,
	}
	tpl, err := pkgtemplate.NewCRDDescriptorGoTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= pkg/resource/" + crd.Names.Snake + "/descriptor.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(pkgResourcePath, crd.Names.Snake, "descriptor.go")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeCRDManagerFactoryGo(sh *schema.Helper, crd *model.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDManagerFactoryGoTemplateVars{
		APIVersion:   latestAPIVersion,
		APIGroup:     sh.GetAPIGroup(),
		ServiceAlias: sh.GetServiceAlias(),
		CRD:          crd,
	}
	tpl, err := pkgtemplate.NewCRDManagerFactoryGoTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= pkg/resource/" + crd.Names.Snake + "/manager_factory.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(pkgResourcePath, crd.Names.Snake, "manager_factory.go")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}


func writeDockerfile(sh *schema.Helper) error {
	var b bytes.Buffer
	vars := &dockertemplate.DockerTemplateVars{
		ServiceAlias: sh.GetServiceAlias(),
	}
	tpl, err := dockertemplate.NewDockerfileTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= Dockerfile ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(optControllerOutputPath, "Dockerfile")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

// getLatestAPIVersion looks in a target output directory to determine what the
// latest Kubernetes API version for CRDs exposed by the generated service
// controller.
func getLatestAPIVersion() (string, error) {
	apisPath := filepath.Join(optControllerOutputPath, "apis")
	versions := []string{}
	subdirs, err := ioutil.ReadDir(apisPath)
	if err != nil {
		return "", err
	}

	for _, subdir := range subdirs {
		versions = append(versions, subdir.Name())
	}
	sort.Slice(versions, func(i, j int) bool {
		return k8sversion.CompareKubeAwareVersionStrings(versions[i], versions[j]) < 0
	})
	return versions[len(versions)-1], nil
}
