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

import (
	"time"

	ackv1alpha1 "github.com/aws/aws-service-operator-k8s/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BookSpec defines the desired state of Book
type BookSpec struct {
	// Name is the Bookstore API Book object's name.
	// If unspecified or empty, it defaults to be "${name}" of k8s Book
	// +optional
	Name *string `json:"name,omitempty"`
	// Title is the Bookstore API Book's title.
	Title *string `json:"title"`
	// Title is the Bookstore API Book's author.
	Author *string `json:"author"`
}

// BookStatus defines the observed state of Book
type BookStatus struct {
	// All CRs managed by ACK will have a common `Status.ACKResourceMetadata`
	// member that is used to contain resource sync state, account ownership,
	// constructed ARN for the resource
	ACKResourceMetadata *ackv1alpha1.ResourceMetadata `json:"ackResourceMetadata"`
	// All CRS managed by ACK have a common `Status.Conditions` member that
	// contains a collection of `ackv1alpha1.Condition` objects that describe
	// the various terminal states of the CR and its backend AWS service API
	// resource
	Conditions []*ackv1alpha1.Condition `json:"conditions"`
	// CreateTime is the datetime the book was created
	CreateTime *time.Time `json:"createTime"`
}

// Book implements sigs.k8s.io/apimachinery/pkg/runtime.Object
type Book struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BookSpec   `json:"spec,omitempty"`
	Status BookStatus `json:"status,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Book{})
}
