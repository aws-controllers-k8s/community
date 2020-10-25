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

	awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"
	"github.com/ghodss/yaml"

	"github.com/aws/aws-controllers-k8s/pkg/util"
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
}

// OperationConfig represents instructions to the ACK code generator to
// specify the overriding values for API operation parameters and its custom implementation.
type OperationConfig struct {
	CustomImplementation string            `json:"custom_implementation,omitempty"`
	OverrideValues       map[string]string `json:"override_values"`
	// SetOutputCustomMethodName provides the name of the custom method on the
	// `resourceManager` struct that will set fields on a `resource` struct
	// depending on the output of the operation.
	SetOutputCustomMethodName string `json:"set_output_custom_method_name,omitempty"`
	// Override for resource name in case of heuristic failure
	// An example of this is correcting stutter when the resource logic doesn't properly determine the resource name
	ResourceName string `json:"resource_name"`
	// Override for operation type in case of heuristic failure
	// An example of this is `Put...` or `Register...` API operations not being correctly classified as `Create` op type
	OperationType string `json:"operation_type"`
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
}

// ResourceConfig represents instructions to the ACK code generator
// for a particular CRD/resource on an AWS service API
type ResourceConfig struct {
	// NameField is the name of the Member of the Create Input shape that
	// represents the name/string identifier field for the resource. If this
	// isn't set, then the generator will look for a field called "Name" or
	// "{Resource}Name" or "{Resource}Id" because, well, because we can never
	// have nice things.
	NameField *string `json:"name_field,omitempty"`
	// UnpackAttributeMapConfig contains instructions for converting a raw
	// `map[string]*string` into real fields on a CRD's Spec or Status object
	UnpackAttributesMapConfig *UnpackAttributesMapConfig `json:"unpack_attributes_map,omitempty"`
	// Exceptions identifies the exception codes for the resource. Some API
	// model files don't contain the ErrorInfo struct that contains the
	// HTTPStatusCode attribute that we usually look for to identify 404 Not
	// Found and other common error types for primary resources, and thus we
	// need these instructions.
	Exceptions *ExceptionsConfig `json:"exceptions,omitempty"`

	// Renames identifies fields in Operations that should be renamed.
	Renames *RenamesConfig `json:"renames,omitempty"`
	// ListOperation contains instructions for the code generator to generate
	// Go code that filters the results of a List operation looking for a
	// singular object. Certain AWS services (e.g. S3's ListBuckets API) have
	// absolutely no way to pass a filter to the operation. Instead, the List
	// operation always returns ALL objects of that type.
	//
	// The ListOperationConfig object enables us to inject some custom code to
	// filter the results of these List operations from within the generated
	// code in sdk.go's sdkFind().
	ListOperation *ListOperationConfig `json:"list_operation,omitempty"`
	// UpdateOperation contains instructions for the code generator to generate
	// Go code for the update operation for the resource. For some APIs, the
	// way that a resource's attributes are updated after creation is, well,
	// very odd. Some APIs have separate API calls for each attribute or set of
	// related attributes of the resource. For example, the ECR API has
	// separate API calls for PutImageScanningConfiguration,
	// PutImageTagMutability, PutLifecyclePolicy and SetRepositoryPolicy. FOr
	// these APIs, we basically need to revert to custom code because there's
	// very little consistency to the APIs that we can use to instruct the code
	// generator :(
	UpdateOperation *UpdateOperationConfig `json:"update_operation,omitempty"`
	// UpdateConditionsCustomMethodName provides the name of the custom method on the
	// `resourceManager` struct that will set Conditions on a `resource` struct
	// depending on the status of the resource.
	UpdateConditionsCustomMethodName string `json:"update_conditions_custom_method_name,omitempty"`

	// SpecFields is a list of instructions about additional Spec fields
	// on this Resource
	SpecFields []*SpecFieldConfig `json:"spec_fields"`
}

// SpecFieldConfig instructs the code generator how to handle an additional
// field in the Resource's SpecFields collection. This additional field can source
// its value from a shape in a different API Operation.
type SpecFieldConfig struct {
	// OperationID refers to the ID of the API Operation where we will
	// determine the field's Go type.
	OperationID string `json:"operation_id,omitempty"`
	// MemberName refers to the name of the member of the
	// Input shape in the Operation identified by OperaitonID that
	// we will take as our additional spec field.
	MemberName string `json:"member_name"`
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
	// FieldConfig instructions for Attributes that should be
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
	Fields map[string]FieldConfig `json:"fields"`
	// SetAttributesSingleAttribute indicates that the SetAttributes API call
	// doesn't actually set multiple attributes but rather must be called
	// multiple times, once for each attribute that needs to change. See SNS
	// SetTopicAttributes API call, which can be compared to the "normal" SNS
	// SetPlatformApplicationAttributes API call which accepts multiple
	// attributes and replaces the supplied attributes map key/values...
	SetAttributesSingleAttribute bool `json:"set_attributes_single_attribute"`
	// GetAttributesInput instructs the code generator how to handle the
	// GetAttributes input shape
	GetAttributesInput *GetAttributesInputConfig `json:"get_attributes_input,omitempty"`
}

// GetAttributesInputConfig is used to instruct the code generator how to
// handle the GetAttributes API operation's Input shape.
type GetAttributesInputConfig struct {
	// Overrides is a map of structures instructing the code generator how to
	// handle the override of a particular field in the Input shape for the
	// GetAttributes operation. The map keys are the names of the field in the
	// Input shape to override.
	Overrides map[string]*MemberConstructorConfig `json:"overrides"`
}

// MemberConstructorConfig contains override instructions for how to handle the
// construction of a particular member for a Shape in the API.
type MemberConstructorConfig struct {
	// Values contains the value or values of the member to always set the
	// member to. If the member's type is a []string, the member is set to the
	// Values list. If the type is a string, the member's value is set to the
	// first list element in the Values list.
	Values []string `json:"values"`
}

// FieldConfig contains instructions to the code generator about how
// to interpret the value of an Attribute and how to map it to a CRD's Spec or
// Status field
type FieldConfig struct {
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
	// Set of aws exception codes that are terminal exceptions for this resource
	TerminalCodes []string `json:"terminal_codes"`
}

// RenamesConfig contains instructions to the code generator how to rename
// fields in various Operation payloads
type RenamesConfig struct {
	// Operations is a map, keyed by Operation ID, of instructions on how to
	// handle renamed fields in Input and Output shapes.
	Operations map[string]*OperationRenamesConfig `json:"operations"`
}

// OperationRenamesConfig contains instructions to the code generator on how to
// rename fields in an Operation's input and output payload shapes
type OperationRenamesConfig struct {
	// InputFields is a map of Input shape fields to renamed field name.
	InputFields map[string]string `json:"input_fields"`
	// OutputFields is a map of Output shape fields to renamed field name.
	OutputFields map[string]string `json:"output_fields"`
}

// ListOperationConfig contains instructions for the code generator to handle
// List operations for service APIs that have no built-in filtering ability and
// whose List Operation always returns all objects.
type ListOperationConfig struct {
	// MatchFields lists the names of fields in the Shape of the
	// list element in the List Operation's Output shape.
	MatchFields []string `json:"match_fields"`
}

// UpdateOperationConfig contains instructions for the code generator to handle
// Update operations for service APIs that have resources that have
// difficult-to-standardize update operations.
type UpdateOperationConfig struct {
	// CustomMethodName is a string for the method name to replace the
	// sdkUpdate() method implementation for this resource
	CustomMethodName string `json:"custom_method_name"`
}

// IsIgnoredOperation returns true if Operation Name is configured to be ignored
// in generator config for the AWS service
func (c *Config) IsIgnoredOperation(operation *awssdkmodel.Operation) bool {
	if c == nil {
		return false
	}
	if operation == nil {
		return true
	}
	return util.InStrings(operation.Name, c.Ignore.Operations)
}

// UnpacksAttributesMap returns true if the underlying API has
// Get{Resource}Attributes/Set{Resource}Attributes API calls that map real,
// schema'd fields to a raw `map[string]*string` for this resource (see SNS and
// SQS APIs)
func (c *Config) UnpacksAttributesMap(resourceName string) bool {
	if c == nil {
		return false
	}
	resGenConfig, found := c.Resources[resourceName]
	return found && resGenConfig.UnpackAttributesMapConfig != nil
}

// SetAttributesSingleAttribute returns true if the supplied resource name has
// a SetAttributes operation that only actually changes a single attribute at a
// time. See: SNS SetTopicAttributes API call, which is entirely different from
// the SNS SetPlatformApplicationAttributes API call, which sets multiple
// attributes at once. :shrug:
func (c *Config) SetAttributesSingleAttribute(resourceName string) bool {
	if c == nil {
		return false
	}
	resGenConfig, found := c.Resources[resourceName]
	if !found || resGenConfig.UnpackAttributesMapConfig == nil {
		return false
	}
	return resGenConfig.UnpackAttributesMapConfig.SetAttributesSingleAttribute
}

// IsIgnoredShape returns true if the supplied shape name should be ignored by the
// code generator, false otherwise
func (c *Config) IsIgnoredShape(shapeName string) bool {
	if c == nil || len(c.Ignore.ShapeNames) == 0 {
		return false
	}
	return util.InStrings(shapeName, c.Ignore.ShapeNames)
}

// OverrideValues gives list of member values to override.
func (c *Config) OverrideValues(operationName string) (map[string]string, bool) {
	if c == nil {
		return nil, false
	}
	oConfig, ok := c.Operations[operationName]
	if !ok {
		return nil, false
	}
	return oConfig.OverrideValues, ok
}

// AdditionSpec gives map of operation and their MemberFields to
// add to spec.
func (c *Config) SpecFieldConfigs(resourceName string) ([]*SpecFieldConfig, bool) {
	if c == nil {
		return nil, false
	}
	resourceConfig, ok := c.Resources[resourceName]
	if !ok {
		return nil, false
	}
	return resourceConfig.SpecFields, ok
}

// IsIgnoredResource returns true if Operation Name is configured to be ignored
// in generator config for the AWS service
func (c *Config) IsIgnoredResource(resourceName string) bool {
	if resourceName == "" {
		return true
	}
	if c == nil {
		return false
	}
	return util.InStrings(resourceName, c.Ignore.ResourceNames)
}

// ResourceInputFieldRename returns the renamed field for a Resource, a
// supplied Operation ID and original field name and whether or not a renamed
// override field name was found
func (c *Config) ResourceInputFieldRename(
	resName string,
	opID string,
	origFieldName string,
) (string, bool) {
	if c == nil {
		return origFieldName, false
	}
	rConfig, ok := c.Resources[resName]
	if !ok {
		return origFieldName, false
	}
	if rConfig.Renames == nil {
		return origFieldName, false
	}
	oRenames, ok := rConfig.Renames.Operations[opID]
	if !ok {
		return origFieldName, false
	}
	renamed, ok := oRenames.InputFields[origFieldName]
	if !ok {
		return origFieldName, false
	}
	return renamed, true
}

// ListOpMatchFieldNames returns a slice of strings representing the field
// names in the List operation's Output shape's element Shape that we should
// check a corresponding value in the target Spec exists.
func (c *Config) ListOpMatchFieldNames(
	resName string,
) []string {
	res := []string{}
	if c == nil {
		return res
	}
	rConfig, found := c.Resources[resName]
	if !found {
		return res
	}
	if rConfig.ListOperation == nil {
		return res
	}
	return rConfig.ListOperation.MatchFields
}

// New returns a new Config object given a supplied
// path to a config file
func New(
	configPath string,
) (*Config, error) {
	gc := Config{}
	contents, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(contents, &gc); err != nil {
		return nil, err
	}
	return &gc, nil
}
