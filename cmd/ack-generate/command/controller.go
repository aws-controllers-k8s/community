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

	"github.com/aws/aws-controllers-k8s/pkg/model"
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
	sdkHelper := model.NewSDKHelper(sdkDir)
	sdkAPI, err := sdkHelper.API(svcAlias)
	if err != nil {
		return err
	}
	sh, err := model.NewHelper(sdkAPI, optGeneratorConfigPath)
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
	if err = writeConfigDirs(sh); err != nil {
		return err
	}
	return nil
}

func writeControllerMainGo(sh *model.Helper) error {
	var b bytes.Buffer
	crdsNames := sh.GetCRDNames()

	// convert CRD names into snake_case to use for package import
	snakeCasedCRDNames := make([]string, 0)
	for _, crdName := range crdsNames {
		snakeCasedCRDNames = append(snakeCasedCRDNames, crdName.Snake)
	}

	vars := &cmdtemplate.ControllerMainTemplateVars{
		APIVersion:         latestAPIVersion,
		ServiceAlias:       sh.GetCleanServiceAlias(),
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

func writeResourcePackage(sh *model.Helper) error {
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
		if err = writeCRDManagerGo(sh, crd); err != nil {
			return err
		}
		if err = writeCRDSDKGo(sh, crd); err != nil {
			return err
		}
	}
	return writeResourcePackageRegistryGo(sh)
}

func writeResourcePackageRegistryGo(sh *model.Helper) error {
	var b bytes.Buffer
	vars := &pkgtemplate.ResourceRegistryGoTemplateVars{
		APIVersion:   latestAPIVersion,
		ServiceAlias: sh.GetCleanServiceAlias(),
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

func writeCRDResourceGo(sh *model.Helper, crd *model.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDResourceGoTemplateVars{
		APIVersion:   latestAPIVersion,
		ServiceAlias: sh.GetCleanServiceAlias(),
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

func writeCRDIdentifiersGo(sh *model.Helper, crd *model.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDIdentifiersGoTemplateVars{
		APIVersion:   latestAPIVersion,
		ServiceAlias: sh.GetCleanServiceAlias(),
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

func writeCRDDescriptorGo(sh *model.Helper, crd *model.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDDescriptorGoTemplateVars{
		APIVersion:   latestAPIVersion,
		APIGroup:     sh.GetAPIGroup(),
		ServiceAlias: sh.GetCleanServiceAlias(),
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

func writeCRDManagerFactoryGo(sh *model.Helper, crd *model.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDManagerFactoryGoTemplateVars{
		APIVersion:   latestAPIVersion,
		APIGroup:     sh.GetAPIGroup(),
		ServiceAlias: sh.GetCleanServiceAlias(),
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

func writeCRDManagerGo(sh *model.Helper, crd *model.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDManagerGoTemplateVars{
		APIVersion:              latestAPIVersion,
		APIGroup:                sh.GetAPIGroup(),
		ServiceAlias:            sh.GetCleanServiceAlias(),
		SDKAPIInterfaceTypeName: sh.GetSDKAPIInterfaceTypeName(),
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

func writeCRDSDKGo(sh *model.Helper, crd *model.CRD) error {
	var b bytes.Buffer
	vars := &pkgtemplate.CRDSDKGoTemplateVars{
		APIVersion:              latestAPIVersion,
		APIGroup:                sh.GetAPIGroup(),
		ServiceAlias:            sh.GetCleanServiceAlias(),
		SDKAPIInterfaceTypeName: sh.GetSDKAPIInterfaceTypeName(),
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

func writeConfigDirs(sh *model.Helper) error {
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
	if err := writeConfigDefaultKustomizationYAML(sh); err != nil {
		return err
	}
	if err := writeConfigControllerKustomizationYAML(sh); err != nil {
		return err
	}
	if err := writeConfigControllerDeploymentYAML(sh); err != nil {
		return err
	}
	if err := writeConfigRBACKustomizationYAML(sh); err != nil {
		return err
	}
	if err := writeConfigRBACClusterRoleBindingYAML(sh); err != nil {
		return err
	}
	return nil
}

func writeConfigDefaultKustomizationYAML(sh *model.Helper) error {
	var b bytes.Buffer
	vars := &configdefaulttemplate.ConfigDefaultKustomizationYAMLTemplateVars{
		ServiceAlias: sh.GetCleanServiceAlias(),
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

func writeConfigControllerKustomizationYAML(sh *model.Helper) error {
	var b bytes.Buffer
	vars := &configcontrollertemplate.ConfigControllerKustomizationYAMLTemplateVars{
		ServiceAlias: sh.GetCleanServiceAlias(),
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

func writeConfigControllerDeploymentYAML(sh *model.Helper) error {
	var b bytes.Buffer
	vars := &configcontrollertemplate.ConfigControllerDeploymentYAMLTemplateVars{
		ServiceAlias: sh.GetCleanServiceAlias(),
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

func writeConfigRBACKustomizationYAML(sh *model.Helper) error {
	var b bytes.Buffer
	vars := &configrbactemplate.ConfigRBACKustomizationYAMLTemplateVars{
		ServiceAlias: sh.GetCleanServiceAlias(),
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

func writeConfigRBACClusterRoleBindingYAML(sh *model.Helper) error {
	var b bytes.Buffer
	vars := &configrbactemplate.ConfigRBACClusterRoleBindingYAMLTemplateVars{
		ServiceAlias: sh.GetCleanServiceAlias(),
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
