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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	k8sversion "k8s.io/apimachinery/pkg/version"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	ackmodel "github.com/aws/aws-controllers-k8s/pkg/model"
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
	latestAPIVersion, err = getLatestAPIVersion()
	if err != nil {
		return err
	}
	g, err := generate.New(
		sdkAPI, latestAPIVersion, optGeneratorConfigPath, optTemplatesDir,
	)
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
	b, err := g.GenerateCmdControllerMainFile()
	if err != nil {
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
	targets := []string{
		"descriptor",
		"identifiers",
		"manager",
		"manager_factory",
		"resource",
		"sdk",
	}
	for _, crd := range crds {
		pkgCRDResourcePath := filepath.Join(pkgResourcePath, crd.Names.Snake)
		if !optDryRun {
			if _, err := ensureDir(pkgCRDResourcePath); err != nil {
				return err
			}
		}
		for _, target := range targets {
			b, err := g.GenerateCRDResourcePackageFile(crd.Names.Original, target)
			if err != nil {
				return err
			}
			if optDryRun {
				fmt.Println("============================= pkg/resource/" + crd.Names.Snake + "/" + target + ".go ======================================")
				fmt.Println(strings.TrimSpace(b.String()))
				return nil
			}
			path := filepath.Join(pkgResourcePath, crd.Names.Snake, target+".go")
			if err := ioutil.WriteFile(path, b.Bytes(), 0666); err != nil {
				return err
			}
		}
	}
	return writeResourcePackageRegistryGo(g)
}

func writeResourcePackageRegistryGo(g *generate.Generator) error {
	b, err := g.GenerateResourceRegistryFile()
	if err != nil {
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
	targets := []string{
		"controller/deployment",
		"controller/kustomization",
		"default/kustomization",
		"rbac/cluster-role-binding",
		"rbac/kustomization",
	}
	for _, target := range targets {
		b, err := g.GenerateConfigYAMLFile(target)
		if err != nil {
			return err
		}
		if optDryRun {
			fmt.Println("============================= config/" + target + ".yaml ======================================")
			fmt.Println(strings.TrimSpace(b.String()))
			return nil
		}
		path := filepath.Join(optControllerOutputPath, "config", target+".yaml")
		if err := ioutil.WriteFile(path, b.Bytes(), 0666); err != nil {
			return err
		}
	}
	return nil
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
