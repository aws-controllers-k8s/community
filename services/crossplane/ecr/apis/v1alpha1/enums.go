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

// Code generated by ack-generate. DO NOT EDIT.

package v1alpha1

type EncryptionType string

const (
	EncryptionType_AES256 EncryptionType = "AES256"
	EncryptionType_KMS EncryptionType = "KMS"
)

type FindingSeverity string

const (
	FindingSeverity_INFORMATIONAL FindingSeverity = "INFORMATIONAL"
	FindingSeverity_LOW FindingSeverity = "LOW"
	FindingSeverity_MEDIUM FindingSeverity = "MEDIUM"
	FindingSeverity_HIGH FindingSeverity = "HIGH"
	FindingSeverity_CRITICAL FindingSeverity = "CRITICAL"
	FindingSeverity_UNDEFINED FindingSeverity = "UNDEFINED"
)

type ImageActionType string

const (
	ImageActionType_EXPIRE ImageActionType = "EXPIRE"
)

type ImageFailureCode string

const (
	ImageFailureCode_InvalidImageDigest ImageFailureCode = "InvalidImageDigest"
	ImageFailureCode_InvalidImageTag ImageFailureCode = "InvalidImageTag"
	ImageFailureCode_ImageTagDoesNotMatchDigest ImageFailureCode = "ImageTagDoesNotMatchDigest"
	ImageFailureCode_ImageNotFound ImageFailureCode = "ImageNotFound"
	ImageFailureCode_MissingDigestAndTag ImageFailureCode = "MissingDigestAndTag"
	ImageFailureCode_ImageReferencedByManifestList ImageFailureCode = "ImageReferencedByManifestList"
	ImageFailureCode_KmsError ImageFailureCode = "KmsError"
)

type ImageTagMutability string

const (
	ImageTagMutability_MUTABLE ImageTagMutability = "MUTABLE"
	ImageTagMutability_IMMUTABLE ImageTagMutability = "IMMUTABLE"
)

type LayerAvailability string

const (
	LayerAvailability_AVAILABLE LayerAvailability = "AVAILABLE"
	LayerAvailability_UNAVAILABLE LayerAvailability = "UNAVAILABLE"
)

type LayerFailureCode string

const (
	LayerFailureCode_InvalidLayerDigest LayerFailureCode = "InvalidLayerDigest"
	LayerFailureCode_MissingLayerDigest LayerFailureCode = "MissingLayerDigest"
)

type LifecyclePolicyPreviewStatus string

const (
	LifecyclePolicyPreviewStatus_IN_PROGRESS LifecyclePolicyPreviewStatus = "IN_PROGRESS"
	LifecyclePolicyPreviewStatus_COMPLETE LifecyclePolicyPreviewStatus = "COMPLETE"
	LifecyclePolicyPreviewStatus_EXPIRED LifecyclePolicyPreviewStatus = "EXPIRED"
	LifecyclePolicyPreviewStatus_FAILED LifecyclePolicyPreviewStatus = "FAILED"
)

type ScanStatus string

const (
	ScanStatus_IN_PROGRESS ScanStatus = "IN_PROGRESS"
	ScanStatus_COMPLETE ScanStatus = "COMPLETE"
	ScanStatus_FAILED ScanStatus = "FAILED"
)

type TagStatus string

const (
	TagStatus_TAGGED TagStatus = "TAGGED"
	TagStatus_UNTAGGED TagStatus = "UNTAGGED"
	TagStatus_ANY TagStatus = "ANY"
)