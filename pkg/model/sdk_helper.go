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
	"io/ioutil"
	"os"
	"path/filepath"

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
func (h *SDKHelper) API(serviceAlias string) (*awssdkmodel.API, error) {
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
		return api, nil
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

// APIVersion returns the API version (e.g. "2012-10-03") for a service API
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
