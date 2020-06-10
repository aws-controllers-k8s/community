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

package resource

import (
	k8srt "k8s.io/apimachinery/pkg/runtime"

	acktypes "github.com/aws/aws-service-operator-k8s/pkg/types"

	// svcapitypes "github.com/aws/aws-sdk-go/service/apis/{{ .AWSServiceVersion}}
	svcapitypes "github.com/aws/aws-service-operator-k8s/services/example/apis/v1alpha1"
	// svcsdk "github.com/aws/aws-sdk-go/service/{{ .AWSServiceAlias }}"
	svcsdk "github.com/aws/aws-service-operator-k8s/services/example/sdk/service/bookstore"
)

// bookResource implements the `aws-service-operator-k8s/pkg/types.AWSResource`
// interface
type bookResource struct {
	// The Kubernetes-native CR representing the resource
	ko *svcapitypes.Book
	// The aws-sdk-go-native representation of the resource
	sdko *svcsdk.BookData
}

// IsBeingDeleted returns true if the Kubernetes resource has a non-zero
// deletion timestemp
func (r *bookResource) IsBeingDeleted() bool {
	return !r.ko.DeletionTimestamp.IsZero()
}

// AccountID returns the AWS account identifier in which the backend AWS
// resource resides
func (r *bookResource) AccountID() acktypes.AWSAccountID {
	// TODO(jaypipes): Returns AWS Account ID from the common metadata that all
	// ACK CRs will share.
	return "example-account-id"
}

// CR returns the Kubernetes custom resource (CR) representation of the
// AWSResource
func (r *bookResource) CR() k8srt.Object {
	return r.ko
}
