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
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	k8sversion "k8s.io/apimachinery/pkg/version"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	ackgenerate "github.com/aws/aws-controllers-k8s/pkg/generate/ack"
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

	if err := ensureSDKRepo(optCacheDir); err != nil {
		return err
	}
	sdkHelper := ackmodel.NewSDKHelper(sdkDir)
	sdkAPI, err := sdkHelper.API(svcAlias)
	if err != nil {
		newSvcAlias, err := FallBackFindServiceID(sdkDir, svcAlias)
		if err != nil {
			return err
		}
		sdkAPI, err = sdkHelper.API(newSvcAlias) // retry with serviceID
		if err != nil {
			return fmt.Errorf("service %s not found", svcAlias)
		}
	}
	latestAPIVersion, err = getLatestAPIVersion()
	if err != nil {
		return err
	}
	g, err := generate.New(
		sdkAPI, latestAPIVersion, optGeneratorConfigPath, ackgenerate.DefaultConfig,
	)
	if err != nil {
		return err
	}
	ts, err := ackgenerate.Controller(g, optTemplatesDir)
	if err != nil {
		return err
	}

	if err = ts.Execute(); err != nil {
		return err
	}

	for path, contents := range ts.Executed() {
		if optDryRun {
			fmt.Printf("============================= %s ======================================\n", path)
			fmt.Println(strings.TrimSpace(contents.String()))
			continue
		}
		outPath := filepath.Join(optControllerOutputPath, path)
		outDir := filepath.Dir(outPath)
		if _, err := ensureDir(outDir); err != nil {
			return err
		}
		if err = ioutil.WriteFile(outPath, contents.Bytes(), 0666); err != nil {
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

// FallBackFindServiceID reads through aws-sdk-go/models/apis/*/*/api-2.json
// Returns ServiceID (as newSuppliedAlias) if supplied service Alias matches with serviceID in api-2.json
// If not a match, return the supllied alias.
func FallBackFindServiceID(sdkDir, svcAlias string) (string, error) {
	basePath := filepath.Join(sdkDir, "models", "apis")
	var files []string
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return svcAlias, err
	}
	for _, file := range files {
		if strings.Contains(file, "api-2.json") {
			f, err := os.Open(file)
			if err != nil {
				return svcAlias, err
			}
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				if strings.Contains(scanner.Text(), "serviceId") {
					getServiceID := strings.Split(scanner.Text(), ":")
					re := regexp.MustCompile(`[," \t]`)
					svcID := strings.ToLower(re.ReplaceAllString(getServiceID[1], ``))
					if svcAlias == svcID {
						getNewSvcAlias := strings.Split(file, string(os.PathSeparator))
						return getNewSvcAlias[len(getNewSvcAlias)-3], nil
					}
				}
			}
		}
	}
	return svcAlias, nil
}
