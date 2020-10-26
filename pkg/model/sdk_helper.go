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

package model

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	ackgenconfig "github.com/aws/aws-controllers-k8s/pkg/generate/config"
	"github.com/aws/aws-controllers-k8s/pkg/names"
	"github.com/aws/aws-controllers-k8s/pkg/util"
	awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"
)

var (
	ErrInvalidVersionDirectory = errors.New(
		"expected to find only directories in api model directory but found non-directory",
	)
	ErrNoValidVersionDirectory = errors.New(
		"no valid version directories found",
	)
	ErrServiceNotFound = errors.New(
		"no such service",
	)
)

// SDKHelper is a helper struct that helps work with the aws-sdk-go models and
// API model loader
type SDKHelper struct {
	basePath string
	loader   *awssdkmodel.Loader
	// Default is "services.k8s.aws"
	APIGroupSuffix string
}

// NewSDKHelper returns a new SDKHelper object
func NewSDKHelper(basePath string) *SDKHelper {
	return &SDKHelper{
		basePath: basePath,
		loader: &awssdkmodel.Loader{
			BaseImport:            basePath,
			IgnoreUnsupportedAPIs: true,
		},
	}
}

// API returns the aws-sdk-go API model for a supplied service alias
func (h *SDKHelper) API(serviceAlias string) (*SDKAPI, error) {
	modelPath, _, err := h.ModelAndDocsPath(serviceAlias)
	if err != nil {
		return nil, err
	}
	apis, err := h.loader.Load([]string{modelPath})
	if err != nil {
		return nil, err
	}
	// apis is a map, keyed by the service alias, of pointers to aws-sdk-go
	// model API objects
	for _, api := range apis {
		// If we don't do this, we can end up with panic()'s like this:
		// panic: assignment to entry in nil map
		// when trying to execute Shape.GoType().
		//
		// Calling API.ServicePackageDoc() ends up resetting the API.imports
		// unexported map variable...
		_ = api.ServicePackageDoc()
		return &SDKAPI{api, nil, nil, h.APIGroupSuffix}, nil
	}
	return nil, ErrServiceNotFound
}

// ModelAndDocsPath returns two string paths to the supplied service alias'
// model and doc JSON files
func (h *SDKHelper) ModelAndDocsPath(
	serviceAlias string,
) (string, string, error) {
	apiVersion, err := h.APIVersion(serviceAlias)
	if err != nil {
		return "", "", err
	}
	versionPath := filepath.Join(
		h.basePath, "models", "apis", serviceAlias, apiVersion,
	)
	modelPath := filepath.Join(versionPath, "api-2.json")
	docsPath := filepath.Join(versionPath, "docs-2.json")
	return modelPath, docsPath, nil
}

// APIVersion returns the API version (e.h. "2012-10-03") for a service API
func (h *SDKHelper) APIVersion(serviceAlias string) (string, error) {
	apiPath := filepath.Join(h.basePath, "models", "apis", serviceAlias)
	versionDirs, err := ioutil.ReadDir(apiPath)
	if err != nil {
		return "", err
	}
	for _, f := range versionDirs {
		version := f.Name()
		fp := filepath.Join(apiPath, version)
		fi, err := os.Lstat(fp)
		if err != nil {
			return "", err
		}
		if !fi.IsDir() {
			return "", ErrInvalidVersionDirectory
		}
		// TODO(jaypipes): handle more than one version? doesn't seem like
		// there is ever more than one.
		return version, nil
	}
	return "", ErrNoValidVersionDirectory
}

// SDKAPI contains an API model for a single AWS service API
type SDKAPI struct {
	API *awssdkmodel.API
	// A map of operation type and resource name to
	// aws-sdk-go/private/model/api.Operation structs
	opMap *OperationMap
	// Map, keyed by original Shape GoTypeElem(), with the values being a
	// renamed type name (due to conflicting names)
	typeRenames map[string]string
	// Default is "services.k8s.aws"
	apiGroupSuffix string
}

// GetPayloads returns a slice of strings of Shape names representing input and
// output request/response payloads
func (a *SDKAPI) GetPayloads() []string {
	res := []string{}
	for _, op := range a.API.Operations {
		res = append(res, op.InputRef.ShapeName)
		res = append(res, op.OutputRef.ShapeName)
	}
	return res
}

// GetOperationMap returns a map, keyed by the operation type and operation
// ID/name, of aws-sdk-go private/model/api.Operation struct pointers
func (a *SDKAPI) GetOperationMap(cfg *ackgenconfig.Config) *OperationMap {
	if a.opMap != nil {
		return a.opMap
	}
	// create an index of Operations by operation types and resource name
	opMap := OperationMap{}
	for opID, op := range a.API.Operations {
		opType, resName := getOpTypeAndResourceName(opID, cfg)
		if _, found := opMap[opType]; !found {
			opMap[opType] = map[string]*awssdkmodel.Operation{}
		}
		opMap[opType][resName] = op
	}
	a.opMap = &opMap
	return &opMap
}

// CRDNames returns a slice of names structs for all top-level resources in the
// API
func (a *SDKAPI) CRDNames(cfg *ackgenconfig.Config) []names.Names {
	opMap := a.GetOperationMap(cfg)
	createOps := (*opMap)[OpTypeCreate]
	crdNames := []names.Names{}
	for crdName := range createOps {
		if cfg.IsIgnoredResource(crdName) {
			continue
		}
		crdNames = append(crdNames, names.New(crdName))
	}
	return crdNames
}

// GetTypeRenames returns a map of original type name to renamed name (some
// type definition names conflict with generated names)
func (a *SDKAPI) GetTypeRenames(cfg *ackgenconfig.Config) map[string]string {
	if a.typeRenames != nil {
		return a.typeRenames
	}

	trenames := map[string]string{}

	payloads := a.GetPayloads()

	for shapeName, shape := range a.API.Shapes {
		if util.InStrings(shapeName, payloads) {
			// Payloads are not type defs
			continue
		}
		if shape.Type != "structure" {
			continue
		}
		if shape.Exception {
			// Neither are exceptions
			continue
		}
		if cfg.IsIgnoredShape(shapeName) {
			continue
		}
		tdefNames := names.New(shapeName)
		if a.HasConflictingTypeName(shapeName, cfg) {
			tdefNames.Camel += ConflictingNameSuffix
			trenames[shapeName] = tdefNames.Camel
		}
	}
	a.typeRenames = trenames
	return trenames
}

// HasConflictingTypeName returns true if the supplied type name will conflict
// with any generated type in the service's API package
func (a *SDKAPI) HasConflictingTypeName(typeName string, cfg *ackgenconfig.Config) bool {
	// First grab the set of CRD struct names and the names of their Spec and
	// Status structs
	cleanTypeName := names.New(typeName).Camel
	crdNames := a.CRDNames(cfg)
	crdResourceNames := []string{}
	crdSpecNames := []string{}
	crdStatusNames := []string{}

	for _, crdName := range crdNames {
		cleanResourceName := crdName.Camel
		crdResourceNames = append(crdResourceNames, cleanResourceName)
		crdSpecNames = append(crdSpecNames, cleanResourceName+"Spec")
		crdStatusNames = append(crdStatusNames, cleanResourceName+"Status")
	}
	return util.InStrings(cleanTypeName, crdResourceNames) ||
		util.InStrings(cleanTypeName, crdSpecNames) ||
		util.InStrings(cleanTypeName, crdStatusNames)
}

// ServiceID returns the exact `metadata.serviceId` attribute for the AWS
// service APi's api-2.json file
func (a *SDKAPI) ServiceID() string {
	if a == nil || a.API == nil {
		return ""
	}
	return awssdkmodel.ServiceID(a.API)
}

// ServiceIDClean returns a lowercased, whitespace-stripped ServiceID
func (a *SDKAPI) ServiceIDClean() string {
	serviceID := strings.ToLower(a.ServiceID())
	return strings.Replace(serviceID, " ", "", -1)
}

func (a *SDKAPI) GetServiceFullName() string {
	if a == nil || a.API == nil {
		return ""
	}
	return a.API.Metadata.ServiceFullName
}

// APIGroup returns the normalized Kubernetes APIGroup for the AWS service API,
// e.g. "sns.services.k8s.aws"
func (a *SDKAPI) APIGroup() string {
	serviceID := a.ServiceIDClean()
	suffix := "services.k8s.aws"
	if a.apiGroupSuffix != "" {
		suffix = a.apiGroupSuffix
	}
	return fmt.Sprintf("%s.%s", serviceID, suffix)
}

// SDKAPIInterfaceTypeName returns the name of the aws-sdk-go primary API
// interface type name.
func (a *SDKAPI) SDKAPIInterfaceTypeName() string {
	if a == nil || a.API == nil {
		return ""
	}
	return a.API.StructName()
}

// Override the operation type and/or resource name if specified in config
func getOpTypeAndResourceName(opID string, cfg *ackgenconfig.Config) (OpType, string) {
	opType, resName := GetOpTypeAndResourceNameFromOpID(opID)

	if cfg != nil {
		if operationConfig, exists := cfg.Operations[opID]; exists {
			if operationConfig.OperationType != "" {
				opType = OpTypeFromString(operationConfig.OperationType)
			}

			if operationConfig.ResourceName != "" {
				resName = operationConfig.ResourceName
			}
		}
	}

	return opType, resName
}
