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

package pet

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8srt "k8s.io/apimachinery/pkg/runtime"

	ackv1alpha1 "github.com/aws/aws-service-operator-k8s/apis/core/v1alpha1"
	acktypes "github.com/aws/aws-service-operator-k8s/pkg/types"

	// svcapitypes "github.com/aws/aws-sdk-go/service/apis/{{ .AWSServiceVersion}}
	svcapitypes "github.com/aws/aws-service-operator-k8s/services/petstore/apis/v1alpha1"
	// svcsdk "github.com/aws/aws-sdk-go/service/{{ .AWSServiceAlias }}"
	svcsdk "github.com/aws/aws-service-operator-k8s/services/petstore/sdk/service/petstore"
)

// petResource implements the `aws-service-operator-k8s/pkg/types.AWSResource`
// interface
type petResource struct {
	// The Kubernetes-native CR representing the resource
	ko *svcapitypes.Pet
	// The aws-sdk-go-native representation of the resource
	sdko *svcsdk.PetData
}

// Identifiers returns an AWSResourceIdentifiers object containing various
// identifying information, including the AWS account ID that owns the
// resource, the resource's AWS Resource Name (ARN)
func (r *petResource) Identifiers() acktypes.AWSResourceIdentifiers {
	return &petResourceIdentifiers{r.ko.Status.ACKResourceMetadata}
}

// IsBeingDeleted returns true if the Kubernetes resource has a non-zero
// deletion timestemp
func (r *petResource) IsBeingDeleted() bool {
	return !r.ko.DeletionTimestamp.IsZero()
}

// RuntimeObject returns the Kubernetes apimachinery/runtime representation of
// the AWSResource
func (r *petResource) RuntimeObject() k8srt.Object {
	return r.ko
}

// MetaObject returns the Kubernetes apimachinery/apis/meta/v1.Object
// representation of the AWSResource
func (r *petResource) MetaObject() metav1.Object {
	return r.ko
}

// RuntimeMetaObject returns an object that implements both the Kubernetes
// apimachinery/runtime.Object and the Kubernetes
// apimachinery/apis/meta/v1.Object interfaces
func (r *petResource) RuntimeMetaObject() acktypes.RuntimeMetaObject {
	return r.ko
}

// Conditions returns the ACK Conditions collection for the AWSResource
func (r *petResource) Conditions() []*ackv1alpha1.Condition {
	return r.ko.Status.Conditions
}
