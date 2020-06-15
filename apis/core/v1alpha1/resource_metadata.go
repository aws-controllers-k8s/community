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

package v1alpha1

// ResourceMetadata is common to all custom resources (CRs) managed by an ACK
// service controller. It is contained in the CR's `Status` member field and
// comprises various status and identifier fields useful to ACK for tracking
// state changes between Kubernetes and the backend AWS service API
type ResourceMetadata struct {
	// ARN is the Amazon Resource Name for the resource. This is a
	// globally-unique identifier and is set only by the ACK service controller
	// once the controller has orchestrated the creation of the resource OR
	// when it has verified that an "adopted" resource (a resource where the
	// ARN annotation was set by the Kubernetes user on the CR) exists and
	// matches the supplied CR's Spec field values.
	ARN *AWSResourceName `json:"arn"`
	// OwnerAccountID is the AWS Account ID of the account that owns the
	// backend AWS service API resource.
	OwnerAccountID *AWSAccountID `json:"ownerAccountID"`
}
