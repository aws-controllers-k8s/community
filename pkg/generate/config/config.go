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

package config

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
)

// Config represents instructions to the ACK code generator for a particular
// AWS service API
type Config struct {
	// Resources contains generator instructions for individual CRDs within an
	// API
	Resources map[string]ResourceConfig `json:"resources"`
	// CRDs to ignore. ACK generator would skip these resources.
	Ignore IgnoreSpec `json:"ignore"`
	// Contains generator instructions for individual API operations.
	Operations map[string]OperationConfig `json:"operations"`
	// PrefixConfig contains the prefixes to access certain fields in the generated
	// Go code.
	PrefixConfig PrefixConfig `json:"prefix_config,omitempty"`
	// IncludeACKMetadata lets you specify whether ACK Metadata should be included
	// in the status. Default is true.
	IncludeACKMetadata bool `json:"include_ack_metadata,omitempty"`
	// SetManyOutputNotFoundErrReturn is the return statement when generated
	// SetManyOutput function fails with NotFound error.
	// Default is "return nil, ackerr.NotFound"
	SetManyOutputNotFoundErrReturn string `json:"set_many_output_notfound_err_return,omitempty"`
}

// IgnoreSpec represents instructions to the ACK code generator to
// ignore operations, resources on an AWS service API
type IgnoreSpec struct {
	// Set of operation IDs/names that should be ignored by the
	// generator when constructing SDK linkage
	Operations []string `json:"operations"`
	// Set of resource names that should be ignored by the
	// generator
	ResourceNames []string `json:"resource_names"`
	// Set of shapes to ignore when constructing API type definitions and
	// associated SDK code for structs that have these shapes as members
	ShapeNames []string `json:"shape_names"`
	// Set of field paths to ignore. The name here should be the original name of
	// the field as it appears in AWS SDK objects. You can refer to a field by
	// giving its "<shape_name>.<field_name>". For example, "CreateApiInput.Name".
	FieldPaths []string `json:"field_paths"`
}

type PrefixConfig struct {
	// SpecField stores the string prefix to use for information that will be
	// sent to AWS. Defaults to `.Spec`
	SpecField string `json:"spec_field,omitempty"`
	// StatusField stores the string prefix to use for information fetched from
	// AWS. Defaults to `.Status`
	StatusField string `json:"status_field,omitempty"`
}

// New returns a new Config object given a supplied
// path to a config file
func New(
	configPath string,
	defaultConfig Config,
) (Config, error) {
	if configPath == "" {
		return defaultConfig, nil
	}
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}
	gc := defaultConfig
	if err = yaml.Unmarshal(content, &gc); err != nil {
		return Config{}, err
	}
	return gc, nil
}
