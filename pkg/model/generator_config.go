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
	"io/ioutil"

	"github.com/ghodss/yaml"
)

// GeneratorConfig represents instructions to the ACK code generator for a
// particular AWS service API
type GeneratorConfig struct {
	// Resources contains generator instructions for individual CRDs within an
	// API
	Resources map[string]ResourceGeneratorConfig `json:"resources"`
	// CRDs to ignore. ACK generator would skip these resources.
	Ignore IgnoreSpec `json:"ignore"`
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
}

// ResourceGeneratorConfig represents instructions to the ACK code generator
// for a particular CRD/resource on an AWS service API
type ResourceGeneratorConfig struct {
	// UnpackAttributeMapConfig contains instructions for converting a raw
	// `map[string]*string` into real fields on a CRD's Spec or Status object
	UnpackAttributesMapConfig *UnpackAttributesMapConfig `json:"unpack_attributes_map,omitempty"`
	// ExceptionsConfig identifies the exception codes for the resource. Some
	// API model files don't contain the ErrorInfo struct that contains the
	// HTTPStatusCode attribute that we usually look for to identify 404 Not
	// Found and other common error types for primary resources, and thus we
	// need these instructions.
	Exceptions *ExceptionsConfig `json:"exceptions,omitempty"`
}

// UnpackAttributesMapConfig informs the code generator that the API follows a
// pattern or using an "Attributes" `map[string]*string` that contains real,
// schema'd fields of the primary resource, and that those fields should be
// "unpacked" from the raw map and into CRD's Spec and Status struct fields.
//
// AWS Simple Notification Service (SNS) and AWS Simple Queue Service (SQS) are
// examples of APIs that use this pattern. For instance, the SNS CreateTopic
// API accepts a parameter called "Attributes" that can contain one of four
// keys:
//
// * DeliveryPolicy – The policy that defines how Amazon SNS retries failed
//   deliveries to HTTP/S endpoints.
// * DisplayName – The display name to use for a topic with SMS subscriptions
// * Policy – The policy that defines who can access your topic.
// * KmsMasterKeyId - The ID of an AWS-managed customer master key (CMK) for
//   Amazon SNS or a custom CMK.
//
// The `CreateTopic` API call **returns** only a single field: the TopicARN.
// But there is a separate `GetTopicAttributes` call that needs to be made that
// returns the above attributes (that are ReadWrite) along with a set of
// key/values that are ReadOnly:
//
// * Owner – The AWS account ID of the topic's owner.
// * SubscriptionsConfirmed – The number of confirmed subscriptions for the
//   topic.
// * SubscriptionsDeleted – The number of deleted subscriptions for the topic.
// * SubscriptionsPending – The number of subscriptions pending confirmation
//   for the topic.
// * TopicArn – The topic's ARN.
// * EffectiveDeliveryPolicy – The JSON serialization of the effective delivery
//   policy, taking system defaults into account.
//
// This structure instructs the code generator about the above real, schema'd
// fields that are masquerading as raw key/value pairs.
type UnpackAttributesMapConfig struct {
	// Fields contains a map, keyed by the original Attribute Key, of
	// FieldGeneratorConfig instructions for Attributes that should be
	// considered actual CRD fields.
	//
	// Some fields are ReadWrite -- i.e. the Kubernetes user has the ability to
	// set/update these fields on their CR -- and therefore go in the Spec
	// struct of the CR
	//
	// Other fields are ReadeOnly -- i.e. the Kubernetes user cannot update the
	// value of these fields.
	//
	// Note that any Attribute keys *not* listed here will be **excluded** from
	// the representation of the CR's Status struct. If there is an Attribute
	// -- e.g. an SNS Topic's `SubscriptionsDeleted` attribute -- that has
	// information that is constantly changing and does not represent
	// information to the ACK service controller that is useful for determining
	// observed versus desired state -- then do NOT list that attribute here.
	Fields map[string]FieldGeneratorConfig `json:"fields"`
}

// FieldGeneratorConfig contains instructions to the code generator about how
// to interpret the value of an Attribute and how to map it to a CRD's Spec or
// Status field
type FieldGeneratorConfig struct {
	// IsReadOnly indicates the field's value can not be set by a Kubernetes
	// user; in other words, the field should go in the CR's Status struct
	IsReadOnly bool `json:"is_read_only"`
	// ContainsOwnerAccountID indicates the field contains the AWS Account ID
	// that owns the resource. This is a special field that we direct to
	// storage in the common `Status.ACKResourceMetadata.OwnerAccountID` field.
	ContainsOwnerAccountID bool `json:"contains_owner_account_id"`
}

// ExceptionsConfig contains instructions to the code generator about how to
// handle the exceptions for the operations on a resource. These instructions
// are necessary for those APIs where the API models do not contain any
// information about the HTTP status codes a particular exception has (or, like
// the EC2 API, where the API model has no information at all about error
// responses for any operation)
type ExceptionsConfig struct {
	// Codes is a map of HTTP status code to the name of the Exception shape
	// that corresponds to that HTTP status code for this resource
	Codes map[int]string `json:"codes"`
}

// NewGeneratorConfig returns a new GeneratorConfig object given a supplied
// path to a config file
func NewGeneratorConfig(
	configPath string,
) (*GeneratorConfig, error) {
	gc := GeneratorConfig{}
	contents, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(contents, &gc); err != nil {
		return nil, err
	}
	return &gc, nil
}
