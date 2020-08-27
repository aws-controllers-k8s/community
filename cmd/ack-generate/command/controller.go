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

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	ackmodel "github.com/aws/aws-controllers-k8s/pkg/model"
	cmdtemplate "github.com/aws/aws-controllers-k8s/pkg/template/cmd"
	configcontrollertemplate "github.com/aws/aws-controllers-k8s/pkg/template/config/controller"
	configdefaulttemplate "github.com/aws/aws-controllers-k8s/pkg/template/config/default"
	configrbactemplate "github.com/aws/aws-controllers-k8s/pkg/template/config/rbac"
	pkgtemplate "github.com/aws/aws-controllers-k8s/pkg/template/pkg"
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
	svcAlias := strings.ToLower(args[0])
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

	if err := ensureSDKRepo(optCacheDir); err != nil {
		return err
	}
	sdkHelper := ackmodel.NewSDKHelper(sdkDir)
	sdkAPI, err := sdkHelper.API(svcAlias)
	if err != nil {
		return err
	}
	g, err := generate.New(sdkAPI, optGeneratorConfigPath)
	if err != nil {
		return err
	}
	latestAPIVersion, err = getLatestAPIVersion()
	if err != nil {
		return err
	}

	crds, err := g.GetCRDs()
	if err != nil {
		return err
	}

	if err = writeControllerMainGo(g, crds); err != nil {
		return err
	}
	if err = writeResourcePackage(g, crds); err != nil {
		return err
	}
	if err = writeConfigDirs(g); err != nil {
		return err
	}
	return nil
}

func writeControllerMainGo(g *generate.Generator, crds []*ackmodel.CRD) error {
	var b bytes.Buffer

	// convert CRD names into snake_case to use for package import
	snakeCasedCRDNames := make([]string, 0)
	for _, crd := range crds {
		snakeCasedCRDNames = append(snakeCasedCRDNames, crd.Names.Snake)
	}

	vars := &cmdtemplate.ControllerMainTemplateVars{
		APIVersion:         latestAPIVersion,
		ServiceAlias:       g.SDKAPI.GetCleanServiceAlias(),
		SnakeCasedCRDNames: snakeCasedCRDNames,
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

func writeResourcePackage(g *generate.Generator, crds []*ackmodel.CRD) error {
	for _, crd := range crds {
		pkgCRDResourcePath := filepath.Join(pkgResourcePath, crd.Names.Snake)
		if !optDryRun {
			if _, err := ensureDir(pkgCRDResourcePath); err != nil {
				return err
			}
		}
		if err := writeCRDResourceGo(g, crd); err != nil {
			return err
		}
		if err := writeCRDIdentifiersGo(g, crd); err != nil {
			return err
		}
		if err := writeCRDDescriptorGo(g, crd); err != nil {
			return err
		}
		if err := writeCRDManagerFactoryGo(g, crd); err != nil {
			return err
		}
		if err := writeCRDManagerGo(g, crd); err != nil {
			return err
		}
		if err := writeCRDSDKGo(g, crd); err != nil {
			return err
		}
	}
	return writeResourcePackageRegistryGo(g)
}

func writeResourcePackageRegistryGo(g *generate.Generator) error {
	var b bytes.Buffer
	vars := &pkgtemplate.ResourceRegistryGoTemplateVars{
		APIVersion:   latestAPIVersion,
		ServiceAlias: g.SDKAPI.GetCleanServiceAlias(),
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

func writeCRDResourceGo(g *generate.Generator, crd *ackmodel.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDResourceGoTemplateVars{
		APIVersion:   latestAPIVersion,
		ServiceAlias: g.SDKAPI.GetCleanServiceAlias(),
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

func writeCRDIdentifiersGo(g *generate.Generator, crd *ackmodel.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDIdentifiersGoTemplateVars{
		APIVersion:   latestAPIVersion,
		ServiceAlias: g.SDKAPI.GetCleanServiceAlias(),
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

func writeCRDDescriptorGo(g *generate.Generator, crd *ackmodel.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDDescriptorGoTemplateVars{
		APIVersion:   latestAPIVersion,
		APIGroup:     g.SDKAPI.GetAPIGroup(),
		ServiceAlias: g.SDKAPI.GetCleanServiceAlias(),
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

func writeCRDManagerFactoryGo(g *generate.Generator, crd *ackmodel.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDManagerFactoryGoTemplateVars{
		APIVersion:   latestAPIVersion,
		APIGroup:     g.SDKAPI.GetAPIGroup(),
		ServiceAlias: g.SDKAPI.GetCleanServiceAlias(),
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

func writeCRDManagerGo(g *generate.Generator, crd *ackmodel.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDManagerGoTemplateVars{
		APIVersion:              latestAPIVersion,
		APIGroup:                g.SDKAPI.GetAPIGroup(),
		ServiceAlias:            g.SDKAPI.GetCleanServiceAlias(),
		SDKAPIInterfaceTypeName: g.SDKAPI.GetSDKAPIInterfaceTypeName(),
		CRD:                     crd,
	}
	tpl, err := pkgtemplate.NewCRDManagerGoTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= pkg/resource/" + crd.Names.Snake + "/manager.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(pkgResourcePath, crd.Names.Snake, "manager.go")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeCRDSDKGo(g *generate.Generator, crd *ackmodel.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDSDKGoTemplateVars{
		APIVersion:              latestAPIVersion,
		APIGroup:                g.SDKAPI.GetAPIGroup(),
		ServiceAlias:            g.SDKAPI.GetCleanServiceAlias(),
		SDKAPIInterfaceTypeName: g.SDKAPI.GetSDKAPIInterfaceTypeName(),
		CRD:                     crd,
	}
	tpl, err := pkgtemplate.NewCRDSDKGoTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= pkg/resource/" + crd.Names.Snake + "/sdk.go ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(pkgResourcePath, crd.Names.Snake, "sdk.go")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeConfigDirs(g *generate.Generator) error {
	configDefaultPath := filepath.Join(optControllerOutputPath, "config", "default")
	configControllerPath := filepath.Join(optControllerOutputPath, "config", "controller")
	configRBACPath := filepath.Join(optControllerOutputPath, "config", "rbac")
	if !optDryRun {
		if _, err := ensureDir(configDefaultPath); err != nil {
			return err
		}
		if _, err := ensureDir(configControllerPath); err != nil {
			return err
		}
		if _, err := ensureDir(configRBACPath); err != nil {
			return err
		}
	}
	if err := writeConfigDefaultKustomizationYAML(g); err != nil {
		return err
	}
	if err := writeConfigControllerKustomizationYAML(g); err != nil {
		return err
	}
	if err := writeConfigControllerDeploymentYAML(g); err != nil {
		return err
	}
	if err := writeConfigRBACKustomizationYAML(g); err != nil {
		return err
	}
	if err := writeConfigRBACClusterRoleBindingYAML(g); err != nil {
		return err
	}
	return nil
}

func writeConfigDefaultKustomizationYAML(g *generate.Generator) error {
	var b bytes.Buffer
	vars := &configdefaulttemplate.ConfigDefaultKustomizationYAMLTemplateVars{
		ServiceAlias: g.SDKAPI.GetCleanServiceAlias(),
	}
	tpl, err := configdefaulttemplate.NewConfigDefaultKustomizationYAMLTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= config/default/kustomization.yaml ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(optControllerOutputPath, "config", "default", "kustomization.yaml")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeConfigControllerKustomizationYAML(g *generate.Generator) error {
	var b bytes.Buffer
	vars := &configcontrollertemplate.ConfigControllerKustomizationYAMLTemplateVars{
		ServiceAlias: g.SDKAPI.GetCleanServiceAlias(),
	}
	tpl, err := configcontrollertemplate.NewConfigControllerKustomizationYAMLTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= config/controller/kustomization.yaml ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(optControllerOutputPath, "config", "controller", "kustomization.yaml")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeConfigControllerDeploymentYAML(g *generate.Generator) error {
	var b bytes.Buffer
	vars := &configcontrollertemplate.ConfigControllerDeploymentYAMLTemplateVars{
		ServiceAlias: g.SDKAPI.GetCleanServiceAlias(),
	}
	tpl, err := configcontrollertemplate.NewConfigControllerDeploymentYAMLTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= config/controller/deployment.yaml ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(optControllerOutputPath, "config", "controller", "deployment.yaml")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeConfigRBACKustomizationYAML(g *generate.Generator) error {
	var b bytes.Buffer
	vars := &configrbactemplate.ConfigRBACKustomizationYAMLTemplateVars{
		ServiceAlias: g.SDKAPI.GetCleanServiceAlias(),
	}
	tpl, err := configrbactemplate.NewConfigRBACKustomizationYAMLTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= config/rbac/kustomization.yaml ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(optControllerOutputPath, "config", "rbac", "kustomization.yaml")
	return ioutil.WriteFile(path, b.Bytes(), 0666)
}

func writeConfigRBACClusterRoleBindingYAML(g *generate.Generator) error {
	var b bytes.Buffer
	vars := &configrbactemplate.ConfigRBACClusterRoleBindingYAMLTemplateVars{
		ServiceAlias: g.SDKAPI.GetCleanServiceAlias(),
	}
	tpl, err := configrbactemplate.NewConfigRBACClusterRoleBindingYAMLTemplate(optTemplatesDir)
	if err != nil {
		return err
	}
	if err := tpl.Execute(&b, vars); err != nil {
		return err
	}
	if optDryRun {
		fmt.Println("============================= config/rbac/cluster-role-binding.yaml ======================================")
		fmt.Println(strings.TrimSpace(b.String()))
		return nil
	}
	path := filepath.Join(optControllerOutputPath, "config", "rbac", "cluster-role-binding.yaml")
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
